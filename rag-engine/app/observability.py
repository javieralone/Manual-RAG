import json
import logging
import time
from contextvars import ContextVar

from prometheus_client import Counter, Gauge, Histogram

trace_id_context: ContextVar[str] = ContextVar("trace_id", default="")

http_requests_total = Counter(
    "rag_engine_http_requests_total",
    "Total HTTP requests handled by the RAG engine.",
    ["method", "path", "status"],
)
http_request_duration = Histogram(
    "rag_engine_http_request_duration_seconds",
    "HTTP request duration in seconds.",
    ["method", "path"],
)
retrieval_duration = Histogram(
    "rag_engine_retrieval_duration_seconds",
    "RAG retrieval duration in seconds.",
)
embedding_duration = Histogram(
    "rag_engine_embedding_duration_seconds",
    "Embedding generation duration in seconds.",
)
qdrant_duration = Histogram(
    "rag_engine_qdrant_duration_seconds",
    "Qdrant query duration in seconds.",
)
chunks_retrieved = Histogram(
    "rag_engine_chunks_retrieved",
    "Number of chunks returned per retrieval.",
    buckets=(0, 1, 2, 3, 5, 10, 20),
)
documents_retrieved = Histogram(
    "rag_engine_documents_retrieved",
    "Number of distinct source documents returned per retrieval.",
    buckets=(0, 1, 2, 3, 5, 10),
)
errors_total = Counter(
    "rag_engine_errors_total",
    "RAG engine errors.",
    ["component"],
)
readiness = Gauge("rag_engine_readiness", "RAG engine readiness: 1 ready, 0 not ready.")


class JsonFormatter(logging.Formatter):
    def format(self, record: logging.LogRecord) -> str:
        payload = {
            "timestamp": self.formatTime(record, "%Y-%m-%dT%H:%M:%S%z"),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
        }
        trace_id = trace_id_context.get()
        if trace_id:
            payload["trace_id"] = trace_id
        return json.dumps(payload, ensure_ascii=False)


def configure_logging() -> None:
    handler = logging.StreamHandler()
    handler.setFormatter(JsonFormatter())
    root = logging.getLogger()
    root.handlers.clear()
    root.addHandler(handler)
    root.setLevel(logging.INFO)


def observe_retrieval(started: float, count: int, documents: int) -> None:
    retrieval_duration.observe(time.perf_counter() - started)
    chunks_retrieved.observe(count)
    documents_retrieved.observe(documents)