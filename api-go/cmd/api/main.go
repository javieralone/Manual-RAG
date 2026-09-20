package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type QueryRequest struct {
	Question string `json:"question"`
}

type QueryResponse struct {
	Answer  string   `json:"answer"`
	Sources []string `json:"sources"`
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Aquí Go orquesta la llamada al motor RAG en Python o consulta directa
	fmt.Printf("Pregunta recibida en Go API: %s\n", req.Question)

	// Respuesta simulada
	resp := QueryResponse{
		Answer:  "Procesando la pregunta a través del backend en Go...",
		Sources: []string{"Manual de lubricación - pág. 12"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/api/v1/query", queryHandler)
	log.Println("API Gateway en Go corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}