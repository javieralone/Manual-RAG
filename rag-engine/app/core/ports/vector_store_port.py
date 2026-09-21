from abc import ABC, abstractmethod
from typing import List
from app.core.domain.schemas import DocumentChunk

class VectorStorePort(ABC):
    @abstractmethod
    def search_similar(self, query_vector: List[float], top_k: int) -> List[DocumentChunk]:
        """Busca fragmentos similares dado un vector de consulta."""
        pass