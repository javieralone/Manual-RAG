# Frontend Manual-RAG

## Descripción

La interfaz web del proyecto está implementada con React + Vite y está pensada como un MVP funcional para:

- iniciar sesión contra el gateway Go,
- elegir la colección a consultar,
- seleccionar modo normal o streaming,
- enviar preguntas en un chat simple,
- revisar el contexto recuperado por RAG.

## Objetivo del MVP

Esta UI cubre el flujo mínimo útil de la feature 7:

- autenticación real con `api-go`,
- consulta al sistema a través del gateway,
- visualización de la respuesta y del contexto,
- soporte para modo normal y modo streaming.

No incluye aún historial persistente ni panel administrativo avanzado.

## Requisitos

- Docker y Docker Compose
- Acceso al servicio `api-go` en `http://localhost:8080`
- Usuario y password válidos configurados en el gateway

## Estructura

```text
apps/web/
├── src/
│   ├── api/
│   │   ├── auth.js
│   │   ├── chat.js
│   │   └── client.js
│   ├── App.jsx
│   ├── index.css
│   └── main.jsx
├── Dockerfile
├── nginx.conf
├── package.json
├── vite.config.js
├── .env.example
├── .dockerignore
├── index.html
└── README.md
```

## Variables de entorno

Se usa la variable:

```env
VITE_API_BASE_URL=http://localhost:8080
```

Contenido de ejemplo en `.env.example`.

## Ejecutar localmente

```bash
cd apps/web
npm install
npm run dev
```

La app queda disponible normalmente en:

```text
http://localhost:5173
```

## Ejecutar con Docker Compose

Desde la raíz del proyecto, con un archivo `.env` configurado:

```bash
docker compose --env-file .env up -d --build frontend
```

La imagen de Compose compila Vite y sirve los archivos estáticos con Nginx en el puerto interno `8080`, publicado como `http://localhost:5173`.

> El MVP actual conserva tokens en almacenamiento del navegador. Antes de exponerlo a Internet, migrar el refresh token a una cookie `HttpOnly`, `Secure` y `SameSite`, tal como se documenta en [production-readiness.md](../../docs/operations/production-readiness.md).

## API consumida

### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{"username":"admin","password":"tu-password"}
```

### Consulta normal

```http
POST /api/v1/query
Authorization: Bearer <token>
Content-Type: application/json

{
  "question": "¿Cómo se realiza el mantenimiento del sistema de lubricación?",
  "collection": "manuales_tecnicos"
}
```

### Consulta streaming

```http
POST /api/v1/query/stream
Authorization: Bearer <token>
Content-Type: application/json

{
  "question": "¿Cómo se realiza el mantenimiento del sistema de lubricación?",
  "collection": "manuales_tecnicos"
}
```

## Nota sobre streaming

El stream usa eventos SSE del gateway. La UI los procesa y los convierte en respuesta incremental en el chat. La corrección de solapamiento de tokens se gestiona en el frontend para evitar duplicaciones como palabras repetidas durante la renderización incremental.

## Estado actual

El MVP ya incluye:

- login funcional con autenticación real del gateway,
- chat con respuesta incremental,
- selector de colección,
- selector de consulta sin stream / con stream,
- panel lateral de contexto recuperado,
- componente Docker listo para despliegue.
