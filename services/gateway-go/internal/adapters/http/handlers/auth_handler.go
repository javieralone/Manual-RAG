package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"api-go/internal/adapters/observability"
	"api-go/internal/adapters/http/response"
	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
)

type AuthHandler struct {
	service ports.AuthService
	logger  *slog.Logger
	metrics *observability.Metrics
	cookie  RefreshCookieConfig
}

type RefreshCookieConfig struct {
	Secure bool
	MaxAge int
}

func NewAuthHandler(service ports.AuthService, logger *slog.Logger, metrics *observability.Metrics, cookie RefreshCookieConfig) *AuthHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuthHandler{service: service, logger: logger, metrics: metrics, cookie: cookie}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSON(r, &request); err != nil {
		writeDecodeError(w, err)
		return
	}

	pair, err := h.service.Login(r.Context(), request.Username, request.Password)
	if err != nil {
		if h.metrics != nil {
			h.metrics.AuthTotal.WithLabelValues("login", "failure").Inc()
		}
		h.logger.Warn("authentication_failed", "remote_addr", r.RemoteAddr)
		writeAuthServiceError(w, err, "credenciales inválidas")
		return
	}
	if h.metrics != nil {
		h.metrics.AuthTotal.WithLabelValues("login", "success").Inc()
	}
	h.setRefreshCookie(w, pair.RefreshToken)
	response.WriteJSON(w, http.StatusOK, pair)
}

func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	request, err := refreshRequestFromCookie(r)
	if err != nil {
		writeDecodeError(w, err)
		return
	}

	pair, err := h.service.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		if h.metrics != nil {
			h.metrics.AuthTotal.WithLabelValues("refresh", "failure").Inc()
		}
		h.logger.Warn("refresh_token_failed", "remote_addr", r.RemoteAddr)
		writeAuthServiceError(w, err, "refresh token inválido o expirado")
		return
	}
	if h.metrics != nil {
		h.metrics.AuthTotal.WithLabelValues("refresh", "success").Inc()
	}
	h.setRefreshCookie(w, pair.RefreshToken)
	response.WriteJSON(w, http.StatusOK, pair)
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	request, err := refreshRequestFromCookie(r)
	if err == nil {
		err = h.service.Logout(r.Context(), request.RefreshToken)
	}
	h.clearRefreshCookie(w)
	if err != nil && !errors.Is(err, domain.ErrInvalidToken) {
		writeAuthServiceError(w, err, "no se pudo cerrar la sesión")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func refreshRequestFromCookie(r *http.Request) (refreshRequest, error) {
	cookie, err := r.Cookie("manual_rag_refresh")
	if err != nil || cookie.Value == "" {
		return refreshRequest{}, err
	}
	return refreshRequest{RefreshToken: cookie.Value}, nil
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{Name: "manual_rag_refresh", Value: token, Path: "/api/v1/auth", HttpOnly: true, Secure: h.cookie.Secure, SameSite: http.SameSiteLaxMode, MaxAge: h.cookie.MaxAge})
}

func (h *AuthHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "manual_rag_refresh", Value: "", Path: "/api/v1/auth", HttpOnly: true, Secure: h.cookie.Secure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
}

func writeAuthServiceError(w http.ResponseWriter, err error, invalidMessage string) {
	if errors.Is(err, domain.ErrSessionUnavailable) {
		response.WriteError(w, http.StatusServiceUnavailable, "servicio de sesión no disponible")
		return
	}
	response.WriteError(w, http.StatusUnauthorized, invalidMessage)
}
