package http

import (
	"net/http"

	"api-go/internal/adapters/http/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter configura y devuelve el enrutador HTTP con sus middlewares y rutas
func NewRouter(queryHandler *handlers.QueryHandler, authHandler *handlers.AuthHandler, healthHandler *handlers.HealthHandler, authenticate func(http.Handler) http.Handler, authorize func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()

	// Registro de endpoints
	protectedQuery := authenticate(authorize(http.HandlerFunc(queryHandler.HandleQuery)))
	mux.Handle("/api/v1/query", protectedQuery)
	mux.HandleFunc("/api/v1/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.HandleRefresh)

	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/ready", healthHandler.Ready)
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}
