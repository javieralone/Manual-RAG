# 🚀 Go API Gateway — RAG System

Un API Gateway resiliente y concurrente desarrollado en **Go**, diseñado bajo los principios de **Clean Architecture** y **SOLID**. Actúa como el punto de entrada orquestador entre los clientes externos, el motor de búsqueda vectorial (**rag-engine** en Python) y el LLM (**Ollama**).

---

## 🏗️ Arquitectura y Diseño

El proyecto sigue una arquitectura hexagonal o **Clean Architecture** dividida en capas concéntricas, garantizando la inversión de dependencias y el desacoplamiento total de frameworks o infraestructura externa.

```text
api-go/
├── cmd/
│   └── api/
│       └── main.go                  # Composition Root & Inyección de Dependencias
└── internal/
    ├── core/                        # NÚCLEO DEL NEGOCIO (Sin dependencias externas)
    │   ├── domain/                  # Entidades puras y errores de negocio
    │   │   └── query.go
    │   ├── ports/                   # Contratos / Interfaces (DIP)
    │   │   ├── rag_port.go
    │   │   ├── llm_port.go
    │   │   └── query_service.go
    │   └── services/                # Casos de uso / Orquestación
    │       └── query_orchestrator.go
    └── adapters/                    # INFRAESTRUCTURA Y DETALLES
        ├── http/                    # Enrutamiento, Handlers y Middlewares
        │   ├── handlers/
        │   ├── middlewares/
        │   └── router.go
        ├── clients/                 # Adaptadores de salida (Python / Ollama)
        └── decorators/              # Decoradores para concurrencia (Worker Pool)
```

## 🧩 Principios SOLID y Patrones Aplicados

### Single Responsibility Principle (SRP)

Cada paquete tiene una responsabilidad acotada:

- `QueryHandler` gestiona el protocolo HTTP.
- `WorkerPoolUseCaseDecorator` maneja los límites de recursos.
- `QueryOrchestrator` contiene la lógica de negocio y orquestación de consultas.

### Dependency Inversion Principle (DIP)

El núcleo (`core`) no depende de implementaciones concretas. Define contratos en `ports/` que son implementados por la capa `adapters/`.

### Open/Closed Principle (OCP) y Patrón Decorator

La limitación de concurrencia se implementa envolviendo el caso de uso principal mediante `WorkerPoolUseCaseDecorator`, sin modificar el código del orquestador.

### Graceful Degradation & Protection

Diseñado para ejecutarse de forma segura en entornos con recursos limitados de CPU y memoria.

---

## ⚡ Concurrencia y Control de Recursos

### Worker Pool (Semáforo mediante canales)

Limita las peticiones concurrentes utilizando un canal con búfer para evitar sobrecargas de RAM y CPU causadas por consultas simultáneas hacia Ollama.

### Context Cancellation & Timeout Middleware

Toda petición HTTP propaga un `context.Context` con un límite configurable de **5 minutos** por defecto.

Si el cliente cancela la solicitud o expira el tiempo establecido:

- Los sockets HTTP se cierran inmediatamente.
- Se interrumpen las operaciones pendientes.
- Se evitan fugas de memoria (*goroutine leaks*).

### Reutilización de HTTP Client

Instancia única de `http.Client` con soporte para:

- Keep-Alive.
- Reutilización de conexiones TCP.
- Menor latencia y consumo de recursos.

---

## 🔌 Endpoints

### POST `/api/v1/query`

Procesa la pregunta del usuario, recupera contexto desde `rag-engine` y genera una respuesta utilizando Ollama.

#### Request Body

```json
{
  "question": "¿Cómo se realiza el mantenimiento del sistema de lubricación?"
}
```

#### Response Body (200 OK)

```json
{
  "question": "¿Cómo se realiza el mantenimiento del sistema de lubricación?",
  "answer": "El mantenimiento requiere revisar los niveles de aceite...",
  "context": [
    {
      "text": "Fragmento de texto recuperado del manual...",
      "score": 0.89,
      "metadata": {}
    }
  ]
}
```

---

### GET `/health`

Endpoint de verificación de estado destinado a Docker y orquestadores.

#### Response Body (200 OK)

```json
{
  "status": "UP"
}
```

### GET `/ready`

Comprueba que `rag-engine` y Ollama están disponibles para atender consultas. Devuelve `503` si alguna dependencia no está lista.

### GET `/metrics`

Expone métricas Prometheus del gateway: peticiones, latencias, errores, autenticación, dependencias, generación y concurrencia.

## Autenticacion y autorizacion

El API Gateway centraliza la autenticacion de usuario. `rag-engine` no recibe ni valida credenciales finales.

### Variables requeridas

Configura estas variables en el entorno del contenedor `api-go` o en un archivo `.env` local que no se versiona:

| Variable | Descripcion |
|---|---|
| `AUTH_JWT_SECRET` | Secreto HS256 del access token, minimo 32 caracteres |
| `AUTH_REFRESH_SECRET` | Secreto HS256 separado para refresh tokens, minimo 32 caracteres |
| `AUTH_ADMIN_USERNAME` | Usuario inicial configurado en el gateway |
| `AUTH_ADMIN_PASSWORD_HASH` | Hash bcrypt del password del usuario inicial |
| `AUTH_ADMIN_ROLES` | Roles separados por coma: `admin`, `operator`, `user` |

Opcionales: `AUTH_ISSUER`, `AUTH_AUDIENCE`, `AUTH_ACCESS_TTL` (por defecto `15m`), `AUTH_REFRESH_TTL` (por defecto `168h`), `OLLAMA_MODEL`, `WORKER_LIMIT`, `HTTP_CLIENT_TIMEOUT` y `REQUEST_TIMEOUT` (por defecto `5m`) y `READINESS_INTERVAL` (por defecto `15s`).

Genera el hash con una herramienta bcrypt confiable, por ejemplo `htpasswd -bnBC 12 "" "tu-password"` y conserva solo el valor despues de los dos puntos. Nunca guardes el password ni los secretos en Git.

### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{"username":"admin","password":"tu-password"}
```

La respuesta contiene `access_token`, `refresh_token`, `token_type` y sus fechas de expiracion.

### Refresh

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{"refresh_token":"..."}
```

Los refresh tokens son JWT stateless y expiran; no existe revocacion persistente hasta incorporar un almacen de sesiones o una lista de revocacion.

### Consulta protegida

```http
POST /api/v1/query
Authorization: Bearer <access_token>
Content-Type: application/json
```

Los roles `admin`, `operator` y `user` pueden consultar. La ausencia de token devuelve `401`; un token valido sin rol permitido devuelve `403`.

## Observabilidad

El gateway genera logs JSON y trazas OpenTelemetry OTLP. El trace context W3C se propaga hacia `rag-engine` y Ollama. El stack completo está documentado en [`observability/README.md`](../observability/README.md).

## Compilacion y ejecucion

Requisitos: Docker y Docker Compose.

### Ejecutar con Docker Compose

Desde la raíz del proyecto:

```bash
docker compose up -d --build api-go
```

### Probar compilación local

```bash
cd api-go

go mod tidy
go build -v ./...
go test ./...
```
