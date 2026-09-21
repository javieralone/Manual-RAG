# Scripts de automatización

Este directorio centraliza las herramientas que permiten a la IA trabajar con flujos ordenados y verificables.

## Scripts disponibles

- `create-feature-branch.sh`: crea una rama siguiendo el patrón `feature/<slug>`.
- `create-feature-branch.ps1`: versión PowerShell para entornos Windows.
- `validate-pr.sh`: ejecuta la validación mínima antes de aceptar una feature o PR.

## Convención recomendada

- Nombres de rama: `feature/<numero>-<descripcion-corta>`
- Ejemplo: `feature/01-base-observabilidad`
- Cada feature debe venir de un objetivo documentado en [features/README.md](../features/README.md)

## Validación obligatoria antes de merge

Se recomienda ejecutar:

```bash
./scripts/validate-pr.sh
```

Esto cubre validación básica de Go, Python y la configuración de Docker Compose.
