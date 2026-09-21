package ports

import (
	"context"
	"api-go/internal/core/domain"
)

type RAGEnginePort interface {
	RetrieveContext(ctx context.Context, question string, topK int) ([]domain.DocumentChunk, error)
}