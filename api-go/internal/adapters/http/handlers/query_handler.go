package handlers

import (
	"encoding/json"
	"net/http"

	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
)

type QueryHandler struct {
	useCase ports.QueryUseCase
}

func NewQueryHandler(useCase ports.QueryUseCase) *QueryHandler {
	return &QueryHandler{useCase: useCase}
}

type HTTPQueryRequest struct {
	Question string `json:"question"`
}

func (h *QueryHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	var req HTTPQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
		http.Error(w, `{"error": "Formato JSON inválido o 'question' vacía"}`, http.StatusBadRequest)
		return
	}

	result, err := h.useCase.ExecuteQuery(r.Context(), req.Question)
	if err != nil {
		if err == domain.ErrEmptyQuestion {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}