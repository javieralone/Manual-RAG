package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"api-go/internal/core/domain"
)

// sseSink implementa ports.StreamSink escribiendo eventos Server-Sent Events directamente
// sobre el ResponseWriter, sin acumular la respuesta completa en memoria.
type sseSink struct {
	w       http.ResponseWriter
	flusher http.Flusher
	once    sync.Once
}

func newSSESink(w http.ResponseWriter, flusher http.Flusher) *sseSink {
	return &sseSink{w: w, flusher: flusher}
}

// commitHeaders difiere el WriteHeader(200) hasta el primer evento, de forma que los fallos
// previos al streaming (p.ej. retrieval fallido) puedan seguir devolviéndose como error HTTP normal.
func (s *sseSink) commitHeaders() {
	s.once.Do(func() {
		s.w.Header().Set("Content-Type", "text/event-stream")
		s.w.Header().Set("Cache-Control", "no-cache")
		s.w.Header().Set("Connection", "keep-alive")
		s.w.WriteHeader(http.StatusOK)
	})
}

func (s *sseSink) writeEvent(name string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	s.commitHeaders()
	if _, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", name, data); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

func (s *sseSink) SendMetadata(chunks []domain.DocumentChunk) error {
	return s.writeEvent("metadata", map[string]any{"context": chunks})
}

func (s *sseSink) SendToken(text string) error {
	return s.writeEvent("token", map[string]string{"text": text})
}

func (s *sseSink) SendComplete(stats domain.StreamStats) error {
	return s.writeEvent("complete", stats)
}

func (s *sseSink) SendError(err error) error {
	return s.writeEvent("error", map[string]string{"error": err.Error()})
}
