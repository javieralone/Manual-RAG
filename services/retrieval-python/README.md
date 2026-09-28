# 🚀 RAG Engine - Technical Manuals Service

Servicio interno de RAG (Retrieval-Augmented Generation) desarrollado en Python con **FastAPI** y **FastMCP**.

Es responsable de:
- Generación de embeddings semánticos.
- Recuperación de información desde la base de datos vectorial Qdrant.
- Exposición de APIs HTTP mediante FastAPI.
- Integración con modelos y agentes mediante Model Context Protocol (MCP).

---

# 🏗️ Arquitectura y Diseño

El proyecto sigue los principios de **Clean Architecture** (Arquitectura Hexagonal) y **SOLID**.

```text
services/retrieval-python/
├── app/
│   ├── main.py                    # Wrapper compatible FastAPI
│   └── mcp_server.py              # Wrapper compatible MCP
├── src/manual_rag/
│   ├── adapters/                  # Adaptadores de infraestructura
│   ├── application/               # Casos de uso
│   ├── domain/                    # DTOs y modelos Pydantic
│   ├── ports/                     # Contratos / interfaces
│   ├── entrypoints/
│   │   ├── http.py                # API FastAPI
│   │   └── mcp.py                 # Servidor MCP
│   ├── bootstrap.py               # Composición lazy compartida
│   └── observability.py
│
├── Dockerfile
├── requirements.txt
└── README.md
```

## Principios Clave y Componentes

### Inversión de Dependencias (DIP)
El servicio de aplicación `RAGService` interactúa únicamente con abstracciones (`EmbeddingPort`, `VectorStorePort`), permitiendo cambiar los adaptadores concretos sin afectar la lógica de negocio.

### Reutilización del Dominio
Tanto FastAPI como FastMCP reutilizan la misma lógica de aplicación (`RAGService`), evitando duplicación de código, instancias múltiples del modelo de embeddings y consumo innecesario de memoria RAM.

### Componentes Core
- **Modelo de Embeddings:** `BAAI/bge-m3` implementado mediante `SentenceTransformers`.
- **Base de Datos Vectorial:** `Qdrant`. La colección por defecto es `generic_manuals`. También soporta colecciones específicas (ej. `manuales_tecnicos`) respetando el formato `^[A-Za-z0-9_-]+$`.

---

# 📑 Endpoints HTTP (FastAPI)

## 1. Health Check & Observabilidad
- `GET /health`: Verifica el estado general de la aplicación.
- `GET /ready`: Verifica que Qdrant está disponible para atender búsquedas (el modelo de embeddings se carga al iniciar el proceso).
- `GET /metrics`: Expone métricas Prometheus de HTTP, embeddings, retrieval y Qdrant.

El servicio escribe logs JSON en `stdout` y propaga el contexto W3C `traceparent`. Las trazas se exportan a Tempo cuando `OTEL_EXPORTER_OTLP_ENDPOINT` está configurado.

### Ejemplo Health Check
```http
GET /health
```
**Respuesta:**
```json
{
  "status": "ok",
  "engine": "RAG Python FastAPI Clean Arch"
}
```

## 2. Búsqueda Vectorial
```http
POST /search
```
**Body:**
```json
{
  "query": "¿Cómo se realiza el mantenimiento del sistema de lubricación?",
  "top_k": 3,
  "collection": "manuales_tecnicos",
  "document_id": "0-lubricacion-mantenimiento",
  "chapter": "2",
  "section": "2.1"
}
```
*Nota: `collection` es opcional (usa `generic_manuals` por defecto). Los filtros `document_id`, `chapter` y `section` son opcionales y se aplican conjuntamente sobre el payload de Qdrant.*

**Respuesta:**
```json
{
  "results": [
    {
      "page": 12,
      "source": "manual_tecnico.pdf",
      "text": "El mantenimiento del sistema de lubricación requiere...",
      "score": 0.895,
      "metadata": {
        "document_id": "0-lubricacion-mantenimiento",
        "chapter": "2",
        "section": "2.1"
      }
    }
  ]
}
```

---

# 🔌 Servidor MCP (Model Context Protocol)

