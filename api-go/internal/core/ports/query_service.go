package ports

import (
	"context"
	"api-go/internal/core/domain"
)

// QueryUseCase defines el contrato del caso de uso principal
type QueryUseCase interface {
	ExecuteQuery(ctx context.Context, question string) (*domain.QueryResponse, error)
}