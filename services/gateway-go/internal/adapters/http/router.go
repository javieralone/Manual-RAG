package http

import (
	"net/http"

	"api-go/internal/adapters/http/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter configura y devuelve el enrutador HTTP con sus middlewares y rutas
func NewRouter(queryHandler *handlers.QueryHandler, authHandler *handlers.AuthHandler, healthHandler *handlers.HealthHandler, ingestionHandler *handlers.IngestionHandler, authenticate func(http.Handler) http.Handler, authorize func(http.Handler) http.Handler, rateLimit func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()

	// Registro de endpoints
	protectedQuery := authenticate(authorize(rateLimit(http.HandlerFunc(queryHandler.HandleQuery))))
	mux.Handle("/api/v1/query", protectedQuery)

	protectedQueryStream := authenticate(authorize(rateLimit(http.HandlerFunc(queryHandler.HandleQueryStream))))
	mux.Handle("/api/v1/query/stream", protectedQueryStream)

	protectedEnqueue := authenticate(authorize(rateLimit(http.HandlerFunc(ingestionHandler.HandleEnqueue))))
	mux.Handle("/api/v1/ingestion/enqueue", protectedEnqueue)
	protectedJobs := authenticate(authorize(rateLimit(http.HandlerFunc(ingestionHandler.HandleJobs))))
	mux.Handle("/api/v1/ingestion/jobs", protectedJobs)
	protectedJob := authenticate(authorize(rateLimit(http.HandlerFunc(ingestionHandler.HandleJob))))
	mux.Handle("/api/v1/ingestion/jobs/{job_id}", protectedJob)
	protectedFailed := authenticate(authorize(rateLimit(http.HandlerFunc(ingestionHandler.HandleFailed))))
	mux.Handle("/api/v1/ingestion/failed", protectedFailed)
	protectedStorage := authenticate(authorize(rateLimit(http.HandlerFunc(ingestionHandler.HandleStorageOptions))))
	mux.Handle("/api/v1/ingestion/storage/options", protectedStorage)
	protectedUpload := authenticate(authorize(rateLimit(http.HandlerFunc(ingestionHandler.HandleUpload))))
	mux.Handle("/api/v1/ingestion/upload", protectedUpload)

	mux.HandleFunc("/api/v1/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.HandleRefresh)
	mux.HandleFunc("/api/v1/auth/logout", authHandler.HandleLogout)

	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/ready", healthHandler.Ready)
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}
