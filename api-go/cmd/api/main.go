package main

import (
	"log"
	"net/http"
	"time"

	adaptersHTTP "api-go/internal/adapters/http"
	"api-go/internal/adapters/http/handlers"
	"api-go/internal/adapters/clients"
	"api-go/internal/core/services"
)

func main() {
	httpClient := &http.Client{Timeout: 300 * time.Second}

	// 1. Adaptadores
	ragAdapter := clients.NewPythonRAGClient("http://rag-engine:8000", httpClient)
	ollamaAdapter := clients.NewOllamaClient("http://host.docker.internal:11434", "qwen2.5:1.5b", httpClient)

	// 2. Core Service
	queryUseCase := services.NewQueryOrchestrator(ragAdapter, ollamaAdapter)

	// 3. Handler & Router HTTP
	queryHandler := handlers.NewQueryHandler(queryUseCase)
	router := adaptersHTTP.NewRouter(queryHandler)

	log.Println("API Gateway corriendo en :8080 bajo Clean Architecture...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}