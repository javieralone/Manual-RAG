package middlewares

import (
	"net/http"

	"api-go/internal/adapters/observability"
	"go.opentelemetry.io/otel"
)

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), observability.HeaderCarrier(r.Header))
		ctx, span := otel.Tracer("manual-rag/api-go").Start(ctx, r.Method+" "+r.URL.Path)
		defer span.End()
		observability.InjectTraceContext(ctx, w.Header())
		traceID := span.SpanContext().TraceID().String()
		ctx = observability.WithTraceID(ctx, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
