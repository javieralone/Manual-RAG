# Production Readiness

## Estado

El sistema queda preparado para una puesta en marcha controlada, pero no debe exponerse directamente a Internet sin un reverse proxy, TLS, gestión de secretos y escaneo de imágenes en CI.

## Implementado

- Healthchecks para Qdrant, retrieval, MCP, gateway y frontend.
- Readiness del gateway contra `rag-engine/ready`.
- Graceful shutdown del gateway con `SHUTDOWN_TIMEOUT`.
- Cierre del proveedor OTLP durante el shutdown de FastAPI.
- Timeouts externalizados para HTTP, Qdrant y readiness.
- Retries limitados con backoff para Qdrant, retrieval y Ollama.
- Circuit breakers por dependencia en el gateway.
- Imágenes Go y Python con versiones base explícitas y usuarios no root.
- Frontend compilado como imagen de producción con Nginx no privilegiado.
- Contraseñas y secretos críticos obligatorios en Compose.
- Variables de resiliencia documentadas en `.env.example`.
- Errores de streaming SSE propagados como evento `error`.

## Configuración mínima

Crear un archivo `.env` fuera del repositorio y reemplazar todos los placeholders de `.env.example`. Validar la configuración con:

```powershell
docker compose --env-file .env config --quiet
```

Para construir o iniciar servicios usando esa configuración:

```powershell
docker compose --env-file .env build
docker compose --env-file .env up -d
```

No ejecutar `docker compose build` sin `--env-file` cuando se utilicen variables obligatorias de producción.

## Riesgos pendientes

Estos puntos deben cerrarse antes de declarar el sistema listo para producción. La lista distingue controles obligatorios de mejoras operativas posteriores.

### Alta prioridad

| Pendiente | Criterio de cierre | Validación |
| --- | --- | --- |
| Red y exposición | Solo frontend/gateway quedan publicados; Qdrant, MCP y observabilidad quedan en red interna o detrás de proxy autenticado. | Revisar `docker compose ps`, reglas del proxy y escaneo de puertos. |
| Secretos | `.env` o Docker/Kubernetes Secrets gestionados fuera de Git, sin placeholders, con rotación definida. | `docker compose --env-file .env config --quiet` y revisión de secretos. |
| Imágenes | Todas las imágenes de aplicación y dependencias base tienen tag y digest aprobados. | `docker image inspect` y política de CI. |
| Vulnerabilidades | Escaneo de imágenes y dependencias sin vulnerabilidades `high`/`critical` aceptadas. | Trivy/Docker Scout, `npm audit` y `govulncheck`. |
| Dependencias Python | Lockfile reproducible con versiones y hashes; instalación verificada desde cero. | Build limpio de `services/retrieval-python`. |
| Tokens web | Refresh token fuera de `localStorage`, en cookie `HttpOnly`, `Secure` y `SameSite`, con protección CSRF. | Prueba de login, refresh, logout y revisión de cookies. |
| Rate limiting | Límite aplicado en proxy o almacenamiento distribuido cuando existan varias réplicas. | Prueba con dos réplicas y métrica de rechazos. |

### Prioridad media

| Pendiente | Criterio de cierre | Validación |
| --- | --- | --- |
| Métricas de resiliencia | Existen métricas de retries, circuitos abiertos y recuperaciones, con labels de baja cardinalidad. | Prometheus y dashboard operativo. |
| Alertas operativas | Alertas para reinicios, memoria, CPU, readiness, worker pool y errores de dependencias. | Prueba controlada de cada alerta. |
| Protección HTTP | Límite de tamaño de body, CSP completa y headers de seguridad revisados. | Tests HTTP y escaneo del frontend. |
| Retención y backups | Política documentada y backup probado para Qdrant, Loki, Tempo, Prometheus y Grafana. | Restauración en entorno aislado. |
| Healthcheck MCP | Healthcheck basado en endpoint de protocolo o endpoint dedicado, no solo conexión TCP. | Smoke test MCP automatizado. |

### Orden recomendado

1. Cerrar secretos, red, TLS y exposición de puertos.
2. Fijar imágenes/dependencias y ejecutar escaneos de vulnerabilidades.
3. Migrar el refresh token y completar headers/CSP.
4. Validar escalado horizontal y rate limiting distribuido.
5. Completar métricas, alertas, retención y restauración.

## Validación realizada

- `docker compose --env-file .env.example config --quiet`
- `go test -count=1 ./...`
- `go build -a ./...`
- `python -m compileall src app scripts tests`
- `npm run build` en `apps/web`

La ejecución local de tests Python requiere instalar `services/retrieval-python/requirements.txt`; el intérprete utilizado durante la revisión no tenía `pydantic`, `prometheus_client` ni `qdrant-client`.

## Checklist antes de producción

- [ ] Secretos generados fuera del repositorio y rotación definida.
- [ ] TLS y reverse proxy configurados.
- [ ] Puertos internos no publicados fuera de la red administrativa.
- [ ] Imágenes escaneadas y fijadas por digest.
- [ ] Dependencias auditadas y lockfiles reproducibles.
- [ ] Backups y retención verificados.
- [ ] Alertas probadas contra Prometheus.
- [ ] Smoke test de `/ready`, `/search`, `/api/v1/query`, streaming y MCP.
- [ ] Prueba de apagado durante una petición y durante un stream SSE.

## Bloqueo de salida

No promover a producción si falla cualquiera de estos controles: secretos gestionados fuera de Git, TLS/proxy, escaneo de vulnerabilidades, backup restaurable, readiness de dependencias o smoke test de consulta y streaming.
