package main

import (
	"log"
	"net/http"
	"time"

	"api-go/internal/adapters/clients"
	"api-go/internal/adapters/decorators"
	adaptersHTTP "api-go/internal/adapters/http"
	"api-go/internal/adapters/http/handlers"
	"api-go/internal/adapters/http/middlewares"
	"api-go/internal/core/services"
)

func main() {
	httpClient := &http.Client{Timeout: 30000 * time.Second}

	// 1. Adaptadores Secundarios
	ragAdapter := clients.NewPythonRAGClient("http://rag-engine:8000", httpClient)
	ollamaAdapter := clients.NewOllamaClient("http://host.docker.internal:11434", "qwen2.5:1.5b", httpClient)

	// 2. Caso de Uso Core
	rawUseCase := services.NewQueryOrchestrator(ragAdapter, ollamaAdapter)

	// 3. Decorador Worker Pool (Máximo 2 peticiones concurrentes a la vez)
	useCaseWithWorkerPool := decorators.NewWorkerPoolUseCaseDecorator(rawUseCase, 2)

	// 4. Handler
	queryHandler := handlers.NewQueryHandler(useCaseWithWorkerPool)

	// 5. Router con Middleware de Timeout (Límite global de 60s por Request)
	router := adaptersHTTP.NewRouter(queryHandler)
	handlerWithMiddleware := middlewares.TimeoutMiddleware(60000 * time.Second)(router)

	log.Println("API Gateway corriendo con WorkerPool y Middleware de Contexto...")
	if err := http.ListenAndServe(":8080", handlerWithMiddleware); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}