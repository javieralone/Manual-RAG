# Production Readiness

## Estado

El sistema queda preparado para una puesta en marcha controlada. El overlay de producción incorpora reverse proxy, TLS y redes internas, pero no debe exponerse a Internet hasta validar certificados en un dominio real y cerrar secretos, CSRF y escaneo en CI.

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
- Overlay `docker-compose.prod.yml` con Nginx como único listener público, Certbot y redes `backend`/`admin` internas.
- Redis para sesiones de refresh con TTL, rotación atómica, revocación y readiness del gateway.
- Cookie de refresh `HttpOnly`, `Secure` y `SameSite=Lax`; el frontend no persiste refresh tokens.
- Límite de body configurable y respuestas JSON de error comunes en el gateway.

## Configuración mínima

Crear un archivo `.env` fuera del repositorio y reemplazar todos los placeholders de `.env.example`. Validar la configuración con:

```powershell
docker compose --env-file .env -f docker-compose.yml -f docker-compose.prod.yml config --quiet
```

Para construir o iniciar servicios usando esa configuración:

```powershell
docker compose --env-file .env -f docker-compose.yml -f docker-compose.prod.yml build
docker compose --env-file .env -f docker-compose.yml -f docker-compose.prod.yml up -d
```

No ejecutar `docker compose build` sin `--env-file` cuando se utilicen variables obligatorias de producción.

## Riesgos pendientes

Estos puntos deben cerrarse antes de declarar el sistema listo para producción. La lista distingue controles obligatorios de mejoras operativas posteriores.

### Alta prioridad

| Pendiente | Criterio de cierre | Validación |
| --- | --- | --- |
| TLS de borde | DNS, puertos `80/443` y emisión/renovación de Certbot funcionan contra el dominio configurado. | Navegación HTTPS, `openssl s_client` y logs de Certbot. |
| Secretos | Docker/Kubernetes Secrets o gestor externo, sin placeholders, con rotación definida. | Revisión de despliegue y rotación de prueba. |
| Imágenes | Todas las imágenes de aplicación y dependencias base tienen tag y digest aprobados. | `docker image inspect` y política de CI. |
| Vulnerabilidades | Escaneo de imágenes y dependencias sin vulnerabilidades `high`/`critical` aceptadas. | Trivy/Docker Scout, `npm audit` y `govulncheck`. |
| Dependencias Python | Lockfile reproducible con versiones y hashes; instalación verificada desde cero. | Build limpio de `services/retrieval-python`. |
| CSRF y CORS | El refresh cookie tiene protección CSRF explícita y CORS usa una allowlist configurada. | Pruebas cross-origin negativas y de login/refresh/logout. |
| Rate limiting | Límite aplicado en proxy o almacenamiento distribuido cuando existan varias réplicas. | Prueba con dos réplicas y métrica de rechazos. |

### Prioridad media

| Pendiente | Criterio de cierre | Validación |
| --- | --- | --- |
| Métricas de resiliencia | Existen métricas de retries, circuitos abiertos y recuperaciones, con labels de baja cardinalidad. | Prometheus y dashboard operativo. |
| Alertas operativas | Alertas para reinicios, memoria, CPU, readiness, worker pool y errores de dependencias. | Prueba controlada de cada alerta. |
| Protección del servicio interno | FastAPI aplica límite de body y sobre de error equivalente al gateway. | Pruebas HTTP de `413` y errores de validación. |
| Retención y backups | Política documentada y backup probado para Qdrant, Loki, Tempo, Prometheus y Grafana. | Restauración en entorno aislado. |
| Healthcheck MCP | Healthcheck basado en endpoint de protocolo o endpoint dedicado, no solo conexión TCP. | Smoke test MCP automatizado. |

### Orden recomendado

1. Validar TLS en dominio real y cerrar CSRF/CORS.
2. Migrar secretos a un gestor externo y definir rotación de JWT.
3. Fijar imágenes/dependencias y ejecutar escaneos de vulnerabilidades.
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

- [ ] Secretos gestionados por un mecanismo externo y rotación definida.
- [ ] TLS emitido y renovado para el dominio público.
- [ ] Despliegue ejecutado con el overlay de producción y puertos internos no publicados.
- [ ] Imágenes escaneadas y fijadas por digest.
- [ ] Dependencias auditadas y lockfiles reproducibles.
- [ ] CSRF y CORS restrictivo validados para cookies de refresh.
- [ ] Backups y retención verificados.
- [ ] Alertas probadas contra Prometheus.
- [ ] Smoke test de `/ready`, `/search`, `/api/v1/query`, streaming y MCP.
- [ ] Prueba de apagado durante una petición y durante un stream SSE.

## Bloqueo de salida

No promover a producción si falla cualquiera de estos controles: secretos gestionados externamente, TLS emitido, CSRF/CORS, escaneo de vulnerabilidades, backup restaurable, readiness de dependencias o smoke test de consulta y streaming.
