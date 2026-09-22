package ports

import (
	"api-go/internal/core/domain"
	"context"
)

// QueryUseCase defines el contrato del caso de uso principal
type QueryUseCase interface {
	ExecuteQuery(ctx context.Context, question string) (*domain.QueryResponse, error)
}

type FilteredQueryUseCase interface {
	ExecuteQueryWithFilters(ctx context.Context, question string, filters domain.QueryFilters) (*domain.QueryResponse, error)
}

// StreamSink recibe los eventos estructurados de una respuesta en streaming.
// Cada método devuelve error si el evento no pudo entregarse (p.ej. cliente desconectado),
// lo que permite abortar el streaming aguas arriba sin reconstruir la respuesta completa.
type StreamSink interface {
	SendMetadata(chunks []domain.DocumentChunk) error
	SendToken(text string) error
	SendComplete(stats domain.StreamStats) error
	SendError(err error) error
}

// QueryStreamUseCase defines el contrato del caso de uso de consulta en streaming
type QueryStreamUseCase interface {
	ExecuteQueryStream(ctx context.Context, question string, sink StreamSink) error
}

type FilteredQueryStreamUseCase interface {
	ExecuteQueryStreamWithFilters(ctx context.Context, question string, filters domain.QueryFilters, sink StreamSink) error
}
