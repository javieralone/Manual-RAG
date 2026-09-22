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

Con GitHub CLI autenticado, desde PowerShell:

```powershell
.\scripts\run-feature.ps1 -Feature 06 -Mode plan
gh run list --workflow feature-plan.yml --limit 1
gh run download <run-id> -n feature-06-plan
```

El workflow lee `docs/roadmap`, genera `implementation-plan.md` y lo publica como artifact.

Para iniciar la implementación:

```powershell
.\scripts\start-feature.ps1 -Feature 06
```

El script anterior ejecuta el workflow en modo `implement`: primero genera el plan y luego crea la Issue que dispara la Automation de Copilot. Para esperar el resultado del workflow:

```powershell
.\scripts\start-feature.ps1 -Feature 06 -Wait
```

En este modo, GitHub Actions crea una Issue con la feature y el plan, etiquetada como `copilot-implementation`. Debes configurar una Automation de Copilot Cloud Agent en GitHub con:

- trigger: issue creada,
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
