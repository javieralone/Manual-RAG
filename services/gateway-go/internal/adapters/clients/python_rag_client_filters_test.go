package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-go/internal/core/domain"
)

func TestPythonRAGClientSendsFiltersAndPreservesMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/search" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var request pythonSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.DocumentID != "manual-a" || request.Chapter != "2" || request.Section != "2.1" || request.Collection != "manuales_tecnicos" {
			t.Fatalf("unexpected filters: %+v", request)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"text":"contenido","score":0.9,"metadata":{"document_id":"manual-a","chapter":"2","section":"2.1"}}]}`))
	}))
	defer server.Close()

	client := NewPythonRAGClient(server.URL, server.Client(), nil)
	chunks, err := client.RetrieveContextWithFilters(context.Background(), "consulta", 3, domain.QueryFilters{
		Collection: "manuales_tecnicos",
		DocumentID: "manual-a",
		Chapter:    "2",
		Section:    "2.1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 1 || chunks[0].Metadata["document_id"] != "manual-a" {
		t.Fatalf("unexpected chunks: %+v", chunks)
	}
}
