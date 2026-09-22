import json
import os
from typing import Dict, Optional

from mcp.server.fastmcp import FastMCP

from app.adapters.bge_embedding_adapter import BGEEmbeddingAdapter
from app.adapters.qdrant_adapter import QdrantAdapter
from app.core.domain.schemas import DEFAULT_COLLECTION, ChunkResult, resolve_collection
from app.services.rag_service import RAGService

QDRANT_HOST = os.getenv("QDRANT_HOST", "localhost")
QDRANT_PORT = int(os.getenv("QDRANT_PORT", 6333))
MCP_PORT = int(os.getenv("MCP_PORT", "8001"))


def _load_domain_collections() -> Dict[str, str]:
    raw_mapping = os.getenv("MCP_DOMAIN_COLLECTIONS", "")
    if not raw_mapping:
        return {"generic_manuals": "generic_manuals"}
    try:
        mapping = json.loads(raw_mapping)
    except json.JSONDecodeError as exc:
        raise ValueError("MCP_DOMAIN_COLLECTIONS debe ser un objeto JSON válido") from exc
    if not isinstance(mapping, dict) or not mapping:
        raise ValueError("MCP_DOMAIN_COLLECTIONS debe ser un objeto JSON no vacío")
    return {str(domain): resolve_collection(str(collection)) for domain, collection in mapping.items()}


embedding_adapter = BGEEmbeddingAdapter(model_name="BAAI/bge-m3")
_service_cache: Dict[str, RAGService] = {}
DOMAIN_COLLECTIONS = _load_domain_collections()


def get_rag_service(collection: Optional[str] = None) -> RAGService:
    collection_name = resolve_collection(collection)
    if collection_name not in _service_cache:
        _service_cache[collection_name] = RAGService(
            embedding_provider=embedding_adapter,
            vector_store=QdrantAdapter(
                host=QDRANT_HOST,
                port=QDRANT_PORT,
                collection_name=collection_name,
            ),
        )
    return _service_cache[collection_name]


def _search(query: str, top_k: int, collection: str, document_id: Optional[str] = None,
            chapter: Optional[str] = None, section: Optional[str] = None) -> str:
    if not query.strip():
        raise ValueError("La consulta no puede estar vacía")
    if not 1 <= top_k <= 20:
        raise ValueError("top_k debe estar entre 1 y 20")
    filters = {
        key: value for key, value in {
            "document_id": document_id, "chapter": chapter, "section": section,
        }.items() if value and value.strip()
    }
    response = get_rag_service(collection).execute_search(
        query_text=query,
        top_k=top_k,
        filters=filters,
        collection_name=resolve_collection(collection),
    )
    if not response.results:
        return "No se encontró información relevante en la colección seleccionada."
    return "\n\n".join(_format_result(chunk, collection) for chunk in response.results)


def _format_result(chunk: ChunkResult, collection: str) -> str:
    metadata = chunk.metadata
    details = [
        f"Colección: {metadata.get('collection', collection)}",
        f"Documento: {metadata.get('document_id', 'Desconocido')}",
        f"Fuente: {chunk.source}",
        f"Página: {chunk.page}",
    ]
    if "part" in metadata:
        details.append(f"Parte: {metadata['part']}")
    return f"--- {' | '.join(details)} | Score: {chunk.score:.4f} ---\n{chunk.text}"


mcp = FastMCP("Technical Manuals RAG Engine", host="0.0.0.0", port=MCP_PORT)


@mcp.tool()
def search_manual(query: str, top_k: int = 3, collection: str = DEFAULT_COLLECTION,
                 document_id: Optional[str] = None, chapter: Optional[str] = None,
                 section: Optional[str] = None) -> str:
    """Busca en la colección indicada; usa generic_manuals si no se especifica."""
    return _search(query, top_k, resolve_collection(collection), document_id, chapter, section)


@mcp.tool()
def search_technical_manuals(query: str, top_k: int = 3, document_id: Optional[str] = None,
                             chapter: Optional[str] = None, section: Optional[str] = None) -> str:
    """Busca exclusivamente en el dominio de manuales técnicos."""
    return _search(query, top_k, DOMAIN_COLLECTIONS["technical_manuals"], document_id, chapter, section)



if __name__ == "__main__":
    mcp.run(transport="streamable-http")
import json
import os
from typing import Dict, Optional

