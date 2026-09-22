import time
from typing import Dict, List, Optional
from qdrant_client import QdrantClient
from qdrant_client import models
from qdrant_client.http.exceptions import UnexpectedResponse
from manual_rag.ports.vector_store_port import VectorStorePort
from manual_rag.domain.schemas import ChunkResult, resolve_collection
from manual_rag.observability import errors_total, qdrant_duration
from opentelemetry import trace

class QdrantAdapter(VectorStorePort):
    def __init__(self, host: str, port: int, collection_name: str = "manuales_tecnicos", timeout: float = 5.0, retries: int = 2, retry_backoff: float = 0.2):
        self._client = QdrantClient(host=host, port=port, timeout=timeout)
        self._collection_name = collection_name
        self._retries = max(0, retries)
        self._retry_backoff = max(0.0, retry_backoff)

    def search_similar(
        self,
        query_vector: List[float],
        top_k: int,
        filters: Optional[Dict[str, str]] = None,
        collection_name: Optional[str] = None,
    ) -> List[ChunkResult]:
        started = time.perf_counter()
        tracer = trace.get_tracer("manual-rag/rag-engine")
        collection = resolve_collection(collection_name or self._collection_name)
        with tracer.start_as_current_span("qdrant.query"):
            try:
                response = None
                for attempt in range(self._retries + 1):
                    try:
                        response = self._client.query_points(
                            collection_name=collection,
                            query=query_vector,
                            limit=top_k,
                            query_filter=self._build_filter(filters),
                        )
                        break
                    except UnexpectedResponse:
                        raise
                    except Exception:
                        if attempt >= self._retries:
                            raise
                        time.sleep(self._retry_backoff * (2**attempt))
            
                results = []
                for point in response.points:
                    payload = point.payload or {}
                    metadata = {
                        key: payload[key]
                        for key in ("page", "source", "document_id", "chapter", "section")
                        if key in payload
                    }
                    metadata["collection"] = payload.get("collection", collection)
                    if "part" in payload:
                        metadata["part"] = payload["part"]
                    results.append(
                        ChunkResult(
                            page=payload.get("page", 0),
                            source=payload.get("source", "Desconocido"),
                            text=payload.get("text", ""),
                            score=float(getattr(point, "score", 0.0)),
                            metadata=metadata,
                        )
                    )
                return results
            except UnexpectedResponse as e:
                if e.status_code == 404:
                    return []
                raise e
            except Exception:
                errors_total.labels(component="qdrant").inc()
                raise
            finally:
                qdrant_duration.observe(time.perf_counter() - started)

    @staticmethod
    def _build_filter(filters: Optional[Dict[str, str]]) -> Optional[models.Filter]:
        if not filters:
            return None

        conditions = [
            models.FieldCondition(
                key=key,
                match=models.MatchValue(value=value),
            )
            for key, value in filters.items()
            if value
        ]
        return models.Filter(must=conditions) if conditions else None

    def check_ready(self) -> None:
        self._client.get_collection(self._collection_name)