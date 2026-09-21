# Manual-RAG workspace guidance

## Project shape
- This repository is a local RAG system composed of `api-go`, `rag-engine`, Qdrant, and Ollama.
- `api-go` is the public HTTP gateway and owns query orchestration, timeouts, cancellation, and concurrency limits.
- `rag-engine` is the Python retrieval service and MCP server. It owns embeddings, vector search, ingestion scripts, and retrieval DTOs.
- Qdrant collection: `manuales_tecnicos`.
- Main runtime endpoints: Go `:8080`, RAG `:8000`, MCP `:8001`, Qdrant `:6333`.

## Architecture invariants
- Keep business logic behind ports/interfaces. HTTP, Ollama, Qdrant, and SentenceTransformers are adapters.
- Preserve the dependency direction: domain/core -> ports -> adapters, never the reverse.
- Keep FastAPI and MCP entrypoints reusing the same `RAGService`; do not duplicate retrieval logic.
- Propagate request cancellation and timeouts across Go outbound calls.
- Treat document files, generated JSON, and `qdrant_storage` as runtime data, not source code.
- Do not expose secrets, local model caches, or vector-store internals in API responses.

## Working rules
- Before editing, inspect the nearest owning service, port, adapter, and test or call site.
- Prefer the smallest change that preserves public contracts and existing naming.
- For cross-service changes, update both sides of the HTTP contract and validate with Docker Compose or focused local checks.
- Use `go build ./...` for Go changes and the repository's Python checks or a focused import/compile check for Python changes.
- Do not rewrite README claims unless the implementation and deployment behavior have been verified.
- Report stale documentation, missing tests, or environment blockers explicitly.

## Delegation
- Use `go-api-specialist` for Go gateway, ports, handlers, clients, middleware, and worker-pool work.
- Use `python-rag-specialist` for embeddings, Qdrant, FastAPI, MCP, ingestion, and Python domain code.
- Use `integration-reviewer` for Docker Compose, cross-service contracts, runtime checks, and code review.
- Use `rag-system-architect` when a request spans more than one service or needs sequencing.
- Use the `observability-stack` skill for Prometheus, Grafana, Loki/Promtail, Tempo/OpenTelemetry, alerts, health/readiness, and trace-to-log work.
