import os
import logging
import time
from fastapi import FastAPI, HTTPException, status, Depends, Request
from fastapi.responses import Response
from prometheus_client import CONTENT_TYPE_LATEST, generate_latest
from app.core.domain.schemas import SearchQuery, SearchResponse
from app.adapters.bge_embedding_adapter import BGEEmbeddingAdapter
from app.adapters.qdrant_adapter import QdrantAdapter
from app.services.rag_service import RAGService
from app.observability import configure_logging, errors_total, http_request_duration, http_requests_total, readiness, trace_id_context
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

QDRANT_HOST = os.getenv("QDRANT_HOST", "localhost")
QDRANT_PORT = int(os.getenv("QDRANT_PORT", 6333))
COLLECTION_NAME = "manuales_tecnicos"

# --- COMPOSITION ROOT (Singletons) ---
configure_logging()

otlp_endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
if otlp_endpoint:
    provider = TracerProvider(resource=Resource.create({"service.name": os.getenv("OTEL_SERVICE_NAME", "rag-engine")}))
    provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter(endpoint=otlp_endpoint, insecure=True)))
    trace.set_tracer_provider(provider)
embedding_adapter = BGEEmbeddingAdapter(model_name="BAAI/bge-m3")
qdrant_adapter = QdrantAdapter(host=QDRANT_HOST, port=QDRANT_PORT, collection_name=COLLECTION_NAME)
rag_service_instance = RAGService(embedding_provider=embedding_adapter, vector_store=qdrant_adapter)

def get_rag_service() -> RAGService:
    return rag_service_instance

app = FastAPI(
    title="RAG Engine Internal API",
    description="Servicio interno en Python para embeddings y búsqueda vectorial.",
    version="1.0.0"
)
FastAPIInstrumentor.instrument_app(app)

@app.middleware("http")
async def observability_middleware(request: Request, call_next):
    span_context = trace.get_current_span().get_span_context()
    trace_id = format(span_context.trace_id, "032x") if span_context.is_valid else ""
    trace_id_context.set(trace_id)
    started = time.perf_counter()
    response = await call_next(request)
    path = request.url.path
    http_requests_total.labels(request.method, path, str(response.status_code)).inc()
    http_request_duration.labels(request.method, path).observe(time.perf_counter() - started)
    return response

@app.get("/health")
def health_check():
    return {"status": "ok", "engine": "RAG Python FastAPI Clean Arch"}

@app.get("/ready")
def readiness_check():
    try:
        qdrant_adapter.check_ready()
        readiness.set(1)
        return {"status": "READY"}
    except Exception:
        readiness.set(0)
        return Response(content='{"status":"NOT_READY"}', status_code=503, media_type="application/json")

@app.get("/metrics")
def metrics():
    return Response(generate_latest(), media_type=CONTENT_TYPE_LATEST)

@app.post("/search", response_model=SearchResponse)
def search_chunks(
    request: SearchQuery, 
    rag_service: RAGService = Depends(get_rag_service)
):
    if not request.query.strip():
        raise HTTPException(status_code=400, detail="La consulta 'query' no puede estar vacía.")

    try:
        return rag_service.execute_search(query_text=request.query, top_k=request.top_k)
    except Exception as e:
        errors_total.labels(component="search").inc()
        logging.getLogger(__name__).exception("rag_search_failed")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Error en el motor RAG"
        )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("app.main:app", host="0.0.0.0", port=8000, reload=True)