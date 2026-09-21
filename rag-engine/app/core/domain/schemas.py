from pydantic import BaseModel, Field
from typing import List, Dict, Any, Optional

class DocumentChunk(BaseModel):
    text: str
    score: float = 0.0
    metadata: Dict[str, Any] = Field(default_factory=dict)

class SearchQuery(BaseModel):
    query: str = Field(..., min_length=1, description="Texto de la consulta")
    top_k: int = Field(default=3, ge=1, le=10)

class SearchResponse(BaseModel):
    results: List[DocumentChunk]