package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIngestionClientProxyForwardsJSONRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/ingestion/enqueue" {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", request.Header.Get("Content-Type"))
		}
		writer.WriteHeader(http.StatusAccepted)
		_, _ = writer.Write([]byte(`{"status":"PENDING"}`))
	}))
	defer server.Close()

	client := NewIngestionClient(server.URL, server.Client())
	status, payload, err := client.Proxy(context.Background(), http.MethodPost, "/ingestion/enqueue", strings.NewReader(`{"local_path":"manual.pdf"}`))
	if err != nil {
		t.Fatalf("proxy returned error: %v", err)
	}
	if status != http.StatusAccepted || string(payload) != `{"status":"PENDING"}` {
		t.Fatalf("unexpected response: %d %s", status, payload)
	}
}

func TestIngestionClientProxyReturnsErrorForServerFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	client := NewIngestionClient(server.URL, server.Client())
	_, _, err := client.Proxy(context.Background(), http.MethodGet, "/ingestion/jobs", nil)
	if err == nil {
		t.Fatal("expected upstream failure")
	}
}
