import time
from typing import Dict, Optional
from manual_rag.ports.embedding_port import EmbeddingPort
from manual_rag.ports.vector_store_port import VectorStorePort
from manual_rag.domain.schemas import ChunkResult, SearchResponse
from manual_rag.observability import observe_retrieval

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
        search_arguments = {"query_vector": vector, "top_k": top_k}
        if filters:
            search_arguments["filters"] = filters
        if collection_name:
            search_arguments["collection_name"] = collection_name
        chunks = self._vector_store.search_similar(**search_arguments)
        observe_retrieval(started, len(chunks), len({chunk.source for chunk in chunks}))

        return SearchResponse(
            query=cleaned_query,
            total_results=len(chunks),
            results=chunks
        )