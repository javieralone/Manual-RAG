#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

AUTO_PUSH=false
TARGET_FEATURE=""

usage() {
  cat <<'EOF'
Uso:
  ./scripts/orchestrate-feature.sh [opciones]

Opciones:
  -f, --feature <nombre>   Feature a ejecutar (ej: 01-base-observabilidad)
  -p, --push               Hace push a la rama si la validación pasa
  -h, --help               Muestra esta ayuda

Ejemplos:
  ./scripts/orchestrate-feature.sh --feature 01-base-observabilidad
  ./scripts/orchestrate-feature.sh --feature 02-evaluacion-calidad-rag --push
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -f|--feature)
      TARGET_FEATURE="$2"
      shift 2
      ;;
    -p|--push)
      AUTO_PUSH=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Opción no reconocida: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ -n "${TARGET_FEATURE}" ]]; then
  FEATURE_PATTERN="${TARGET_FEATURE}"
else
  FEATURE_PATTERN="*"
fi

FEATURES=( $(printf '%s\n' "$(find docs/roadmap -maxdepth 1 -type f -name '*.md' ! -name 'README.md' | sort)" | sed 's#^docs/roadmap/##; s#\.md$##' | head -n 100) )

if [[ ${#FEATURES[@]} -eq 0 ]]; then
  echo "No se encontraron features en docs/roadmap/*.md" >&2
  exit 1
fi

if [[ -n "$TARGET_FEATURE" ]]; then
  FOUND=false
  for feature in "${FEATURES[@]}"; do
    if [[ "$feature" == "$TARGET_FEATURE" ]]; then
      FOUND=true
      break
    fi
  done
  if [[ "$FOUND" != true ]]; then
    echo "La feature '$TARGET_FEATURE' no existe." >&2
    exit 1
  fi
fi

if [[ -n "$TARGET_FEATURE" ]]; then
  SELECTED_FEATURE="$TARGET_FEATURE"
else
  SELECTED_FEATURE="${FEATURES[0]}"
fi

CURRENT_BRANCH="$(git branch --show-current 2>/dev/null || true)"
if [[ -z "$CURRENT_BRANCH" ]]; then
  echo "No hay una rama activa. Creando la rama de la feature..."
  git checkout -b "feature/$SELECTED_FEATURE"
else
  if [[ "$CURRENT_BRANCH" != "feature/$SELECTED_FEATURE" && "$CURRENT_BRANCH" != "feature/"* ]]; then
    echo "La rama actual es '$CURRENT_BRANCH'. Se recomienda trabajar sobre una rama feature."
  fi
fi

if [[ "$CURRENT_BRANCH" != "feature/$SELECTED_FEATURE" ]]; then
  if git rev-parse --verify "feature/$SELECTED_FEATURE" >/dev/null 2>&1; then
    git checkout "feature/$SELECTED_FEATURE"
  else
    git checkout -b "feature/$SELECTED_FEATURE"
  fi
fi

echo "Ejecutando feature: $SELECTED_FEATURE"

if ! ./scripts/validate-pr.sh; then
  echo "La validación falló. La feature no se marca como completada." >&2
  exit 1
fi

echo "La validación pasó correctamente."

if [[ "$AUTO_PUSH" == true ]]; then
  git add -A
  if ! git diff --cached --quiet; then
    git commit -m "feat: ${SELECTED_FEATURE}"
  fi
  git push --set-upstream origin "feature/$SELECTED_FEATURE"
  echo "Push realizado en la rama feature/$SELECTED_FEATURE"
else
  echo "Validación OK. Si quieres hacer push, reejecuta con --push."
fi

exit 0
