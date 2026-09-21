from typing import List
from sentence_transformers import SentenceTransformer
from app.core.ports.embedding_port import EmbeddingPort

class SentenceTransformerAdapter(EmbeddingPort):
    def __init__(self, model_name: str = "paraphrase-multilingual-MiniLM-L12-v2"):
        # Cargamos el modelo ligero optimizado para CPU/8GB RAM
        self.model = SentenceTransformer(model_name)

    def generate_embedding(self, text: str) -> List[float]:
        embedding = self.model.encode(text, convert_to_numpy=True)
        return embedding.tolist()