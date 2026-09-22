# Scripts de automatización

Este directorio centraliza las herramientas que permiten a la IA trabajar con flujos ordenados y verificables.

## Scripts disponibles

- `create-feature-branch.sh`: crea una rama siguiendo el patrón `feature/<slug>`.
- `create-feature-branch.ps1`: versión PowerShell para entornos Windows.
- `validate-pr.sh`: ejecuta la validación mínima antes de aceptar una feature o PR.
- `orchestrate-feature.sh`: orquesta una feature completa: crea/selecciona rama, valida, y opcionalmente hace push.
- `orchestrate-feature.ps1`: versión PowerShell del orquestador.
- `run-feature.ps1`: dispara el workflow de GitHub Actions para generar el plan de una feature.
- `start-feature.ps1`: genera el plan y dispara la implementación en un único comando.
- `resolve-feature.py`: resuelve una feature del roadmap y genera un plan inicial reproducible.

## Convención recomendada

- Nombres de rama: `feature/<numero>-<descripcion-corta>`
- Ejemplo: `feature/04-mejora-ocr`
- Cada feature pendiente debe venir de un objetivo documentado en [docs/roadmap/README.md](../docs/roadmap/README.md)

## Flujo de ejecución

```bash
./scripts/orchestrate-feature.sh --feature 04-mejora-ocr --push
```

Esto hace lo siguiente:

1. crea o selecciona la rama `feature/04-mejora-ocr`
2. ejecuta la validación general del proyecto
3. si la validación pasa y se proporciona `--push`, hace commit y push
4. no avanza a la siguiente feature si la validación falla

## Generar un plan desde GitHub Actions

Primero crea manualmente un Issue en GitHub con el texto completo de la feature. Luego, con GitHub CLI autenticado, desde PowerShell:

```powershell
.\scripts\start-feature.ps1 -Issue 123 -Wait
```

El comando:

1. envía el número del Issue a GitHub Actions,
2. descarga su título y cuerpo,
3. genera `implementation-plan.md`,
4. publica el plan como comentario en el mismo Issue,
5. agrega la etiqueta `copilot-implementation`.

Para consultar el artifact:

```powershell
gh run list --workflow feature-plan.yml --limit 1
gh run download <run-id> -n feature-123-plan
```

Para que Copilot implemente el Issue, después de que aparezca el comentario del plan, abre el Issue en GitHub y selecciona **Start task** o la acción equivalente de Copilot Cloud Agent. También puedes configurar una Automation compatible con tu cuenta para Issues que tengan esta etiqueta:

- trigger: el evento de Issue soportado por tu configuración,
- filtro: `label:copilot-implementation`,
- herramientas: modificar archivos, ejecutar tests, crear branch y Pull Request,
- instrucción: implementar la Issue, ejecutar validaciones y abrir el PR contra `dev`.

Copilot toma esa Issue, implementa en una rama y abre el PR. El merge sigue requiriendo revisión humana y los checks de CI.

## Validación obligatoria antes de merge

Se recomienda ejecutar:

```bash
./scripts/validate-pr.sh
```

Esto cubre validación básica de Go, Python y la configuración de Docker Compose.

En Windows, `orchestrate-feature.ps1` usa `validate-pr.sh`, por lo que requiere Git Bash disponible en el `PATH`. El script PowerShell no sustituye ese validador Bash.
