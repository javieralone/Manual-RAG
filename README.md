# Manual-RAG

Sistema de preguntas y respuestas basado en **Retrieval-Augmented Generation (RAG)** para consultar manuales técnicos utilizando lenguaje natural.

El proyecto combina:

- Un **API Gateway desarrollado en Go**.
- Un **motor RAG desarrollado en Python**.
- **Qdrant** como base de datos vectorial.
- **Sentence Transformers** para generar embeddings.
- **Ollama** para generar respuestas mediante un modelo local.
- Un servidor **MCP** para permitir la integración con agentes y herramientas compatibles.

---

## 📚 Descripción general

Manual-RAG permite cargar manuales técnicos, procesarlos, dividirlos en fragmentos, generar embeddings y almacenarlos en Qdrant.

Posteriormente, un usuario puede realizar preguntas como:

> ¿Cómo se realiza el mantenimiento del sistema de lubricación?

El sistema busca los fragmentos más relevantes del manual y los utiliza como contexto para generar una respuesta mediante un modelo de lenguaje.

El flujo principal es:

```text
Usuario
   │
   ▼
API Gateway en Go
   │
   ▼
Motor RAG en Python
   │
   ├── Generación de embedding de la pregunta
   ├── Búsqueda semántica en Qdrant
   └── Recuperación del contexto relevante
   │
   ▼
Ollama
   │
   ▼
Respuesta generada
```

---

## 🏗️ Arquitectura del proyecto

Manual-RAG está diseñado como una solución distribuida orientada a servicios, con separación clara entre entrada HTTP, recuperación semántica, almacenamiento vectorial y generación de respuestas.

Además, el repositorio incluye una interfaz web de consulta en React + Vite para autenticación, selección de colección y chat con soporte para consulta normal y streaming.

### Diagrama de alto nivel

```text
Usuario / Cliente
      │
      ▼
api-go (Gateway HTTP en Go)
      │
      ├── Valida autenticación y autorización
      ├── Aplica rate limiting y timeouts
      ├── Orquesta la consulta
      ▼
rag-engine (Motor RAG en Python)
      │
      ├── Genera embedding de la pregunta
      ├── Consulta Qdrant por similitud semántica
      ├── Recupera contextos relevantes
      └── Devuelve el contexto y metadatos
      │
      ▼
Qdrant (base vectorial)
      │
      ▼
Ollama (modelo local de lenguaje)
      │
      ▼
Respuesta final al cliente
```

Además, el proyecto incluye:

- `mcp-server`: exposición del sistema a clientes MCP compatibles.
- `prometheus`, `grafana`, `loki`, `promtail` y `tempo`: stack de observabilidad para métricas, logs y trazas.

### Estructura del repositorio

