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
├── api-go/                          # Gateway HTTP y orquestación de consultas
│   ├── cmd/
│   │   └── api/
│   │       └── main.go
│   ├── internal/
│   │   ├── core/
│   │   │   ├── domain/
│   │   │   ├── ports/
│   │   │   └── services/
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
├── rag-engine/                      # Servicio de recuperación y embeddings
│   ├── app/
│   │   ├── adapters/
│   │   ├── core/
│   │   ├── main.py
│   │   ├── mcp_server.py
│   │   ├── observability.py
│   │   └── __init__.py
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
├── observability/                   # Métricas, trazas y logs
│   ├── grafana/
│   ├── loki/
│   ├── prometheus/
│   ├── promtail/
│   └── tempo/
│
├── documents/                      # Manuales originales cargados al sistema
├── output/                         # Artefactos generados durante procesamiento
├── qdrant_storage/                 # Persistencia del índice vectorial
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

### `POST /query/stream`

Misma consulta que `/api/v1/query`, pero la respuesta de Ollama se transmite en tiempo real mediante **Server-Sent Events** (`metadata` → `token`* → `complete`/`error`), sin que el Gateway reconstruya la respuesta completa. Requiere el mismo `Bearer <access_token>`. Detalle completo, diagrama de secuencia y ejemplos en [`api-go/README.md`](api-go/README.md#post-querystream--streaming-en-tiempo-real-sse).

```bash
curl -N --no-buffer -X POST http://localhost:8080/query/stream \
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
rag-engine/app/adapters/bge_embedding_adapter.py
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
rag-engine/scripts/
```

La ingesta usa estas carpetas:

```text
documents/
├── new/        # PDFs pendientes
├── reading/    # PDF en procesamiento
└── completed/  # PDFs procesados correctamente
```

Para manuales divididos en varias partes, usa el mismo identificador y numera cada PDF:

```text
documents/new/manual-reparaciones-valiant__parte-001.pdf
documents/new/manual-reparaciones-valiant__parte-002.pdf
```

Ejecuta el orquestador desde la raíz del proyecto:

```bash
python rag-engine/scripts/process_manual.py
```

`process_manual.py` toma todos los PDFs de `documents/new`, mueve cada uno a `reading`, ejecuta OCR, indexación y carga en Qdrant, y lo mueve a `completed` solo si las tres etapas terminan correctamente. Si falla, vuelve a `new` y registra el error en `logs/ingestion.log`. Un lock impide ejecutar dos ingestas simultáneas.

Los nombres con formato `<document_id>__parte-<numero>.pdf` permiten agrupar partes del mismo manual en la colección `manuales_tecnicos` mediante `document_id`. Para depurar un archivo concreto también se puede usar `--pdf`, sin aplicar el movimiento de estados:

```bash
python rag-engine/scripts/process_manual.py --pdf documents/manual-escaneado.pdf
```

El OCR admite un PDF y una salida alternativos, además de ajustar la resolución:

```bash
python rag-engine/scripts/ocr_manual.py \
  --pdf documents/manual-escaneado.pdf \
  --output output/manual_pages.json \
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
rag-engine/app/mcp_server.py
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

---

## 🧪 Desarrollo local

### Ejecutar el motor Python

Crea y activa un entorno virtual:

```bash
cd rag-engine

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
uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
```

El servidor MCP se inicia por separado en el puerto `8001`:

```bash
python -m app.mcp_server
```

### Ejecutar el API Gateway en Go

```bash
cd api-go
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
| `ollama` | Generación de respuestas mediante un LLM | Ejecuta el modelo de lenguaje en local para redactar respuestas precisas utilizando el contexto recuperado. |
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

El motor RAG usa `QDRANT_HOST=qdrant`, `QDRANT_PORT=6333`, la colección `manuales_tecnicos` y el modelo de embeddings `BAAI/bge-m3`. Ollama no es un servicio de Compose: debe estar disponible en el equipo host mediante `host.docker.internal:11434`.

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
documents/
output/
qdrant_storage/
```

El directorio `qdrant_storage/` se monta directamente desde el host, por lo que conserva los vectores al reiniciar los contenedores. `docker compose down -v` elimina los volúmenes nombrados, pero no borra ese directorio; elimínalo manualmente solo si quieres reconstruir el índice.

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

## 🛣️ Próximas mejoras

El roadmap del proyecto puede enfocarse en tres grandes líneas: mejor experiencia de usuario, escalabilidad y calidad del RAG.

### Experiencia y acceso

- Desarrollar una interfaz web para consultar el sistema sin usar curl o clientes HTTP.
- Añadir paneles de administración para revisar consultas, contexto recuperado y métricas de uso.
- Mejorar la experiencia de autenticación y autorización para múltiples roles y permisos.

### Calidad del RAG

- Añadir evaluación automática de calidad de respuestas mediante groundedness, relevancia y precisión del contexto.
- Soportar múltiples colecciones o índices por familia de manuales.
- Mejorar la extracción de texto OCR para documentos escaneados o con baja calidad.

### Operabilidad y escalabilidad

- Añadir una cola de trabajos para indexación asíncrona y procesamiento por lotes.
- Extender las pruebas de integración con Docker Compose para validar flujos completos en entorno real.
- Mejorar la observabilidad con dashboards más específicos por consulta, tiempo de recuperación y latencia de Ollama.
- Explorar estrategias de caché y reindexación incremental para reducir tiempos de respuesta y carga.

---

## 📈 Observabilidad

El sistema incluye observabilidad end-to-end:

- `/health`, `/ready` y `/metrics` en `api-go` y `rag-engine`.
- Métricas Prometheus de tráfico, latencias, errores, autenticación, concurrencia, retrieval, embeddings, Qdrant y Ollama.
- Logs JSON en Go y Python.
- Promtail recoge los logs Docker y los envía a Loki con labels estables (`service`, `container`, `project`, `environment`, `level`).
- Trazas OpenTelemetry OTLP con destino Tempo y propagación W3C entre gateway, RAG engine y Ollama.
- Stack Docker Compose con Prometheus, Grafana, Loki y Tempo.
- Dashboard provisionado en `observability/grafana/dashboards/`.
- Alertas base en `observability/prometheus/alerts.yml` para disponibilidad, latencia, errores, Ollama y Qdrant.

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
