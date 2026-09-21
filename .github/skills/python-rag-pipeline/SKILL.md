---
name: python-rag-pipeline
description: 'Use when changing rag-engine Python or RAG retrieval: FastAPI, FastMCP, RAGService, embeddings, SentenceTransformers, BGE, Qdrant/vector search, Pydantic schemas, /search, /ready, /metrics, MCP search_manual, OCR, indexing, ingestion, or Python tests.'
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
- Vector store: Qdrant collection `manuales_tecnicos`.
- HTTP service: port `8000`; MCP service: port `8001`.
- Runtime data lives under `documents/`, `output/`, and `qdrant_storage/`.
