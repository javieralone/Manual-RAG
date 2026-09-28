# Frontend Manual-RAG

## Descripción General

La interfaz web del proyecto está implementada con **React + Vite** y actúa como la interfaz de usuario (MVP) para interactuar con el backend RAG. Ofrece un flujo completo de consulta y visualización de contextos recuperados mediante autenticación real.

### Funcionalidades del MVP
- **Autenticación real:** Inicio de sesión e integración contra el gateway Go (`api-go`).
- **Consultas flexibles:** Selector de colecciones a consultar y alternancia entre respuesta normal o streaming en tiempo real.
- **Interfaz interactiva:** Chat simple para consultas y panel lateral dedicado a la inspección del contexto RAG recuperado.
- **Procesamiento SSE:** Renderizado incremental de respuestas por streaming con control frontend para evitar solapamientos de tokens.

> **Nota de alcance:** Esta versión MVP no incluye aún historial persistente de conversaciones ni panel de administración avanzado.

---

## Requisitos Previos

- **Node.js y npm** (para ejecuciones locales)
- **Docker y Docker Compose** (para entornos en contenedores)
- Acceso al servicio `api-go` ejecutándose en `http://localhost:8080`
- Credenciales válidas (usuario y contraseña) configuradas en el gateway

---

## Estructura del Proyecto

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

---

## Configuración y Variables de Entorno

El proyecto utiliza variables de entorno para configurar la URL del gateway. Consulta `.env.example` para la referencia de variables.

```env
VITE_API_BASE_URL=http://localhost:8080
```

---

## Guía de Ejecución

### 1. Desarrollo Local
```bash
cd apps/web
npm install
npm run dev
```
La aplicación quedará disponible en `http://localhost:5173`.

### 2. Despliegue con Docker Compose
Desde la raíz del proyecto y habiendo preparado el archivo `.env`:

```bash
docker compose --env-file .env up -d --build frontend
```
El contenedor compila la aplicación mediante Vite y sirve los estáticos optimizados a través de **Nginx** en el puerto interno `8080`, publicado hacia el exterior en `http://localhost:5173`.

> ⚠️ **Consideración de Seguridad para Producción:**
> El MVP actual almacena los tokens JWT en el almacenamiento del navegador. Antes de pasar a producción, se debe migrar el *refresh token* a una cookie con atributos `HttpOnly`, `Secure` y `SameSite`, tal como se documenta en [production-readiness.md](../../docs/operations/production-readiness.md).

---

## Integración de API

El frontend interactúa con los siguientes endpoints expuestos por el gateway:

### Autenticación (`POST /api/v1/auth/login`)
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "tu-password"
}
```

### Consulta Normal (`POST /api/v1/query`)
```http
POST /api/v1/query
Authorization: Bearer <token>
Content-Type: application/json

{
  "question": "¿Cómo se realiza el mantenimiento del sistema de lubricación?",
  "collection": "manuales_tecnicos"
}
```

### Consulta con Streaming (`POST /api/v1/query/stream`)
```http
POST /api/v1/query/stream
Authorization: Bearer <token>
Content-Type: application/json

{
  "question": "¿Cómo se realiza del mantenimiento del sistema de lubricación?",
  "collection": "manuales_tecnicos"
}
```

*Nota técnica sobre streaming:* Las respuestas en tiempo real se consumen mediante **Server-Sent Events (SSE)** del gateway. El frontend procesa estos eventos en tiempo real e implementa una lógica de desduplicación para prevenir la repetición o el solapamiento visual de tokens durante el renderizado.