```text
Manual-RAG/
├── .vscode/                         # Configuración del cliente MCP
│   └── mcp.json
├── services/gateway-go/             # Gateway HTTP y orquestación de consultas
│   ├── cmd/
│   │   └── api/
│   │       └── main.go
│   ├── internal/
│   │   ├── core/
│   │   │   ├── domain/
│   │   │   ├── ports/
│   │   ├── application/
│   │   │   └── use_cases/
│   │   ├── adapters/
│   │   │   ├── auth/
│   │   │   ├── clients/
│   │   │   ├── decorators/
│   │   │   ├── http/
│   │   │   └── observability/
│   │   └── README.md
│   ├── Dockerfile
│   ├── go.mod
│   └── README.md
│
├── services/retrieval-python/       # Servicio de recuperación y embeddings
│   ├── app/
│   │   ├── main.py                    # Wrapper compatible FastAPI
│   │   └── mcp_server.py              # Wrapper compatible MCP
│   ├── src/manual_rag/
│   │   ├── adapters/
│   │   ├── application/
│   │   ├── domain/
│   │   ├── ports/
│   │   ├── entrypoints/
│   │   │   ├── http.py
│   │   │   └── mcp.py
│   │   ├── bootstrap.py
│   │   └── observability.py
│   ├── scripts/
│   │   ├── index_manual.py
│   │   ├── ocr_manual.py
│   │   ├── query_rag.py
│   │   ├── test_rag_direct.py
│   │   └── upload_to_qdrant.py
│   ├── tests/
│   │   └── test_observability.py
│   ├── requirements.txt
│   ├── Dockerfile
│   └── README.md
│
├── apps/web/                        # UI React + Vite para login y chat
│   ├── src/
│   ├── Dockerfile
│   ├── package.json
│   ├── vite.config.js
│   ├── .env.example
│   └── README.md
│
├── deploy/observability/            # Métricas, trazas y logs
│   ├── grafana/
│   ├── loki/
│   ├── prometheus/
│   ├── promtail/
│   └── tempo/
│
├── data/
│   ├── documents/                  # Manuales originales cargados al sistema
│   ├── artifacts/                 # Artefactos generados durante procesamiento
│   └── local/qdrant/               # Persistencia del índice vectorial
├── .env.example                    # Plantilla de configuración
├── .env.local                      # Configuración local con secretos (ignorada)
├── .gitignore
├── docker-compose.yml              # Orquestación de servicios
├── README.md
└── .github/
```

### Componentes y responsabilidades

| Componente | Rol principal |
|---|---|
| `api-go` | Gateway público, autenticación, timeout, rate limiting y orquestación |
| `rag-engine` | Generación de embeddings, búsqueda semántica y recuperación de contexto |
| `frontend` | UI React + Vite para login, chat y selección de colección/modo |
| `mcp-server` | Exposición del servicio RAG a clientes MCP |
| `qdrant` | Base vectorial para búsqueda por similitud |
| `ollama` | Generación final de respuesta a partir del contexto recuperado |
| `prometheus`, `grafana`, `loki`, `tempo` | Observabilidad centralizada del sistema |

> La implementación sigue una arquitectura limpia (Clean Architecture) con dominio, puertos y adaptadores bien definidos, evitando acoplamiento directo entre la lógica de negocio y los servicios externos.

---

## 🧩 Principios de diseño

El proyecto sigue principios de arquitectura hexagonal y de **Clean Architecture** para mantener el sistema fácil de extender, testear y operar.

### Separación de responsabilidades

Cada capa tiene una misión concreta:

- La API en Go recibe y valida requests HTTP.
- El motor Python realiza la recuperación semántica y la lógica RAG.
- Qdrant almacena y consulta embeddings para similitud.
- Ollama genera la respuesta final con contexto relevante.
- Los contratos entre capas se expresan mediante interfaces o puertos.

### Inversión de dependencias

El núcleo del sistema depende de abstracciones, no de implementaciones concretas.

Esto permite cambiar el modelo de embeddings, el cliente de búsqueda o la infraestructura sin reescribir la lógica principal. Un ejemplo claro es el adaptador de embeddings que implementa el contrato `EmbeddingPort`.

### Control de concurrencia y resiliencia

El gateway Go implementa control de concurrencia para evitar saturar la memoria y el CPU cuando hay varias consultas simultáneas.

Además, las rutas de consulta aplican rate limiting por IP y usuario autenticado. Los parámetros `RATE_LIMIT_ENABLED`, `RATE_LIMIT_REQUESTS` y `RATE_LIMIT_WINDOW` permiten ajustar la protección. Si se excede la cuota, el sistema responde con `429 Too Many Requests`.

### Cancelación y timeouts

Las llamadas HTTP y los flujos de consulta usan contextos con timeout para evitar bloqueos, liberar recursos y responder correctamente cuando el cliente cancela una solicitud.

---

## ⚙️ Requisitos

Para ejecutar el proyecto necesitas:

- Docker
- Docker Compose
- Git
- Ollama, si se ejecuta fuera de Docker
- Un modelo de lenguaje compatible con Ollama
- Suficiente memoria RAM para ejecutar el modelo y el modelo de embeddings

