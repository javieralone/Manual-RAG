#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "[1/4] Ejecutando tests de Go..."
(cd api-go && go test ./...)

echo "[2/4] Compilando API Go..."
(cd api-go && go build ./...)

echo "[3/4] Validando compilación Python..."
python -m compileall rag-engine/app

echo "[4/4] Validando archivo Docker Compose..."
docker compose --env-file .env.example config --quiet

echo "Validación completada correctamente."
