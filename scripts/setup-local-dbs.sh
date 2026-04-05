#!/usr/bin/env bash
set -euo pipefail

echo "Setting up local Postgres databases..."
export PGPASSWORD=mudropass

psql -h localhost -p 5432 -U mudro mudro_main -c 'SELECT 1' >/dev/null 2>&1 || (
  echo 'Creating main database...'; createdb mudro_main -U mudro || true
)

psql -h localhost -p 5433 -U mudro mudro_casino -c 'SELECT 1' >/dev/null 2>&1 || (
  echo 'Creating casino database...'; createdb mudro_casino -U mudro || true
)

echo "Databases ready. Apply migrations as needed using your existing migration tooling." 
