import time
from typing import List
from app.observability import embedding_duration
from sentence_transformers import SentenceTransformer
from app.core.ports.embedding_port import EmbeddingPort
from opentelemetry import trace

class BGEEmbeddingAdapter(EmbeddingPort):
    def __init__(self, model_name: str = "BAAI/bge-m3"):
        import logging
        logging.getLogger(__name__).info("embedding_model_loading", extra={"model": model_name})
        self._model = SentenceTransformer(model_name)

    def generate_embedding(self, text: str) -> List[float]:
        started = time.perf_counter()
        tracer = trace.get_tracer("manual-rag/rag-engine")
        with tracer.start_as_current_span("embeddings.generate"):
            try:
                return self._model.encode(text).tolist()
            finally:
                embedding_duration.observe(time.perf_counter() - started)