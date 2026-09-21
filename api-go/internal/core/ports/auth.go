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
	ParseRefreshToken(token string) (domain.Identity, error)
}

type AuthService interface {
	Login(ctx context.Context, username string, password string) (domain.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error)
}
