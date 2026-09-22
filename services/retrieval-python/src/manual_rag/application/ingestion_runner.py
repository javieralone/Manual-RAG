import hashlib
import shutil
import subprocess
import sys
import uuid
from datetime import datetime, timezone
from pathlib import Path

from manual_rag.domain.ingestion import IngestionRequest, IngestionResult, JobStatus


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as file:
        for block in iter(lambda: file.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def resolve_collection(object_key: str, explicit: str | None = None) -> str:
    if explicit:
        return explicit
    parts = Path(object_key).parts
    return parts[1] if len(parts) > 2 and parts[0] == "manuals" else "generic_manuals"


class IngestionRunner:
    def __init__(self, workspace_root: Path, scripts_dir: Path, python_executable: str = sys.executable):
        self.workspace_root = workspace_root
        self.scripts_dir = scripts_dir
        self.python_executable = python_executable

    def run(self, request: IngestionRequest) -> IngestionResult:
        source = request.pdf_path.resolve()
        if not source.is_file() or source.suffix.lower() != ".pdf":
            raise FileNotFoundError(f"El archivo no existe o no es PDF: {source}")

        job_id = request.job_id or str(uuid.uuid4())
        collection = resolve_collection(request.object_key, request.collection)
        workspace = self.workspace_root / job_id
        workspace.mkdir(parents=True, exist_ok=False)
        started_at = datetime.now(timezone.utc)
        file_sha256 = sha256_file(source)
        input_pdf = workspace / source.name
        pages_path = workspace / "manual_pages.json"
        chunks_path = workspace / "manual_chunks.json"
        shutil.copy2(source, input_pdf)

        try:
            commands = [
                [self.python_executable, str(self.scripts_dir / "process_manual_opt.py"),
                 "--pdf", str(input_pdf), "--collection", collection,
                 "--pages-output", str(pages_path), "--chunks-output", str(chunks_path),
                 "--job-id", job_id, "--object-key", request.object_key,
                 "--file-sha256", file_sha256],
            ]
            for command in commands:
                subprocess.run(command, check=True)
            status = JobStatus.COMPLETED
        except subprocess.TimeoutExpired:
            status = JobStatus.TIMEOUT
            raise
        except Exception:
            status = JobStatus.FAILED
            raise
        finally:
            finished_at = datetime.now(timezone.utc)
            shutil.rmtree(workspace, ignore_errors=True)

        return IngestionResult(job_id, collection, file_sha256, status, workspace, started_at, finished_at)
