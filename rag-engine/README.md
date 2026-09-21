# 🧠 Python RAG Engine — Vector Retrieval Service

Servicio micro-componente desarrollado en **Python** con **FastAPI**, diseñado bajo **Clean Architecture** y principios **SOLID**. Actúa como el motor de recuperación vectorial (*Retrieval Engine*) dentro de la arquitectura RAG, interactuando con **Qdrant** para la búsqueda por similitud semántica y con **SentenceTransformers** para la generación de *embeddings*.

---

## 🏗️ Arquitectura y Diseño

El proyecto utiliza una arquitectura de capas concéntricas basada en puertos y adaptadores (*Hexagonal Architecture*), asegurando el desacoplamiento total de los frameworks y la base de datos vectorial mediante **Clases Base Abstractas (`abc.ABC`)**.

```text
rag-engine/
├── app/
│   ├── main.py                          # Entrypoint HTTP (FastAPI) y Composition Root
│   ├── core/                            # DOMINIO Y PUERTOS (Interfaces Abstractas)
│   │   ├── domain/
│   │   │   └── schemas.py               # DTOs y modelos de Pydantic v2
│   │   └── ports/
│   │       ├── vector_store_port.py     # Interfaz abstracta para Qdrant u otras DBs
│   │       └── embedding_port.py        # Interfaz abstracta para modelos de embeddings
│   ├── services/                        # CASOS DE USO (Lógica pura de RAG)
│   │   └── rag_service.py               # Recuperación de contexto sin dependencias
│   └── adapters/                        # ADAPTADORES / INFRAESTRUCTURA
│       ├── qdrant_adapter.py            # Implementación concreta de VectorStorePort
│       └── sentence_transformer_adapter.py # Implementación concreta de EmbeddingPort
├── requirements.txt
└── Dockerfile

🧩 Principios SOLID y Decisiones Técnicas
Single Responsibility Principle (SRP): Cada clase tiene un único propósito. El QdrantAdapter solo conoce la API del cliente de Qdrant; RAGService se encarga exclusivamente de la orquestación del retrieval; y SentenceTransformerAdapter convierte texto a vectores.

Dependency Inversion Principle (DIP): El caso de uso (RAGService) no depende de la implementación concreta de la base de datos ni del modelo de NLP. Interactúa únicamente a través de los contratos VectorStorePort y EmbeddingPort.

Gestión de Recursos (RAM & CPU Optimization):

Patrón Singleton: El modelo de embeddings (paraphrase-multilingual-MiniLM-L12-v2) y la conexión a Qdrant se instancian una sola vez al arrancar la aplicación y se inyectan a través del contenedor de dependencias de FastAPI. Esto evita re-cargar el modelo PyTorch en memoria durante cada petición HTTP.

Eficiencia en Entornos Reducidos: Seleccionado específicamente para funcionar de manera ágil bajo límites estrictos de hardware (8GB RAM / CPU local).

🔌 Endpoints HTTP
POST /search
Realiza la búsqueda de los fragmentos de texto más relevantes (top-k) dada una consulta en lenguaje natural.

Request Body:

JSON
{
  "query": "¿Cómo se realiza el mantenimiento del sistema de lubricación?",
  "top_k": 3
}
Response Body (200 OK):

JSON
{
  "results": [
    {
      "text": "El sistema de lubricación requiere cambio de aceite cada 5,000 km...",
      "score": 0.8921,
      "metadata": {
        "source": "manual_tecnico.pdf",
        "page": 12
      }
    }
  ]
}
GET /health
Verificación de estado (Health Check) para Docker y el orquestador principal.

Response Body (200 OK):

JSON
{
  "status": "UP"
}

🚀 Compilación y Ejecución
Requisitos
Docker y Docker Compose

Python 3.11+ (si deseas ejecutarlo fuera de Docker)

Ejecutar con Docker Compose
Desde la raíz del proyecto global:

Bash
docker compose up -d --build rag-engine
Ejecutar Localmente para Desarrollo
Bash
cd rag-engine
python -m venv venv
source venv/bin/activate  # En Windows: .\venv\Scripts\activate
pip install -r requirements.txt
uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
🤖 Roadmap & Integración MCP (Model Context Protocol)
Este microservicio constituye la base de la capa de conocimientos (Knowledge Layer). La siguiente fase del proyecto incluye:

Exposición como MCP Server: Encapsular los puertos de búsqueda vectorial para que funcionen como Tools estandarizadas bajo la especificación MCP (Model Context Protocol).

Capacidad de Ingesta: Extender las interfaces para permitir la indexación dinámica y troceado (chunking) de nuevos documentos en Qdrant.


<ElicitationsGroup message="¿Qué te gustaría hacer ahora con rag-engine?">
  <Elicitation label="Escribir el código completo del refactor" query="Genera el código completo de los archivos del nuevo rag-engine en Python para aplicar esta estructura."/>
  <Elicitation label="Diseñar la integración con MCP Server" query="Explícame paso a paso cómo convertir esta capa de búsqueda en un servidor MCP real."/>
</ElicitationsGroup>