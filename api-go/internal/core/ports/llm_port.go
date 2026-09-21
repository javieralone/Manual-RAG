package ports

import (
	"context"
	"api-go/internal/core/domain"
)

// LLMClientPort abstrae la generación de texto (Ollama u otros LLMs)
type LLMClientPort interface {
	GenerateAnswer(ctx context.Context, question string, context []domain.DocumentChunk) (string, error)
}