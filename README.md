# Manual-RAG

Sistema de preguntas y respuestas basado en **Retrieval-Augmented Generation (RAG)** para consultar manuales técnicos utilizando lenguaje natural.

El proyecto es una solución distribuida orientada a servicios que combina:

* **API Gateway** desarrollado en **Go**.

* **Motor RAG** desarrollado en **Python** (FastAPI / MCP).

* **Frontend Web** desarrollado en **React + Vite**.

* **Qdrant** como base de datos vectorial.

* **Sentence Transformers** (`BAAI/bge-m3`) para embeddings.

* **Ollama** para generación de respuestas mediante modelos de lenguaje locales.

* **MinIO + Redis** para la ingesta de documentos asíncrona.

* **Servidor MCP** para integración con agentes y herramientas compatibles.

* **Stack de Observabilidad**: Prometheus, Grafana, Loki, Promtail y Tempo.

## 📚 Descripción general

Manual-RAG permite cargar manuales técnicos en formato PDF, procesarlos, extraer su contenido mediante OCR optimizado, dividirlos en fragmentos, generar embeddings e indexarlos en Qdrant.

Posteriormente, el usuario realiza preguntas en lenguaje natural como:

> ¿Cómo se realiza el mantenimiento del sistema de lubricación?

El sistema busca semánticamente los fragmentos más relevantes y los utiliza como contexto para generar una respuesta precisa mediante un modelo de lenguaje.

## 🏗️ Arquitectura y Flujo del Sistema

### Diagrama de alto nivel

```
               ┌───────────────────────────────┐
               │    Usuario / Cliente Web      │
               └───────────────┬───────────────┘
                               │
                               ▼
               ┌───────────────────────────────┐
               │  api-go (Gateway HTTP / Go)   │
               │  - Auth (JWT/Refresh)         │
               │  - Rate Limiting & Concurrencia│
               └───────────────┬───────────────┘
                               │
                               ▼
               ┌───────────────────────────────┐
               │  rag-engine (Motor RAG Python)│
               │  - Generación de embeddings   │
               │  - Búsqueda semántica         │
               └───────┬───────────────┬───────┘
                       │               │
                       ▼               ▼
          ┌─────────────────┐   ┌───────────────┐
          │     Qdrant      │   │    Ollama     │
          │ (Base Vectorial)│   │ (Modelo LLM)  │
          └─────────────────┘   └───────┬───────┘
                                        │
                                        ▼
                                Resposta Final / Stream

```

### Componentes y Responsabilidades

| Componente | Descripción / Rol |
| --- | --- |
| `api-go` | Gateway HTTP público, autenticación JWT, rate limiting, control de concurrencia y orquestación de solicitudes. |
| `rag-engine` | Servicio Python con Clean Architecture encargado de la generación de embeddings (`BAAI/bge-m3`) y recuperación semántica en Qdrant. |
| `apps/web` | Interfaz de usuario en React + Vite para login, selección de colecciones y chat (soporta respuestas estándar y SSE streaming). |
| `mcp-server` | Servidor compatible con Model Context Protocol (MCP) para exponer herramientas de consulta a agentes externos (ej. Copilot Chat). |
| `ingestion-api` / `worker` | Pipeline asíncrono basado en Redis DB 2 y MinIO para procesar, aplicar OCR e indexar documentos pesados. |
| `qdrant` | Base de datos vectorial para búsqueda por similitud de coseno/distancia. |
| `ollama` (externo) | Ejecutor local del modelo de lenguaje (ej. `qwen2.5:1.5b`), accesible vía host. |
| **Observabilidad** | `prometheus` (métricas), `grafana` (paneles), `loki`/`promtail` (logs centralizados) y `tempo` (trazado distribuido OTLP). |

> **Nota de Arquitectura:** El backend está implementado bajo los principios de **Clean Architecture** (Arquitectura Hexagonal), separando el dominio y los casos de uso de la infraestructura mediante puertos y adaptadores.

## 📂 Estructura del Repositorio

