package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBodyLimitMiddlewareRejectsOversizedContentLength(t *testing.T) {
	called := false
	handler := BodyLimitMiddleware(8)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/query", strings.NewReader("123456789"))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if called {
		t.Fatal("expected oversized request to be rejected before the handler")
	}
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, response.Code)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}
	if body := response.Body.String(); body != "{\"error\":\"cuerpo de solicitud demasiado grande\"}\n" {
		t.Fatalf("unexpected response body %q", body)
	}
}

func TestBodyLimitMiddlewareCapsStreamedBodies(t *testing.T) {
	handler := BodyLimitMiddleware(8)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Fatal("expected streamed body to exceed limit")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/query", strings.NewReader("123456789"))
	request.ContentLength = -1
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected downstream handler to observe the capped body, got %d", response.Code)
	}
}