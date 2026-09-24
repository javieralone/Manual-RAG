package handlers

import (
	"bytes"
	"io"
	"net/http"

	"api-go/internal/adapters/clients"
	"api-go/internal/adapters/http/response"
)

type IngestionHandler struct {
	client *clients.IngestionClient
}

func NewIngestionHandler(client *clients.IngestionClient) *IngestionHandler {
	return &IngestionHandler{client: client}
}

func (h *IngestionHandler) HandleEnqueue(w http.ResponseWriter, r *http.Request) {
	h.proxy(w, r, http.MethodPost, "/ingestion/enqueue", "application/json")
}

func (h *IngestionHandler) HandleJobs(w http.ResponseWriter, r *http.Request) {
	h.proxy(w, r, http.MethodGet, "/ingestion/jobs", "")
}

func (h *IngestionHandler) HandleJob(w http.ResponseWriter, r *http.Request) {
	h.proxy(w, r, http.MethodGet, "/ingestion/jobs/"+r.PathValue("job_id"), "")
}

func (h *IngestionHandler) HandleFailed(w http.ResponseWriter, r *http.Request) {
	h.proxy(w, r, http.MethodGet, "/ingestion/failed", "")
}

func (h *IngestionHandler) HandleStorageOptions(w http.ResponseWriter, r *http.Request) {
	h.proxy(w, r, http.MethodGet, "/ingestion/storage/options", "")
}

func (h *IngestionHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	h.proxy(w, r, http.MethodPost, "/ingestion/upload", r.Header.Get("Content-Type"))
}

func (h *IngestionHandler) proxy(w http.ResponseWriter, r *http.Request, method, path, contentType string) {
	var body io.Reader
	if r.Body != nil {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
			return
		}
		body = bytes.NewReader(payload)
	}
	status, payload, err := h.client.Proxy(r.Context(), method, path, contentType, body)
	if err != nil {
		response.WriteError(w, http.StatusBadGateway, "servicio de ingesta no disponible")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}