```
Manual-RAG/
├── apps/
│   └── web/                         # Frontend React + Vite
├── services/
│   ├── gateway-go/                  # API Gateway en Go
│   └── retrieval-python/            # Motor RAG, scripts e ingesta en Python
│       ├── app/                     # Wrappers FastAPI y MCP
│       ├── src/manual_rag/          # Arquitectura Hexagonal (domain, ports, adapters)
│       └── scripts/                 # Scripts de indexación, OCR y evaluación
├── deploy/
│   └── observability/               # Configuración de Grafana, Loki, Prometheus, Tempo
├── data/
│   ├── documents/                   # Estructura de documentos (new, reading, completed)
│   ├── artifacts/                   # Artefactos temporales de procesamiento
│   └── local/qdrant/                # Persistencia local de datos vectoriales
├── .vscode/                         # Configuración del servidor MCP para clientes
├── docker-compose.yml               # Orquestación para entorno de desarrollo
├── docker-compose.prod.yml          # Overlay para despliegue en producción con proxy
└── .env.example                     # Plantilla de variables de entorno

```

## ⚙️ Requisitos previos

* **Docker** y **Docker Compose**

* **Git**

* **Ollama** instalado y ejecutándose en el equipo host (`http://localhost:11434`) con un modelo previamente descargado (ej. `ollama run qwen2.5:1.5b`).

* Recurso recomendado: mínimo 8 GB de RAM (para soportar los modelos de embedding e inferencia simultáneamente).

Opcional (para desarrollo local sin Docker):

* **Go 1.22+**

* **Python 3.11+**

* **Tesseract OCR** (si se ejecutan scripts de OCR directamente en la máquina host)

## 🚀 Despliegue e Instalación

### 1. Clonar el repositorio y preparar credenciales

```
git clone https://github.com/javieralone/Manual-RAG.git
cd Manual-RAG

# Copiar el archivo de variables de entorno
cp .env.example .env.local

```

> ⚠️ **Importante:** Edita `.env.local` y asigna valores reales a las variables marcadas como `replace-with-*` (`AUTH_JWT_SECRET`, `AUTH_REFRESH_SECRET`, `AUTH_ADMIN_PASSWORD_HASH`, etc.) antes de continuar.

### 2. Iniciar el entorno de desarrollo

```
docker compose --env-file .env.local up -d --build

```

### 3. Iniciar el entorno de producción (Overlay)

Para entornos productivos que requieran un reverse proxy y gestión SSL con Certbot (publica en puertos `80` y `443` únicamente):

```
docker compose --env-file .env.local -f docker-compose.yml -f docker-compose.prod.yml up -d --build

```

*(Requiere configurar `DOMAIN` y `CERTBOT_EMAIL` en las variables de entorno).*

## 📦 Ingesta e Indexación de Documentos

El procesamiento de manuales sigue este flujo de estados de archivo dentro del directorio `data/documents/`:

```
data/documents/
├── new/<nombre_coleccion>/       # Documentos entrantes
├── reading/<nombre_coleccion>/   # En proceso activo de OCR / indexación
└── completed/<nombre_coleccion>/ # Procesados con éxito

```

### Método 1: Ingesta Asíncrona (Recomendado vía API)

El sistema integra un pipeline asíncrono con `ingestion-api` (disponible en `http://localhost:8002`), `ingestion-worker`, Redis y MinIO:

* **Encolar trabajo:** `POST http://localhost:8002/ingestion/enqueue`

* **Consultar estado:** `GET http://localhost:8002/ingestion/jobs/{job_id}`

* **Listar errores:** `GET http://localhost:8002/ingestion/failed`

### Método 2: Ingesta Directa vía CLI

Puedes ejecutar manualmente el script de optimización desde la raíz del proyecto sin pasar por la cola de Redis:

```
python services/retrieval-python/scripts/process_manual_opt.py \
  --fast-ocr \
  --workers 2 \
  --memory-mode disk

```

Para procesar un PDF o colección específica:

```
python services/retrieval-python/scripts/process_manual_opt.py \
  --pdf data/documents/new/manuales_tecnicos/manual-01.pdf \
  --collection manuales_tecnicos

```

## 🔌 Uso de la API HTTP

### 1. Autenticación (`POST /api/v1/auth/login`)

Obtención del token Bearer necesario para operar los endpoints protegidos:

```
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<tu_password>"}'

```

### 2. Consulta estándar (`POST /api/v1/query`)

```
curl -X POST http://localhost:8080/api/v1/query \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "¿Cómo se realiza el mantenimiento del sistema de lubricación?"
  }'

```

### 3. Consulta en tiempo real (`POST /api/v1/query/stream`)

