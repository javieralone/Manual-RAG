import os
from functools import lru_cache
from pathlib import Path

from manual_rag.adapters.bge_embedding_adapter import BGEEmbeddingAdapter
from manual_rag.adapters.qdrant_adapter import QdrantAdapter
from manual_rag.application.rag_service import RAGService
from manual_rag.domain.schemas import DEFAULT_COLLECTION, resolve_collection

QDRANT_HOST = os.getenv("QDRANT_HOST", "localhost")
QDRANT_PORT = int(os.getenv("QDRANT_PORT", 6333))
QDRANT_TIMEOUT = float(os.getenv("QDRANT_TIMEOUT_SECONDS", "5"))
QDRANT_RETRIES = int(os.getenv("QDRANT_RETRY_ATTEMPTS", "2"))
QDRANT_RETRY_BACKOFF = float(os.getenv("QDRANT_RETRY_BACKOFF_SECONDS", "0.2"))
DEFAULT_QDRANT_COLLECTION = os.getenv("DEFAULT_COLLECTION", DEFAULT_COLLECTION)


@lru_cache(maxsize=1)
def get_embedding_adapter() -> BGEEmbeddingAdapter:
    return BGEEmbeddingAdapter(model_name="BAAI/bge-m3")


@lru_cache(maxsize=16)
def get_qdrant_adapter(collection: str = DEFAULT_QDRANT_COLLECTION) -> QdrantAdapter:
    return QdrantAdapter(
        host=QDRANT_HOST,
        port=QDRANT_PORT,
        collection_name=resolve_collection(collection),
        timeout=QDRANT_TIMEOUT,
        retries=QDRANT_RETRIES,
        retry_backoff=QDRANT_RETRY_BACKOFF,
    )


@lru_cache(maxsize=16)
def get_rag_service(collection: str = DEFAULT_QDRANT_COLLECTION) -> RAGService:
    collection_name = resolve_collection(collection)
    return RAGService(
        embedding_provider=get_embedding_adapter(),
        vector_store=get_qdrant_adapter(collection_name),
    )


@lru_cache(maxsize=1)
def get_ingestion_job_service():
    import redis
    from rq import Queue

    from manual_rag.adapters.redis_ingestion_job_store import RedisIngestionJobStore
    from manual_rag.adapters.minio_object_storage import MinioObjectStorage
    from manual_rag.application.ingestion_jobs import IngestionJobService

    connection = redis.Redis.from_url(os.getenv("INGESTION_REDIS_URL", "redis://localhost:6379/2"))
    store = RedisIngestionJobStore(
        connection,
        prefix=os.getenv("INGESTION_REDIS_PREFIX", "manual-rag:ingestion"),
    )
    queue = Queue(os.getenv("INGESTION_QUEUE", "ingestion"), connection=connection)
    return IngestionJobService(
        store=store,
        queue=queue,
        local_root=Path(os.getenv("INGESTION_LOCAL_ROOT", "data/documents/new")),
        storage=MinioObjectStorage.from_environment(),
    )