import json
from typing import Any

from manual_rag.domain.ingestion import IngestionJob


class RedisIngestionJobStore:
    """Redis JSON records isolated from gateway session keys by DB and prefix."""

    def __init__(self, client: Any, prefix: str = "manual-rag:ingestion"):
        self.client = client
        self.prefix = prefix.rstrip(":")

    def _job_key(self, job_id: str) -> str:
        return f"{self.prefix}:job:{job_id}"

    def _identity_key(self, collection: str, file_sha256: str) -> str:
        return f"{self.prefix}:identity:{collection}:{file_sha256}"

    def get(self, job_id: str) -> IngestionJob | None:
        raw = self.client.get(self._job_key(job_id))
        if raw is None:
            return None
        if isinstance(raw, bytes):
            raw = raw.decode("utf-8")
        return IngestionJob.from_dict(json.loads(raw))

    def save(self, job: IngestionJob) -> None:
        self.client.set(self._job_key(job.job_id), json.dumps(job.to_dict()))

    def claim_identity(self, collection: str, file_sha256: str, job_id: str) -> str | None:
        key = self._identity_key(collection, file_sha256)
        claimed = self.client.set(key, job_id, nx=True)
        if claimed:
            return None
        existing = self.client.get(key)
        return existing.decode("utf-8") if isinstance(existing, bytes) else existing

    def list_jobs(self) -> list[IngestionJob]:
        jobs = []
        for key in self.client.scan_iter(match=f"{self.prefix}:job:*"):
            job_id = key.decode("utf-8").rsplit(":", 1)[-1] if isinstance(key, bytes) else key.rsplit(":", 1)[-1]
            job = self.get(job_id)
            if job:
                jobs.append(job)
        return sorted(jobs, key=lambda job: job.created_at or 0, reverse=True)