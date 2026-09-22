# Manual-RAG workspace guidance

## Project shape
- This repository is a local RAG system composed of `services/gateway-go`, `services/retrieval-python`, `apps/web`, Qdrant, and Ollama.
- `api-go` is the public HTTP gateway and owns query orchestration, timeouts, cancellation, and concurrency limits.
- `rag-engine` is the Python retrieval service and MCP server. It owns embeddings, vector search, ingestion scripts, and retrieval DTOs.
- `frontend` is the React + Vite UI for login, chat, collection selection and streaming query mode.
- Qdrant default collection: `generic_manuals`; `manuales_tecnicos` is an explicit supported collection and the default target for the `technical_manuals` MCP domain.
- Main runtime endpoints: Go `:8080`, RAG `:8000`, MCP `:8001`, Qdrant `:6333`, frontend `:5173`.
- Feature 01 ingestion adds a separate operational flow around `services/retrieval-python`: Redis RQ, MinIO/local input, isolated per-job workspaces, job state, retries/DLQ, and Qdrant provenance metadata. It must remain compatible with the existing folder-based scripts during transition.

## Architecture invariants
- Keep business logic behind ports/interfaces. HTTP, Ollama, Qdrant, and SentenceTransformers are adapters.
- Preserve the dependency direction: domain/core -> ports -> adapters, never the reverse.
- Keep FastAPI and MCP entrypoints reusing the same `RAGService`; do not duplicate retrieval logic.
- Propagate request cancellation and timeouts across Go outbound calls.
- Treat document files, generated JSON, and `qdrant_storage` as runtime data, not source code.
- Do not expose secrets, local model caches, or vector-store internals in API responses.
- Do not mix ingestion queue/state keys with gateway refresh-session keys in Redis; use a separate database or explicit prefixes and document the choice.
- Do not claim feature 01 is complete without testing idempotency, retry exhaustion/DLQ, cleanup on failure, and metadata on uploaded chunks.

## Working rules
- Before editing, inspect the nearest owning service, port, adapter, and test or call site.
- Prefer the smallest change that preserves public contracts and existing naming.
- For cross-service changes, update both sides of the HTTP contract and validate with Docker Compose or focused local checks.
- For UI work in `apps/web`, keep the React + Vite app aligned with the API Gateway contract and document any assumptions in `apps/web/README.md`.
- Use `go build ./...` for Go changes and the repository's Python checks or a focused import/compile check for Python changes.
- For ingestion changes, validate script CLI compatibility, Python compile/import checks, and `docker compose --env-file .env.example config --quiet` before model-dependent or full-stack checks.
- Validate frontend changes with `npm run build` before declaring the UI ready.
- Do not rewrite README claims unless the implementation and deployment behavior have been verified.
- Report stale documentation, missing tests, or environment blockers explicitly.

## Delegation
- Use `rag-system-architect` first for tasks mentioning more than one service, architecture, contracts, authentication across services, deployment sequencing, or broad feature work.
- Use `go-api-specialist` for Go, `api-go`, domain, ports, services, handlers, JWT auth, clients, middleware, timeouts, worker pools, or Go tests.
- Use `python-rag-specialist` for Python, `rag-engine`, FastAPI, FastMCP, RAGService, embeddings, Qdrant, OCR, indexing, schemas, or Python tests.
- Use `observability-specialist` for metrics, Prometheus, Grafana, dashboards, logs, Loki, Promtail, Alloy, traces, Tempo, OpenTelemetry, OTLP, TraceQL, LogQL, alerts, readiness, scraping, or telemetry correlation.
- Use `integration-reviewer` for Docker Compose, Dockerfiles, startup, runtime smoke tests, cross-service contracts, health checks, resource limits, and regression review.
- Use `ingestion-specialist` for Redis RQ, MinIO, ingestion API, job lifecycle, worker isolation, retries/DLQ, deduplication, pipeline parameterization, or Qdrant ingestion metadata.

## Automatic task routing
- Select the most specific specialist based on the task keywords above before editing.
- When multiple routing categories match, use `rag-system-architect` to coordinate and delegate; do not duplicate the same investigation in several agents.
- Always load the relevant skill after selecting the agent: `go-clean-architecture`, `python-rag-pipeline`, `observability-stack`, or `docker-integration-validation`.
- For feature 01, always load `ingestion-pipeline`; combine it with `python-rag-pipeline`, `docker-integration-validation`, `go-clean-architecture`, or `observability-stack` only when those boundaries are touched.
