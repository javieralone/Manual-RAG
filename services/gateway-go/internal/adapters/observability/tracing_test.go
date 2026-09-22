package observability

import (
	"context"
	"net/http"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestInjectTraceContextWritesW3CTraceparent(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		SpanID:     trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithRemoteSpanContext(context.Background(), spanContext)
	headers := make(http.Header)
	InjectTraceContext(ctx, headers)
	if headers.Get("traceparent") == "" {
		t.Fatal("expected traceparent header")
	}
}
