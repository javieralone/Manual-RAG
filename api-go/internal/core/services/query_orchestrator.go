package services

import (
	"context"
	"strings"
	"time"

	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
)

type QueryOrchestrator struct {
	ragClient ports.RAGEnginePort
	llmClient ports.LLMClientPort
}

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

	// 1. Obtener fragmentos (pasa el context recibido desde el middleware)
	chunks, err := s.ragClient.RetrieveContext(ctx, cleanQuestion, 3)
	if err != nil {
		return nil, err
	}

	// 2. Generar respuesta con Ollama
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

// ExecuteQueryStream orquesta la misma consulta pero reenvía el retrieval y cada token de Ollama
// al sink en tiempo real, sin acumular la respuesta completa en memoria.
func (s *QueryOrchestrator) ExecuteQueryStream(ctx context.Context, question string, sink ports.StreamSink) error {
	started := time.Now()
	cleanQuestion := strings.TrimSpace(question)
	if cleanQuestion == "" {
		return sink.SendError(domain.ErrEmptyQuestion)
	}

	chunks, err := s.ragClient.RetrieveContext(ctx, cleanQuestion, 3)
	if err != nil {
		return sink.SendError(err)
	}

	if err := sink.SendMetadata(chunks); err != nil {
		return err
	}

	var firstTokenAt time.Time
	firstTokenSeen := false
	onToken := func(token string) error {
		if !firstTokenSeen {
			firstTokenSeen = true
			firstTokenAt = time.Now()
		}
		return sink.SendToken(token)
	}

	tokenCount, err := s.llmClient.GenerateAnswerStream(ctx, cleanQuestion, chunks, onToken)
	if err != nil {
		return sink.SendError(err)
	}

	stats := domain.StreamStats{
		TotalDurationMs: time.Since(started).Milliseconds(),
		TokenCount:      tokenCount,
	}
	if firstTokenSeen {
		stats.TimeToFirstTokenMs = firstTokenAt.Sub(started).Milliseconds()
	}

	return sink.SendComplete(stats)
}