import os
from functools import lru_cache

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