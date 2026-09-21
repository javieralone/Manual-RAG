package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"api-go/internal/core/domain"
)

type pythonSearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}

type pythonSearchResponse struct {
	Results []domain.DocumentChunk `json:"results"`
}

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

func (c *PythonRAGClient) RetrieveContext(ctx context.Context, query string, topK int) ([]domain.DocumentChunk, error) {
	reqBody, err := json.Marshal(pythonSearchRequest{
		Query: query,
		TopK:  topK,
	})
	if err != nil {
		return nil, fmt.Errorf("error serializando request para rag-engine: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/search", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creando request HTTP a rag-engine: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error conectando con rag-engine: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rag-engine devolvió un status no esperado: %d", resp.StatusCode)
	}

	var searchResp pythonSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta de rag-engine: %w", err)
	}

	// Si Results es nil (sin datos en Qdrant), inicializamos como slice vacío para no enviar 'null' en el JSON
	if searchResp.Results == nil {
		return []domain.DocumentChunk{}, nil
	}

	return searchResp.Results, nil
}