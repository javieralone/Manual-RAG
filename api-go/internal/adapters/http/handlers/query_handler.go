package handlers

import (
	"encoding/json"
	"net/http"

	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
)

// QueryService agrupa los casos de uso normal y en streaming que consume este handler.
type QueryService interface {
	ports.QueryUseCase
	ports.QueryStreamUseCase
}

type QueryHandler struct {
	useCase QueryService
}

func NewQueryHandler(useCase QueryService) *QueryHandler {
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

// HandleQueryStream expone POST /query/stream: retransmite metadata, tokens y el evento final
// de Ollama vía Server-Sent Events, actuando como proxy sin reconstruir la respuesta completa.
func (h *QueryHandler) HandleQueryStream(w http.ResponseWriter, r *http.Request) {
	var req HTTPQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
		http.Error(w, `{"error": "Formato JSON inválido o 'question' vacía"}`, http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error": "streaming no soportado por el servidor"}`, http.StatusInternalServerError)
		return
	}

	sink := newSSESink(w, flusher)
	_ = h.useCase.ExecuteQueryStream(r.Context(), req.Question, sink)
}

