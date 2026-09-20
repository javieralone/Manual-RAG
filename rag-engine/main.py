import os
import sys
from pathlib import Path
from typing import List, Optional
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from sentence_transformers import SentenceTransformer
from qdrant_client import QdrantClient

# 1. Definir rutas y constantes
BASE_DIR = Path(__file__).resolve().parent.parent
COLLECTION_NAME = "manuales_tecnicos"
QDRANT_HOST = os.getenv("QDRANT_HOST", "localhost")
QDRANT_PORT = int(os.getenv("QDRANT_PORT", 6333))

# 2. Inicializar FastAPI
app = FastAPI(
    title="RAG Engine Internal API",
    description="Servicio interno en Python para embeddings y búsqueda vectorial en Qdrant.",
    version="1.0.0"
)

# 3. Modelos de datos (Pydantic)
class SearchRequest(BaseModel):
    query: str = Field(..., description="Pregunta o texto a buscar en el manual")
    top_k: int = Field(default=5, ge=1, le=20, description="Cantidad de fragmentos a recuperar")

class ChunkResult(BaseModel):
    page: int
    source: str
    text: str
    score: float

class SearchResponse(BaseModel):
    query: str
    total_results: int
    results: List[ChunkResult]

# 4. Cargar clientes y modelos globales al iniciar
print(f"Conectando a Qdrant en {QDRANT_HOST}:{QDRANT_PORT}...")
try:
    qdrant_client = QdrantClient(host=QDRANT_HOST, port=QDRANT_PORT)
    qdrant_client.get_collections()
except Exception as e:
    print(f"[ERROR] No se pudo conectar a Qdrant: {e}")
    sys.exit(1)

print("Cargando modelo BGE-M3...")
embedding_model = SentenceTransformer("BAAI/bge-m3")

# 5. Endpoints
@app.get("/health")
def health_check():
    """Endpoint para verificar que el servicio está vivo."""
    return {"status": "ok", "engine": "RAG Python FastAPI"}

@app.post("/search", response_model=SearchResponse)
def search_chunks(request: SearchRequest):
    """
    Recibe una consulta, genera su embedding con BGE-M3 y
    busca los fragmentos más relevantes en Qdrant.
    """
    if not request.query.strip():
        raise HTTPException(status_code=400, detail="La consulta 'query' no puede estar vacía.")

    try:
        # Generar vector denso
        query_vector = embedding_model.encode(request.query).tolist()

        # Consultar en Qdrant
        response = qdrant_client.query_points(
            collection_name=COLLECTION_NAME,
            query=query_vector,
            limit=request.top_k
        )

        results = []
        for point in response.points:
            payload = point.payload or {}
            results.append(
                ChunkResult(
                    page=payload.get("page", 0),
                    source=payload.get("source", "Desconocido"),
                    text=payload.get("text", ""),
                    score=point.score if hasattr(point, "score") else 0.0
                )
            )

        return SearchResponse(
            query=request.query,
            total_results=len(results),
            results=results
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error en el motor RAG: {str(e)}")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)