package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"api-go/internal/core/domain"
)

type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
	issuer        string
	audience      string
	accessTTL     time.Duration
	refreshTTL    time.Duration
	now           func() time.Time
}

type jwtClaims struct {
	Roles     []string `json:"roles"`
	TokenType string   `json:"token_type"`
	jwt.RegisteredClaims
}

func NewJWTService(accessSecret string, refreshSecret string, issuer string, audience string, accessTTL time.Duration, refreshTTL time.Duration) (*JWTService, error) {
	if len(accessSecret) < 32 || len(refreshSecret) < 32 {
		return nil, errors.New("los secretos JWT deben tener al menos 32 caracteres")
	}
	if issuer == "" || audience == "" || accessTTL <= 0 || refreshTTL <= 0 {
		return nil, errors.New("configuración JWT inválida")
	}
	return &JWTService{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		issuer:        issuer,
		audience:      audience,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		now:           time.Now,
	}, nil
}

func (s *JWTService) IssueTokenPair(user domain.User) (domain.TokenPair, error) {
	now := s.now().UTC()
	accessExpiresAt := now.Add(s.accessTTL)
	refreshExpiresAt := now.Add(s.refreshTTL)

	accessToken, err := s.issue(user, "access", s.accessSecret, accessExpiresAt, now)
	if err != nil {
		return domain.TokenPair{}, err
	}
	refreshSessionID, err := newSessionID()
	if err != nil {
		return domain.TokenPair{}, err
	}
	refreshToken, err := s.issue(user, "refresh", s.refreshSecret, refreshExpiresAt, now, refreshSessionID)
	if err != nil {
		return domain.TokenPair{}, err
	}

	return domain.TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		TokenType:             "Bearer",
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshTokenExpiresAt: refreshExpiresAt,
		RefreshSessionID:      refreshSessionID,
	}, nil
}

func (s *JWTService) issue(user domain.User, tokenType string, secret []byte, expiresAt time.Time, now time.Time, sessionID ...string) (string, error) {
	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, string(role))
	}
	claims := jwtClaims{
		Roles:     roles,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   user.Username,
			Audience:  jwt.ClaimStrings{s.audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	if len(sessionID) > 0 {
		claims.ID = sessionID[0]
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (s *JWTService) ParseRefreshToken(rawToken string) (domain.RefreshToken, error) {
	claims, err := s.parse(rawToken, "refresh", s.refreshSecret)
	if err != nil || claims.RegisteredClaims.ID == "" || claims.ExpiresAt == nil {
		return domain.RefreshToken{}, domain.ErrInvalidToken
	}
	return domain.RefreshToken{
		Identity:  claims.identity(),
		SessionID: claims.RegisteredClaims.ID,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

func (s *JWTService) parse(rawToken string, expectedType string, secret []byte) (*jwtClaims, error) {
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, domain.ErrInvalidToken
		}
		return secret, nil
	}, jwt.WithIssuer(s.issuer), jwt.WithAudience(s.audience), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid || claims.TokenType != expectedType || claims.Subject == "" {
		return nil, domain.ErrInvalidToken
	}
	return claims, nil
}

func (claims *jwtClaims) identity() domain.Identity {
	roles := make([]domain.Role, 0, len(claims.Roles))
	for _, role := range claims.Roles {
		roles = append(roles, domain.Role(role))
	}
	return domain.Identity{Username: claims.Subject, Roles: roles}
}

func (s *JWTService) ParseAccessToken(rawToken string) (domain.Identity, error) {
	claims, err := s.parse(rawToken, "access", s.accessSecret)
	if err != nil {
		return domain.Identity{}, err
	}
	return claims.identity(), nil
}

func newSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
