package services

import (
	"context"
	"strings"

	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
)

type QueryOrchestrator struct {
	ragClient ports.RAGEnginePort
	llmClient ports.LLMClientPort
}

// NewQueryOrchestrator aplica inyección de dependencias por constructor
func NewQueryOrchestrator(rag ports.RAGEnginePort, llm ports.LLMClientPort) *QueryOrchestrator {
	return &QueryOrchestrator{
		ragClient: rag,
		llmClient: llm,
	}
}

func (s *QueryOrchestrator) ExecuteQuery(ctx context.Context, question string) (*domain.QueryResponse, error) {
	cleanQuestion := strings.TrimSpace(question)
	if cleanQuestion == "" {
		return nil, domain.ErrEmptyQuestion
	}

	// 1. Recuperar fragmentos desde la interfaz RAG
	chunks, err := s.ragClient.RetrieveContext(ctx, cleanQuestion, 3)
	if err != nil {
		return nil, err
	}

	// 2. Generar la respuesta desde la interfaz LLM
	answer, err := s.llmClient.GenerateAnswer(ctx, cleanQuestion, chunks)
	if err != nil {
		return nil, err
	}

	return &domain.QueryResponse{
		Question: cleanQuestion,
		Answer:   answer,
		Context:  chunks,
	}, nil
}