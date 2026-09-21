---
name: go-clean-architecture
description: 'Use when changing api-go Go code: Go, Clean Architecture, domain, ports, services, JWT authentication, authorization, query orchestration, HTTP handlers, RAG/Ollama clients, context cancellation, timeouts, worker pools, metrics, tracing, or Go tests.'
argument-hint: 'Describe the Go gateway behavior or failing check.'
---
# Go Clean Architecture

## Workflow
1. Trace the request through handler, service, port, and adapter.
2. Identify the layer that owns the behavior before editing.
3. Keep domain and ports free of infrastructure imports.
4. Preserve context propagation through clients and middleware.
5. Add focused tests around the changed use case or contract.
6. Run `go build ./...` and the narrowest relevant `go test` command from `api-go`.

## Contract checklist
- Empty or invalid questions are rejected by the use-case boundary.
- HTTP concerns stay in handlers and middleware.
- Python and Ollama failures remain distinguishable from validation errors.
- Concurrency limits do not leak goroutines or block cancellation.
- JSON changes are intentional and reflected at both integration ends.
