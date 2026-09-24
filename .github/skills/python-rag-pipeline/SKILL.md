---
name: python-rag-pipeline
description: 'Use when changing the Python retrieval service or RAG pipeline: FastAPI, FastMCP, RAGService, embeddings, SentenceTransformers, BGE, Qdrant/vector search, Pydantic schemas, /search, /ready, /metrics, MCP search_manual, OCR, indexing, ingestion, or Python tests.'
argument-hint: 'Describe the retrieval, ingestion, or MCP behavior to change.'
---
# Python RAG Pipeline

## Workflow
1. Trace the entrypoint into `RAGService`, ports, and concrete adapters.
2. Confirm the schema and collection contract before editing.
3. Keep embedding and vector-store libraries behind ports.
4. Reuse `RAGService` from FastAPI and MCP.
5. Validate empty queries, `top_k`, result metadata, and adapter failures.
6. Run a focused Python import/compile or test check and report missing runtime dependencies.

## Runtime facts
- Embedding model: `BAAI/bge-m3` through the embedding adapter.
- Vector store: Qdrant collection `generic_manuals` by default; `manuales_tecnicos` remains an explicitly selectable collection and the default technical MCP domain target.
- HTTP service: port `8000`; MCP service: port `8001`.
- Runtime data lives under `data/documents/`, `data/artifacts/`, and `data/local/qdrant/`.
- Feature 01 must preserve the legacy folder-based flow while adding dynamic CLI paths for per-job PDF/pages/chunks; do not let worker jobs write shared runtime artifacts.
- Upload metadata is a cross-cutting contract: `job_id`, `minio_object_key`, `file_sha256`, `ingested_at`, `collection`, `document_id`, `part`, `page`, and `source` must survive into Qdrant payloads.
- Route queue orchestration, job persistence, MinIO adapters, retries, DLQ, and cleanup lifecycle to `ingestion-specialist` and use the dedicated `ingestion-pipeline` skill.
