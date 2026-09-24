from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any


class JobStatus(str, Enum):
    PENDING = "PENDING"
    PROCESSING = "PROCESSING"
    COMPLETED = "COMPLETED"
    SKIPPED = "SKIPPED"
    TIMEOUT = "TIMEOUT"
    FAILED = "FAILED"
    FAILED_PERMANENTLY = "FAILED_PERMANENTLY"


@dataclass(frozen=True)
class IngestionRequest:
    pdf_path: Path | None = None
    collection: str | None = None
    object_key: str = ""
    job_id: str = ""
    bucket: str = ""


@dataclass(frozen=True)
class IngestionResult:
    job_id: str
    collection: str
    file_sha256: str
    status: JobStatus
    workspace: Path
    started_at: datetime
    finished_at: datetime


@dataclass
class IngestionJob:
    job_id: str
    status: JobStatus
    collection: str
    file_sha256: str
    object_key: str = ""
    bucket: str = ""
    local_path: str = ""
    temporary_local_path: bool = False
    created_at: datetime | None = None
    started_at: datetime | None = None
    finished_at: datetime | None = None
    retry_count: int = 0
    worker_id: str = ""
    error: str = ""
    traceback: str = ""
    result_summary: dict[str, Any] | None = None

    def to_dict(self) -> dict[str, Any]:
        return {
            "job_id": self.job_id,
            "status": self.status.value,
            "collection": self.collection,
            "file_sha256": self.file_sha256,
            "object_key": self.object_key,
            "bucket": self.bucket,
            "local_path": self.local_path,
            "temporary_local_path": self.temporary_local_path,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "started_at": self.started_at.isoformat() if self.started_at else None,
            "finished_at": self.finished_at.isoformat() if self.finished_at else None,
            "retry_count": self.retry_count,
            "worker_id": self.worker_id,
            "error": self.error,
            "traceback": self.traceback,
            "result_summary": self.result_summary,
        }

    @classmethod
    def from_dict(cls, value: dict[str, Any]) -> "IngestionJob":
        def parse_timestamp(raw: str | None) -> datetime | None:
            return datetime.fromisoformat(raw) if raw else None

        return cls(
            job_id=value["job_id"],
            status=JobStatus(value["status"]),
            collection=value["collection"],
            file_sha256=value["file_sha256"],
            object_key=value.get("object_key", ""),
            bucket=value.get("bucket", ""),
            local_path=value.get("local_path", ""),
            temporary_local_path=bool(value.get("temporary_local_path", False)),
            created_at=parse_timestamp(value.get("created_at")),
            started_at=parse_timestamp(value.get("started_at")),
            finished_at=parse_timestamp(value.get("finished_at")),
            retry_count=int(value.get("retry_count", 0)),
            worker_id=value.get("worker_id", ""),
            error=value.get("error", ""),
            traceback=value.get("traceback", ""),
            result_summary=value.get("result_summary"),
        )
