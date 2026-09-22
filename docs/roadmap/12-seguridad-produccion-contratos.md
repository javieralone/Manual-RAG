# 12 - Seguridad de producción y contratos

## Prioridad

P0/P1 - bloqueador para exposición a Internet.

## Estado de base implementado

- `docker-compose.prod.yml` publica únicamente Nginx en `80/443`; backend, Redis y observabilidad se separan en redes internas.
- Nginx usa Certbot para certificados y aplica HSTS, CSP, `Permissions-Policy` y proxy sin buffering para SSE.
- El gateway limita el tamaño de body y usa respuestas JSON de error comunes.
- Redis almacena sesiones de refresh con TTL, rotación de uso único y revocación por logout.
- El refresh token se transporta en cookie `HttpOnly`, `Secure` y `SameSite=Lax`.

## Pendientes

### P0: despliegue y cadena de suministro

- Validar DNS, puertos públicos, emisión inicial y renovación de Certbot con un dominio real.
- Migrar secretos a Docker Secrets o un gestor externo; documentar y probar rotación dual de claves JWT.
- Fijar imágenes por digest.
- Generar lockfile Python con hashes.
- Añadir a CI `govulncheck`, `pip-audit`, auditoría npm, SBOM y escaneo de imágenes con fallo para severidades aceptadas como bloqueantes.

### P1: protección de sesión y contratos

- Añadir protección CSRF explícita a refresh y logout, con pruebas cross-origin negativas.
- Sustituir el fallback CORS permisivo por una allowlist configurable.
- Publicar y versionar el OpenAPI de retrieval.
- Añadir pruebas contractuales de `/search` entre el proveedor FastAPI y el cliente Go para éxitos, validaciones y errores.
- Aplicar un límite de body y sobre de error equivalente en FastAPI.

## Criterio de cierre

- El despliegue productivo emite y renueva un certificado válido; un escaneo externo confirma que sólo `80/443` están accesibles.
- Secretos y rotación se prueban sin valores versionados en Git.
- CI bloquea vulnerabilidades por la política acordada y usa dependencias reproducibles.
- Las rutas de cookie rechazan peticiones cross-origin sin evidencia CSRF válida.
- Un cambio incompatible en el contrato FastAPI-Go falla en CI.