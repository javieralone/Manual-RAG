package http

import (
	"net/http"

	"api-go/internal/adapters/http/handlers"
)

// NewRouter configura y devuelve el enrutador HTTP con sus middlewares y rutas
func NewRouter(queryHandler *handlers.QueryHandler, authHandler *handlers.AuthHandler, authenticate func(http.Handler) http.Handler, authorize func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()

	// Registro de endpoints
	protectedQuery := authenticate(authorize(http.HandlerFunc(queryHandler.HandleQuery)))
	mux.Handle("/api/v1/query", protectedQuery)
	mux.HandleFunc("/api/v1/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.HandleRefresh)

	// Healthcheck simple para Docker / Orquestadores
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "UP"}`))
	})

	return mux
}
