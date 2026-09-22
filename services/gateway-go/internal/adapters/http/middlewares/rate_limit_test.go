package middlewares

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-go/internal/adapters/observability"
	"api-go/internal/core/domain"
	"github.com/prometheus/client_golang/prometheus"
)

func TestRateLimitMiddlewareRejectsIP(t *testing.T) {
	metrics := observability.NewMetrics(prometheus.NewRegistry())
	handler := RateLimitMiddleware(1, time.Minute, metrics, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRequest(http.MethodPost, "/api/v1/query", nil)
	first.RemoteAddr = "192.0.2.10:1234"
	second := httptest.NewRequest(http.MethodPost, "/api/v1/query", nil)
	second.RemoteAddr = "192.0.2.10:5678"

	firstResponse := httptest.NewRecorder()
	handler.ServeHTTP(firstResponse, first)
	secondResponse := httptest.NewRecorder()
	handler.ServeHTTP(secondResponse, second)

	if firstResponse.Code != http.StatusNoContent {
		t.Fatalf("expected first request to pass, got %d", firstResponse.Code)
	}
	if secondResponse.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request to be rate limited, got %d", secondResponse.Code)
	}
	if secondResponse.Header().Get("Retry-After") != "60" {
		t.Fatalf("expected Retry-After 60, got %q", secondResponse.Header().Get("Retry-After"))
	}
}

func TestRateLimitMiddlewareRejectsUserAcrossIPs(t *testing.T) {
	metrics := observability.NewMetrics(prometheus.NewRegistry())
	handler := RateLimitMiddleware(1, time.Minute, metrics, testLogger())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRequest(http.MethodPost, "/api/v1/query", nil)
	first.RemoteAddr = "192.0.2.10:1234"
	first = first.WithContext(context.WithValue(first.Context(), identityContextKey{}, domain.Identity{Username: "operator"}))
	second := httptest.NewRequest(http.MethodPost, "/api/v1/query", nil)
	second.RemoteAddr = "192.0.2.11:1234"
	second = second.WithContext(context.WithValue(second.Context(), identityContextKey{}, domain.Identity{Username: "operator"}))

	firstResponse := httptest.NewRecorder()
	handler.ServeHTTP(firstResponse, first)
	secondResponse := httptest.NewRecorder()
	handler.ServeHTTP(secondResponse, second)

	if secondResponse.Code != http.StatusTooManyRequests {
		t.Fatalf("expected user limit to reject second IP, got %d", secondResponse.Code)
	}
}

func TestClientIPSupportsIPv6(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.RemoteAddr = "[2001:db8::1]:443"
	if got := clientIP(request); got != "2001:db8::1" {
		t.Fatalf("expected normalized IPv6 address, got %q", got)
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
