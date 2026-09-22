package use_cases

import (
	"context"
	"errors"
	"testing"
	"time"

	"api-go/internal/core/domain"
)

type fakeUserRepository struct {
	user domain.User
	err  error
	seen string
}

func (r *fakeUserRepository) FindByUsername(_ context.Context, username string) (domain.User, error) {
	r.seen = username
	return r.user, r.err
}

type fakePasswordHasher struct {
	err error
}

func (h fakePasswordHasher) Compare(_, _ string) error {
	return h.err
}

type fakeTokenService struct {
	pair       domain.TokenPair
	identity   domain.Identity
	refresh    domain.RefreshToken
	issueCalls int
	parseCalls int
	parseError error
}

func (s *fakeTokenService) IssueTokenPair(_ domain.User) (domain.TokenPair, error) {
	s.issueCalls++
	return s.pair, nil
}

func (s *fakeTokenService) ParseAccessToken(_ string) (domain.Identity, error) {
	return s.identity, s.parseError
}

func (s *fakeTokenService) ParseRefreshToken(_ string) (domain.RefreshToken, error) {
	s.parseCalls++
	return s.refresh, s.parseError
}

type fakeRefreshSessionStore struct {
	createErr error
	rotateErr error
	rotated   bool
	revoked   string
}

func (s *fakeRefreshSessionStore) Create(_ context.Context, _ domain.RefreshSession) error {
	return s.createErr
}

func (s *fakeRefreshSessionStore) Rotate(_ context.Context, _ domain.RefreshSession, _ domain.RefreshSession) (bool, error) {
	return s.rotated, s.rotateErr
}

func (s *fakeRefreshSessionStore) Revoke(_ context.Context, sessionID string) error {
	s.revoked = sessionID
	return nil
}

func TestAuthServiceLoginReturnsGenericErrorForInvalidCredentials(t *testing.T) {
	service := NewAuthService(
		&fakeUserRepository{err: errors.New("not found")},
		fakePasswordHasher{},
		&fakeTokenService{},
		&fakeRefreshSessionStore{},
	)

	_, err := service.Login(context.Background(), "missing", "password")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestAuthServiceLoginAndRefreshIssueTokens(t *testing.T) {
	tokens := &fakeTokenService{
		pair: domain.TokenPair{AccessToken: "access", RefreshToken: "refresh", RefreshSessionID: "next", RefreshTokenExpiresAt: time.Now().Add(time.Hour)},
		refresh: domain.RefreshToken{Identity: domain.Identity{Username: "admin", Roles: []domain.Role{domain.RoleAdmin}}, SessionID: "current", ExpiresAt: time.Now().Add(time.Hour)},
	}
	repository := &fakeUserRepository{user: domain.User{Username: "admin", PasswordHash: "hash"}}
	sessions := &fakeRefreshSessionStore{rotated: true}
	service := NewAuthService(repository, fakePasswordHasher{}, tokens, sessions)

	if _, err := service.Login(context.Background(), " admin ", "password"); err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if _, err := service.Refresh(context.Background(), "refresh"); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if repository.seen != "admin" || tokens.issueCalls != 2 || tokens.parseCalls != 1 {
		t.Fatalf("unexpected calls: username=%q issue=%d parse=%d", repository.seen, tokens.issueCalls, tokens.parseCalls)
	}
}

func TestAuthServiceLogoutRevokesRefreshSession(t *testing.T) {
	tokens := &fakeTokenService{
		refresh: domain.RefreshToken{Identity: domain.Identity{Username: "admin"}, SessionID: "session-id", ExpiresAt: time.Now().Add(time.Hour)},
	}
	sessions := &fakeRefreshSessionStore{}
	service := NewAuthService(&fakeUserRepository{}, fakePasswordHasher{}, tokens, sessions)

	if err := service.Logout(context.Background(), "refresh"); err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	if sessions.revoked != "session-id" {
		t.Fatalf("expected session revocation, got %q", sessions.revoked)
	}
}
