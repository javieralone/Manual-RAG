package handlers

import (
	"net/http"

	"api-go/internal/adapters/http/response"
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
	Collection string `json:"collection,omitempty"`
	DocumentID string `json:"document_id,omitempty"`
	Chapter    string `json:"chapter,omitempty"`
	Section    string `json:"section,omitempty"`
}

func (h *QueryHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	var req HTTPQueryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeDecodeError(w, err)
		return
	}
	if req.Question == "" {
		response.WriteError(w, http.StatusBadRequest, "Formato JSON inválido o 'question' vacía")
		return
	}

	result, err := h.executeQuery(r, req)
	if err != nil {
		if err == domain.ErrEmptyQuestion {
			response.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}

// HandleQueryStream expone POST /api/v1/query/stream: retransmite metadata, tokens y el evento final
// de Ollama vía Server-Sent Events, actuando como proxy sin reconstruir la respuesta completa.
func (h *QueryHandler) HandleQueryStream(w http.ResponseWriter, r *http.Request) {
	var req HTTPQueryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeDecodeError(w, err)
		return
	}
	if req.Question == "" {
		response.WriteError(w, http.StatusBadRequest, "Formato JSON inválido o 'question' vacía")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		response.WriteError(w, http.StatusInternalServerError, "streaming no soportado por el servidor")
		return
	}

	sink := newSSESink(w, flusher)
	if err := h.executeQueryStream(r, req, sink); err != nil {
		if !sink.headersCommitted() {
			response.WriteError(w, http.StatusInternalServerError, "Error procesando el streaming")
			return
		}
		_ = sink.SendError(err)
	}
}

func (h *QueryHandler) executeQuery(r *http.Request, req HTTPQueryRequest) (*domain.QueryResponse, error) {
	filters := domain.QueryFilters{Collection: req.Collection, DocumentID: req.DocumentID, Chapter: req.Chapter, Section: req.Section}
	if filtered, ok := h.useCase.(ports.FilteredQueryUseCase); ok {
		return filtered.ExecuteQueryWithFilters(r.Context(), req.Question, filters)
	}
	return h.useCase.ExecuteQuery(r.Context(), req.Question)
}

func (h *QueryHandler) executeQueryStream(r *http.Request, req HTTPQueryRequest, sink ports.StreamSink) error {
	filters := domain.QueryFilters{Collection: req.Collection, DocumentID: req.DocumentID, Chapter: req.Chapter, Section: req.Section}
	if filtered, ok := h.useCase.(ports.FilteredQueryStreamUseCase); ok {
		return filtered.ExecuteQueryStreamWithFilters(r.Context(), req.Question, filters, sink)
	}
	return h.useCase.ExecuteQueryStream(r.Context(), req.Question, sink)
}
