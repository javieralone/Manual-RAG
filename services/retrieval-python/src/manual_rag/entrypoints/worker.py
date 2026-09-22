import os
import socket
import subprocess
from pathlib import Path
from typing import Any

from manual_rag.adapters.redis_ingestion_job_store import RedisIngestionJobStore
from manual_rag.application.ingestion_jobs import mark_job_failure, mark_job_processing
from manual_rag.application.ingestion_runner import IngestionRunner
from manual_rag.domain.ingestion import IngestionJob, IngestionRequest


def _store() -> RedisIngestionJobStore:
    import redis

    return RedisIngestionJobStore(
        redis.Redis.from_url(os.getenv("INGESTION_REDIS_URL", "redis://localhost:6379/2")),
        prefix=os.getenv("INGESTION_REDIS_PREFIX", "manual-rag:ingestion"),
    )


def process_ingestion_job(job_id: str, serialized_job: dict[str, Any]) -> dict[str, Any]:
    store = _store()
    job = store.get(job_id) or IngestionJob.from_dict(serialized_job)
    mark_job_processing(store, job, f"{socket.gethostname()}:{os.getpid()}")
    request = IngestionRequest(
        pdf_path=Path(job.local_path),
        collection=job.collection,
        object_key=job.object_key,
        bucket=job.bucket,
        job_id=job.job_id,
    )
    runner = IngestionRunner(
        Path(os.getenv("INGESTION_WORKSPACE_ROOT", "/tmp/manual-rag/jobs")),
        Path(os.getenv("INGESTION_SCRIPTS_DIR", "services/retrieval-python/scripts")),
    )
    try:
        result = runner.run(request)
        job.status = result.status
        job.finished_at = result.finished_at
        job.result_summary = {"collection": result.collection, "file_sha256": result.file_sha256}
        store.save(job)
        return job.to_dict()
    except Exception as error:
        current_job = None
        try:
            from rq import get_current_job
            current_job = get_current_job()
        except ImportError:
            pass
        retries_left = getattr(current_job, "retries_left", 0) if current_job else 0
        mark_job_failure(
            store,
            job,
            error,
            permanent=retries_left <= 0,
            timed_out=isinstance(error, subprocess.TimeoutExpired),
        )
        raise


def main() -> None:
    import redis
    from rq import Connection, Queue, Worker

    connection = redis.Redis.from_url(os.getenv("INGESTION_REDIS_URL", "redis://localhost:6379/2"))
    queue = Queue(os.getenv("INGESTION_QUEUE", "ingestion"), connection=connection)
    with Connection(connection):
        Worker([queue], name=os.getenv("INGESTION_WORKER_NAME", socket.gethostname())).work()


if __name__ == "__main__":
    main()