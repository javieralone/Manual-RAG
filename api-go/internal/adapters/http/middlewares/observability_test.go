package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"api-go/internal/adapters/observability"
	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricsMiddlewareRecordsStatus(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := observability.NewMetrics(registry)
	handler := MetricsMiddleware(metrics)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", response.Code)
	}
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("gathering metrics: %v", err)
	}
	for _, family := range metricFamilies {
		if family.GetName() == "api_go_http_requests_total" && len(family.GetMetric()) == 1 {
			return
		}
	}
	t.Fatal("expected one HTTP request metric")
}
