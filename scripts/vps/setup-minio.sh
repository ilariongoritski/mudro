#!/usr/bin/env bash
set -euo pipefail

# Load local environment when running from the repository root.
if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

MINIO_ALIAS="mudro"
MINIO_URL="${MINIO_URL:-http://localhost:9000}"
: "${MINIO_ROOT_USER:?Set MINIO_ROOT_USER in the environment or .env}"
: "${MINIO_ROOT_PASSWORD:?Set MINIO_ROOT_PASSWORD in the environment or .env}"
BUCKET_NAME="${MINIO_BUCKET:-media}"

mc alias set "$MINIO_ALIAS" "$MINIO_URL" "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"
mc mb "$MINIO_ALIAS/$BUCKET_NAME" || true

# This bucket is intended for public media URLs. Do not use this for private uploads.
mc anonymous set public "$MINIO_ALIAS/$BUCKET_NAME"

echo "MinIO bucket '$BUCKET_NAME' is ready at $MINIO_URL/$BUCKET_NAME"
