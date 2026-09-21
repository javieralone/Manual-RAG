package observability

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	RequestsTotal      *prometheus.CounterVec
	RequestDuration    *prometheus.HistogramVec
	AuthTotal          *prometheus.CounterVec
	DependencyTotal    *prometheus.CounterVec
	DependencyDuration *prometheus.HistogramVec
	GenerationDuration prometheus.Histogram
	GeneratedTokens    prometheus.Histogram
	TimeToFirstToken   prometheus.Histogram
	ErrorsTotal        *prometheus.CounterVec
	WorkerInFlight     prometheus.Gauge
	WorkerRejections   prometheus.Counter
	Readiness          prometheus.Gauge
	concurrentRequests atomic.Int64
}

type traceIDKey struct{}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	traceID, _ := ctx.Value(traceIDKey{}).(string)
	return traceID
}

func NewMetrics(registerer prometheus.Registerer) *Metrics {
	m := &Metrics{
		RequestsTotal:      prometheus.NewCounterVec(prometheus.CounterOpts{Name: "api_go_http_requests_total", Help: "Total HTTP requests handled by the Go gateway."}, []string{"method", "route", "status"}),
		RequestDuration:    prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "api_go_http_request_duration_seconds", Help: "HTTP request duration in seconds."}, []string{"method", "route"}),
		AuthTotal:          prometheus.NewCounterVec(prometheus.CounterOpts{Name: "api_go_authentication_total", Help: "Authentication attempts."}, []string{"operation", "result"}),
		DependencyTotal:    prometheus.NewCounterVec(prometheus.CounterOpts{Name: "api_go_dependency_requests_total", Help: "Requests to downstream dependencies."}, []string{"dependency", "result"}),
		DependencyDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "api_go_dependency_duration_seconds", Help: "Downstream dependency latency."}, []string{"dependency"}),
		GenerationDuration: prometheus.NewHistogram(prometheus.HistogramOpts{Name: "api_go_generation_duration_seconds", Help: "Ollama generation duration in seconds."}),
		GeneratedTokens:    prometheus.NewHistogram(prometheus.HistogramOpts{Name: "api_go_generated_tokens_total", Help: "Tokens reported by Ollama per generation."}),
		TimeToFirstToken:   prometheus.NewHistogram(prometheus.HistogramOpts{Name: "api_go_time_to_first_token_seconds", Help: "Time to first token; completion time for non-streaming Ollama."}),
		ErrorsTotal:        prometheus.NewCounterVec(prometheus.CounterOpts{Name: "api_go_errors_total", Help: "Classified gateway errors."}, []string{"component", "kind"}),
		WorkerInFlight:     prometheus.NewGauge(prometheus.GaugeOpts{Name: "api_go_worker_pool_in_flight", Help: "Queries currently using a worker."}),
		WorkerRejections:   prometheus.NewCounter(prometheus.CounterOpts{Name: "api_go_worker_pool_rejections_total", Help: "Queries rejected because the worker pool is full."}),
		Readiness:          prometheus.NewGauge(prometheus.GaugeOpts{Name: "api_go_readiness", Help: "Gateway readiness: 1 ready, 0 not ready."}),
	}
	registerer.MustRegister(m.RequestsTotal, m.RequestDuration, m.AuthTotal, m.DependencyTotal, m.DependencyDuration, m.GenerationDuration, m.GeneratedTokens, m.TimeToFirstToken, m.ErrorsTotal, m.WorkerInFlight, m.WorkerRejections, m.Readiness)
	return m
}

func (m *Metrics) ObserveRequest(method, route string, status int, started time.Time) {
	m.RequestsTotal.WithLabelValues(method, route, strconv.Itoa(status)).Inc()
	m.RequestDuration.WithLabelValues(method, route).Observe(time.Since(started).Seconds())
}

func (m *Metrics) IncConcurrent() { m.concurrentRequests.Add(1) }
func (m *Metrics) DecConcurrent() { m.concurrentRequests.Add(-1) }
