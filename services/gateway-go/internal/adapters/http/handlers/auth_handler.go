package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"api-go/internal/adapters/observability"
	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
)

type AuthHandler struct {
	service ports.AuthService
	logger  *slog.Logger
	metrics *observability.Metrics
}

func NewAuthHandler(service ports.AuthService, logger *slog.Logger, metrics *observability.Metrics) *AuthHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuthHandler{service: service, logger: logger, metrics: metrics}
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
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAuthError(w, http.StatusBadRequest, "formato JSON inválido")
		return
	}

	pair, err := h.service.Login(r.Context(), request.Username, request.Password)
	if err != nil {
		if h.metrics != nil {
			h.metrics.AuthTotal.WithLabelValues("login", "failure").Inc()
		}
		h.logger.Warn("authentication_failed", "remote_addr", r.RemoteAddr)
		writeAuthError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
	}
	if h.metrics != nil {
		h.metrics.AuthTotal.WithLabelValues("login", "success").Inc()
	}
	writeAuthJSON(w, http.StatusOK, pair)
}

func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAuthError(w, http.StatusBadRequest, "formato JSON inválido")
		return
	}

	pair, err := h.service.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		if h.metrics != nil {
			h.metrics.AuthTotal.WithLabelValues("refresh", "failure").Inc()
		}
		h.logger.Warn("refresh_token_failed", "remote_addr", r.RemoteAddr)
		status := http.StatusUnauthorized
		if !errors.Is(err, domain.ErrInvalidToken) {
			status = http.StatusInternalServerError
		}
		writeAuthError(w, status, "refresh token inválido o expirado")
		return
	}
	if h.metrics != nil {
		h.metrics.AuthTotal.WithLabelValues("refresh", "success").Inc()
	}
	writeAuthJSON(w, http.StatusOK, pair)
}

func writeAuthJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	writeAuthJSON(w, status, map[string]string{"error": message})
}
