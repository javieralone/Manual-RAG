package middlewares

import (
	"net/http"
	"time"

	"api-go/internal/adapters/observability"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(body)
}

// Flush permite que las respuestas en streaming (SSE) atraviesen este middleware.
func (w *statusWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func MetricsMiddleware(metrics *observability.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			writer := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(writer, r)
			route := r.URL.Path
			metrics.ObserveRequest(r.Method, route, writer.statusOrOK(), started)
		})
	}
}

func (w *statusWriter) statusOrOK() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}
