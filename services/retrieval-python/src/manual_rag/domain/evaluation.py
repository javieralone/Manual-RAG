from typing import List, Optional
from pydantic import BaseModel, Field


class EvalCase(BaseModel):
    id: str
    query: str
    expected_sources: List[str] = Field(default_factory=list)
    expected_keywords: List[str] = Field(default_factory=list)
    top_k: int = Field(default=5, ge=1, le=20)


class EvalThresholds(BaseModel):
    min_context_recall: float = 0.7
    min_context_precision: float = 0.5
    min_groundedness: float = 0.5
    max_latency_seconds: float = 5.0


class EvalResult(BaseModel):
    case_id: str
    query: str
    context_precision: float
    context_recall: float
    groundedness: float
    latency_seconds: float
    retrieved_sources: List[str]
    answer: Optional[str] = None
    passed: bool


class EvalReport(BaseModel):
    results: List[EvalResult]
    total_cases: int
    passed_cases: int
    pass_rate: float
    avg_context_precision: float
    avg_context_recall: float
    avg_groundedness: float
    avg_latency_seconds: float
