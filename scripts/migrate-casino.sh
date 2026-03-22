#!/usr/bin/env bash
set -euo pipefail

CASINO_DB_NAME="${CASINO_DB_NAME:-mudro_casino}"
CASINO_MIGRATION="${CASINO_MIGRATION:-services/casino/migrations/001_init.sql}"

if docker compose ps -q casino-db 2>/dev/null | grep -q .; then
  cat "${CASINO_MIGRATION}" | docker compose exec -T casino-db psql -U postgres -d "${CASINO_DB_NAME}" -X -v ON_ERROR_STOP=1
  exit 0
fi

psql "${CASINO_DSN:-postgres://postgres:postgres@localhost:5434/${CASINO_DB_NAME}?sslmode=disable}" -X -v ON_ERROR_STOP=1 -f "${CASINO_MIGRATION}"
