import os
from fastapi import FastAPI, HTTPException, status, Depends
from app.core.domain.schemas import SearchQuery, SearchResponse
from app.adapters.bge_embedding_adapter import BGEEmbeddingAdapter
from app.adapters.qdrant_adapter import QdrantAdapter
from app.services.rag_service import RAGService

QDRANT_HOST = os.getenv("QDRANT_HOST", "localhost")
QDRANT_PORT = int(os.getenv("QDRANT_PORT", 6333))
COLLECTION_NAME = "manuales_tecnicos"

# --- COMPOSITION ROOT (Singletons) ---
embedding_adapter = BGEEmbeddingAdapter(model_name="BAAI/bge-m3")
qdrant_adapter = QdrantAdapter(host=QDRANT_HOST, port=QDRANT_PORT, collection_name=COLLECTION_NAME)
rag_service_instance = RAGService(embedding_provider=embedding_adapter, vector_store=qdrant_adapter)

def get_rag_service() -> RAGService:
    return rag_service_instance

app = FastAPI(
    title="RAG Engine Internal API",
    description="Servicio interno en Python para embeddings y búsqueda vectorial.",
    version="1.0.0"
)

@app.get("/health")
def health_check():
    return {"status": "ok", "engine": "RAG Python FastAPI Clean Arch"}

@app.post("/search", response_model=SearchResponse)
def search_chunks(
    request: SearchQuery, 
    rag_service: RAGService = Depends(get_rag_service)
):
    if not request.query.strip():
        raise HTTPException(status_code=400, detail="La consulta 'query' no puede estar vacía.")

    try:
        return rag_service.execute_search(query_text=request.query, top_k=request.top_k)
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Error en el motor RAG: {str(e)}"
        )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("app.main:app", host="0.0.0.0", port=8000, reload=True)