Ubicación del punto de entrada: `src/manual_rag/entrypoints/mcp.py` (el wrapper `app/mcp_server.py` delega en él).

Se expone en `http://localhost:8001/mcp` al ejecutar con Docker Compose. Mantiene una caché de `RAGService` por colección y comparte el proveedor de embeddings.

### Herramientas Expuestas:
- `search_manual(query: str, top_k: int = 3, collection: str = "generic_manuals") -> str`: Mantiene compatibilidad, usa `generic_manuals` por defecto y permite seleccionar colección explícita.
- `search_technical_manuals(query: str, top_k: int = 3) -> str`: Consulta la colección mapeada al dominio `technical_manuals`.

*Ambas herramientas aceptan los filtros opcionales `document_id`, `chapter` y `section`, devolviendo colección, documento, fuente, página y parte.*

### Mapeo de Dominios (`MCP_DOMAIN_COLLECTIONS`):
```json
{"technical_manuals": "manuales_tecnicos"}
```
*La clave `technical_manuals` debe estar presente en el mapa.*

**Ejecución local:**
```bash
PYTHONPATH=src python -m manual_rag.entrypoints.mcp
```

---

# ⚙️ Ingesta Asíncrona de Documentos

El Compose incluye los servicios `ingestion-api` (publicada en `http://localhost:8002`) e `ingestion-worker`.

### Endpoints de Ingesta:
- `POST /ingestion/enqueue` (Alias: `POST /ingestion/jobs`): Crear un trabajo de ingesta.
- `GET /ingestion/jobs` y `GET /ingestion/jobs/{job_id}`: Listar y consultar estado de trabajos.
- `GET /ingestion/failed`: Consultar trabajos fallidos o agotados.

### Worker:
Iniciado mediante `python -m manual_rag.entrypoints.worker`, procesa la cola de Redis configurada (`INGESTION_QUEUE`). Orquesta el pipeline en un workspace aislado por trabajo.

---

# 🛠️ OCR, Procesamiento e Indexación Manual

Para procesamiento local o pruebas directas, se utiliza el orquestador `services/retrieval-python/scripts/process_manual_opt.py`.

### Estructura de Directorios
Los documentos en PDF deben ubicarse en `data/documents/new/<nombre_coleccion>/`:
```text
data/documents/new/manuales_tecnicos/<nombre-manual>__parte-001.pdf
data/documents/reading/manuales_tecnicos/<nombre-manual>__parte-001.pdf
data/documents/completed/manuales_tecnicos/<nombre-manual>__parte-001.pdf
```

### Ejecución del Orquestador
```powershell
python services/retrieval-python/scripts/process_manual_opt.py `
  --fast-ocr `
  --workers 2 `
  --memory-mode disk
