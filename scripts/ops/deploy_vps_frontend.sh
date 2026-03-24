#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/srv/mudro}"
WEB_ROOT="${WEB_ROOT:-/var/www/mudro/frontend}"
NGINX_SITE="${NGINX_SITE:-/etc/nginx/sites-available/mudro}"
NGINX_ENABLED="${NGINX_ENABLED:-/etc/nginx/sites-enabled/mudro}"
NGINX_TEMPLATE="${PROJECT_DIR}/scripts/ops/mudro.nginx.conf"

if [ "$(id -u)" -eq 0 ]; then
  SUDO=()
else
  SUDO=(sudo)
fi

if [ ! -d "${PROJECT_DIR}" ]; then
  echo "project dir not found: ${PROJECT_DIR}" >&2
  exit 1
fi

if [ ! -d "${PROJECT_DIR}/frontend/dist" ]; then
  echo "frontend dist not found: ${PROJECT_DIR}/frontend/dist" >&2
  exit 1
fi

"${SUDO[@]}" apt-get update >/tmp/mudro-nginx-apt-update.log 2>&1
"${SUDO[@]}" apt-get install -y nginx rsync >/tmp/mudro-nginx-install.log 2>&1

"${SUDO[@]}" install -d -m 755 "${WEB_ROOT}"
"${SUDO[@]}" rsync -a --delete "${PROJECT_DIR}/frontend/dist/" "${WEB_ROOT}/"

"${SUDO[@]}" install -D -m 644 "${NGINX_TEMPLATE}" "${NGINX_SITE}"
"${SUDO[@]}" ln -sfn "${NGINX_SITE}" "${NGINX_ENABLED}"
"${SUDO[@]}" rm -f /etc/nginx/sites-enabled/default

"${SUDO[@]}" nginx -t
"${SUDO[@]}" systemctl enable nginx >/dev/null 2>&1
"${SUDO[@]}" systemctl restart nginx

if command -v ufw >/dev/null 2>&1; then
  "${SUDO[@]}" ufw allow 80/tcp >/dev/null 2>&1 || true
fi

curl -fsS http://127.0.0.1/ >/dev/null
curl -fsS http://127.0.0.1/healthz >/dev/null

echo "[ok] nginx serves frontend from ${WEB_ROOT}"
echo "[ok] /healthz is proxied to 127.0.0.1:8080"
