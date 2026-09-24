import tempfile
import unittest
import subprocess
from datetime import datetime, timezone
from pathlib import Path
from unittest.mock import patch

from manual_rag.application.ingestion_jobs import IngestionJobService, mark_job_failure
from manual_rag.domain.ingestion import IngestionJob, IngestionRequest, IngestionResult, JobStatus


class FakeStore:
    def __init__(self):
        self.jobs = {}
        self.identities = {}

    def get(self, job_id):
        return self.jobs.get(job_id)

    def save(self, job):
        self.jobs[job.job_id] = job

    def claim_identity(self, collection, file_sha256, job_id):
        key = (collection, file_sha256)
        if key in self.identities:
            return self.identities[key]
        self.identities[key] = job_id
        return None

    def list_jobs(self):
        return list(self.jobs.values())


class FakeQueue:
    def __init__(self):
        self.calls = []

    def enqueue(self, *args, **kwargs):
        self.calls.append((args, kwargs))


class FailingQueue:
    def enqueue(self, *_args, **_kwargs):
        raise RuntimeError("queue unavailable")


class FakeStorage:
    def download(self, bucket, object_key, destination):
        self.bucket = bucket
        self.object_key = object_key
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_bytes(b"minio fixture")


class IngestionJobServiceTests(unittest.TestCase):
    def test_worker_persists_completed_and_removes_temporary_source(self):
        from manual_rag.entrypoints import worker

        with tempfile.TemporaryDirectory() as temp_dir:
            source = Path(temp_dir) / "downloaded.pdf"
            source.write_bytes(b"downloaded fixture")
            job = IngestionJob(
                "job-completed",
                JobStatus.PENDING,
                "generic_manuals",
                "sha",
                local_path=str(source),
                temporary_local_path=True,
            )
            store = FakeStore()
            store.save(job)
            finished = datetime.now(timezone.utc)
            result = IngestionResult(
                job_id=job.job_id,
                collection=job.collection,
                file_sha256=job.file_sha256,
                status=JobStatus.COMPLETED,
                workspace=Path(temp_dir) / "workspace",
                started_at=finished,
                finished_at=finished,
            )

            class FakeRunner:
                def __init__(self, *_args, **_kwargs):
                    pass

                def run(self, _request):
                    return result

            with patch.object(worker, "_store", return_value=store), patch.object(worker, "IngestionRunner", FakeRunner):
                completed = worker.process_ingestion_job(job.job_id, job.to_dict())

            self.assertEqual(completed["status"], JobStatus.COMPLETED.value)
            self.assertEqual(store.get(job.job_id).status, JobStatus.COMPLETED)
            self.assertFalse(source.exists())

    def test_enqueue_downloads_minio_object_when_local_path_is_absent(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            store = FakeStore()
            queue = FakeQueue()
            storage = FakeStorage()
            service = IngestionJobService(store, queue, root, storage=storage)

            job = service.submit(IngestionRequest(
                bucket="manuals",
                object_key="manuals/manuales_tecnicos/manual.pdf",
            ))

            self.assertEqual(job.collection, "manuales_tecnicos")
            self.assertEqual(storage.bucket, "manuals")
            self.assertEqual(storage.object_key, "manuals/manuales_tecnicos/manual.pdf")
            self.assertTrue(Path(job.local_path).is_file())
            self.assertEqual(len(queue.calls), 1)

    def test_minio_download_is_removed_when_enqueue_fails(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            storage = FakeStorage()
            service = IngestionJobService(FakeStore(), FailingQueue(), root, storage=storage)

            with self.assertRaisesRegex(RuntimeError, "queue unavailable"):
                service.submit(IngestionRequest(bucket="manuals", object_key="generic_manuals/manual.pdf"))

            self.assertEqual(list((root / ".minio-cache").glob("*.pdf")), [])

    def test_enqueue_and_duplicate_return_persisted_job(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            source = root / "manual.pdf"
            source.write_bytes(b"fixture")
            store = FakeStore()
            queue = FakeQueue()
            service = IngestionJobService(store, queue, root)

            first = service.submit(IngestionRequest(source))
            duplicate = service.submit(IngestionRequest(source))

            self.assertEqual(first.job_id, duplicate.job_id)
            self.assertEqual(first.status, JobStatus.PENDING)
            self.assertEqual(len(queue.calls), 1)
            self.assertEqual(queue.calls[0][1]["job_timeout"], 1800)

    def test_failure_exhaustion_is_terminal_and_keeps_traceback(self):
        store = FakeStore()
        job = IngestionJob("job-1", JobStatus.PROCESSING, "generic_manuals", "sha")
        store.save(job)
        try:
            raise RuntimeError("pipeline failed")
        except RuntimeError as error:
            mark_job_failure(store, job, error, permanent=True)

        self.assertEqual(store.get("job-1").status, JobStatus.FAILED_PERMANENTLY)
        self.assertEqual(store.get("job-1").error, "pipeline failed")
        self.assertIn("RuntimeError", store.get("job-1").traceback)

    def test_retryable_timeout_has_timeout_status(self):
        store = FakeStore()
        job = IngestionJob("job-2", JobStatus.PROCESSING, "generic_manuals", "sha")
        mark_job_failure(store, job, subprocess.TimeoutExpired("pipeline", 1), permanent=False, timed_out=True)

        self.assertEqual(store.get("job-2").status, JobStatus.TIMEOUT)


if __name__ == "__main__":
    unittest.main()