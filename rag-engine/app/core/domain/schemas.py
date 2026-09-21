from typing import Dict, List, Optional
from pydantic import BaseModel, Field

class ChunkResult(BaseModel):
    page: int = Field(default=0)
    source: str = Field(default="Desconocido")
    text: str = Field(default="")
    score: float = Field(default=0.0)
    metadata: Dict[str, object] = Field(default_factory=dict)

class SearchQuery(BaseModel):
    query: str = Field(..., description="Pregunta o texto a buscar en el manual")
    top_k: int = Field(default=5, ge=1, le=20, description="Cantidad de fragmentos a recuperar")
    document_id: Optional[str] = Field(default=None)
    chapter: Optional[str] = Field(default=None)
    section: Optional[str] = Field(default=None)

class SearchResponse(BaseModel):
    query: str
    total_results: int
    results: List[ChunkResult]