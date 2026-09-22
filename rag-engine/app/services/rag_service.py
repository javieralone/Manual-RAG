import time
from typing import Dict, Optional
from app.core.ports.embedding_port import EmbeddingPort
from app.core.ports.vector_store_port import VectorStorePort
from app.core.domain.schemas import ChunkResult, SearchResponse
from app.observability import observe_retrieval

class RAGService:
    def __init__(self, embedding_provider: EmbeddingPort, vector_store: VectorStorePort):
        self._embedding_provider = embedding_provider
        self._vector_store = vector_store

    def execute_search(
        self,
        query_text: str,
        top_k: int,
        filters: Optional[Dict[str, str]] = None,
        collection_name: Optional[str] = None,
    ) -> SearchResponse:
        cleaned_query = query_text.strip()
        if not cleaned_query:
            return SearchResponse(query=query_text, total_results=0, results=[])

        started = time.perf_counter()
        vector = self._embedding_provider.generate_embedding(cleaned_query)
        if filters:
            chunks = self._vector_store.search_similar(
                query_vector=vector,
                top_k=top_k,
                filters=filters,
                collection_name=collection_name,
            )
        else:
            chunks = self._vector_store.search_similar(
                query_vector=vector,
                top_k=top_k,
                collection_name=collection_name,
            )
        observe_retrieval(started, len(chunks), len({chunk.source for chunk in chunks}))

        return SearchResponse(
            query=cleaned_query,
            total_results=len(chunks),
            results=chunks
        )