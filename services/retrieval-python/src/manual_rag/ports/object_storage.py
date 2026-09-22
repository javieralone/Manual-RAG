from pathlib import Path
from typing import Protocol


class ObjectStorage(Protocol):
    def download(self, bucket: str, object_key: str, destination: Path) -> None: ...