package domain

import (
	"errors"
	"time"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleUser     Role = "user"
)

var (
	ErrInvalidCredentials = errors.New("credenciales inválidas")
	ErrInvalidToken       = errors.New("token inválido")
	ErrInsufficientRole   = errors.New("permisos insuficientes")
	ErrSessionUnavailable = errors.New("sesión no disponible")
)

type User struct {
	Username     string
	PasswordHash string
	Roles        []Role
}

type Identity struct {
	Username string
	Roles    []Role
}

type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"-"`
	TokenType             string    `json:"token_type"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	RefreshSessionID      string    `json:"-"`
}

type RefreshToken struct {
	Identity  Identity
	SessionID string
	ExpiresAt time.Time
}

type RefreshSession struct {
	ID        string
	Username  string
	ExpiresAt time.Time
}
