package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
)

type identityContextKey struct{}

func AuthenticationMiddleware(tokens ports.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if len(header) < 7 || !strings.EqualFold(header[:6], "Bearer") || header[6] != ' ' {
				writeJSONError(w, http.StatusUnauthorized, "autenticación requerida")
				return
			}

			identity, err := tokens.ParseAccessToken(strings.TrimSpace(header[7:]))
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "token inválido o expirado")
				return
			}

			ctx := context.WithValue(r.Context(), identityContextKey{}, identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAnyRole(allowed ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := IdentityFromContext(r.Context())
			if !ok || !hasAllowedRole(identity.Roles, allowed) {
				writeJSONError(w, http.StatusForbidden, domain.ErrInsufficientRole.Error())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func IdentityFromContext(ctx context.Context) (domain.Identity, bool) {
	identity, ok := ctx.Value(identityContextKey{}).(domain.Identity)
	return identity, ok
}

func hasAllowedRole(actual []domain.Role, allowed []domain.Role) bool {
	for _, actualRole := range actual {
		for _, allowedRole := range allowed {
			if actualRole == allowedRole {
				return true
			}
		}
	}
	return false
}

func CORSMiddleware(allowedOrigins ...string) func(http.Handler) http.Handler {
	origins := allowedOrigins
	if len(origins) == 0 {
		origins = []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://frontend:5173", "http://rag_frontend:5173"}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				for _, allowedOrigin := range origins {
					if origin == allowedOrigin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Access-Control-Allow-Credentials", "true")
						break
					}
				}
				if origin != "" && !contains(origins, origin) {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
