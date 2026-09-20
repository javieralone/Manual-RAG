package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os" // <--- Agrega este import
)

// Estructuras para comunicarse con FastAPI (Python RAG)
type RAGSearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}

type ChunkResult struct {
	Page   int    `json:"page"`
	Source string `json:"source"`
	Text   string `json:"text"`
}

type RAGSearchResponse struct {
	Results []ChunkResult `json:"results"`
}

// Estructuras para comunicarse con Ollama
type OllamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaGenerateResponse struct {
	Response string `json:"response"`
}

// Servicio RAG principal en Go
type RAGService struct {
	PythonEngineURL string
	OllamaURL       string
	ModelName       string
}

func NewRAGService() *RAGService {
	pythonURL := os.Getenv("PYTHON_ENGINE_URL")
	if pythonURL == "" {
		pythonURL = "http://localhost:8000"
	}

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	return &RAGService{
		PythonEngineURL: pythonURL,
		OllamaURL:       ollamaURL,
		ModelName:       "qwen2.5:7b",
	}
}

// QueryRAG orquesta la consulta completa
func (s *RAGService) QueryRAG(userQuestion string) (string, []string, error) {
	// 1. Pedir contexto al motor de Python (FastAPI)
	reqBody, _ := json.Marshal(RAGSearchRequest{
		Query: userQuestion,
		TopK:  3,
	})

	resp, err := http.Post(s.PythonEngineURL+"/search", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", nil, fmt.Errorf("error conectando con RAG Engine Python: %w", err)
	}
	defer resp.Body.Close()

	var ragResp RAGSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&ragResp); err != nil {
		return "", nil, fmt.Errorf("error decodificando respuesta de Python: %w", err)
	}

	if len(ragResp.Results) == 0 {
		return "No se encontró información relevante en los manuales.", nil, nil
	}

	// 2. Construir el contexto y las fuentes
	var contextText string
	var sources []string
	for _, chunk := range ragResp.Results {
		contextText += fmt.Sprintf("[Página %d]: %s\n\n", chunk.Page, chunk.Text)
		sources = append(sources, fmt.Sprintf("Página %d", chunk.Page))
	}

	// 3. Crear el Prompt para Ollama
	prompt := fmt.Sprintf(`Eres un asistente técnico especializado en manuales de mantenimiento y lubricación.
Responde a la pregunta del usuario utilizando ÚNICAMENTE la información provista en los fragmentos del manual a continuación.
Si la respuesta no se encuentra en el texto provisto, indica claramente que la información no está disponible.

FRAGMENTOS DEL MANUAL:
%s

PREGUNTA DEL USUARIO:
%s

RESPUESTA DETALLADA:`, contextText, userQuestion)

	// 4. Enviar a Ollama
	ollamaReq, _ := json.Marshal(OllamaGenerateRequest{
		Model:  s.ModelName,
		Prompt: prompt,
		Stream: false,
	})

	ollamaResp, err := http.Post(s.OllamaURL+"/api/generate", "application/json", bytes.NewBuffer(ollamaReq))
	if err != nil {
		return "", nil, fmt.Errorf("error conectando con Ollama: %w", err)
	}
	defer ollamaResp.Body.Close()

	bodyBytes, _ := io.ReadAll(ollamaResp.Body)
	var finalGen OllamaGenerateResponse
	json.Unmarshal(bodyBytes, &finalGen)

	return finalGen.Response, sources, nil
}