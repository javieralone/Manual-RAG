package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteDecodeErrorRespondsWith413ForBodyLimit(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/query", strings.NewReader(`{"question":"123456789"}`))
	response := httptest.NewRecorder()
	request.Body = http.MaxBytesReader(response, request.Body, 8)

	var payload HTTPQueryRequest
	err := decodeJSON(request, &payload)
	writeDecodeError(response, err)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, response.Code)
	}
	if body := response.Body.String(); body != "{\"error\":\"cuerpo de solicitud demasiado grande\"}\n" {
		t.Fatalf("unexpected response body %q", body)
	}
}