package use_cases

import (
	"context"
	"errors"
	"testing"

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

func (s *fakeTokenService) ParseRefreshToken(_ string) (domain.Identity, error) {
	s.parseCalls++
	return s.identity, s.parseError
}

func TestAuthServiceLoginReturnsGenericErrorForInvalidCredentials(t *testing.T) {
	service := NewAuthService(
		&fakeUserRepository{err: errors.New("not found")},
		fakePasswordHasher{},
		&fakeTokenService{},
	)

	_, err := service.Login(context.Background(), "missing", "password")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestAuthServiceLoginAndRefreshIssueTokens(t *testing.T) {
	tokens := &fakeTokenService{
		pair:     domain.TokenPair{AccessToken: "access", RefreshToken: "refresh"},
		identity: domain.Identity{Username: "admin", Roles: []domain.Role{domain.RoleAdmin}},
	}
	repository := &fakeUserRepository{user: domain.User{Username: "admin", PasswordHash: "hash"}}
	service := NewAuthService(repository, fakePasswordHasher{}, tokens)

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
