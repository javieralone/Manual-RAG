from abc import ABC, abstractmethod


class AnswerGeneratorPort(ABC):
    @abstractmethod
    def generate_answer(self, query: str, context: str) -> str:
        """Genera una respuesta a partir de la consulta y el contexto recuperado."""
        pass
