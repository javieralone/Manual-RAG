package http

import (
	"net/http"

	"api-go/internal/adapters/http/handlers"
)

// NewRouter configura y devuelve el enrutador HTTP con sus middlewares y rutas
func NewRouter(queryHandler *handlers.QueryHandler) http.Handler {
	mux := http.NewServeMux()

	// Registro de endpoints
	mux.HandleFunc("/api/v1/query", queryHandler.HandleQuery)

	// Healthcheck simple para Docker / Orquestadores
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "UP"}`))
	})

	return mux
}