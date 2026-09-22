# Proxy de producción

El despliegue público usa Nginx y Certbot con:

- Sólo los puertos `80` y `443` publicados por `proxy`.
- Servicios de aplicación en la red Docker interna `backend`.
- Observabilidad aislada en la red interna `admin`.
- Refresh token en cookie `HttpOnly`, `Secure` y `SameSite=Lax`.

Configure `DOMAIN`, `CERTBOT_EMAIL` y secretos en un gestor de secretos o en el entorno del host. No los copie al repositorio.

Inicie producción con:

```powershell
docker compose --env-file .env.local -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

El DNS de `DOMAIN` debe apuntar al host y los puertos `80` y `443` deben ser accesibles antes de iniciar Certbot. Nginx inicia con un certificado temporal de un día; Certbot lo reemplaza y Nginx recarga certificados como máximo en cinco minutos.