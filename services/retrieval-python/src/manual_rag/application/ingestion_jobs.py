import hashlib
import os
import traceback as traceback_module
import uuid
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

from manual_rag.application.ingestion_runner import resolve_collection
from manual_rag.domain.ingestion import IngestionJob, IngestionRequest, JobStatus
from manual_rag.domain.schemas import resolve_collection as validate_collection
from manual_rag.ports.ingestion_job_store import IngestionJobStore
from manual_rag.ports.object_storage import ObjectStorage


def _sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


class IngestionJobService:
    def __init__(self, store: IngestionJobStore, queue: Any, local_root: Path, storage: ObjectStorage | None = None):
        self.store = store
        self.queue = queue
        self.local_root = local_root.resolve()
        self.storage = storage

    def submit(self, request: IngestionRequest) -> IngestionJob:
        job_id = request.job_id or str(uuid.uuid4())
        source = request.pdf_path.resolve() if request.pdf_path else None
        if source is None:
            if self.storage is None or not request.bucket or not request.object_key:
                raise ValueError("Se requiere pdf_path local o bucket y object_key MinIO")
            source = self.local_root / ".minio-cache" / f"{job_id}.pdf"
            self.storage.download(request.bucket, request.object_key, source)
        if not source.is_relative_to(self.local_root):
            raise ValueError("El PDF debe estar dentro de INGESTION_LOCAL_ROOT")
        if not source.is_file() or source.suffix.lower() != ".pdf":
            raise ValueError("El archivo local debe existir y ser PDF")

        collection = validate_collection(resolve_collection(request.object_key, request.collection))
        file_sha256 = _sha256_file(source)
        existing_id = self.store.claim_identity(collection, file_sha256, job_id)
        if existing_id:
            existing = self.store.get(existing_id)
            if existing:
                return existing

        now = datetime.now(timezone.utc)
        job = IngestionJob(
            job_id=job_id,
            status=JobStatus.PENDING,
            collection=collection,
            file_sha256=file_sha256,
            object_key=request.object_key,
            bucket=request.bucket,
            local_path=str(source),
            created_at=now,
        )
        self.store.save(job)
        self.queue.enqueue(
            "manual_rag.entrypoints.worker.process_ingestion_job",
            job.job_id,
            job.to_dict(),
            job_timeout=int(os.getenv("INGESTION_JOB_TIMEOUT_SECONDS", "1800")),
            retry=self._retry_policy(),
        )
        return job

    @staticmethod
    def _retry_policy() -> Any:
        try:
            from rq import Retry
        except ImportError:
            return None
        return Retry(max=int(os.getenv("INGESTION_MAX_RETRIES", "2")), interval=[10, 30, 90])


def mark_job_processing(store: IngestionJobStore, job: IngestionJob, worker_id: str) -> IngestionJob:
    job.status = JobStatus.PROCESSING
    job.started_at = datetime.now(timezone.utc)
    job.worker_id = worker_id
    store.save(job)
    return job


def mark_job_failure(
    store: IngestionJobStore,
    job: IngestionJob,
    error: Exception,
    permanent: bool,
    timed_out: bool = False,
) -> None:
    job.status = JobStatus.FAILED_PERMANENTLY if permanent else (JobStatus.TIMEOUT if timed_out else JobStatus.FAILED)
    job.error = str(error)
    job.traceback = traceback_module.format_exc()
    job.finished_at = datetime.now(timezone.utc)
    job.retry_count += 1
    store.save(job)