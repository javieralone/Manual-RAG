package auth

import (
	"testing"
	"time"

	"api-go/internal/core/domain"
)

func TestJWTServiceSeparatesAccessAndRefreshTokens(t *testing.T) {
	service, err := NewJWTService(
		"access-secret-with-at-least-32-characters",
		"refresh-secret-with-at-least-32-characters",
		"manual-rag-api",
		"manual-rag-client",
		15*time.Minute,
		24*time.Hour,
	)
	if err != nil {
		t.Fatalf("creating JWT service: %v", err)
	}

	pair, err := service.IssueTokenPair(domain.User{Username: "admin", Roles: []domain.Role{domain.RoleAdmin}})
	if err != nil {
		t.Fatalf("issuing tokens: %v", err)
	}
	identity, err := service.ParseAccessToken(pair.AccessToken)
	if err != nil || identity.Username != "admin" {
		t.Fatalf("parsing access token: identity=%+v err=%v", identity, err)
	}
	if _, err := service.ParseAccessToken(pair.RefreshToken); err == nil {
		t.Fatal("refresh token must not be accepted as an access token")
	}
}

func TestJWTServiceRejectsShortSecrets(t *testing.T) {
	if _, err := NewJWTService("short", "also-short", "issuer", "audience", time.Minute, time.Hour); err == nil {
		t.Fatal("expected short secrets to be rejected")
	}
}
