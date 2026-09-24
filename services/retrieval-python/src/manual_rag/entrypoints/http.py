import logging
import os
import re
import time
from contextlib import asynccontextmanager
from pathlib import Path

from fastapi import FastAPI, File, Form, HTTPException, Request, UploadFile, status
from fastapi.responses import Response
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from prometheus_client import CONTENT_TYPE_LATEST, generate_latest

from pydantic import BaseModel, Field

from manual_rag.bootstrap import (
    DEFAULT_QDRANT_COLLECTION,
    get_ingestion_job_service,
    get_qdrant_adapter,
    get_rag_service,
)
from manual_rag.domain.ingestion import IngestionRequest
from manual_rag.domain.schemas import SearchQuery, SearchResponse
from manual_rag.observability import (
    configure_logging,
    errors_total,
    http_request_duration,
    http_requests_total,
    readiness,
    trace_id_context,
)

configure_logging()
otlp_endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
if otlp_endpoint:
    provider = TracerProvider(resource=Resource.create({"service.name": os.getenv("OTEL_SERVICE_NAME", "rag-engine")}))
    provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter(endpoint=otlp_endpoint, insecure=True)))
    trace.set_tracer_provider(provider)


@asynccontextmanager
async def lifespan(_app: FastAPI):
    yield
    if otlp_endpoint:
        current_provider = trace.get_tracer_provider()
        shutdown = getattr(current_provider, "shutdown", None)
        if shutdown:
            shutdown()

app = FastAPI(
    title="RAG Engine Internal API",
    description="Servicio interno en Python para embeddings y búsqueda vectorial.",
    version="1.0.0",
    lifespan=lifespan,
)
FastAPIInstrumentor.instrument_app(app)


class IngestionSubmission(BaseModel):
    local_path: str | None = Field(default=None, description="PDF local dentro de INGESTION_LOCAL_ROOT")
    collection: str | None = None
    bucket: str = ""
    object_key: str = ""


PDF_NAME_PATTERN = re.compile(r"^.+__parte-\d{3}\.pdf$", re.IGNORECASE)


def validate_pdf_name(filename: str) -> None:
    if not PDF_NAME_PATTERN.fullmatch(filename):
        raise ValueError("El PDF debe llamarse <nombre_manual>__parte-xxx.pdf")


@app.post("/ingestion/enqueue", status_code=202)
@app.post("/ingestion/jobs", status_code=202, include_in_schema=False)
def enqueue_ingestion(request: IngestionSubmission):
    try:
        job = get_ingestion_job_service().submit(IngestionRequest(
            pdf_path=Path(request.local_path) if request.local_path else None,
            collection=request.collection,
            bucket=request.bucket,
            object_key=request.object_key,
        ))
        return job.to_dict()
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except Exception:
        logging.getLogger(__name__).exception("ingestion_enqueue_failed")
        raise HTTPException(status_code=503, detail="No se pudo encolar la ingesta")


@app.get("/ingestion/storage/options")
def ingestion_storage_options():
    try:
        storage = get_ingestion_job_service().storage
        buckets = storage.list_buckets()
        return {"buckets": [{"name": bucket, "object_keys": storage.list_object_keys(bucket)} for bucket in buckets]}
    except Exception:
        logging.getLogger(__name__).exception("ingestion_storage_options_failed")
        raise HTTPException(status_code=503, detail="No se pudo consultar el almacenamiento")


@app.post("/ingestion/upload", status_code=202)
async def upload_ingestion(
    file: UploadFile = File(...),
    bucket: str = Form(...),
    object_key: str = Form(...),
):
    filename = (file.filename or "").strip()
    if not filename.lower().endswith(".pdf"):
        raise HTTPException(status_code=400, detail="Solo se permiten archivos PDF")
    try:
        validate_pdf_name(filename)
        if Path(object_key).name != filename:
            raise ValueError("El object_key debe terminar con el nombre original del PDF")
        storage = get_ingestion_job_service().storage
        content_type = file.content_type or "application/pdf"
        length = 0
        contents = await file.read()
        length = len(contents)
        from io import BytesIO
        storage.upload(bucket, object_key, BytesIO(contents), length, content_type)
        return enqueue_ingestion(IngestionSubmission(bucket=bucket, object_key=object_key))
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except Exception:
        logging.getLogger(__name__).exception("ingestion_upload_failed")
        raise HTTPException(status_code=503, detail="No se pudo subir el documento")


@app.get("/ingestion/jobs/{job_id}")
def ingestion_status(job_id: str):
    try:
        job = get_ingestion_job_service().store.get(job_id)
    except Exception:
        logging.getLogger(__name__).exception("ingestion_status_failed")
        raise HTTPException(status_code=503, detail="No se pudo consultar la ingesta")
    if job is None:
        raise HTTPException(status_code=404, detail="Trabajo de ingesta no encontrado")
    return job.to_dict()


@app.get("/ingestion/jobs")
def ingestion_jobs():
    try:
        return [job.to_dict() for job in get_ingestion_job_service().store.list_jobs()]
    except Exception:
        logging.getLogger(__name__).exception("ingestion_list_failed")
        raise HTTPException(status_code=503, detail="No se pudieron consultar las ingestas")


@app.get("/ingestion/failed")
def failed_ingestion_jobs():
    try:
        failed_statuses = {"FAILED", "FAILED_PERMANENTLY", "TIMEOUT"}
        return [
            job.to_dict()
            for job in get_ingestion_job_service().store.list_jobs()
            if job.status.value in failed_statuses
        ]
    except Exception:
        logging.getLogger(__name__).exception("ingestion_failed_list_failed")
        raise HTTPException(status_code=503, detail="No se pudieron consultar las ingestas fallidas")


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
        get_qdrant_adapter(DEFAULT_QDRANT_COLLECTION).check_ready()
        readiness.set(1)
        return {"status": "READY"}
    except Exception:
        readiness.set(0)
        return Response(content='{"status":"NOT_READY"}', status_code=503, media_type="application/json")


@app.get("/metrics")
def metrics():
    return Response(generate_latest(), media_type=CONTENT_TYPE_LATEST)


@app.post("/search", response_model=SearchResponse)
def search_chunks(request: SearchQuery):
    if not request.query.strip():
        raise HTTPException(status_code=400, detail="La consulta 'query' no puede estar vacía.")

    try:
        collection_name = request.resolved_collection
        filters = {
            key: value
            for key, value in {
                "document_id": request.document_id,
                "chapter": request.chapter,
                "section": request.section,
            }.items()
            if value
        }
        return get_rag_service(collection_name).execute_search(
            query_text=request.query,
            top_k=request.top_k,
            filters=filters,
            collection_name=collection_name,
        )
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except Exception:
        errors_total.labels(component="search").inc()
        logging.getLogger(__name__).exception("rag_search_failed")
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail="Error en el motor RAG")