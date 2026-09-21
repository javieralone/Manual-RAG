from abc import ABC, abstractmethod
from typing import List
from app.core.domain.schemas import ChunkResult

class VectorStorePort(ABC):
    @abstractmethod
    def search_similar(self, query_vector: List[float], top_k: int) -> List[ChunkResult]:
        """Recupera los puntos más similares dada una representación vectorial."""
        pass