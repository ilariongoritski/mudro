#!/usr/bin/env bash
set -euo pipefail

MAIN_DB_NAME="${MAIN_DB_NAME:-gallery}"
MAIN_CASINO_MIGRATIONS=(
  "migrations/012_casino.sql"
  "migrations/017_casino_emoji_v2.sql"
  "migrations/018_casino_stabilization.sql"
)

apply_file() {
  local file="$1"
  echo "==> applying ${file}"

  if docker compose ps -q db 2>/dev/null | grep -q .; then
    cat "${file}" | docker compose exec -T db psql -U postgres -d "${MAIN_DB_NAME}" -X -v ON_ERROR_STOP=1
    return
  fi

  psql "${DSN:-postgres://postgres:postgres@localhost:5433/${MAIN_DB_NAME}?sslmode=disable}" -X -v ON_ERROR_STOP=1 -f "${file}"
}

for file in "${MAIN_CASINO_MIGRATIONS[@]}"; do
  apply_file "${file}"
done
