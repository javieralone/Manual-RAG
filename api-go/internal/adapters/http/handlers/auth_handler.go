package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
)

type AuthHandler struct {
	service ports.AuthService
	logger  *log.Logger
}

func NewAuthHandler(service ports.AuthService, logger *log.Logger) *AuthHandler {
	if logger == nil {
		logger = log.Default()
	}
	return &AuthHandler{service: service, logger: logger}
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
		h.logger.Printf("fallo de autenticación desde %s", r.RemoteAddr)
		writeAuthError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
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
		h.logger.Printf("fallo de refresh token desde %s", r.RemoteAddr)
		status := http.StatusUnauthorized
		if !errors.Is(err, domain.ErrInvalidToken) {
			status = http.StatusInternalServerError
		}
		writeAuthError(w, status, "refresh token inválido o expirado")
		return
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
