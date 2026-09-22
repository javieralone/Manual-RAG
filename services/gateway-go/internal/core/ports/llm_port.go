package ports

import (
	"context"
	"api-go/internal/core/domain"
)

// LLMClientPort abstrae la generación de texto (Ollama u otros LLMs)
type LLMClientPort interface {
	GenerateAnswer(ctx context.Context, question string, context []domain.DocumentChunk) (string, error)
	// GenerateAnswerStream invoca onToken por cada fragmento de texto recibido; devuelve el total de tokens generados.
	GenerateAnswerStream(ctx context.Context, question string, context []domain.DocumentChunk, onToken func(token string) error) (tokenCount int, err error)
}