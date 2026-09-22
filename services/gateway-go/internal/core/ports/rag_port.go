package ports

import (
	"api-go/internal/core/domain"
	"context"
)

type RAGEnginePort interface {
	RetrieveContext(ctx context.Context, question string, topK int) ([]domain.DocumentChunk, error)
}

type FilteredRAGEnginePort interface {
	RetrieveContextWithFilters(ctx context.Context, question string, topK int, filters domain.QueryFilters) ([]domain.DocumentChunk, error)
}
