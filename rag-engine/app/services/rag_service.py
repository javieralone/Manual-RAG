from typing import List
from app.core.ports.embedding_port import EmbeddingPort
from app.core.ports.vector_store_port import VectorStorePort
from app.core.domain.schemas import ChunkResult, SearchResponse

class RAGService:
    def __init__(self, embedding_provider: EmbeddingPort, vector_store: VectorStorePort):
        self._embedding_provider = embedding_provider
        self._vector_store = vector_store

    def execute_search(self, query_text: str, top_k: int) -> SearchResponse:
        cleaned_query = query_text.strip()
        if not cleaned_query:
            return SearchResponse(query=query_text, total_results=0, results=[])

        # 1. Generar embedding vía puerto
        vector = self._embedding_provider.generate_embedding(cleaned_query)

        # 2. Consultar almacenamiento vectorial vía puerto
        chunks = self._vector_store.search_similar(query_vector=vector, top_k=top_k)

        return SearchResponse(
            query=cleaned_query,
            total_results=len(chunks),
            results=chunks
        )