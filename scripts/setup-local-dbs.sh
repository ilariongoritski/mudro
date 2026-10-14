#!/usr/bin/env bash
set -euo pipefail

: "${POSTGRES_PASSWORD:?Set POSTGRES_PASSWORD to the password configured for the local databases}"
export PGPASSWORD="$POSTGRES_PASSWORD"

echo "Checking local PostgreSQL databases..."

psql -h localhost -p 5432 -U mudro mudro_main -c 'SELECT 1' >/dev/null 2>&1 || {
  echo "Main database is not reachable. Start the deprecated local stack and create mudro_main first." >&2
  exit 1
}

psql -h localhost -p 5433 -U mudro mudro_casino -c 'SELECT 1' >/dev/null 2>&1 || {
  echo "Casino database is not reachable. Start the deprecated local stack and create mudro_casino first." >&2
  exit 1
}

echo "Both local databases are reachable. Apply migrations with the project migration tooling."
