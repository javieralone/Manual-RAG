from mcp.server.fastmcp import FastMCP
from sentence_transformers import SentenceTransformer
from qdrant_client import QdrantClient

# Crear el servidor MCP
mcp = FastMCP("Technical Manuals RAG Engine")

# Inicializar clientes (Qdrant debe estar corriendo)
client = QdrantClient(host="localhost", port=6333)
model = SentenceTransformer("BAAI/bge-m3")

@mcp.tool()
def search_manual(query: str, top_k: int = 3) -> str:
    """
    Busca fragmentos relevantes en los manuales técnicos guardados en Qdrant.
    Usa esta herramienta cuando necesites consultar especificaciones técnicas,
    mantenimiento o lubricación.
    """
    query_vector = model.encode(query).tolist()

    response = client.query_points(
        collection_name="manuales_tecnicos",
        query=query_vector,
        limit=top_k
    )

    if not response.points:
        return "No se encontró información relevante en los manuales."

    results = []
    for point in response.points:
        page = point.payload.get("page", "?")
        text = point.payload.get("text", "")
        source = point.payload.get("source", "Manual")
        results.append(f"--- Documento: {source} (Página {page}) ---\n{text}")

    return "\n\n".join(results)

if __name__ == "__main__":
    mcp.run()
