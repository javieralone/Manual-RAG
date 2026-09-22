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
}

func NewAuthService(users ports.UserRepository, passwordHash ports.PasswordHasher, tokens ports.TokenService) *AuthService {
	return &AuthService{users: users, passwordHash: passwordHash, tokens: tokens}
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

	return s.tokens.IssueTokenPair(user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return domain.TokenPair{}, domain.ErrInvalidToken
	}

	identity, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return domain.TokenPair{}, domain.ErrInvalidToken
	}

	user, err := s.users.FindByUsername(ctx, identity.Username)
	if err != nil {
		return domain.TokenPair{}, domain.ErrInvalidToken
	}

	return s.tokens.IssueTokenPair(user)
}
