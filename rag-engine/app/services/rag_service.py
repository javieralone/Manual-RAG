from typing import List
from app.core.ports.embedding_port import EmbeddingPort
from app.core.ports.vector_store_port import VectorStorePort
from app.core.domain.schemas import DocumentChunk

class RAGService:
    def __init__(self, embedding_provider: EmbeddingPort, vector_store: VectorStorePort):
        self._embedding_provider = embedding_provider
        self._vector_store = vector_store

    def retrieve_relevant_chunks(self, query_text: str, top_k: int = 3) -> List[DocumentChunk]:
        # 1. Convertir texto a vector usando el puerto de embeddings
        query_vector = self._embedding_provider.generate_embedding(query_text)
        
        # 2. Buscar vectores cercanos usando el puerto de la base vectorial
        chunks = self._vector_store.search_similar(query_vector, top_k)
        
        return chunks