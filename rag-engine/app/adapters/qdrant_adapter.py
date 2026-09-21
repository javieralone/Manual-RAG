from typing import List
from qdrant_client import QdrantClient
from qdrant_client.http.exceptions import UnexpectedResponse
from app.core.ports.vector_store_port import VectorStorePort
from app.core.domain.schemas import DocumentChunk

class QdrantAdapter(VectorStorePort):
    def __init__(self, host: str = "qdrant", port: int = 6333, collection_name: str = "manuales_tecnicos"):
        self.client = QdrantClient(host=host, port=port)
        self.collection_name = collection_name

    def search_similar(self, query_vector: List[float], top_k: int) -> List[DocumentChunk]:
        try:
            # Usamos query_points en lugar del método obsoleto search()
            response = self.client.query_points(
                collection_name=self.collection_name,
                query=query_vector,
                limit=top_k
            )
            
            results = []
            for hit in response.points:
                payload = hit.payload or {}
                # Mapeo flexible del campo de texto principal del chunk
                text = payload.get("text") or payload.get("content") or payload.get("page_content") or ""
                
                results.append(
                    DocumentChunk(
                        text=text,
                        score=hit.score,
                        metadata=payload.get("metadata", {})
                    )
                )
            return results
        except UnexpectedResponse as e:
            if e.status_code == 404:
                return []
            raise e
        except Exception as e:
            # Si el método fallback antiguo sigue disponible en tu versión:
            try:
                search_result = self.client.search(
                    collection_name=self.collection_name,
                    query_vector=query_vector,
                    limit=top_k
                )
                results = []
                for hit in search_result:
                    payload = hit.payload or {}
                    text = payload.get("text") or payload.get("content") or ""
                    results.append(
                        DocumentChunk(
                            text=text,
                            score=hit.score,
                            metadata=payload.get("metadata", {})
                        )
                    )
                return results
            except Exception:
                return []