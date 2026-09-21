package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"api-go/internal/core/domain"
)

type PythonRAGClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewPythonRAGClient(baseURL string, httpClient *http.Client) *PythonRAGClient {
	return &PythonRAGClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *PythonRAGClient) RetrieveContext(ctx context.Context, question string, topK int) ([]domain.DocumentChunk, error) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"query": question,
		"top_k": topK,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/search", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error conectando con rag-engine: %w", err)
	}
	defer resp.Body.Close()

	var payload struct {
		Results []domain.DocumentChunk `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload.Results, nil
}