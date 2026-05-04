#!/usr/bin/env bash
set -euo pipefail

CASINO_DB_NAME="${CASINO_DB_NAME:-mudro_casino}"
CASINO_MIGRATIONS_DIR="${CASINO_MIGRATIONS_DIR:-services/casino/migrations}"
MODE="${1:-apply}"

list_files() {
  find "${CASINO_MIGRATIONS_DIR}" -maxdepth 1 -type f -name "*.sql" ! -name "*.down.sql" -print | sort
}

compose() {
  if [ -n "${CASINO_COMPOSE_FILE:-}" ]; then
    docker compose -f "${CASINO_COMPOSE_FILE}" "$@"
    return
  fi
  docker compose "$@"
}

apply_file() {
  local file="$1"
  echo "==> applying ${file}"

  if compose ps -q casino-db 2>/dev/null | grep -q .; then
    cat "${file}" | compose exec -T casino-db psql -U postgres -d "${CASINO_DB_NAME}" -X -v ON_ERROR_STOP=1
    return
  fi

  psql "${CASINO_DSN:-postgres://postgres:postgres@localhost:5434/${CASINO_DB_NAME}?sslmode=disable}" -X -v ON_ERROR_STOP=1 -f "${file}"
}

case "${MODE}" in
  --list|list)
    list_files
    exit 0
    ;;
  --dry-run|dry-run)
    list_files | sed 's/^/==> would apply /'
    exit 0
    ;;
  apply)
    ;;
  *)
    echo "usage: $0 [--list|--dry-run]" >&2
    exit 2
    ;;
esac

while IFS= read -r file; do
  apply_file "${file}"
done < <(list_files)
