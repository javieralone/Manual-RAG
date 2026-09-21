from typing import List
from qdrant_client import QdrantClient
from app.core.ports.vector_store_port import VectorStorePort
from app.core.domain.schemas import DocumentChunk

class QdrantAdapter(VectorStorePort):
    def __init__(self, host: str = "qdrant", port: int = 6333, collection_name: str = "documents"):
        self.client = QdrantClient(host=host, port=port)
        self.collection_name = collection_name

    def search_similar(self, query_vector: List[float], top_k: int) -> List[DocumentChunk]:
        search_result = self.client.search(
            collection_name=self.collection_name,
            query_vector=query_vector,
            limit=top_k
        )
        
        results = []
        for hit in search_result:
            results.append(
                DocumentChunk(
                    text=hit.payload.get("text", ""),
                    score=hit.score,
                    metadata=hit.payload.get("metadata", {})
                )
            )
        return results