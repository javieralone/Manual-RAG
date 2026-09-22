package middlewares

import (
	"net/http"

	"api-go/internal/adapters/http/response"
)

func BodyLimitMiddleware(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > limit {
				response.WriteError(w, http.StatusRequestEntityTooLarge, "cuerpo de solicitud demasiado grande")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}