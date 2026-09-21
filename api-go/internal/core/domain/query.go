package domain

import "errors"

var ErrEmptyQuestion = errors.New("la pregunta no puede estar vacía")

type DocumentChunk struct {
	Text     string                 `json:"text"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

type QueryResponse struct {
	Question string          `json:"question"`
	Answer   string          `json:"answer"`
	Context  []DocumentChunk `json:"context"`
}