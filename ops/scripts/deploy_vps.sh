#!/bin/bash
set -euo pipefail
cd /opt/mudro

echo "==> Pulling latest code..."
git pull origin main 2>/dev/null || echo "  (pull failed, using existing code)"

echo "==> Checking frontend changes..."
if git diff --name-only HEAD@{1} HEAD 2>/dev/null | grep -q "^frontend/"; then
  echo "  Frontend changed, rebuilding..."
  cd frontend
  npm ci --production=false 2>/dev/null || npm install 2>/dev/null
  npm run build
  cd /opt/mudro
  echo "  Frontend rebuilt."
else
  echo "  No frontend changes."
fi

echo "==> Applying new migrations..."
# Main DB
for f in $(find migrations/ -name "*.sql" ! -name "*.down.sql" | sort); do
  docker compose -f docker-compose.prod.yml exec -T db psql -U postgres -d gallery -v ON_ERROR_STOP=0 < "$f" > /dev/null 2>&1 || true
done
# Casino DB
for f in $(find services/casino/migrations/ -name "*.sql" ! -name "*.down.sql" | sort); do
  docker compose -f docker-compose.prod.yml exec -T casino-db psql -U postgres -d mudro_casino -v ON_ERROR_STOP=0 < "$f" > /dev/null 2>&1 || true
done


echo "==> Importing movie catalog data..."
if [ -f out/movie_catalog_data.sql ]; then
  echo "  Restoring movie catalog data from SQL dump..."
  docker compose -f docker-compose.prod.yml exec -T db psql -U postgres -d gallery -v ON_ERROR_STOP=0 < out/movie_catalog_data.sql > /dev/null 2>&1 || \
    echo "  (movie catalog data restore skipped or already up to date)"
  echo "  Movie catalog data restored."
elif [ -f out/movie-catalog.slim.json ]; then
  echo "  Importing movie catalog data from slim.json..."
  docker compose -f docker-compose.prod.yml run --rm \
    -e MOVIE_CATALOG_DB_DSN="${MUDRO_APP_DSN}" \
    -e MOVIE_CATALOG_IMPORT_FILE="/app/out/movie-catalog.slim.json" \
    -v "$(pwd)/out/movie-catalog.slim.json:/app/out/movie-catalog.slim.json:ro" \
    -v "$(pwd):/app" \
    --no-deps \
    movie-catalog /usr/local/bin/go run ./tools/importers/moviecatalogimport/cmd 2>&1 | tail -3 || \
    echo "  (movie import skipped or failed)"
  echo "  Movie catalog data imported."
else
  echo "  No movie catalog data file found, skipping import."
fi
echo "==> Rebuilding services..."
docker compose -f docker-compose.prod.yml build --parallel 2>&1 | tail -3

echo "==> Restarting services..."
docker compose -f docker-compose.prod.yml up -d 2>&1 | tail -5

echo "==> Waiting for services..."
sleep 15
docker compose -f docker-compose.prod.yml ps --format "table {{.Name}}\t{{.Status}}"

echo "==> Reload nginx..."
nginx -t 2>&1 && systemctl reload nginx

echo "==> Health check..."

# movie-catalog
curl -s http://localhost:8091/healthz | grep -q "ok" && echo "  movie-catalog: OK" || echo "  movie-catalog: FAIL"
# bff-web movie catalog proxy
curl -s "http://localhost:8086/api/movie-catalog/genres" | grep -q "items" && echo "  bff-web movie-catalog proxy: OK" || echo "  bff-web movie-catalog proxy: FAIL"
# main API
curl -s http://localhost/healthz || echo "HEALTH FAIL"
echo ""
echo "DONE! Site: https://222.167.208.10.nip.io"
