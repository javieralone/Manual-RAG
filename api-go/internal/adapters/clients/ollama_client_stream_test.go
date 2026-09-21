package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-go/internal/core/domain"
)

func ndjsonStreamHandler(lines []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher := w.(http.Flusher)
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		for _, line := range lines {
			fmt.Fprintln(w, line)
			flusher.Flush()
		}
	}
}

func TestOllamaClientGenerateAnswerStreamForwardsTokensInOrder(t *testing.T) {
	server := httptest.NewServer(ndjsonStreamHandler([]string{
		`{"response":"Hola","done":false}`,
		`{"response":" mundo","done":false}`,
		`{"done":true,"eval_count":2}`,
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "test-model", &http.Client{}, nil)

	var got []string
	tokenCount, err := client.GenerateAnswerStream(context.Background(), "pregunta", []domain.DocumentChunk{}, func(token string) error {
		got = append(got, token)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokenCount != 2 {
		t.Fatalf("expected tokenCount=2, got %d", tokenCount)
	}
	want := []string{"Hola", " mundo"}
	if len(got) != len(want) {
		t.Fatalf("expected tokens %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected tokens %v, got %v", want, got)
		}
	}
}

func TestOllamaClientGenerateAnswerStreamAbortsOnSinkError(t *testing.T) {
	server := httptest.NewServer(ndjsonStreamHandler([]string{
		`{"response":"Hola","done":false}`,
		`{"response":" mundo","done":false}`,
		`{"done":true,"eval_count":2}`,
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "test-model", &http.Client{}, nil)

	sinkErr := errors.New("cliente desconectado")
	calls := 0
	_, err := client.GenerateAnswerStream(context.Background(), "pregunta", []domain.DocumentChunk{}, func(_ string) error {
		calls++
		return sinkErr
	})
	if !errors.Is(err, sinkErr) {
		t.Fatalf("expected sink error to propagate, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected stream to stop after first failing token, got %d calls", calls)
	}
}

func TestOllamaClientGenerateAnswerStreamRespectsContextCancellation(t *testing.T) {
	blockUntilCancel := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher := w.(http.Flusher)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"response":"Hola","done":false}`)
		flusher.Flush()
		<-r.Context().Done()
		close(blockUntilCancel)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "test-model", &http.Client{}, nil)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := client.GenerateAnswerStream(ctx, "pregunta", []domain.DocumentChunk{}, func(_ string) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected an error after context cancellation")
	}

	select {
	case <-blockUntilCancel:
	case <-time.After(time.Second):
		t.Fatal("server did not observe client cancellation")
	}
}
