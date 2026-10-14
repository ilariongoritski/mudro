#!/usr/bin/env bash
set -euo pipefail

echo "=== MUDRO PostgreSQL Backup to MinIO ==="

DB_USER="${POSTGRES_USER:-postgres}"
: "${POSTGRES_PASSWORD:?Set POSTGRES_PASSWORD in the environment}"
DB_PASS="$POSTGRES_PASSWORD"
DB_NAME="${POSTGRES_DB:-gallery}"
DB_HOST="${POSTGRES_HOST:-127.0.0.1}"
DB_PORT="${POSTGRES_PORT:-5433}"

MINIO_ENDPOINT="${MINIO_ENDPOINT:-http://127.0.0.1:9000}"
: "${MINIO_ROOT_USER:?Set MINIO_ROOT_USER in the environment}"
: "${MINIO_ROOT_PASSWORD:?Set MINIO_ROOT_PASSWORD in the environment}"
MINIO_USER="$MINIO_ROOT_USER"
MINIO_PASS="$MINIO_ROOT_PASSWORD"
MINIO_BUCKET="${MINIO_BUCKET:-backups}"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
FILENAME="dump_${DB_NAME}_${TIMESTAMP}.sql.gz"
LOCAL_PATH="/tmp/${FILENAME}"
trap 'rm -f "$LOCAL_PATH"' EXIT

echo "1. Creating database dump..."
export PGPASSWORD="$DB_PASS"
pg_dump -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" "$DB_NAME" | gzip > "$LOCAL_PATH"

echo "2. Installing/configuring MinIO Client (mc) if needed..."
if ! command -v mc >/dev/null 2>&1; then
    curl -fL --retry 3 -o mc https://dl.min.io/client/mc/release/linux-amd64/mc
    chmod +x mc
    sudo mv mc /usr/local/bin/
fi

mc alias set myminio "$MINIO_ENDPOINT" "$MINIO_USER" "$MINIO_PASS"

echo "3. Creating bucket if not exists..."
mc mb "myminio/$MINIO_BUCKET" 2>/dev/null || true

echo "4. Uploading backup to MinIO..."
mc cp "$LOCAL_PATH" "myminio/$MINIO_BUCKET/$FILENAME"

echo "5. Applying retention policy (keep last 30 days)..."
mc ilm add --expire-days 30 "myminio/$MINIO_BUCKET" || echo "ILM config failed; configure retention manually."

echo "Backup completed successfully!"
