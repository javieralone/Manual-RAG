import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from manual_rag.application.ingestion_runner import IngestionRunner, resolve_collection, sha256_file
from manual_rag.domain.ingestion import IngestionRequest, JobStatus


class IngestionRunnerTests(unittest.TestCase):
    def test_resolves_collection_from_minio_key(self):
        self.assertEqual(resolve_collection("manuals/manuales_tecnicos/manual.pdf"), "manuales_tecnicos")
        self.assertEqual(resolve_collection("manual.pdf"), "generic_manuals")

    def test_runs_in_isolated_workspace_and_cleans_up(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            source = root / "manual.pdf"
            source.write_bytes(b"pdf fixture")
            runner = IngestionRunner(root / "jobs", root / "scripts", python_executable="python")

            with patch("manual_rag.application.ingestion_runner.subprocess.run") as run:
                result = runner.run(IngestionRequest(source, object_key="manuals/generic_manuals/manual.pdf", job_id="job-1"))

            command = run.call_args.args[0]
            self.assertEqual(result.status, JobStatus.COMPLETED)
            self.assertEqual(result.collection, "generic_manuals")
            self.assertEqual(result.file_sha256, sha256_file(source))
            self.assertIn(str(root / "jobs" / "job-1" / "manual.pdf"), command)
            self.assertFalse((root / "jobs" / "job-1").exists())

    def test_runner_uses_explicit_collection_before_object_key(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            source = root / "manual.pdf"
            source.write_bytes(b"pdf fixture")
            runner = IngestionRunner(root / "jobs", root / "scripts", python_executable="python")

            with patch("manual_rag.application.ingestion_runner.subprocess.run") as run:
                result = runner.run(IngestionRequest(
                    source,
                    collection="custom_collection",
                    object_key="manuals/manuales_tecnicos/manual.pdf",
                    job_id="job-3",
                ))

            self.assertEqual(result.collection, "custom_collection")
            self.assertIn("custom_collection", run.call_args.args[0])

    def test_cleans_up_after_pipeline_failure(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            source = root / "manual.pdf"
            source.write_bytes(b"pdf fixture")
            runner = IngestionRunner(root / "jobs", root / "scripts")

            with patch("manual_rag.application.ingestion_runner.subprocess.run", side_effect=RuntimeError("failed")):
                with self.assertRaises(RuntimeError):
                    runner.run(IngestionRequest(source, job_id="job-2"))

            self.assertFalse((root / "jobs" / "job-2").exists())


if __name__ == "__main__":
    unittest.main()