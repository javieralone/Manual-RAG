package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-go/internal/core/domain"
)

type fakeMiddlewareTokens struct {
	identity domain.Identity
	err      error
}

func (t fakeMiddlewareTokens) IssueTokenPair(domain.User) (domain.TokenPair, error) {
	return domain.TokenPair{}, nil
}

func (t fakeMiddlewareTokens) ParseAccessToken(string) (domain.Identity, error) {
	return t.identity, t.err
}

func (t fakeMiddlewareTokens) ParseRefreshToken(string) (domain.RefreshToken, error) {
	return domain.RefreshToken{Identity: t.identity}, t.err
}

func TestAuthenticationMiddlewareRequiresBearerToken(t *testing.T) {
	handler := AuthenticationMiddleware(fakeMiddlewareTokens{})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, req)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestRequireAnyRoleRejectsInsufficientRole(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(contextWithIdentity(req, domain.Identity{Username: "user", Roles: []domain.Role{domain.RoleUser}}))
	protected := RequireAnyRole(domain.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	response := httptest.NewRecorder()

	protected.ServeHTTP(response, req)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
}

func contextWithIdentity(req *http.Request, identity domain.Identity) context.Context {
	return context.WithValue(req.Context(), identityContextKey{}, identity)
}
