package services

import (
	"context"
	"errors"
	"testing"

	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
)

type fakeRAGClient struct {
	chunks []domain.DocumentChunk
	err    error
}

func (c *fakeRAGClient) RetrieveContext(_ context.Context, _ string, _ int) ([]domain.DocumentChunk, error) {
	return c.chunks, c.err
}

type fakeLLMStreamClient struct {
	tokens []string
	err    error
}

func (c *fakeLLMStreamClient) GenerateAnswer(_ context.Context, _ string, _ []domain.DocumentChunk) (string, error) {
	return "", nil
}

func (c *fakeLLMStreamClient) GenerateAnswerStream(_ context.Context, _ string, _ []domain.DocumentChunk, onToken func(string) error) (int, error) {
	for _, tok := range c.tokens {
		if err := onToken(tok); err != nil {
			return 0, err
		}
	}
	return len(c.tokens), c.err
}

type recordingSink struct {
	events []string
	tokens []string
	stats  domain.StreamStats
	errMsg string
}

func (s *recordingSink) SendMetadata(_ []domain.DocumentChunk) error {
	s.events = append(s.events, "metadata")
	return nil
}

func (s *recordingSink) SendToken(text string) error {
	s.events = append(s.events, "token")
	s.tokens = append(s.tokens, text)
	return nil
}

func (s *recordingSink) SendComplete(stats domain.StreamStats) error {
	s.events = append(s.events, "complete")
	s.stats = stats
	return nil
}

func (s *recordingSink) SendError(err error) error {
	s.events = append(s.events, "error")
	s.errMsg = err.Error()
	return nil
}

func TestExecuteQueryStreamEmitsMetadataTokensThenComplete(t *testing.T) {
	rag := &fakeRAGClient{chunks: []domain.DocumentChunk{{Text: "chunk-1"}}}
	llm := &fakeLLMStreamClient{tokens: []string{"Hola", " mundo"}}
	orchestrator := NewQueryOrchestrator(rag, llm)
	sink := &recordingSink{}

	if err := orchestrator.ExecuteQueryStream(context.Background(), "  pregunta  ", sink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantEvents := []string{"metadata", "token", "token", "complete"}
	if len(sink.events) != len(wantEvents) {
		t.Fatalf("expected events %v, got %v", wantEvents, sink.events)
	}
	for i, e := range wantEvents {
		if sink.events[i] != e {
			t.Fatalf("expected events %v, got %v", wantEvents, sink.events)
		}
	}
	if sink.stats.TokenCount != 2 {
		t.Fatalf("expected token count 2, got %d", sink.stats.TokenCount)
	}
}

func TestExecuteQueryStreamRejectsEmptyQuestion(t *testing.T) {
	orchestrator := NewQueryOrchestrator(&fakeRAGClient{}, &fakeLLMStreamClient{})
	sink := &recordingSink{}

	if err := orchestrator.ExecuteQueryStream(context.Background(), "   ", sink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sink.events) != 1 || sink.events[0] != "error" {
		t.Fatalf("expected a single error event, got %v", sink.events)
	}
	if sink.errMsg != domain.ErrEmptyQuestion.Error() {
		t.Fatalf("expected empty question error, got %q", sink.errMsg)
	}
}

func TestExecuteQueryStreamSendsErrorOnRetrievalFailure(t *testing.T) {
	rag := &fakeRAGClient{err: errors.New("rag-engine caído")}
	orchestrator := NewQueryOrchestrator(rag, &fakeLLMStreamClient{})
	sink := &recordingSink{}

	if err := orchestrator.ExecuteQueryStream(context.Background(), "pregunta", sink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sink.events) != 1 || sink.events[0] != "error" {
		t.Fatalf("expected a single error event, got %v", sink.events)
	}
}

func TestExecuteQueryStreamSendsErrorOnGenerationFailure(t *testing.T) {
	rag := &fakeRAGClient{chunks: []domain.DocumentChunk{{Text: "chunk-1"}}}
	llm := &fakeLLMStreamClient{tokens: []string{"Hola"}, err: errors.New("ollama caído")}
	orchestrator := NewQueryOrchestrator(rag, llm)
	sink := &recordingSink{}

	if err := orchestrator.ExecuteQueryStream(context.Background(), "pregunta", sink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantEvents := []string{"metadata", "token", "error"}
	if len(sink.events) != len(wantEvents) {
		t.Fatalf("expected events %v, got %v", wantEvents, sink.events)
	}
}

var _ ports.StreamSink = (*recordingSink)(nil)
