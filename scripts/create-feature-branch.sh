#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "Uso: $0 <feature-name>"
  echo "Ejemplo: $0 01-base-observabilidad"
  exit 1
fi

RAW_NAME="$1"
SLUG="${RAW_NAME,,}"
SLUG="${SLUG//[^a-z0-9-]/-}"
SLUG="${SLUG//--/-}"
SLUG="${SLUG#-}"
SLUG="${SLUG%-}"

if [ -z "$SLUG" ]; then
  echo "El nombre de la feature no es válido."
  exit 1
fi

BRANCH="feature/$SLUG"

git checkout -b "$BRANCH"
echo "Rama creada: $BRANCH"