Transmite la respuesta del LLM mediante **Server-Sent Events (SSE)** token por token:

```
curl -N --no-buffer -X POST http://localhost:8080/api/v1/query/stream \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"question": "¿Cómo se realiza el mantenimiento del sistema de lubricación?"}'

```

## 🤖 Integración con Agentes vía MCP (Model Context Protocol)

El motor expone un servidor MCP en el puerto `8001` (`http://localhost:8001/mcp`) documentado en `.vscode/mcp.json`.

Herramientas (`tools`) disponibles para clientes y agentes compatibles:

* `search_manual(query, top_k, collection, document_id, chapter, section)`: Búsqueda flexible en una colección explícita.

* `search_technical_manuals(query, top_k, document_id, chapter, section)`: Búsqueda orientada al dominio de manuales técnicos.

## 🌐 Tabla de Puertos y Servicios

| Servicio | Puerto Local | Función |
| --- | --- | --- |
| **API Gateway** (`api-go`) | `8080` | Punto de entrada único para peticiones HTTP/REST. |
| **RAG Engine** (`rag-engine`) | `8000` | Servicio interno Python de inferencia y vectorizado. |
| **MCP Server** | `8001` | Servidor MCP para integración con agentes IA. |
| **Ingestion API** | `8002` | API para encolar y gestionar trabajos de ingesta. |
| **Frontend Web** | `5173` | UI del usuario (React + Vite). |
| **Qdrant** | `6333` / `6334` | Base vectorial (HTTP / gRPC). |
| **MinIO** | `9000` / `9001` | Object Storage (API / Console). |
| **Grafana** | `3000` | Visualización de métricas y dashboards. |
| **Prometheus** | `9090` | Monitorización y recolección de métricas. |
| **Loki** | `3100` | Recolección centralizada de logs. |
| **Tempo** | `3200` | Trazado distribuido (OTLP `4317`/`4318`). |
| **Ollama** *(Host)* | `11434` | Inferencia LLM ejecutada localmente en la máquina host. |

## 📈 Observabilidad y Diagnóstico

El sistema cuenta con un pipeline completo de telemetría:

* Health Checks públicos en `/health` y `/ready`.

* Métricas expuestas en `/metrics` en formato Prometheus.

* Logs estructurados en formato JSON.

* Trazas distribuidas propagadas vía headers W3C (`traceparent`).

Para acceder a los paneles de monitoreo:

* **Grafana**: `http://localhost:3000` (incluye dashboards preconfigurados en `deploy/observability/grafana/dashboards/`).

* **Prometheus**: `http://localhost:9090`.

### Solución de problemas comunes

* **Contenedores no arrancan o fallan al conectar:** Revisa logs específicos con `docker compose logs -f <nombre_servicio>`.

* **Lentitud en primera consulta:** Ocurre mientras se descarga o inicializa en memoria el modelo de embedding `BAAI/bge-m3`.

* **Ollama no responde / Connection Refused:** Verifica que Ollama esté corriendo en la máquina host y responda a `curl http://localhost:11434/api/tags`.

* **Qdrant no devuelve contexto:** Asegúrate de que los documentos se hayan movido a la carpeta `completed/` y que la búsqueda apunte a la colección correcta (`generic_manuals` o `manuales_tecnicos`).

## 🛣️ Roadmap y Próximas Mejoras

* \[x\] Soporte multi-colección y streaming SSE.

* \[x\] Pipeline asíncrono de ingesta con Redis y MinIO.

* \[x\] Integración de servidor MCP.

* \[ \] Panel de administración UI para gestión de usuarios, historial de consultas y reindexación (ver [docs/roadmap/07-ui-consulta-admin.md](docs/roadmap/07-ui-consulta-admin.md)).

* \[ \] Ampliación de pruebas de integración End-to-End en CI/CD (ver [docs/roadmap/08-tests-integracion.md](docs/roadmap/08-tests-integracion.md)).

* \[ \] Caching de embedding para consultas frecuentes (ver [docs/roadmap/11-cache-reindexacion.md](docs/roadmap/11-cache-reindexacion.md)).

## 📄 Licencia y Créditos

Desarrollado por [@javieralone](https://github.com/javieralone?utm_source=gemini).

Repositorio Oficial: [https://github.com/javieralone/Manual-RAG](https://github.com/javieralone/Manual-RAG?utm_source=gemini)