```
*El script mueve el PDF a `reading`, ejecuta OCR, genera chunks, los carga en Qdrant y finalmente lo mueve a `completed`. Si falla, lo devuelve a `new` y registra el error en `../logs/ingestion.log`.*

### Carga por Partes y Documentos Escaneados
- **Partes múltiples:** Asignar un identificador común (ej. `__parte-001.pdf`, `__parte-002.pdf`). Los IDs generados en Qdrant son deterministas, por lo que reintentar una parte no duplica puntos.
- **OCR Manual para Escaneos:**
  ```bash
  python scripts/ocr_manual_opt.py \
    --pdf data/documents/new/manuales_tecnicos/<nombre-manual>__parte-001.pdf \
    --output data/artifacts/manual_pages.json \
    --collection manuales_tecnicos \
    --dpi 200 \
    --language spa
  ```
  Tras regenerar el JSON, ejecutar `index_manual.py` y `upload_to_qdrant.py --collection <nombre_coleccion>`.

*Nota: Para aplicar o actualizar filtros sobre datos existentes, se debe re-ejecutar `index_manual.py` y `upload_to_qdrant.py`.*

---

# ✅ Evaluación de Calidad del RAG

El script `scripts/evaluate_rag.py` valida la calidad de recuperación y respuesta utilizando un dataset de referencia (`eval/golden_dataset.json`) y thresholds definidos en `eval/thresholds.json`.

### Métricas Mencionadas:
- `context_precision`: Fracción de fragmentos recuperados pertenecientes a fuentes esperadas.
- `context_recall`: Fracción de fuentes esperadas encontradas en la recuperación.
- `groundedness`: Cobertura de palabras clave esperadas en el contexto o respuesta.
- `latency_seconds`: Tiempo de respuesta por consulta.

### Comandos de Evaluación
```bash
cd rag-engine
python scripts/evaluate_rag.py --skip-generation   # Solo prueba recuperación (sin Ollama)
python scripts/evaluate_rag.py                     # Incluye generación de respuestas con Ollama
```
*Genera un reporte en `../data/artifacts/evaluations/eval_report_<timestamp>.json`.*

### Integración CI/CD
El workflow `.github/workflows/rag-evaluation.yml` ejecuta en Pull Requests un fixture determinista (`eval/ci_fixture.json`) contra los thresholds estrictos de `eval/ci_thresholds.json`, bloqueando la integración en caso de no cumplirlos.

---

# ⚙️ Configuración y Variables de Entorno

## Requisitos Previos
- Docker y Docker Compose
- Qdrant
- Python 3.10+

## Tabla de Variables de Entorno

| Categoría | Variable | Descripción | Valor por defecto |
|-----------|----------|-------------|-------------------|
| **Qdrant / General** | `QDRANT_HOST` | Host de la base de datos Qdrant | `qdrant` |
| | `QDRANT_PORT` | Puerto HTTP de Qdrant | `6333` |
| | `DEFAULT_COLLECTION` | Colección por defecto | `generic_manuals` |
| | `OMP_NUM_THREADS` | Hilos para CPU | `2` |
| | `MKL_NUM_THREADS` | Hilos MKL | `2` |
| **MCP** | `MCP_PORT` | Puerto del servidor MCP | `8001` |
| | `MCP_DOMAIN_COLLECTIONS` | Mapa JSON de dominio MCP a colección | `{"technical_manuals":"manuales_tecnicos"}` |
| **Ingesta Asíncrona** | `INGESTION_REDIS_URL` | URL de conexión a Redis | `redis://localhost:6379/2` |
| | `INGESTION_REDIS_PREFIX` | Prefijo de claves Redis | - |
| | `INGESTION_QUEUE` | Nombre de la cola de ingesta | - |
| | `INGESTION_JOB_TIMEOUT_SECONDS` | Tiempo límite por trabajo | - |
| | `INGESTION_PIPELINE_TIMEOUT_SECONDS` | Tiempo límite del pipeline | - |
| | `INGESTION_MAX_RETRIES` | Reintentos máximos | - |
| **Almacenamiento (MinIO)** | `MINIO_ENDPOINT` | Endpoint de MinIO | - |
| | `MINIO_ROOT_USER` | Usuario Root | - |
| | `MINIO_ROOT_PASSWORD` | Contraseña Root | - |
| | `MINIO_BUCKET` | Bucket de almacenamiento | - |

---

# 🚀 Despliegue y Flujo del Sistema

## Despliegue con Docker Compose
```powershell
# Construir y levantar servicios principales
docker compose --env-file .env.local up -d --build rag-engine mcp-server

# Ver logs
docker logs -f rag_engine
```

## Servicios Docker Expuestos
- **Qdrant (`http://localhost:6333`):** Base de datos vectorial.
- **RAG Engine (`http://localhost:8000`):** Servicio FastAPI principal.
- **MCP Server (`http://localhost:8001`):** Servidor MCP.
- **Ingestion API (`http://localhost:8002`):** API de encolamiento de ingestas.
- **API Gateway (Go) (`http://localhost:8080`):** Coordinador frontal de peticiones.

## 📊 Diagrama del Flujo de Consulta

```text
Usuario
   │
   ▼
API Gateway (Go)
   │
   ▼
RAG Engine (FastAPI)
   │
   ├── Genera embedding
   │
   ▼
Qdrant (Base de Datos Vectorial)
   │
   ▼
Chunks relevantes recuperados
   │
   ▼
LLM (Generación de Respuesta)
   │
   ▼
Respuesta final al usuario
```

## 🧪 Pruebas Rápidas (cURL)
```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mantenimiento de lubricacion",
    "top_k": 3
  }'
```