Opcionalmente:

- Go 1.22 o superior
- Python 3.11 o superior
- Qdrant instalado localmente

---

## 🚀 Ejecución con Docker Compose

Clona el repositorio:

```bash
git clone https://github.com/javieralone/Manual-RAG.git
cd Manual-RAG
```

Construye y levanta los servicios:

```bash
docker compose up -d --build
```

Comprueba el estado de los contenedores:

```bash
docker compose ps
```

Consulta los logs:

```bash
docker compose logs -f
```

Para detener el sistema:

```bash
docker compose down
```

Para detenerlo y eliminar los volúmenes:

```bash
docker compose down -v
```

> El primer arranque puede tardar debido a la descarga del modelo de embeddings y del modelo de lenguaje.

---

## 🔌 API

### `GET /health`

Comprueba si el API Gateway está disponible.

Respuesta:

```json
{
  "status": "UP"
}
```

### `POST /api/v1/query`

Realiza una consulta autenticada sobre los manuales indexados. Requiere un token de acceso en la cabecera `Authorization`.

Petición:

```json
{
  "question": "¿Cómo se realiza el mantenimiento del sistema de lubricación?"
}
```

Respuesta:

```json
{
  "question": "¿Cómo se realiza el mantenimiento del sistema de lubricación?",
  "answer": "El mantenimiento requiere revisar los niveles de aceite...",
  "context": [
    {
      "text": "Fragmento relevante recuperado del manual.",
      "score": 0.89,
      "metadata": {}
    }
  ]
}
```

Ejemplo utilizando `curl`:

```bash
curl -X POST http://localhost:8080/api/v1/query \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "¿Cómo se realiza el mantenimiento del sistema de lubricación?"
  }'
```

> Si el puerto configurado en `docker-compose.yml` es diferente, reemplaza `8080` por el puerto correspondiente.

### `POST /api/v1/query/stream`

