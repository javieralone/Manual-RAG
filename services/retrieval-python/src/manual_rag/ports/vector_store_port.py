from abc import ABC, abstractmethod
from typing import Dict, List, Optional
from manual_rag.domain.schemas import ChunkResult

class VectorStorePort(ABC):
    @abstractmethod
    def search_similar(
        self,
        query_vector: List[float],
        top_k: int,
        filters: Optional[Dict[str, str]] = None,
        collection_name: Optional[str] = None,
    ) -> List[ChunkResult]:
        """Recupera los puntos más similares dada una representación vectorial."""
        pass