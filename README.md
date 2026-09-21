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

```text
Manual-RAG/
├── api-go/                         # API Gateway desarrollado en Go
│   ├── cmd/
│   │   └── api/
│   │       └── main.go             # Punto de entrada de la aplicación
│   ├── internal/
│   │   ├── core/
│   │   │   ├── domain/             # Entidades y reglas de negocio
│   │   │   ├── ports/              # Interfaces del dominio
│   │   │   └── services/           # Casos de uso y orquestación
│   │   └── adapters/
│   │       ├── http/               # Handlers, router y middlewares
│   │       ├── clients/            # Clientes para Python y Ollama
│   │       └── decorators/         # Control de concurrencia
│   ├── Dockerfile
│   ├── go.mod
│   └── README.md
│
├── rag-engine/                     # Motor RAG desarrollado en Python
│   ├── app/
│   │   ├── adapters/               # Implementaciones concretas
│   │   ├── core/
│   │   │   ├── domain/             # Entidades del dominio
│   │   │   ├── ports/              # Interfaces y contratos
│   │   │   └── services/           # Casos de uso RAG
│   │   └── infrastructure/         # Configuración e infraestructura
│   ├── scripts/
│   │   ├── index_manual.py         # Indexación de manuales
│   │   ├── ocr_manual.py           # Procesamiento OCR
│   │   ├── query_rag.py            # Consultas al sistema RAG
│   │   └── upload_to_qdrant.py     # Carga de vectores
│   ├── main.py                     # API interna con FastAPI
│   ├── mcp_server.py               # Servidor MCP
│   ├── requirements.txt
│   └── Dockerfile
│
├── documents/                      # Manuales y documentos originales
├── output/                         # Archivos procesados y resultados
├── qdrant_storage/                 # Almacenamiento persistente de Qdrant
├── .vscode/
│   └── mcp.json                    # Configuración del servidor MCP
├── docker-compose.yml              # Orquestación de servicios
└── README.md
```

---

## 🧩 Principios de diseño

El proyecto sigue una arquitectura hexagonal y principios de **Clean Architecture**.

### Separación de responsabilidades

Cada componente tiene una responsabilidad específica:

- La API en Go gestiona las peticiones HTTP.
- El motor Python gestiona la recuperación de información.
- Qdrant almacena y consulta los vectores.
- Ollama genera la respuesta final.
- Los contratos se definen mediante interfaces o puertos.

### Dependency Inversion Principle

El núcleo de la aplicación depende de abstracciones y no de implementaciones concretas.

Por ejemplo, el adaptador de embeddings implementa el contrato definido por `EmbeddingPort`, permitiendo cambiar el modelo de embeddings sin modificar la lógica principal del sistema.

### Control de concurrencia

El API Gateway utiliza un mecanismo de control de concurrencia para evitar que demasiadas consultas simultáneas saturen la memoria o el procesador.

### Cancelación y timeout

Las peticiones HTTP utilizan contextos con timeout para evitar conexiones bloqueadas y liberar recursos cuando el cliente cancela una solicitud.

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

Realiza una consulta sobre los manuales indexados.

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
  -H "Content-Type: application/json" \
  -d '{
    "question": "¿Cómo se realiza el mantenimiento del sistema de lubricación?"
  }'
```

> Si el puerto configurado en `docker-compose.yml` es diferente, reemplaza `8080` por el puerto correspondiente.

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

Ejemplos:

```bash
python rag-engine/scripts/ocr_manual.py
python rag-engine/scripts/index_manual.py
python rag-engine/scripts/upload_to_qdrant.py
```

Los nombres y parámetros exactos pueden variar según la configuración del proyecto.

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
rag-engine/mcp_server.py
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

Inicia el servicio:

```bash
uvicorn main:app --host 0.0.0.0 --port 8001 --reload
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

| Servicio | Responsabilidad |
|---|---|
| `api-go` | API Gateway y punto de entrada público |
| `rag-engine` | Recuperación semántica y API interna |
| `qdrant` | Base de datos vectorial |
| `ollama` | Generación de respuestas mediante un LLM |
| `manual-rag` | Servidor MCP para agentes |

---

## 🔐 Variables de configuración

La configuración puede variar según el entorno. Algunos valores habituales son:

```env
RAG_ENGINE_URL=http://rag-engine:8001
OLLAMA_URL=http://ollama:11434
QDRANT_URL=http://qdrant:6333
QDRANT_COLLECTION=manuals
EMBEDDING_MODEL=BAAI/bge-m3
```

No incluyas claves privadas, tokens ni credenciales directamente en el repositorio.

Para desarrollo local puedes utilizar un archivo `.env`:

```bash
cp .env.example .env
```

Si el proyecto no incluye `.env.example`, crea el archivo `.env` siguiendo las variables definidas en `docker-compose.yml`.

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

El directorio `qdrant_storage/` debe persistirse mediante un volumen para evitar perder los vectores al reiniciar los contenedores.

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

- Añadir autenticación para la API.
- Incorporar streaming de respuestas.
- Añadir métricas y observabilidad.
- Crear una interfaz web para realizar consultas.
- Añadir evaluación automática de la calidad de las respuestas.
- Incorporar filtros por documento, capítulo o sección.
- Añadir soporte para múltiples colecciones.
- Mejorar el procesamiento OCR de manuales escaneados.
- Añadir pruebas de integración con Docker Compose.
- Incorporar una cola de trabajos para la indexación de documentos.

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
