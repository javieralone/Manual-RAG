package clients

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"api-go/internal/adapters/observability"
	"api-go/internal/core/domain"
	"go.opentelemetry.io/otel"
)

type OllamaClient struct {
	baseURL    string
	modelName  string
	httpClient *http.Client
	metrics    *observability.Metrics
	breaker    *circuitBreaker
	retries    int
	backoff    time.Duration
}

func NewOllamaClient(baseURL string, modelName string, httpClient *http.Client, metrics *observability.Metrics, values ...ResilienceConfig) *OllamaClient {
	config := configuredResilience(values)
	return &OllamaClient{
		baseURL:    baseURL,
		modelName:  modelName,
		httpClient: httpClient,
		metrics:    metrics,
		breaker:    newCircuitBreaker(config.MaxFailures, config.ResetAfter),
		retries:    config.Retries,
		backoff:    config.Backoff,
	}
}

// Request y Response según la API nativa de Ollama (/api/generate)
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response  string `json:"response"`
	EvalCount int    `json:"eval_count"`
}

// ollamaStreamLine representa cada línea NDJSON devuelta por Ollama cuando Stream=true
type ollamaStreamLine struct {
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	EvalCount int    `json:"eval_count"`
}

func buildPrompt(question string, chunks []domain.DocumentChunk) string {
	var contextBuilder strings.Builder
	for i, chunk := range chunks {
		contextBuilder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, chunk.Text))
	}

	return fmt.Sprintf(
		"Eres un asistente técnico especializado. Responde de forma concisa y precisa basándote ÚNICAMENTE en el siguiente contexto.\n\n"+
			"CONTEXTO:\n%s\n"+
			"PREGUNTA:\n%s\n\n"+
			"RESPUESTA:",
		contextBuilder.String(),
		question,
	)
}

func (c *OllamaClient) GenerateAnswer(ctx context.Context, question string, chunks []domain.DocumentChunk) (string, error) {
	ctx, span := otel.Tracer("manual-rag/api-go").Start(ctx, "ollama /api/generate")
	defer span.End()
	// 1. Construir el prompt inyectando el contexto RAG
	prompt := buildPrompt(question, chunks)

	// 2. Preparar la petición HTTP
	reqPayload := ollamaRequest{
		Model:  c.modelName,
		Prompt: prompt,
		Stream: false, // Setear en false para recibir la respuesta completa de una vez
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("error serializando payload para Ollama: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("error creando request para Ollama: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	observability.InjectTraceContext(ctx, req.Header)

	// 3. Ejecutar la llamada
	started := time.Now()
	if !c.breaker.allow() {
		return "", ErrCircuitOpen
	}
	var resp *http.Response
	for attempt := 0; attempt <= c.retries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < http.StatusInternalServerError {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if attempt < c.retries {
			if waitErr := retryDelay(ctx, c.backoff, attempt); waitErr != nil {
				return "", waitErr
			}
		}
	}
	if c.metrics != nil {
		c.metrics.DependencyDuration.WithLabelValues("ollama").Observe(time.Since(started).Seconds())
	}
	if err != nil {
		c.breaker.failure()
		if c.metrics != nil {
			c.metrics.DependencyTotal.WithLabelValues("ollama", "error").Inc()
		}
		return "", fmt.Errorf("error de conexión con Ollama en %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.breaker.failure()
		if c.metrics != nil {
			c.metrics.DependencyTotal.WithLabelValues("ollama", "error").Inc()
		}
		return "", fmt.Errorf("Ollama devolvió un estado no esperado: %d", resp.StatusCode)
	}
	c.breaker.success()
	if c.metrics != nil {
		c.metrics.DependencyTotal.WithLabelValues("ollama", "success").Inc()
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("error deserializando respuesta de Ollama: %w", err)
	}
	if c.metrics != nil {
		elapsed := time.Since(started).Seconds()
		c.metrics.GenerationDuration.Observe(elapsed)
		c.metrics.TimeToFirstToken.Observe(elapsed)
		c.metrics.GeneratedTokens.Observe(float64(ollamaResp.EvalCount))
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

// GenerateAnswerStream llama a Ollama con Stream=true y reenvía cada fragmento de texto vía onToken,
// sin acumular la respuesta completa en memoria.
func (c *OllamaClient) GenerateAnswerStream(ctx context.Context, question string, chunks []domain.DocumentChunk, onToken func(token string) error) (int, error) {
	ctx, span := otel.Tracer("manual-rag/api-go").Start(ctx, "ollama /api/generate (stream)")
	defer span.End()

	prompt := buildPrompt(question, chunks)
	reqPayload := ollamaRequest{
		Model:  c.modelName,
		Prompt: prompt,
		Stream: true,
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return 0, fmt.Errorf("error serializando payload para Ollama: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return 0, fmt.Errorf("error creando request para Ollama: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	observability.InjectTraceContext(ctx, req.Header)

	started := time.Now()
	if !c.breaker.allow() {
		return 0, ErrCircuitOpen
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.breaker.failure()
		if c.metrics != nil {
			c.metrics.DependencyTotal.WithLabelValues("ollama", "error").Inc()
		}
		return 0, fmt.Errorf("error de conexión con Ollama en %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.breaker.failure()
		if c.metrics != nil {
			c.metrics.DependencyTotal.WithLabelValues("ollama", "error").Inc()
		}
		return 0, fmt.Errorf("Ollama devolvió un estado no esperado: %d", resp.StatusCode)
	}

	// Ollama devuelve un objeto JSON por línea (NDJSON) cuando stream=true
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var firstTokenObserved bool
	tokenCount := 0
	var evalCount int

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var chunk ollamaStreamLine
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue // línea corrupta: se ignora sin abortar el stream
		}

		if chunk.Response != "" {
			if !firstTokenObserved {
				firstTokenObserved = true
				if c.metrics != nil {
					c.metrics.TimeToFirstToken.Observe(time.Since(started).Seconds())
				}
			}
			tokenCount++
			if err := onToken(chunk.Response); err != nil {
				// El sink falló (p.ej. cliente desconectado): abortamos sin marcar error de dependencia
				return tokenCount, err
			}
		}

		if chunk.Done {
			evalCount = chunk.EvalCount
			break
		}
	}

	if err := scanner.Err(); err != nil {
		c.breaker.failure()
		if c.metrics != nil {
			c.metrics.DependencyTotal.WithLabelValues("ollama", "error").Inc()
		}
		return tokenCount, fmt.Errorf("error leyendo stream de Ollama: %w", err)
	}
	c.breaker.success()

	if c.metrics != nil {
		elapsed := time.Since(started).Seconds()
		c.metrics.DependencyDuration.WithLabelValues("ollama").Observe(elapsed)
		c.metrics.DependencyTotal.WithLabelValues("ollama", "success").Inc()
		c.metrics.GenerationDuration.Observe(elapsed)
		if evalCount > 0 {
			c.metrics.GeneratedTokens.Observe(float64(evalCount))
		} else {
			c.metrics.GeneratedTokens.Observe(float64(tokenCount))
		}
	}

	return tokenCount, nil
}
