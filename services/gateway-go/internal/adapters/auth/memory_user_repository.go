package auth

import (
	"context"
	"errors"

	"api-go/internal/core/domain"
)

var errUserNotFound = errors.New("usuario no encontrado")

type InMemoryUserRepository struct {
	users map[string]domain.User
}

func NewInMemoryUserRepository(users ...domain.User) *InMemoryUserRepository {
	byUsername := make(map[string]domain.User, len(users))
	for _, user := range users {
		byUsername[user.Username] = user
	}
	return &InMemoryUserRepository{users: byUsername}
}

func (r *InMemoryUserRepository) FindByUsername(_ context.Context, username string) (domain.User, error) {
	user, ok := r.users[username]
	if !ok {
		return domain.User{}, errUserNotFound
	}
	return user, nil
}
