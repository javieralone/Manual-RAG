import time
from typing import List
from qdrant_client import QdrantClient
from qdrant_client.http.exceptions import UnexpectedResponse
from app.core.ports.vector_store_port import VectorStorePort
from app.core.domain.schemas import ChunkResult
from app.observability import errors_total, qdrant_duration

class QdrantAdapter(VectorStorePort):
    def __init__(self, host: str, port: int, collection_name: str = "manuales_tecnicos"):
        self._client = QdrantClient(host=host, port=port)
        self._collection_name = collection_name

    def search_similar(self, query_vector: List[float], top_k: int) -> List[ChunkResult]:
        started = time.perf_counter()
        try:
            response = self._client.query_points(
                collection_name=self._collection_name,
                query=query_vector,
                limit=top_k
            )
            
            results = []
            for point in response.points:
                payload = point.payload or {}
                results.append(
                    ChunkResult(
                        page=payload.get("page", 0),
                        source=payload.get("source", "Desconocido"),
                        text=payload.get("text", ""),
                        score=float(getattr(point, "score", 0.0))
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

    def check_ready(self) -> None:
        self._client.get_collection(self._collection_name)