# 🚀 RAG Engine - Technical Manuals Service

Servicio interno de RAG (Retrieval-Augmented Generation) desarrollado en Python con **FastAPI** y **FastMCP**.

Es responsable de:

- Generación de embeddings semánticos.
- Recuperación de información desde Qdrant.
- Exposición de APIs HTTP mediante FastAPI.
- Integración mediante Model Context Protocol (MCP).

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

## Principios Clave

### Inversión de Dependencias (DIP)

El servicio de aplicación `RAGService` interactúa únicamente con abstracciones:

- `EmbeddingPort`
- `VectorStorePort`

Los adaptadores concretos pueden cambiar sin afectar a la lógica de negocio.

### Reutilización del Dominio

Tanto FastAPI como FastMCP reutilizan la misma lógica de aplicación (`RAGService`), evitando:

- Duplicación de código.
- Instancias múltiples del modelo de embeddings.
- Consumo innecesario de memoria RAM.

### Modelo de Embeddings

```text
BAAI/bge-m3
```

implementado mediante:

```text
SentenceTransformers
```

### Base de Datos Vectorial

```text
Qdrant
```

La colección por defecto es:

```text
generic_manuals
```

Y también puede usarse explícitamente `manuales_tecnicos` u otra colección válida con el mismo formato `^[A-Za-z0-9_-]+$`.

---

# 🛠️ Requisitos Previos

- Docker
- Docker Compose
- Qdrant
- Python 3.10 o superior

---

# ⚙️ Variables de Entorno

| Variable | Descripción | Valor por defecto |
|-----------|------------|------------------|
| `QDRANT_HOST` | Host de Qdrant | `qdrant` |
| `QDRANT_PORT` | Puerto HTTP de Qdrant | `6333` |
| `DEFAULT_COLLECTION` | Colección usada cuando no se indica otra | `generic_manuals` |
| `MCP_DOMAIN_COLLECTIONS` | Mapa JSON de dominio MCP a colección Qdrant | `{"technical_manuals":"manuales_tecnicos"}` |
| `MCP_PORT` | Puerto del servidor MCP | `8001` |
| `OMP_NUM_THREADS` | Hilos para CPU | `2` |
| `MKL_NUM_THREADS` | Hilos MKL | `2` |

---

# 🚀 Despliegue con Docker Compose

Construir y levantar el servicio:

```powershell
docker compose up -d --build rag-engine mcp-server
```

Consultar logs:

```powershell
docker logs -f rag_engine
```

---

# 📑 Endpoints HTTP (FastAPI)

## Health Check

### Request

```http
GET /health
```

### Response

```json
{
  "status": "ok",
  "engine": "RAG Python FastAPI Clean Arch"
}
```

## Readiness y métricas

- `GET /ready`: verifica que Qdrant está disponible para atender búsquedas. El modelo de embeddings se carga al iniciar el proceso.
- `GET /metrics`: expone métricas Prometheus de HTTP, embeddings, retrieval y Qdrant.

El servicio escribe logs JSON en stdout y propaga el contexto W3C `traceparent`. Las trazas se exportan a Tempo cuando `OTEL_EXPORTER_OTLP_ENDPOINT` está configurado.

---

## Búsqueda Vectorial

### Request

```http
POST /search
```

### Body

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

`collection` es opcional. Si no se envía, el servicio usa `generic_manuals`. Los filtros `document_id`, `chapter` y `section` son opcionales y se aplican conjuntamente sobre el payload de Qdrant dentro de la colección seleccionada.

## OCR e indexación

Los PDFs pendientes deben colgarse en `../data/documents/new/<nombre_coleccion>/`. El orquestador oficial es `process_manual_opt.py`:

```powershell
python scripts/process_manual_opt.py `
  --fast-ocr `
  --workers 2 `
  --memory-mode disk
