from typing import List, Optional
from pydantic import BaseModel, Field

class ChunkResult(BaseModel):
    page: int = Field(default=0)
    source: str = Field(default="Desconocido")
    text: str = Field(default="")
    score: float = Field(default=0.0)

class SearchQuery(BaseModel):
    query: str = Field(..., description="Pregunta o texto a buscar en el manual")
    top_k: int = Field(default=5, ge=1, le=20, description="Cantidad de fragmentos a recuperar")

class SearchResponse(BaseModel):
    query: str
    total_results: int
    results: List[ChunkResult]