Misma consulta que `/api/v1/query`, pero la respuesta de Ollama se transmite en tiempo real mediante **Server-Sent Events** (`metadata` → `token`* → `complete`/`error`), sin que el Gateway reconstruya la respuesta completa. Requiere el mismo `Bearer <access_token>`. Detalle completo, diagrama de secuencia y ejemplos en [`services/gateway-go/README.md`](services/gateway-go/README.md#post-querystream--streaming-en-tiempo-real-sse).

```bash
curl -N --no-buffer -X POST http://localhost:8080/api/v1/query/stream \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"question": "¿Cómo se realiza el mantenimiento del sistema de lubricación?"}'
```

### Autenticación

Obtén un par de tokens con las credenciales configuradas para el usuario administrador:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<password>"}'
```

Renueva el acceso cuando expire el token:

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

Los endpoints `/health`, `/ready` y `/metrics` son públicos. La consulta requiere `Bearer <access_token>`.

---

## 🧠 Motor de embeddings

El motor RAG utiliza `SentenceTransformer` y el modelo:

```text
BAAI/bge-m3
```

Este modelo transforma el texto en vectores numéricos que permiten realizar búsquedas semánticas.

La implementación se encuentra en:

```text
services/retrieval-python/src/manual_rag/adapters/bge_embedding_adapter.py
```

El modelo se carga una sola vez cuando se inicia el servicio para evitar volver a descargarlo o inicializarlo en cada consulta.

---

## 📦 Indexación de documentos

Antes de realizar consultas, los manuales deben ser procesados e indexados.

El proceso general es:

```text
Documento PDF
   │
   ▼
Extracción de texto u OCR
   │
   ▼
División en fragmentos
   │
   ▼
Generación de embeddings
   │
   ▼
Almacenamiento en Qdrant
```

Los scripts relacionados con este proceso se encuentran en:

```text
services/retrieval-python/scripts/
```

La ingesta usa estas carpetas:

```text
data/documents/
├── new/        # PDFs pendientes
├── reading/    # PDF en procesamiento
└── completed/  # PDFs procesados correctamente
```

Para manuales divididos en varias partes, usa el mismo identificador y numera cada PDF:

```text
data/documents/new/manual-reparaciones-valiant__parte-001.pdf
data/documents/new/manual-reparaciones-valiant__parte-002.pdf
```

Ejecuta el orquestador oficial desde la raíz del proyecto:

```powershell
python services/retrieval-python/scripts/process_manual_opt.py `
  --fast-ocr `
  --workers 2 `
  --memory-mode disk
```

`process_manual_opt.py` toma un PDF cada vez de `data/documents/new/<nombre_coleccion>`, lo mueve a `reading`, ejecuta OCR, indexación y carga en Qdrant, y lo mueve a `completed` solo si todas las etapas terminan correctamente. Si falla o se interrumpe, vuelve a `new` y registra el error en `logs/ingestion.log`.

La ingesta sigue el patrón de carpetas `data/documents/new/<nombre_coleccion>/`, `data/documents/reading/<nombre_coleccion>/` y `data/documents/completed/<nombre_coleccion>/`. El nombre de la colección se deriva de la carpeta y se conserva durante todo el ciclo de vida del archivo.

Ejemplo de estructura:

```text
data/documents/
├── new/
│   ├── manuales_tecnicos/
│   │   ├── <nombre-manual>__parte-001.pdf
│   │   └── <nombre-manual>__parte-002.pdf
│   └── generic_manuals/
│       └── <nombre-manual>__parte-001.pdf
├── reading/
│   ├── manuales_tecnicos/
│   └── generic_manuals/
└── completed/
    ├── manuales_tecnicos/
    └── generic_manuals/
```

El orquestador oficial es `process_manual_opt.py` y conserva los parámetros de OCR y rendimiento:

```powershell
python services/retrieval-python/scripts/process_manual_opt.py `
  --fast-ocr `
  --workers 2 `
  --memory-mode disk
```

También puedes procesar un PDF concreto o forzar una colección:

```powershell
python services/retrieval-python/scripts/process_manual_opt.py --pdf data/documents/new/manuales_tecnicos/<nombre-manual>__parte-001.pdf --collection manuales_tecnicos
```

Si la colección no se indica, se usa la inferida desde la carpeta `data/documents/new/<nombre_coleccion>/`. Cuando no hay carpeta compatible, el valor por defecto es `generic_manuals`.

El OCR admite configuración adicional de salida, resolución y lenguaje:

```bash
python services/retrieval-python/scripts/ocr_manual_opt.py \
  --pdf data/documents/new/manuales_tecnicos/<nombre-manual>__parte-001.pdf \
  --output data/artifacts/manual_pages.json \
  --collection manuales_tecnicos \
  --dpi 200 \
  --language spa
```

Cada página se convierte a escala de grises, se mejora el contraste y se limpia antes de ejecutar Tesseract. Si el texto preprocesado tiene peor calidad que el resultado directo, se conserva el resultado directo como fallback. La variable `TESSERACT_CMD` permite indicar explícitamente el binario cuando no está disponible en el `PATH`.

Para medir el impacto sobre recuperación, ejecuta la evaluación antes y después de regenerar los chunks y subirlos a Qdrant:

```bash
cd rag-engine
python scripts/evaluate_rag.py --skip-generation
```

---

## 🔎 Flujo de una consulta

Cuando llega una pregunta al sistema:

1. El usuario envía una petición al API Gateway.
2. El Gateway valida la solicitud.
3. La pregunta se envía al motor RAG.
4. Se genera el embedding de la pregunta.
5. Qdrant busca los fragmentos más similares.
6. El contexto recuperado se envía al modelo de Ollama.
7. Ollama genera la respuesta.
8. El Gateway devuelve la respuesta junto con el contexto utilizado.

---

## 🤖 Integración MCP

El proyecto incluye un servidor MCP en:

```text
services/retrieval-python/src/manual_rag/entrypoints/mcp.py
```

La configuración para clientes compatibles se encuentra en:

```text
.vscode/mcp.json
```

Configuración actual:

```json
{
  "servers": {
    "manual-rag": {
      "type": "http",
      "url": "http://localhost:8001/mcp"
    }
  }
}
```

Esto permite que agentes compatibles con MCP puedan consultar el sistema RAG mediante herramientas externas.

El servidor expone actualmente estas tools:

```text
search_manual(query, top_k=3, collection="generic_manuals", document_id=None, chapter=None, section=None)
search_technical_manuals(query, top_k=3, document_id=None, chapter=None, section=None)
```

`search_manual` permite consultar cualquier colección válida de forma explícita. `search_technical_manuals` fija la colección del dominio técnico mediante `MCP_DOMAIN_COLLECTIONS`. Las consultas vacías y los valores de `top_k` fuera del rango `1..20` se rechazan.

Para probarlo desde Copilot Chat, activa el servidor `manual-rag` y solicita, por ejemplo:

```text
Usa search_technical_manuals para buscar el procedimiento de mantenimiento del sistema de lubricación.
```

---

## 🧪 Desarrollo local

### Ejecutar el motor Python

Crea y activa un entorno virtual:

```bash
cd services/retrieval-python

python -m venv .venv
```

En Linux o macOS:

```bash
source .venv/bin/activate
```

En Windows:

```powershell
.venv\Scripts\activate
```

Instala las dependencias:

```bash
pip install -r requirements.txt
```

Inicia el servicio FastAPI:

```bash
PYTHONPATH=src uvicorn manual_rag.entrypoints.http:app --host 0.0.0.0 --port 8000 --reload
```

El servidor MCP se inicia por separado en el puerto `8001`:

```bash
PYTHONPATH=src python -m manual_rag.entrypoints.mcp
```

La validación final de arquitectura está documentada en [docs/architecture/final-validation.md](docs/architecture/final-validation.md). La puesta en producción requiere además cerrar el checklist de [docs/operations/production-readiness.md](docs/operations/production-readiness.md).

### Ejecutar el API Gateway en Go

```bash
cd services/gateway-go
go mod tidy
go build -v ./...
go run ./cmd/api
```

Para ejecutar las pruebas:

```bash
go test ./...
```

---

## 🗂️ Servicios principales

| Servicio | Responsabilidad | Descripción / Función |
| :--- | :--- | :--- |
| `api-go` | API Gateway y punto de entrada público | Gestiona peticiones HTTP, valida solicitudes, aplica control de concurrencia y orquesta la comunicación con el motor RAG. |
| `rag-engine` | Recuperación semántica y API interna | Genera embeddings de preguntas con `bge-m3`, realiza búsquedas vectoriales y construye el contexto para el LLM. |
| `qdrant` | Base de datos vectorial | Almacena los vectores semánticos de los manuales y ejecuta búsquedas de similitud en tiempo real. |
| `mcp-server` | Servidor MCP para agentes | Expone las herramientas y capacidades del sistema RAG para integrarse con clientes y agentes compatibles con MCP. |
| Ollama (externo) | Generación de respuestas mediante un LLM | Debe ejecutarse en el equipo host; `api-go` lo alcanza mediante `host.docker.internal:11434`. No es un servicio definido en Docker Compose. |
| `prometheus` | Recolector de métricas | Mide en tiempo real la latencia, tráfico, tasa de errores y disponibilidad de los componentes de la aplicación. |
| `grafana` | Visualización y paneles | Dashboard unificado que muestra gráficas de rendimiento, alertas activas y logs de la infraestructura. |
| `loki` | Almacenamiento y gestión de logs | Agrupa y centraliza los registros de texto emitidos por las aplicaciones para diagnosticar fallos y errores. |
| `tempo` | Tracing / Rastreo distribuido | Mide el tiempo de ejecución exacto y el recorrido de las peticiones entre los distintos microservicios. |

---

## 🔐 Variables de configuración

La configuración puede variar según el entorno. Docker Compose utiliza estas variables principales:

```env
PYTHON_ENGINE_URL=http://rag-engine:8000
OLLAMA_URL=http://host.docker.internal:11434
OLLAMA_MODEL=qwen2.5:1.5b
AUTH_JWT_SECRET=<secreto>
AUTH_REFRESH_SECRET=<secreto>
AUTH_ADMIN_USERNAME=admin
AUTH_ADMIN_PASSWORD_HASH=<hash-bcrypt>
WORKER_LIMIT=2
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=60
RATE_LIMIT_WINDOW=1m
HTTP_CLIENT_TIMEOUT=5m
REQUEST_TIMEOUT=5m
```

El motor RAG usa `QDRANT_HOST=qdrant`, `QDRANT_PORT=6333`, el valor por defecto `generic_manuals` para la colección y el modelo de embeddings `BAAI/bge-m3`. La colección puede seleccionarse explícitamente desde la API, MCP o la ingesta, y `manuales_tecnicos` sigue funcionando como una colección independiente y compatible. Ollama no es un servicio de Compose: debe estar disponible en el equipo host mediante `host.docker.internal:11434`.

El servidor MCP usa `MCP_PORT=8001` y el mapa `MCP_DOMAIN_COLLECTIONS` para asociar tools de dominio con colecciones Qdrant. El mapa debe incluir la clave `technical_manuals`; en el Compose actual apunta a `manuales_tecnicos`.

No incluyas claves privadas, tokens ni credenciales directamente en el repositorio. Las variables `AUTH_JWT_SECRET`, `AUTH_REFRESH_SECRET`, `AUTH_ADMIN_USERNAME` y `AUTH_ADMIN_PASSWORD_HASH` son obligatorias al iniciar `api-go`.

Para desarrollo local copia `.env.example` a `.env.local`, completa los secretos y usa ese archivo explícitamente con Docker Compose:

```powershell
Copy-Item .env.example .env.local
docker compose --env-file .env.local up -d --build
```

`.env.example` es apto para versionar; `.env.local` y `.env` están ignorados porque pueden contener secretos. Si ya usas `.env`, puedes continuar con `docker compose up`; para el archivo local separado debes indicar siempre `--env-file .env.local`.

### Puertos

| Servicio | Puerto local |
|---|---:|
| API Gateway | `8080` |
| RAG Engine | `8000` |
| MCP | `8001` |
| Qdrant | `6333` (HTTP), `6334` (gRPC) |
| Grafana | `3000` |
| Prometheus | `9090` |
| Loki | `3100` |
| Tempo | `3200`, OTLP `4317`/`4318` |

---

## 🩺 Solución de problemas

### Los contenedores no arrancan

Consulta los logs:

```bash
docker compose logs -f
```

También puedes reconstruir las imágenes:

```bash
docker compose down
docker compose build --no-cache
docker compose up -d
```

### El modelo de embeddings tarda en cargar

La primera ejecución descarga el modelo `BAAI/bge-m3`. Este proceso puede tardar y requiere conexión a Internet.

### Qdrant no encuentra resultados

Comprueba que:

- Los documentos hayan sido indexados.
- La colección exista en Qdrant.
- El modelo utilizado para indexar sea el mismo utilizado para consultar.
- El tamaño de los vectores sea compatible.
- El volumen de Qdrant esté correctamente montado.

### Ollama no responde

Comprueba que el servicio esté funcionando:

```bash
curl http://localhost:11434/api/tags
```

También verifica que el modelo requerido esté descargado:

```bash
ollama list
```

---

## 📁 Persistencia de datos

Los siguientes directorios pueden contener datos generados por el sistema:

```text
data/documents/
data/artifacts/
data/local/qdrant/
```

El directorio `data/local/qdrant/` se monta directamente desde el host, por lo que conserva los vectores al reiniciar los contenedores. `docker compose down -v` elimina los volúmenes nombrados, pero no borra ese directorio; elimínalo manualmente solo si quieres reconstruir el índice.

---

## 🛡️ Buenas prácticas

- No subir documentos confidenciales al repositorio.
- No incluir credenciales en archivos versionados.
- Utilizar variables de entorno para la configuración.
- Mantener separados los documentos originales y los datos procesados.
- Limitar el número de consultas simultáneas.
- Configurar timeouts para las llamadas entre servicios.
- Monitorizar el consumo de memoria de los modelos.
- Utilizar volúmenes persistentes para Qdrant.

---

## 🛣️ Estado y próximas mejoras

### Capacidades ya implementadas

- Interfaz web React + Vite con login, selección de colección, chat y consulta normal o streaming.
- Evaluación automática del RAG con métricas de precisión, recall, groundedness y latencia.
- OCR optimizado para manuales escaneados, con control de resolución, idioma y uso de memoria.
- Soporte multi-colección y multi-manual, con filtros por documento, capítulo y sección.
- Servidor MCP con herramientas por dominio y selección explícita de colección.
- Observabilidad base con métricas Prometheus, logs JSON, trazas OpenTelemetry, Grafana, Loki y Tempo.

### Pendientes

- Añadir un [panel administrativo operativo](docs/roadmap/07-ui-consulta-admin.md) para historial de consultas, contexto recuperado, métricas y gestión avanzada de usuarios y permisos.
- Añadir una [cola de trabajos](docs/roadmap/06-cola-indexacion.md) para indexación asíncrona y procesamiento por lotes.
- Extender las [pruebas de integración](docs/roadmap/08-tests-integracion.md) con Docker Compose para validar el flujo completo entre gateway, RAG, Qdrant y Ollama.
- Mejorar los [dashboards operativos](docs/roadmap/10-dashboards-operativos.md) con vistas específicas de recuperación, tiempo hasta el primer token y latencia de Ollama.
- Explorar estrategias de [caché y reindexación incremental](docs/roadmap/11-cache-reindexacion.md) para reducir tiempos de respuesta y carga.

---

## 📈 Observabilidad

La revisión de preparación productiva y su checklist están en [docs/operations/production-readiness.md](docs/operations/production-readiness.md).

El sistema incluye observabilidad end-to-end:

- `/health`, `/ready` y `/metrics` en `api-go` y `rag-engine`.
- Métricas Prometheus de tráfico, latencias, errores, autenticación, concurrencia, retrieval, embeddings, Qdrant y Ollama.
- Logs JSON en Go y Python.
- Promtail recoge los logs Docker y los envía a Loki con labels estables (`service`, `container`, `project`, `environment`, `level`).
- Trazas OpenTelemetry OTLP con destino Tempo y propagación W3C entre gateway, RAG engine y Ollama.
- Stack Docker Compose con Prometheus, Grafana, Loki y Tempo.
- Dashboard provisionado en `deploy/observability/grafana/dashboards/`.
- Alertas base en `deploy/observability/prometheus/alerts.yml` para disponibilidad, latencia, errores, Ollama y Qdrant.

URLs locales:

| Servicio | URL |
|---|---|
| Prometheus | `http://localhost:9090` |
| Grafana | `http://localhost:3000` |
| Loki | `http://localhost:3100` |
| Tempo | `http://localhost:3200` |

Para levantar todo:

```bash
docker compose up -d --build
```

---

## 📄 Licencia

Este proyecto se distribuye bajo la licencia definida en el repositorio.

Si no existe una licencia específica, todos los derechos quedan reservados por el autor.

---

## 👤 Autor

Desarrollado por [@javieralone](https://github.com/javieralone).

Repositorio:

[https://github.com/javieralone/Manual-RAG](https://github.com/javieralone/Manual-RAG)
````
