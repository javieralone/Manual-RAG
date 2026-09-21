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

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
