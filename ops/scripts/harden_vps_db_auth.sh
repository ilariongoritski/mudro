#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${PROJECT_DIR:-$(cd "${SCRIPT_DIR}/../.." && pwd)}"
APP_USER="${MUDRO_DB_APP_USER:-mudro_app}"
APP_PASSWORD="${MUDRO_DB_APP_PASSWORD:?MUDRO_DB_APP_PASSWORD is required}"
SUPERUSER_PASSWORD="${MUDRO_DB_SUPERUSER_PASSWORD:?MUDRO_DB_SUPERUSER_PASSWORD is required}"
PROD_COMPOSE_FILE="${PROD_COMPOSE_FILE:-${PROJECT_DIR}/docker-compose.prod.yml}"
SYSTEMD_RUNTIME_DIR="${MUDRO_SYSTEMD_RUNTIME_DIR:-/etc/mudro/runtime}"
SYSTEMD_ENV_FILE="${SYSTEMD_RUNTIME_DIR}/mudro-api.env"
FIREWALL_UNIT="/etc/systemd/system/mudro-db-firewall.service"
APP_DSN="postgres://${APP_USER}:${APP_PASSWORD}@localhost:5433/gallery?sslmode=disable"
JWT_SECRET_OVERRIDE="${MUDRO_API_JWT_SECRET:-}"

cd "${PROJECT_DIR}"

read_existing_env_value() {
  local key="$1"
  local file="$2"
  if [[ ! -f "${file}" ]]; then
    return 1
  fi

  grep -E "^${key}=" "${file}" | tail -n 1 | cut -d '=' -f 2-
}

JWT_SECRET_VALUE="${JWT_SECRET_OVERRIDE}"
if [[ -z "${JWT_SECRET_VALUE}" ]]; then
  JWT_SECRET_VALUE="$(read_existing_env_value "JWT_SECRET" "${SYSTEMD_ENV_FILE}" || true)"
fi
if [[ -z "${JWT_SECRET_VALUE}" || "${JWT_SECRET_VALUE}" == "change-me" ]]; then
  echo "[error] set MUDRO_API_JWT_SECRET or provide an existing non-placeholder JWT_SECRET in ${SYSTEMD_ENV_FILE}" >&2
  exit 1
fi

POSTGRES_PASSWORD="${SUPERUSER_PASSWORD}" docker compose -f "${PROD_COMPOSE_FILE}" up -d db >/dev/null

SQL_FILE="$(mktemp)"
cat > "${SQL_FILE}" <<SQL
ALTER ROLE postgres WITH PASSWORD '${SUPERUSER_PASSWORD}';
DO \$\$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${APP_USER}') THEN
    CREATE ROLE ${APP_USER} LOGIN PASSWORD '${APP_PASSWORD}' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
  ELSE
    ALTER ROLE ${APP_USER} WITH LOGIN PASSWORD '${APP_PASSWORD}' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
  END IF;
END \$\$;
GRANT CONNECT ON DATABASE gallery TO ${APP_USER};
GRANT USAGE ON SCHEMA public TO ${APP_USER};
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ${APP_USER};
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO ${APP_USER};
ALTER DEFAULT PRIVILEGES FOR USER postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${APP_USER};
ALTER DEFAULT PRIVILEGES FOR USER postgres IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO ${APP_USER};
SQL

cat "${SQL_FILE}" | POSTGRES_PASSWORD="${SUPERUSER_PASSWORD}" docker compose -f "${PROD_COMPOSE_FILE}" exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1 >/dev/null
rm -f "${SQL_FILE}"

install -d -m 755 "${SYSTEMD_RUNTIME_DIR}"
cat > "${SYSTEMD_ENV_FILE}" <<EOF
MUDRO_ENV=production
MUDRO_ROOT=/opt/mudro/app
DSN=${APP_DSN}
API_ADDR=:8080
JWT_SECRET=${JWT_SECRET_VALUE}
CASINO_SERVICE_URL=http://127.0.0.1:8082
MEDIA_ROOT=/var/lib/mudro/media
EOF
chmod 600 "${SYSTEMD_ENV_FILE}"

rm -f /etc/systemd/system/mudro-api.service.d/10-dsn.conf
rmdir /etc/systemd/system/mudro-api.service.d 2>/dev/null || true

cat > "${FIREWALL_UNIT}" <<'EOF'
[Unit]
Description=Mudro DB firewall guard
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/bin/sh -lc 'iptables -C INPUT ! -i lo -p tcp --dport 5433 -j DROP 2>/dev/null || iptables -I INPUT ! -i lo -p tcp --dport 5433 -j DROP'
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOF

ONLY=api NO_START=1 bash "${SCRIPT_DIR}/install_mudro_systemd.sh"
systemctl enable --now mudro-db-firewall.service >/dev/null
systemctl restart mudro-api
POSTGRES_PASSWORD="${SUPERUSER_PASSWORD}" docker compose -f "${PROD_COMPOSE_FILE}" up -d db >/dev/null

for service in mudro-bot.service mudro-reporter.service; do
  if systemctl list-unit-files "${service}" >/dev/null 2>&1; then
    if systemctl is-active --quiet "${service}"; then
      systemctl restart "${service}"
    fi
  fi
done

curl -fsS http://127.0.0.1:8080/healthz >/dev/null

echo "[ok] mudro-api DSN switched to ${APP_USER}"
echo "[ok] mudro-api runtime env synced to ${SYSTEMD_ENV_FILE}"
echo "[ok] docker db port now follows compose bind on localhost only"
echo "[ok] inbound tcp/5433 is blocked outside loopback by systemd-managed iptables rule"
echo "[ok] mudro-api healthz passed"
