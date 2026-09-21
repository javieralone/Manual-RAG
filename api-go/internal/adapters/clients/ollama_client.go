package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"api-go/internal/core/domain"
)

type OllamaClient struct {
	baseURL    string
	modelName  string
	httpClient *http.Client
}

func NewOllamaClient(baseURL string, modelName string, httpClient *http.Client) *OllamaClient {
	return &OllamaClient{
		baseURL:    baseURL,
		modelName:  modelName,
		httpClient: httpClient,
	}
}

// Request y Response según la API nativa de Ollama (/api/generate)
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func (c *OllamaClient) GenerateAnswer(ctx context.Context, question string, chunks []domain.DocumentChunk) (string, error) {
	// 1. Construir el prompt inyectando el contexto RAG
	var contextBuilder strings.Builder
	for i, chunk := range chunks {
		contextBuilder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, chunk.Text))
	}

	prompt := fmt.Sprintf(
		"Eres un asistente técnico especializado. Responde de forma concisa y precisa basándote ÚNICAMENTE en el siguiente contexto.\n\n"+
			"CONTEXTO:\n%s\n"+
			"PREGUNTA:\n%s\n\n"+
			"RESPUESTA:",
		contextBuilder.String(),
		question,
	)

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

	// 3. Ejecutar la llamada
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error de conexión con Ollama en %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama devolvió un estado no esperado: %d", resp.StatusCode)
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("error deserializando respuesta de Ollama: %w", err)
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}