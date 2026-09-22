package domain

import "errors"

var ErrEmptyQuestion = errors.New("la pregunta no puede estar vacía")

type DocumentChunk struct {
	Text     string                 `json:"text"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

type QueryFilters struct {
	Collection string `json:"collection,omitempty"`
	DocumentID string `json:"document_id,omitempty"`
	Chapter    string `json:"chapter,omitempty"`
	Section    string `json:"section,omitempty"`
}

func (f QueryFilters) IsEmpty() bool {
	return f.Collection == "" && f.DocumentID == "" && f.Chapter == "" && f.Section == ""
}

type QueryResponse struct {
	Question string          `json:"question"`
	Answer   string          `json:"answer"`
	Context  []DocumentChunk `json:"context"`
}

// StreamStats resume el timing de una respuesta en streaming, enviado en el evento "complete".
type StreamStats struct {
	TotalDurationMs    int64 `json:"total_duration_ms"`
	TimeToFirstTokenMs int64 `json:"time_to_first_token_ms"`
	TokenCount         int   `json:"token_count"`
}