```

La colección se deriva de la carpeta y se mantiene a lo largo del flujo:

```text
../data/documents/new/manuales_tecnicos/<nombre-manual>__parte-001.pdf
../data/documents/reading/manuales_tecnicos/<nombre-manual>__parte-001.pdf
../data/documents/completed/manuales_tecnicos/<nombre-manual>__parte-001.pdf
```

`process_manual_opt.py` mueve el PDF a `reading`, ejecuta OCR, genera chunks, carga en Qdrant y solo al final lo mueve a `completed`. Si falla, lo devuelve a `new` y registra el error en `../logs/ingestion.log`.

Para varias partes del mismo manual usa un identificador común:

```text
../data/documents/new/manuales_tecnicos/<nombre-manual>__parte-001.pdf
../data/documents/new/manuales_tecnicos/<nombre-manual>__parte-002.pdf
```

Las partes se cargan dentro de la colección indicada con el mismo `document_id` y un número de parte distinto. Los IDs de Qdrant son deterministas, así que reintentar una parte no duplica sus puntos.

Para manuales escaneados, el script OCR permite configurar el documento, la salida, la resolución y la colección:

```bash
python scripts/ocr_manual_opt.py \
  --pdf ../data/documents/new/manuales_tecnicos/<nombre-manual>__parte-001.pdf \
  --output ../data/artifacts/manual_pages.json \
  --collection manuales_tecnicos \
  --dpi 200 \
  --language spa
```

Después de regenerar el JSON, ejecuta `index_manual.py` y `upload_to_qdrant.py --collection <nombre_coleccion>`. Para comparar precisión y recall, guarda un reporte con `python scripts/evaluate_rag.py --skip-generation` antes y después del reprocesamiento.

### Response

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

El archivo:

```text
app/mcp_server.py
```

expone las herramientas MCP:

```python
search_manual(query: str, top_k: int = 3, collection: str = "generic_manuals") -> str
search_technical_manuals(query: str, top_k: int = 3) -> str
```

`search_manual` conserva la compatibilidad existente, usa `generic_manuals` por defecto y permite seleccionar una colección explícita. `search_technical_manuals` consulta la colección configurada para el dominio `technical_manuals`. Ambas tools aceptan los filtros opcionales `document_id`, `chapter` y `section`, y devuelven colección, documento, fuente, página y parte cuando están disponibles.

El mapa de dominios se configura con `MCP_DOMAIN_COLLECTIONS`, por ejemplo:

```json
{"technical_manuals":"manuales_tecnicos"}
```

La clave `technical_manuals` debe estar presente en el mapa porque la tool `search_technical_manuals` la consulta directamente.

El servidor comparte el proveedor de embeddings y mantiene una caché de `RAGService` por colección. MCP se expone en `http://localhost:8001/mcp` cuando se ejecuta con Docker Compose.

Iniciar localmente:

```bash
PYTHONPATH=src python -m manual_rag.entrypoints.mcp
```

---

# 🧪 Pruebas con cURL

```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mantenimiento de lubricacion",
    "top_k": 3
  }'
```

---

# ✅ Evaluación de calidad del RAG

`scripts/evaluate_rag.py` ejecuta un conjunto de consultas de referencia (`eval/golden_dataset.json`) contra Qdrant y Ollama, y mide:

- `context_precision`: fracción de fragmentos recuperados que provienen de una fuente esperada.
- `context_recall`: fracción de fuentes esperadas encontradas entre los resultados.
- `groundedness`: cobertura de palabras clave esperadas en la respuesta generada (o en el contexto, si se omite la generación).
- `latency_seconds`: tiempo de recuperación por consulta.

Los thresholds de aprobación se definen en `eval/thresholds.json`. El valor por defecto de `max_latency_seconds` está calibrado para hardware local de desarrollo (sin GPU); en CI o producción conviene bajarlo.

```bash
cd rag-engine
python scripts/evaluate_rag.py --skip-generation   # solo recuperación, sin Ollama
python scripts/evaluate_rag.py                     # incluye generación de respuesta con Ollama
```

El runner guarda un reporte JSON en `../data/artifacts/evaluations/eval_report_<timestamp>.json` y termina con código de salida distinto de cero si algún caso no supera los thresholds, para poder integrarlo en la pipeline de validación.

---

# 🐳 Servicios Docker

## Qdrant

```text
http://localhost:6333
```

Base de datos vectorial utilizada para almacenar embeddings.

## RAG Engine

```text
http://localhost:8000
```

Servicio FastAPI encargado de:

- Generar embeddings.
- Consultar Qdrant.
- Recuperar contexto relevante.

## API Gateway (Go)

```text
http://localhost:8080
```

Servicio frontal encargado de coordinar las consultas entre el cliente, el motor RAG y el LLM.

---

# 📊 Flujo de Consulta

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
Qdrant
   │
   ▼
Chunks relevantes
   │
   ▼
LLM
   │
   ▼
Respuesta final
```

Para que los filtros funcionen sobre datos existentes, hay que volver a ejecutar `index_manual.py` y `upload_to_qdrant.py` después de cambiar los metadatos.