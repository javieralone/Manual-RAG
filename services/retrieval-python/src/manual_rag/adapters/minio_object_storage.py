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

    def list_buckets(self) -> list[str]:
        return sorted(bucket.name for bucket in self.client.list_buckets())

    def list_object_keys(self, bucket: str) -> list[str]:
        if not bucket:
            return []
        return sorted(item.object_name for item in self.client.list_objects(bucket, recursive=True))

    def upload(self, bucket: str, object_key: str, source: object, length: int, content_type: str) -> None:
        if not bucket or not object_key:
            raise ValueError("bucket y object_key son obligatorios")
        if not self.client.bucket_exists(bucket):
            self.client.make_bucket(bucket)
        self.client.put_object(bucket, object_key, source, length=length, content_type=content_type)