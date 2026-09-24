---
name: ingestion-pipeline
description: "Use when implementing or validating Manual-RAG feature 01 ingestion: Redis RQ queues and workers, MinIO/local PDF input, job lifecycle APIs, retries and DLQ, SHA256 deduplication, isolated temp workspaces, pipeline CLI parameterization, Qdrant provenance metadata, or ingestion Docker Compose integration."
argument-hint: "Describe the ingestion flow, job failure, contract, or runtime check to implement."
---
# Manual-RAG Feature 01 Ingestion

## Routing
Use `ingestion-specialist` for feature-specific implementation. Use `rag-system-architect` when the change spans ingestion, Go gateway contracts, deployment, or observability. Use `integration-reviewer` for runtime Compose validation and `observability-specialist` for metrics, logs, and traces.

## Contract invariants
- Public gateway remains the system boundary; ingestion orchestration belongs in the Python retrieval service or a dedicated Python ingestion API, not in Go query orchestration.
- Keep gateway refresh-session Redis usage isolated from ingestion queue/state keys, preferably with a separate Redis database or explicit prefixes.
- Supported collection resolution is explicit collection, object-key prefix, then `generic_manuals`. `manuales_tecnicos` remains valid.
- Job states are `PENDING`, `PROCESSING`, `COMPLETED`, `SKIPPED`, `TIMEOUT`, `FAILED`, and `FAILED_PERMANENTLY`.
- A job record includes UUID, source identity, collection, SHA256, timestamps, retry count, worker ID, error/traceback, and result summary.
- RQ timeout and retries must be configured at enqueue/worker boundaries, with exponential or declared backoff and a failed registry treated as the DLQ.
- Deduplication is durable and scoped to `(collection, file_sha256)`; concurrent duplicate submissions need an atomic claim or equivalent race-safe check.
- Each job owns `/tmp/jobs/{job_id}` or the configured equivalent. All PDF, OCR pages, chunks, logs, and intermediate files stay inside it and cleanup runs in `finally`.
- Existing OCR, chunking, embedding, and upload logic is reused. Scripts receive dynamic input/output paths and must not mutate global document folders during a worker job.
- Every uploaded Qdrant chunk has `job_id`, `minio_object_key`, `file_sha256`, `ingested_at`, `collection`, `document_id`, `part`, `page`, and `source`.

## Implementation sequence
1. Define typed job/status/input/result contracts and the persistence boundary.
2. Add storage adapters for MinIO and the local transition path, including size/type validation and SHA256.
3. Parameterize the three existing pipeline scripts and add an orchestration function that owns the isolated workspace.
4. Add RQ enqueue/worker configuration, timeout, retries, worker ID, and failed registry/DLQ behavior.
5. Expose enqueue, list/status, and failed-job endpoints with stable JSON and authorization at the public boundary.
6. Add Compose services, healthchecks, volumes, environment variables, and separate Redis namespace/DB.
7. Add structured logs/metrics/traces and validate the complete flow with deterministic fixtures before a real PDF/model run.

## Validation checklist
- Contract tests cover collection resolution, invalid input, SHA256, duplicate jobs, all terminal transitions, retry exhaustion, and cleanup.
- Script tests prove dynamic paths and metadata without invoking the embedding model.
- `python -m compileall services/retrieval-python/app services/retrieval-python/scripts` passes.
- `docker compose --env-file .env.example config --quiet` passes, and service dependency ordering/readiness is checked.
- A smoke flow verifies enqueue -> processing -> completed, status lookup, failed/DLQ visibility, duplicate skip, Qdrant metadata, and no leftover workspace.
- Never validate by modifying committed documents, generated artifacts, or `data/local/qdrant`.