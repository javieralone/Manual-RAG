package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"api-go/internal/adapters/observability"
	"api-go/internal/core/domain"
	"go.opentelemetry.io/otel"
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
	metrics    *observability.Metrics
}

func NewPythonRAGClient(baseURL string, httpClient *http.Client, metrics *observability.Metrics) *PythonRAGClient {
	return &PythonRAGClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		metrics:    metrics,
	}
}

func (c *PythonRAGClient) RetrieveContext(ctx context.Context, query string, topK int) ([]domain.DocumentChunk, error) {
	ctx, span := otel.Tracer("manual-rag/api-go").Start(ctx, "rag-engine /search")
	defer span.End()
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
	observability.InjectTraceContext(ctx, req.Header)

	started := time.Now()
	resp, err := c.httpClient.Do(req)
	if c.metrics != nil {
		c.metrics.DependencyDuration.WithLabelValues("rag-engine").Observe(time.Since(started).Seconds())
	}
	if err != nil {
		if c.metrics != nil {
			c.metrics.DependencyTotal.WithLabelValues("rag-engine", "error").Inc()
		}
		return nil, fmt.Errorf("error conectando con rag-engine: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if c.metrics != nil {
			c.metrics.DependencyTotal.WithLabelValues("rag-engine", "error").Inc()
		}
		return nil, fmt.Errorf("rag-engine devolvió un status no esperado: %d", resp.StatusCode)
	}
	if c.metrics != nil {
		c.metrics.DependencyTotal.WithLabelValues("rag-engine", "success").Inc()
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
