import re
from typing import Dict, List, Optional
from pydantic import BaseModel, Field

DEFAULT_COLLECTION = "generic_manuals"

class ChunkResult(BaseModel):
    page: int = Field(default=0)
    source: str = Field(default="Desconocido")
    text: str = Field(default="")
    score: float = Field(default=0.0)
    metadata: Dict[str, object] = Field(default_factory=dict)

class SearchQuery(BaseModel):
    query: str = Field(..., description="Pregunta o texto a buscar en el manual")
    top_k: int = Field(default=5, ge=1, le=20, description="Cantidad de fragmentos a recuperar")
    collection: Optional[str] = Field(default=None, description="Colección Qdrant a consultar")
    document_id: Optional[str] = Field(default=None)
    chapter: Optional[str] = Field(default=None)
    section: Optional[str] = Field(default=None)

    @property
    def resolved_collection(self) -> str:
        value = (self.collection or DEFAULT_COLLECTION).strip()
        if not value or value == "default_collection":
            return DEFAULT_COLLECTION
        if not re.fullmatch(r"[A-Za-z0-9_-]+", value):
            raise ValueError("La colección debe contener solo letras, números, guion y guion bajo")
        return value

class SearchResponse(BaseModel):
    query: str
    total_results: int
    results: List[ChunkResult]