from mcp.server.fastmcp import FastMCP

from app.adapters.bge_embedding_adapter import BGEEmbeddingAdapter
from app.adapters.qdrant_adapter import QdrantAdapter
from app.core.domain.schemas import DEFAULT_COLLECTION, ChunkResult, resolve_collection
from app.services.rag_service import RAGService

QDRANT_HOST = os.getenv("QDRANT_HOST", "localhost")
QDRANT_PORT = int(os.getenv("QDRANT_PORT", 6333))
MCP_PORT = int(os.getenv("MCP_PORT", "8001"))


def _load_domain_collections() -> Dict[str, str]:
    raw_mapping = os.getenv("MCP_DOMAIN_COLLECTIONS", "")
    if not raw_mapping:
        return {
            "generic_manuals": "generic_manuals",
        }
    try:
        mapping = json.loads(raw_mapping)
    except json.JSONDecodeError as exc:
        raise ValueError("MCP_DOMAIN_COLLECTIONS debe ser un objeto JSON válido") from exc
    if not isinstance(mapping, dict) or not mapping:
        raise ValueError("MCP_DOMAIN_COLLECTIONS debe ser un objeto JSON no vacío")
    return {str(domain): resolve_collection(str(collection)) for domain, collection in mapping.items()}


embedding_adapter = BGEEmbeddingAdapter(model_name="BAAI/bge-m3")
_service_cache: Dict[str, RAGService] = {}
DOMAIN_COLLECTIONS = _load_domain_collections()


def get_rag_service(collection: Optional[str] = None) -> RAGService:
    collection_name = resolve_collection(collection)
    if collection_name not in _service_cache:
        vector_store = QdrantAdapter(
            host=QDRANT_HOST,
            port=QDRANT_PORT,
            collection_name=collection_name,
        )
        _service_cache[collection_name] = RAGService(
            embedding_provider=embedding_adapter,
            vector_store=vector_store,
        )
    return _service_cache[collection_name]


def _search(
    query: str,
    top_k: int,
    collection: str,
    document_id: Optional[str] = None,
    chapter: Optional[str] = None,
    section: Optional[str] = None,
) -> str:
    if not query.strip():
        raise ValueError("La consulta no puede estar vacía")
    if not 1 <= top_k <= 20:
        raise ValueError("top_k debe estar entre 1 y 20")
    filters = {
        key: value
        for key, value in {
            "document_id": document_id,
            "chapter": chapter,
            "section": section,
        }.items()
        if value and value.strip()
    }
    response = get_rag_service(collection).execute_search(
        query_text=query,
        top_k=top_k,
        filters=filters,
        collection_name=resolve_collection(collection),
    )
    if not response.results:
        return "No se encontró información relevante en la colección seleccionada."
    return "\n\n".join(_format_result(chunk, collection) for chunk in response.results)


def _format_result(chunk: ChunkResult, collection: str) -> str:
    metadata = chunk.metadata
    details = [
        f"Colección: {metadata.get('collection', collection)}",
        f"Documento: {metadata.get('document_id', 'Desconocido')}",
        f"Fuente: {chunk.source}",
        f"Página: {chunk.page}",
    ]
    if "part" in metadata:
        details.append(f"Parte: {metadata['part']}")
    return f"--- {' | '.join(details)} | Score: {chunk.score:.4f} ---\n{chunk.text}"


mcp = FastMCP("Technical Manuals RAG Engine", host="0.0.0.0", port=MCP_PORT)


@mcp.tool()
def search_manual(
    query: str,
    top_k: int = 3,
    collection: str = DEFAULT_COLLECTION,
    document_id: Optional[str] = None,
    chapter: Optional[str] = None,
    section: Optional[str] = None,
) -> str:
    """Busca en la colección indicada; usa generic_manuals si no se especifica."""
    return _search(query, top_k, resolve_collection(collection), document_id, chapter, section)


@mcp.tool()
def search_technical_manuals(
    query: str,
    top_k: int = 3,
    document_id: Optional[str] = None,
    chapter: Optional[str] = None,
    section: Optional[str] = None,
) -> str:
    """Busca exclusivamente en el dominio de manuales técnicos."""
    return _search(query, top_k, DOMAIN_COLLECTIONS["technical_manuals"], document_id, chapter, section)


if __name__ == "__main__":
    mcp.run(transport="streamable-http")
