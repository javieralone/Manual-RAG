package use_cases

import (
	"context"
	"strings"

	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
)

type AuthService struct {
	users        ports.UserRepository
	passwordHash ports.PasswordHasher
	tokens       ports.TokenService
	sessions     ports.RefreshSessionStore
}

func NewAuthService(users ports.UserRepository, passwordHash ports.PasswordHasher, tokens ports.TokenService, sessions ports.RefreshSessionStore) *AuthService {
	return &AuthService{users: users, passwordHash: passwordHash, tokens: tokens, sessions: sessions}
}

func (s *AuthService) Login(ctx context.Context, username string, password string) (domain.TokenPair, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return domain.TokenPair{}, domain.ErrInvalidCredentials
	}

	user, err := s.users.FindByUsername(ctx, username)
	if err != nil || s.passwordHash.Compare(user.PasswordHash, password) != nil {
		return domain.TokenPair{}, domain.ErrInvalidCredentials
	}

	pair, err := s.tokens.IssueTokenPair(user)
	if err != nil {
		return domain.TokenPair{}, err
	}
	if err := s.sessions.Create(ctx, refreshSession(pair, user.Username)); err != nil {
		return domain.TokenPair{}, domain.ErrSessionUnavailable
	}
	return pair, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return domain.TokenPair{}, domain.ErrInvalidToken
	}

	current, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return domain.TokenPair{}, domain.ErrInvalidToken
	}

	user, err := s.users.FindByUsername(ctx, current.Identity.Username)
	if err != nil {
		return domain.TokenPair{}, domain.ErrInvalidToken
	}

	pair, err := s.tokens.IssueTokenPair(user)
	if err != nil {
		return domain.TokenPair{}, err
	}
	rotated, err := s.sessions.Rotate(ctx, domain.RefreshSession{ID: current.SessionID, Username: current.Identity.Username, ExpiresAt: current.ExpiresAt}, refreshSession(pair, user.Username))
	if err != nil {
		return domain.TokenPair{}, domain.ErrSessionUnavailable
	}
	if !rotated {
		return domain.TokenPair{}, domain.ErrInvalidToken
	}
	return pair, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	current, err := s.tokens.ParseRefreshToken(strings.TrimSpace(refreshToken))
	if err != nil {
		return domain.ErrInvalidToken
	}
	if err := s.sessions.Revoke(ctx, current.SessionID); err != nil {
		return domain.ErrSessionUnavailable
	}
	return nil
}

func refreshSession(pair domain.TokenPair, username string) domain.RefreshSession {
	return domain.RefreshSession{ID: pair.RefreshSessionID, Username: username, ExpiresAt: pair.RefreshTokenExpiresAt}
}
