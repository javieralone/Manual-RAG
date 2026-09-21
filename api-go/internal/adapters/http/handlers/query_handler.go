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
	Question   string `json:"question"`
	DocumentID string `json:"document_id,omitempty"`
	Chapter    string `json:"chapter,omitempty"`
	Section    string `json:"section,omitempty"`
}

func (h *QueryHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	var req HTTPQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
		http.Error(w, `{"error": "Formato JSON inválido o 'question' vacía"}`, http.StatusBadRequest)
		return
	}

	result, err := h.executeQuery(r, req)
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
	_ = h.executeQueryStream(r, req, sink)
}

func (h *QueryHandler) executeQuery(r *http.Request, req HTTPQueryRequest) (*domain.QueryResponse, error) {
	filters := domain.QueryFilters{DocumentID: req.DocumentID, Chapter: req.Chapter, Section: req.Section}
	if filtered, ok := h.useCase.(ports.FilteredQueryUseCase); ok {
		return filtered.ExecuteQueryWithFilters(r.Context(), req.Question, filters)
	}
	return h.useCase.ExecuteQuery(r.Context(), req.Question)
}

func (h *QueryHandler) executeQueryStream(r *http.Request, req HTTPQueryRequest, sink ports.StreamSink) error {
	filters := domain.QueryFilters{DocumentID: req.DocumentID, Chapter: req.Chapter, Section: req.Section}
	if filtered, ok := h.useCase.(ports.FilteredQueryStreamUseCase); ok {
		return filtered.ExecuteQueryStreamWithFilters(r.Context(), req.Question, filters, sink)
	}
	return h.useCase.ExecuteQueryStream(r.Context(), req.Question, sink)
}
