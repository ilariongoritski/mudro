#!/bin/bash
# setup-minio.sh - Configure MinIO for Mudro

# Load env
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

MINIO_ALIAS="mudro"
MINIO_URL=${MINIO_URL:-"http://localhost:9000"}
MINIO_ROOT_USER=${MINIO_ROOT_USER:-"admin"}
MINIO_ROOT_PASSWORD=${MINIO_ROOT_PASSWORD:-"admin123"}
BUCKET_NAME="media"

echo "Configuring mc alias..."
mc alias set $MINIO_ALIAS $MINIO_URL $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD

echo "Creating bucket: $BUCKET_NAME..."
mc mb $MINIO_ALIAS/$BUCKET_NAME || true

echo "Setting bucket policy to public read..."
mc anonymous set public $MINIO_ALIAS/$BUCKET_NAME

echo "MinIO setup complete. Bucket '$BUCKET_NAME' is ready at $MINIO_URL/$BUCKET_NAME"
