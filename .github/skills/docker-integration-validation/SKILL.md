---
name: docker-integration-validation
description: 'Use when validating Manual-RAG across Docker Compose services: Qdrant, rag-engine, MCP, api-go, Ollama connectivity, health checks, ports, environment variables, and end-to-end query flow.'
argument-hint: 'Describe the service flow or integration failure to validate.'
---
# Docker Integration Validation

## Workflow
1. Inspect `docker-compose.yml` and the affected Dockerfiles.
2. Run `docker compose config` to catch malformed or unresolved configuration.
3. Build or start only the affected services when possible.
4. Check `/health` endpoints before testing query flow.
5. Test `rag-engine /search`, then `api-go /api/v1/query` so failures are localized.
6. Check logs and resource assumptions before changing application code.

## Service contract
- Qdrant: `6333`.
- RAG HTTP: `8000`.
- MCP: `8001`.
- Go API: `8080`.
- Go uses `PYTHON_ENGINE_URL` and `OLLAMA_URL`.
- RAG uses `QDRANT_HOST`, `QDRANT_PORT`, and bounded CPU thread settings.

## Reporting
Separate configuration errors, application errors, unavailable dependencies, and data/model initialization failures. Do not alter persistent vector data during validation.
