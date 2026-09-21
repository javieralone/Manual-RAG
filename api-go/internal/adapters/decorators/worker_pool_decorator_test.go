package decorators

import (
	"context"
	"sync"
	"testing"

	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
)

type blockingQueryService struct {
	release chan struct{}
	started chan struct{}
}

func (s *blockingQueryService) ExecuteQuery(_ context.Context, _ string) (*domain.QueryResponse, error) {
	return nil, nil
}

func (s *blockingQueryService) ExecuteQueryStream(_ context.Context, _ string, sink ports.StreamSink) error {
	close(s.started)
	<-s.release
	return sink.SendComplete(domain.StreamStats{})
}

type noopSink struct {
	errMsg string
	mu     sync.Mutex
}

func (s *noopSink) SendMetadata(_ []domain.DocumentChunk) error { return nil }
func (s *noopSink) SendToken(_ string) error                    { return nil }
func (s *noopSink) SendComplete(_ domain.StreamStats) error      { return nil }
func (s *noopSink) SendError(err error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errMsg = err.Error()
	return nil
}

func TestWorkerPoolDecoratorStreamRejectsWhenPoolFull(t *testing.T) {
	wrapped := &blockingQueryService{release: make(chan struct{}), started: make(chan struct{})}
	decorator := NewWorkerPoolUseCaseDecorator(wrapped, 1, nil, nil)

	go func() {
		_ = decorator.ExecuteQueryStream(context.Background(), "pregunta", &noopSink{})
	}()
	<-wrapped.started // el único worker está ocupado

	rejectedSink := &noopSink{}
	if err := decorator.ExecuteQueryStream(context.Background(), "otra pregunta", rejectedSink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rejectedSink.mu.Lock()
	got := rejectedSink.errMsg
	rejectedSink.mu.Unlock()
	if got != ErrServerBusy.Error() {
		t.Fatalf("expected ErrServerBusy event, got %q", got)
	}

	close(wrapped.release)
}
