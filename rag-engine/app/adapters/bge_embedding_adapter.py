from typing import List
from sentence_transformers import SentenceTransformer
from app.core.ports.embedding_port import EmbeddingPort

class BGEEmbeddingAdapter(EmbeddingPort):
    def __init__(self, model_name: str = "BAAI/bge-m3"):
        print(f"Cargando modelo de embeddings: {model_name}...")
        self._model = SentenceTransformer(model_name)

    def generate_embedding(self, text: str) -> List[float]:
        return self._model.encode(text).tolist()