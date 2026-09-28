# Frontend Manual-RAG

Interfaz web para **Manual-RAG**, desarrollada con **React + Vite** y servida con **Nginx** en producción[cite: 1]. Permite interactuar con el sistema de consulta RAG mediante una interfaz de chat interactiva[cite: 1].

---

## 🚀 Funcionalidades

- **Autenticación real:** Inicio de sesión contra la API Gateway Go (`api-go`)[cite: 1].
- **Chat interactivo:** Envío de preguntas sobre los manuales indexados[cite: 1].
- **Soporte Streaming (SSE):** Elección entre respuesta estándar o transmisión token por token en tiempo real[cite: 1].
- **Selector de colección:** Permite seleccionar la base de conocimiento/colección a consultar en Qdrant[cite: 1].
- **Panel de contexto:** Visualización detallada del contenido y metadatos recuperados por el RAG para construir la respuesta[cite: 1].

---

## 🏗️ Estructura del proyecto

```text
apps/web/
├── src/
│   ├── api/          # Clientes HTTP y conexión con la API (auth, chat)
│   │   ├── auth.js
│   │   ├── chat.js
│   │   └── client.js
│   ├── App.jsx       # Componente principal de la interfaz
│   ├── index.css     # Estilos globales
│   └── main.jsx      # Punto de entrada de React
├── Dockerfile        # Build multi-stage (Vite + Nginx)
├── nginx.conf        # Configuración de Nginx para producción
├── package.json
├── vite.config.js
└── .env.example
```[cite: 1]

---

## ⚙️ Configuración y requisitos

### Requisitos previos
- Node.js (v18+) o Docker[cite: 1]
- Servicio `api-go` ejecutándose (por defecto en `http://localhost:8080`)[cite: 1]

### Variables de entorno
Crea un archivo `.env` basado en `.env.example`[cite: 1]:

```env
VITE_API_BASE_URL=http://localhost:8080
```[cite: 1]

---

## 🛠️ Ejecución

### Desarrollo local (Node.js)

```bash
cd apps/web
npm install
npm run dev
```[cite: 1]

Accede desde el navegador a: `http://localhost:5173`[cite: 1]

### Despliegue con Docker Compose

Para ejecutar el frontend containerizado junto al stack:

```bash
docker compose --env-file .env up -d --build frontend
```[cite: 1]

La imagen compila los estáticos con Vite y los sirve con Nginx en el puerto público `5173`[cite: 1].

---

## 🔌 Endpoints consumidos

El frontend interactúa con los siguientes endpoints del Gateway:

| Endpoint | Método | Descripción |
| :--- | :---: | :--- |
| `/api/v1/auth/login` | `POST` | Autenticación de usuario y obtención de token JWT[cite: 1]. |
| `/api/v1/query` | `POST` | Consulta RAG estándar (espera la respuesta completa)[cite: 1]. |
| `/api/v1/query/stream` | `POST` | Consulta RAG con **Server-Sent Events (SSE)** para streaming incremental[cite: 1]. |

> **Nota sobre el Streaming:** Los eventos SSE son procesados en el cliente para renderizar la respuesta de forma progresiva. El frontend gestiona la deduplicación de tokens para evitar palabras repetidas durante el renderizado en tiempo real[cite: 1].

---

## 🔒 Consideraciones de seguridad

⚠️ **Aviso de seguridad para producción:** En la versión actual los tokens se almacenan en la memoria local del navegador[cite: 1]. Antes de desplegar en entornos de producción expuestos a Internet, se debe migrar la gestión del *refresh token* a cookies `HttpOnly`, `Secure` y `SameSite` (consulta la guía [production-readiness.md](../../docs/operations/production-readiness.md))[cite: 1].
