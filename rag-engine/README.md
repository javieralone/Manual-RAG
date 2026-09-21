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
rag-engine/
├── app/
│   ├── adapters/                  # Adaptadores de infraestructura
│   │   ├── sentence_transformer_adapter.py
│   │   └── qdrant_adapter.py
│   │
│   ├── core/
│   │   ├── domain/
│   │   │   └── schemas.py         # DTOs y modelos Pydantic
│   │   │
│   │   └── ports/
│   │       ├── embedding_port.py
│   │       └── vector_store_port.py
│   │
│   ├── services/
│   │   └── rag_service.py         # Casos de uso
│   │
│   ├── main.py                    # API FastAPI
│   └── mcp_server.py              # Servidor MCP
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

Colección:

```text
manuales_tecnicos
```

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
| `OMP_NUM_THREADS` | Hilos para CPU | `2` |
| `MKL_NUM_THREADS` | Hilos MKL | `2` |

---

# 🚀 Despliegue con Docker Compose

Construir y levantar el servicio:

```powershell
docker compose up -d --build rag-engine
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
  "status": "UP"
}
```

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
  "top_k": 3
}
```

### Response

```json
{
  "results": [
    {
      "page": 12,
      "source": "manual_tecnico.pdf",
      "text": "El mantenimiento del sistema de lubricación requiere...",
      "score": 0.895
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

expone la herramienta MCP:

```python
search_manual(query: str, top_k: int = 3) -> str
```

Iniciar localmente:

```bash
python -m app.mcp_server
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