#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/projects/mudro}"
WEB_ROOT="${WEB_ROOT:-/var/www/mudro/frontend}"
NGINX_SITE="${NGINX_SITE:-/etc/nginx/sites-available/mudro}"
NGINX_ENABLED="${NGINX_ENABLED:-/etc/nginx/sites-enabled/mudro}"
NGINX_TEMPLATE="${PROJECT_DIR}/ops/scripts/mudro.nginx.conf"

if [ ! -d "${PROJECT_DIR}" ]; then
  echo "project dir not found: ${PROJECT_DIR}" >&2
  exit 1
fi

if [ ! -d "${PROJECT_DIR}/frontend/dist" ]; then
  echo "frontend dist not found: ${PROJECT_DIR}/frontend/dist" >&2
  exit 1
fi

apt-get update >/tmp/mudro-nginx-apt-update.log 2>&1
apt-get install -y nginx >/tmp/mudro-nginx-install.log 2>&1

install -d -m 755 "${WEB_ROOT}"
rsync -a --delete "${PROJECT_DIR}/frontend/dist/" "${WEB_ROOT}/"

install -D -m 644 "${NGINX_TEMPLATE}" "${NGINX_SITE}"
ln -sfn "${NGINX_SITE}" "${NGINX_ENABLED}"
rm -f /etc/nginx/sites-enabled/default

nginx -t
systemctl enable nginx >/dev/null 2>&1
systemctl restart nginx

if command -v ufw >/dev/null 2>&1; then
  ufw allow 80/tcp >/dev/null 2>&1 || true
fi

curl -fsS http://127.0.0.1/ >/dev/null
curl -fsS http://127.0.0.1/healthz >/dev/null

echo "[ok] nginx serves frontend from ${WEB_ROOT}"
echo "[ok] /healthz is proxied to 127.0.0.1:8080"
