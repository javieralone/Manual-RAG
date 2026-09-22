package ports

import (
	"context"

	"api-go/internal/core/domain"
)

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (domain.User, error)
}

type PasswordHasher interface {
	Compare(hash string, password string) error
}

type TokenService interface {
	IssueTokenPair(user domain.User) (domain.TokenPair, error)
	ParseAccessToken(token string) (domain.Identity, error)
	ParseRefreshToken(token string) (domain.RefreshToken, error)
}

type RefreshSessionStore interface {
	Create(ctx context.Context, session domain.RefreshSession) error
	Rotate(ctx context.Context, current domain.RefreshSession, replacement domain.RefreshSession) (bool, error)
	Revoke(ctx context.Context, sessionID string) error
}

type AuthService interface {
	Login(ctx context.Context, username string, password string) (domain.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}
