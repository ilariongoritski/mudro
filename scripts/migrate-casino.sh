#!/usr/bin/env bash
set -euo pipefail

CASINO_DB_NAME="${CASINO_DB_NAME:-mudro_casino}"
CASINO_MIGRATIONS_DIR="${CASINO_MIGRATIONS_DIR:-services/casino/migrations}"

apply_file() {
  local file="$1"
  echo "==> applying ${file}"

  if docker compose ps -q casino-db 2>/dev/null | grep -q .; then
    cat "${file}" | docker compose exec -T casino-db psql -U postgres -d "${CASINO_DB_NAME}" -X -v ON_ERROR_STOP=1
    return
  fi

  psql "${CASINO_DSN:-postgres://postgres:postgres@localhost:5434/${CASINO_DB_NAME}?sslmode=disable}" -X -v ON_ERROR_STOP=1 -f "${file}"
}

for file in "${CASINO_MIGRATIONS_DIR}"/*.sql; do
  apply_file "${file}"
done
