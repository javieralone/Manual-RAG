from fastapi import FastAPI, Depends, HTTPException, status
from app.core.domain.schemas import SearchQuery, SearchResponse
from app.core.ports.embedding_port import EmbeddingPort
from app.core.ports.vector_store_port import VectorStorePort
from app.adapters.sentence_transformer_adapter import SentenceTransformerAdapter
from app.adapters.qdrant_adapter import QdrantAdapter
from app.services.rag_service import RAGService

app = FastAPI(title="RAG Engine Python", version="1.0.0")

# --- CONTENEDORES DE DEPENDENCIAS (Singletons para evitar recargar el modelo en RAM) ---
_embedding_adapter = SentenceTransformerAdapter(model_name="paraphrase-multilingual-MiniLM-L12-v2")
_vector_store_adapter = QdrantAdapter(
    host="qdrant", 
    port=6333, 
    collection_name="manuales_tecnicos"
)

def get_rag_service() -> RAGService:
    return RAGService(
        embedding_provider=_embedding_adapter,
        vector_store=_vector_store_adapter
    )

# --- ENDPOINTS ---
@app.post("/search", response_model=SearchResponse, status_code=status.HTTP_200_OK)
def search(
    payload: SearchQuery, 
    rag_service: RAGService = Depends(get_rag_service)
):
    try:
        results = rag_service.retrieve_relevant_chunks(
            query_text=payload.query, 
            top_k=payload.top_k
        )
        return SearchResponse(results=results)
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Error en la búsqueda vectorial: {str(e)}"
        )

@app.get("/health")
def health():
    return {"status": "UP"}