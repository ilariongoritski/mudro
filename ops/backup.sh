#!/bin/bash
set -eo pipefail

echo "=== MUDRO PostgreSQL Backup to MinIO ==="

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_DIR="${ENV_DIR:-${PROJECT_DIR}/env}"

set -a
[ -f "${ENV_DIR}/db.env" ] && . "${ENV_DIR}/db.env"
[ -f "${ENV_DIR}/storage.env" ] && . "${ENV_DIR}/storage.env"
set +a

DB_USER="${POSTGRES_USER:-postgres}"
DB_PASS="${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
DB_NAME="${POSTGRES_DB:-gallery}"
DB_HOST="${POSTGRES_HOST:-127.0.0.1}"
DB_PORT="${POSTGRES_PORT:-5433}"

MINIO_ENDPOINT="http://127.0.0.1:9000"
MINIO_USER="${MINIO_ROOT_USER:?MINIO_ROOT_USER is required}"
MINIO_PASS="${MINIO_ROOT_PASSWORD:?MINIO_ROOT_PASSWORD is required}"
MINIO_BUCKET="backups"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
FILENAME="dump_${DB_NAME}_${TIMESTAMP}.sql.gz"
LOCAL_PATH="/tmp/${FILENAME}"

echo "1. Creating database dump..."
export PGPASSWORD="$DB_PASS"
pg_dump -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" "$DB_NAME" | gzip > "$LOCAL_PATH"

echo "2. Installing/configuring MinIO Client (mc) if needed..."
if ! command -v mc &> /dev/null; then
    curl -O https://dl.min.io/client/mc/release/linux-amd64/mc
    chmod +x mc
    sudo mv mc /usr/local/bin/
fi

mc alias set myminio "$MINIO_ENDPOINT" "$MINIO_USER" "$MINIO_PASS"

echo "3. Creating bucket if not exists..."
mc mb myminio/${MINIO_BUCKET} 2>/dev/null || true

echo "4. Uploading backup to MinIO..."
mc cp "$LOCAL_PATH" myminio/${MINIO_BUCKET}/${FILENAME}

echo "5. Cleanup local file..."
rm -f "$LOCAL_PATH"

echo "6. Applying retention policy (keep last 30 days)..."
mc ilm add --expire-days 30 myminio/${MINIO_BUCKET} || echo "ILM config failed, you may need to set it manually."

echo "Backup completed successfully!"
