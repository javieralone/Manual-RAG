import os
from pathlib import Path


class MinioObjectStorage:
    def __init__(
        self,
        endpoint: str,
        access_key: str,
        secret_key: str,
        secure: bool = False,
    ):
        from minio import Minio

        self.client = Minio(endpoint, access_key=access_key, secret_key=secret_key, secure=secure)

    @classmethod
    def from_environment(cls) -> "MinioObjectStorage":
        return cls(
            endpoint=os.getenv("MINIO_ENDPOINT", "localhost:9000"),
            access_key=os.getenv("MINIO_ROOT_USER", "manuals"),
            secret_key=os.getenv("MINIO_ROOT_PASSWORD", ""),
            secure=os.getenv("MINIO_SECURE", "false").lower() == "true",
        )

    def download(self, bucket: str, object_key: str, destination: Path) -> None:
        if not bucket or not object_key:
            raise ValueError("bucket y object_key son obligatorios para una descarga MinIO")
        if not self.client.bucket_exists(bucket):
            self.client.make_bucket(bucket)
        destination.parent.mkdir(parents=True, exist_ok=True)
        self.client.fget_object(bucket, object_key, str(destination))