package main

import (
	"log"
	"net/http"

	"api-go/internal/handlers"
	"api-go/internal/services"
)

func main() {
	// 1. Inicializar el servicio RAG (se comunica con FastAPI en :8000 y Ollama en :11434)
	ragService := services.NewRAGService()

	// 2. Inicializar el Handler
	queryHandler := handlers.NewHandler(ragService)

	// 3. Registrar los Endpoints
	http.HandleFunc("/api/v1/query", queryHandler.HandleQuery)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API Go Gateway funcionando ok."))
	})

	log.Println("==================================================")
	log.Println("🚀 Go API Gateway corriendo en http://localhost:8080")
	log.Println("📌 Endpoint RAG: POST http://localhost:8080/api/v1/query")
	log.Println("==================================================")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error al iniciar el servidor Go: %v", err)
	}
}