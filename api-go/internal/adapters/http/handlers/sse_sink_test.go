package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"api-go/internal/core/domain"
)

func TestSSESinkDefersHeaderCommitUntilFirstEvent(t *testing.T) {
	recorder := httptest.NewRecorder()
	sink := newSSESink(recorder, recorder)

	if recorder.Body.Len() != 0 {
		t.Fatalf("expected no body written before first event, got %q", recorder.Body.String())
	}

	if err := sink.SendMetadata([]domain.DocumentChunk{{Text: "chunk"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Fatalf("expected text/event-stream content type, got %q", contentType)
	}

	body := recorder.Body.String()
	if !strings.HasPrefix(body, "event: metadata\ndata: ") {
		t.Fatalf("unexpected event framing: %q", body)
	}
	if !strings.HasSuffix(body, "\n\n") {
		t.Fatalf("expected event to end with a blank line, got %q", body)
	}
}

func TestSSESinkWritesTokenAndCompleteEvents(t *testing.T) {
	recorder := httptest.NewRecorder()
	sink := newSSESink(recorder, recorder)

	_ = sink.SendToken("Hola")
	_ = sink.SendComplete(domain.StreamStats{TokenCount: 1})

	body := recorder.Body.String()
	if !strings.Contains(body, "event: token\ndata: {\"text\":\"Hola\"}\n\n") {
		t.Fatalf("expected token event in body, got %q", body)
	}
	if !strings.Contains(body, "event: complete\ndata: ") {
		t.Fatalf("expected complete event in body, got %q", body)
	}
}

func TestSSESinkWritesErrorEvent(t *testing.T) {
	recorder := httptest.NewRecorder()
	sink := newSSESink(recorder, recorder)

	_ = sink.SendError(errTest{"boom"})

	body := recorder.Body.String()
	if !strings.Contains(body, `event: error`) || !strings.Contains(body, `"error":"boom"`) {
		t.Fatalf("expected error event with message, got %q", body)
	}
}

type errTest struct{ msg string }

func (e errTest) Error() string { return e.msg }
