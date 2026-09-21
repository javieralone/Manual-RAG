# Scripts de automatización

Este directorio centraliza las herramientas que permiten a la IA trabajar con flujos ordenados y verificables.

## Scripts disponibles

- `create-feature-branch.sh`: crea una rama siguiendo el patrón `feature/<slug>`.
- `create-feature-branch.ps1`: versión PowerShell para entornos Windows.
- `validate-pr.sh`: ejecuta la validación mínima antes de aceptar una feature o PR.
- `orchestrate-feature.sh`: orquesta una feature completa: crea/selecciona rama, valida, y opcionalmente hace push.
- `orchestrate-feature.ps1`: versión PowerShell del orquestador.

## Convención recomendada

- Nombres de rama: `feature/<numero>-<descripcion-corta>`
- Ejemplo: `feature/01-base-observabilidad`
- Cada feature debe venir de un objetivo documentado en [features/README.md](../features/README.md)

## Flujo de ejecución

```bash
./scripts/orchestrate-feature.sh --feature 01-base-observabilidad --push
```

Esto hace lo siguiente:

1. crea o selecciona la rama `feature/01-base-observabilidad`
2. ejecuta la validación general del proyecto
3. si la validación pasa, hace commit y push
4. no avanza a la siguiente feature si la validación falla

## Validación obligatoria antes de merge

Se recomienda ejecutar:

```bash
./scripts/validate-pr.sh
```

Esto cubre validación básica de Go, Python y la configuración de Docker Compose.
