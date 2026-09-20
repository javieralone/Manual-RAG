package handlers

import (
	"encoding/json"
	"net/http"

	"api-go/internal/services"
)

type QueryRequest struct {
	Question string `json:"question"`
}

type QueryResponse struct {
	Answer  string   `json:"answer"`
	Sources []string `json:"sources"`
}

type Handler struct {
	ragService *services.RAGService
}

func NewHandler(ragService *services.RAGService) *Handler {
	return &Handler{ragService: ragService}
}

func (h *Handler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido. Usa POST.", http.StatusMethodNotAllowed)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
		http.Error(w, "Formato JSON inválido o 'question' vacía", http.StatusBadRequest)
		return
	}

	// Ejecutar la orquestación a través del servicio
	answer, sources, err := h.ragService.QueryRAG(req.Question)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := QueryResponse{
		Answer:  answer,
		Sources: sources,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}