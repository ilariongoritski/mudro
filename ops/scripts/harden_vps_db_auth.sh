#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/projects/mudro}"
APP_USER="${MUDRO_DB_APP_USER:-mudro_app}"
APP_PASSWORD="${MUDRO_DB_APP_PASSWORD:?MUDRO_DB_APP_PASSWORD is required}"
SUPERUSER_PASSWORD="${MUDRO_DB_SUPERUSER_PASSWORD:?MUDRO_DB_SUPERUSER_PASSWORD is required}"
ENV_FILE="${PROJECT_DIR}/.env"
SERVICE_DROPIN_DIR="/etc/systemd/system/mudro-api.service.d"
SERVICE_DROPIN_FILE="${SERVICE_DROPIN_DIR}/10-dsn.conf"
FIREWALL_UNIT="/etc/systemd/system/mudro-db-firewall.service"
APP_DSN="postgres://${APP_USER}:${APP_PASSWORD}@localhost:5433/gallery?sslmode=disable"

cd "${PROJECT_DIR}"

python3 - <<'PY' "${ENV_FILE}" "${APP_DSN}" "${SUPERUSER_PASSWORD}"
from pathlib import Path
import sys

env_path = Path(sys.argv[1])
app_dsn = sys.argv[2]
superuser_password = sys.argv[3]

lines = []
if env_path.exists():
    lines = env_path.read_text(encoding="utf-8").splitlines()

def upsert(key: str, value: str) -> None:
    prefix = key + "="
    for idx, line in enumerate(lines):
        if line.startswith(prefix):
            lines[idx] = prefix + value
            return
    lines.append(prefix + value)

upsert("DSN", app_dsn)
upsert("POSTGRES_PASSWORD", superuser_password)
env_path.write_text("\n".join(lines).rstrip() + "\n", encoding="utf-8")
PY

docker compose up -d db >/dev/null

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

cat "${SQL_FILE}" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1 >/dev/null
rm -f "${SQL_FILE}"

install -d -m 755 "${SERVICE_DROPIN_DIR}"
cat > "${SERVICE_DROPIN_FILE}" <<EOF
[Service]
Environment=DSN=${APP_DSN}
EOF

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

systemctl daemon-reload
systemctl enable --now mudro-db-firewall.service >/dev/null
systemctl restart mudro-api
docker compose up -d db >/dev/null

for service in mudro-bot.service mudro-reporter.service; do
  if systemctl list-unit-files "${service}" >/dev/null 2>&1; then
    if systemctl is-active --quiet "${service}"; then
      systemctl restart "${service}"
    fi
  fi
done

curl -fsS http://127.0.0.1:8080/healthz >/dev/null

echo "[ok] mudro-api DSN switched to ${APP_USER}"
echo "[ok] docker db port now follows compose bind on localhost only"
echo "[ok] inbound tcp/5433 is blocked outside loopback by systemd-managed iptables rule"
echo "[ok] mudro-api healthz passed"
