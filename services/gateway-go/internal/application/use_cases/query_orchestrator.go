package use_cases

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
	return s.ExecuteQueryWithFilters(ctx, question, domain.QueryFilters{})
}

func (s *QueryOrchestrator) ExecuteQueryWithFilters(ctx context.Context, question string, filters domain.QueryFilters) (*domain.QueryResponse, error) {
	cleanQuestion := strings.TrimSpace(question)
	if cleanQuestion == "" {
		return nil, domain.ErrEmptyQuestion
	}

	// 1. Obtener fragmentos (pasa el context recibido desde el middleware)
	chunks, err := s.retrieveContext(ctx, cleanQuestion, 3, filters)
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
	return s.ExecuteQueryStreamWithFilters(ctx, question, domain.QueryFilters{}, sink)
}

func (s *QueryOrchestrator) ExecuteQueryStreamWithFilters(ctx context.Context, question string, filters domain.QueryFilters, sink ports.StreamSink) error {
	started := time.Now()
	cleanQuestion := strings.TrimSpace(question)
	if cleanQuestion == "" {
		return sink.SendError(domain.ErrEmptyQuestion)
	}

	chunks, err := s.retrieveContext(ctx, cleanQuestion, 3, filters)
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

func (s *QueryOrchestrator) retrieveContext(ctx context.Context, question string, topK int, filters domain.QueryFilters) ([]domain.DocumentChunk, error) {
	if !filters.IsEmpty() {
		if filtered, ok := s.ragClient.(ports.FilteredRAGEnginePort); ok {
			return filtered.RetrieveContextWithFilters(ctx, question, topK, filters)
		}
	}
	return s.ragClient.RetrieveContext(ctx, question, topK)
}
