#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/projects/mudro}"
WEB_ROOT="${WEB_ROOT:-/var/www/mudro/frontend}"
CERTBOT_WEBROOT="${CERTBOT_WEBROOT:-/var/www/mudro/certbot}"
NGINX_SITE="${NGINX_SITE:-/etc/nginx/sites-available/mudro}"
NGINX_ENABLED="${NGINX_ENABLED:-/etc/nginx/sites-enabled/mudro}"
HTTP_TEMPLATE="${PROJECT_DIR}/ops/nginx/mudro.conf"
HTTPS_TEMPLATE="${PROJECT_DIR}/ops/nginx/mudro.https.conf.template"

MUDRO_DOMAIN="${MUDRO_DOMAIN:-}"
MUDRO_LETSENCRYPT_EMAIL="${MUDRO_LETSENCRYPT_EMAIL:-}"

if [ -z "${MUDRO_DOMAIN}" ]; then
  echo "MUDRO_DOMAIN is required" >&2
  exit 1
fi

if [ -z "${MUDRO_LETSENCRYPT_EMAIL}" ]; then
  echo "MUDRO_LETSENCRYPT_EMAIL is required" >&2
  exit 1
fi

if ! [[ "${MUDRO_DOMAIN}" =~ ^[A-Za-z0-9.-]+$ ]]; then
  echo "MUDRO_DOMAIN contains unsupported characters" >&2
  exit 1
fi

for path in "${HTTP_TEMPLATE}" "${HTTPS_TEMPLATE}" "${WEB_ROOT}"; do
  if [ ! -e "${path}" ]; then
    echo "required path not found: ${path}" >&2
    exit 1
  fi
done

apt-get update >/tmp/mudro-https-apt-update.log 2>&1
apt-get install -y nginx certbot >/tmp/mudro-https-apt-install.log 2>&1

install -d -m 755 "${CERTBOT_WEBROOT}"
install -D -m 644 "${HTTP_TEMPLATE}" "${NGINX_SITE}"
ln -sfn "${NGINX_SITE}" "${NGINX_ENABLED}"
rm -f /etc/nginx/sites-enabled/default

nginx -t
systemctl enable nginx >/dev/null 2>&1
systemctl restart nginx

certbot certonly \
  --webroot \
  --webroot-path "${CERTBOT_WEBROOT}" \
  --domain "${MUDRO_DOMAIN}" \
  --email "${MUDRO_LETSENCRYPT_EMAIL}" \
  --agree-tos \
  --no-eff-email \
  --non-interactive

sed "s/{{MUDRO_DOMAIN}}/${MUDRO_DOMAIN}/g" "${HTTPS_TEMPLATE}" >"${NGINX_SITE}.tmp"
install -m 644 "${NGINX_SITE}.tmp" "${NGINX_SITE}"
rm -f "${NGINX_SITE}.tmp"

nginx -t
systemctl reload nginx

if command -v ufw >/dev/null 2>&1; then
  if [ "${MUDRO_CONFIRM_FIREWALL:-}" = "1" ]; then
    ufw allow 80/tcp >/dev/null 2>&1 || true
    ufw allow 443/tcp >/dev/null 2>&1 || true
  else
    echo "[skip] UFW changes skipped. Run with MUDRO_CONFIRM_FIREWALL=1 after explicit approval."
  fi
fi

systemctl list-timers --all | grep -E 'certbot|snap.certbot' || true
curl -fsS "https://${MUDRO_DOMAIN}/healthz" >/dev/null

echo "[ok] HTTPS enabled for ${MUDRO_DOMAIN}"
echo "[ok] /healthz is proxied to 127.0.0.1:8080"
