from abc import ABC, abstractmethod
from typing import List

class EmbeddingPort(ABC):
    @abstractmethod
    def generate_embedding(self, text: str) -> List[float]:
        """Genera la representación vectorial densa de un texto."""
        pass