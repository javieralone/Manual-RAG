---
name: observability-stack
description: 'Use when changing Manual-RAG observability: Prometheus metrics, Grafana dashboards, Loki/Promtail logs, Tempo/OpenTelemetry traces, health/readiness endpoints, alerts, Docker Compose monitoring, or trace-to-log correlation.'
argument-hint: 'Describe the observability signal or integration failure to implement.'
---
# Manual-RAG Observability

## Workflow
1. Inspect current metrics, loggers, tracing providers, endpoints, Compose services, datasources, dashboards, and alert rules.
2. Keep instrumentation in adapters/middleware and composition roots; do not import observability frameworks into domain or ports.
3. Use low-cardinality labels. Never label or log passwords, JWTs, Authorization headers, full prompts, questions, document text, or model responses.
4. Propagate W3C `traceparent` across Go, FastAPI, Qdrant, and Ollama where the dependency supports it.
5. Validate each signal independently: endpoint, Prometheus target, Loki query, Tempo trace, and Grafana datasource.
6. Run focused Go/Python checks before Docker smoke tests.

## Operational checklist
- `/health` is cheap liveness; `/ready` checks dependencies; `/metrics` is Prometheus format.
- JSON logs include service context and optional trace ID without sensitive payloads.
- Promtail/Alloy sends Docker logs to Loki and uses stable labels only.
- Tempo receives OTLP on `4317` or `4318` and Grafana has trace-to-log links.
- Alerts use metric names that exist and are tested against Prometheus targets.
- Non-streaming LLM calls must not claim a real TTFT; implement streaming before reporting first-token latency.
