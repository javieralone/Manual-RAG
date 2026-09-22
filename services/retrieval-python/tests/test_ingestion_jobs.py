import tempfile
import unittest
import subprocess
from pathlib import Path

from manual_rag.application.ingestion_jobs import IngestionJobService, mark_job_failure
from manual_rag.domain.ingestion import IngestionJob, IngestionRequest, JobStatus


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


class FakeStorage:
    def download(self, bucket, object_key, destination):
        self.bucket = bucket
        self.object_key = object_key
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_bytes(b"minio fixture")


class IngestionJobServiceTests(unittest.TestCase):
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