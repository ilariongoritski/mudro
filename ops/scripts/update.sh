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
for f in $(find migrations/ -name "*.sql" ! -name "*.down.sql" | sort); do
  docker compose -f docker-compose.prod.yml exec -T db psql -U postgres -d gallery -v ON_ERROR_STOP=0 < "$f" > /dev/null 2>&1 || true
done
for f in $(find services/casino/migrations/ -name "*.sql" ! -name "*.down.sql" | sort); do
  docker compose -f docker-compose.prod.yml exec -T casino-db psql -U postgres -d mudro_casino -v ON_ERROR_STOP=0 < "$f" > /dev/null 2>&1 || true
done

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
curl -s http://localhost/healthz || echo "HEALTH FAIL"
echo ""
echo "DONE! Site: https://222.167.208.10.nip.io"
