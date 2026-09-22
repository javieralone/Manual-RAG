import os
from mcp.server.fastmcp import FastMCP
from app.adapters.bge_embedding_adapter import BGEEmbeddingAdapter
from app.adapters.qdrant_adapter import QdrantAdapter
from app.services.rag_service import RAGService

QDRANT_HOST = os.getenv("QDRANT_HOST", "localhost")
QDRANT_PORT = int(os.getenv("QDRANT_PORT", 6333))
DEFAULT_COLLECTION = os.getenv("DEFAULT_COLLECTION", "generic_manuals")

# --- COMPOSITION ROOT ---
embedding_adapter = BGEEmbeddingAdapter(model_name="BAAI/bge-m3")
qdrant_adapter = QdrantAdapter(host=QDRANT_HOST, port=QDRANT_PORT, collection_name=DEFAULT_COLLECTION)
rag_service = RAGService(embedding_provider=embedding_adapter, vector_store=qdrant_adapter)

mcp = FastMCP(
    "Technical Manuals RAG Engine",
    host="0.0.0.0",
    port=int(os.getenv("MCP_PORT", "8001")),
)

@mcp.tool()
def search_manual(query: str, top_k: int = 3, collection: str = DEFAULT_COLLECTION) -> str:
    """
    Busca fragmentos relevantes en los manuales técnicos guardados en Qdrant.
    Usa esta herramienta cuando necesites consultar especificaciones técnicas,
    mantenimiento o lubricación.
    """
    collection_name = collection.strip() if collection else DEFAULT_COLLECTION
    if collection_name == "default_collection":
        collection_name = DEFAULT_COLLECTION
    search_response = rag_service.execute_search(query_text=query, top_k=top_k, collection_name=collection_name)

    if not search_response.results:
        return "No se encontró información relevante en los manuales."

    results = []
    for chunk in search_response.results:
        results.append(f"--- Documento: {chunk.source} (Página {chunk.page}) ---\n{chunk.text}")

    return "\n\n".join(results)

if __name__ == "__main__":
    mcp.run(transport="streamable-http")