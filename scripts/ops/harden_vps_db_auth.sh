#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/srv/mudro}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"
APP_USER="${MUDRO_DB_APP_USER:-mudro_app}"
APP_PASSWORD="${MUDRO_DB_APP_PASSWORD:?MUDRO_DB_APP_PASSWORD is required}"
SUPERUSER_PASSWORD="${MUDRO_DB_SUPERUSER_PASSWORD:?MUDRO_DB_SUPERUSER_PASSWORD is required}"
ENV_DIR="${PROJECT_DIR}/env"
COMMON_ENV_FILE="${ENV_DIR}/common.env"
DB_ENV_FILE="${ENV_DIR}/db.env"
API_ENV_FILE="${ENV_DIR}/api.env"
AGENT_ENV_FILE="${ENV_DIR}/agent.env"
REPORTER_ENV_FILE="${ENV_DIR}/reporter.env"
SERVICE_DROPIN_DIR="/etc/systemd/system/mudro-api.service.d"
SERVICE_DROPIN_FILE="${SERVICE_DROPIN_DIR}/10-dsn.conf"
FIREWALL_UNIT="/etc/systemd/system/mudro-db-firewall.service"
APP_DSN="postgres://${APP_USER}:${APP_PASSWORD}@localhost:5433/gallery?sslmode=disable"
COMPOSE=(docker compose -f "${COMPOSE_FILE}")

if [ "$(id -u)" -eq 0 ]; then
  SUDO=()
else
  SUDO=(sudo)
fi

cd "${PROJECT_DIR}"
install -d -m 755 "${ENV_DIR}"

python3 - <<'PY' "${COMMON_ENV_FILE}" "${DB_ENV_FILE}" "${API_ENV_FILE}" "${AGENT_ENV_FILE}" "${REPORTER_ENV_FILE}" "${APP_DSN}" "${SUPERUSER_PASSWORD}"
from pathlib import Path
import sys

common_env_path = Path(sys.argv[1])
db_env_path = Path(sys.argv[2])
api_env_path = Path(sys.argv[3])
agent_env_path = Path(sys.argv[4])
reporter_env_path = Path(sys.argv[5])
app_dsn = sys.argv[6]
superuser_password = sys.argv[7]

def load_lines(path: Path) -> list[str]:
    if path.exists():
        return path.read_text(encoding="utf-8").splitlines()
    return []

def upsert(lines: list[str], key: str, value: str) -> None:
    prefix = key + "="
    for idx, line in enumerate(lines):
        if line.startswith(prefix):
            lines[idx] = prefix + value
            return
    lines.append(prefix + value)

common_lines = load_lines(common_env_path)
db_lines = load_lines(db_env_path)
api_lines = load_lines(api_env_path)
agent_lines = load_lines(agent_env_path)
reporter_lines = load_lines(reporter_env_path)

upsert(common_lines, "MUDRO_ROOT", "/app")
upsert(db_lines, "POSTGRES_PASSWORD", superuser_password)
upsert(api_lines, "DSN", app_dsn)
upsert(agent_lines, "DSN", app_dsn)
upsert(reporter_lines, "DSN", app_dsn)

common_env_path.write_text("\n".join(common_lines).rstrip() + "\n", encoding="utf-8")
db_env_path.write_text("\n".join(db_lines).rstrip() + "\n", encoding="utf-8")
api_env_path.write_text("\n".join(api_lines).rstrip() + "\n", encoding="utf-8")
agent_env_path.write_text("\n".join(agent_lines).rstrip() + "\n", encoding="utf-8")
reporter_env_path.write_text("\n".join(reporter_lines).rstrip() + "\n", encoding="utf-8")
PY

"${COMPOSE[@]}" up -d db >/dev/null

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

cat "${SQL_FILE}" | "${COMPOSE[@]}" exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1 >/dev/null
rm -f "${SQL_FILE}"

TMP_DROPIN="$(mktemp)"
cat > "${TMP_DROPIN}" <<EOF
[Service]
Environment=DSN=${APP_DSN}
EOF
"${SUDO[@]}" install -d -m 755 "${SERVICE_DROPIN_DIR}"
"${SUDO[@]}" install -D -m 644 "${TMP_DROPIN}" "${SERVICE_DROPIN_FILE}"
rm -f "${TMP_DROPIN}"

TMP_FIREWALL_UNIT="$(mktemp)"
cat > "${TMP_FIREWALL_UNIT}" <<'EOF'
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
"${SUDO[@]}" install -D -m 644 "${TMP_FIREWALL_UNIT}" "${FIREWALL_UNIT}"
rm -f "${TMP_FIREWALL_UNIT}"

"${SUDO[@]}" systemctl daemon-reload
"${SUDO[@]}" systemctl enable --now mudro-db-firewall.service >/dev/null

if "${SUDO[@]}" systemctl list-unit-files mudro-api.service >/dev/null 2>&1; then
  "${SUDO[@]}" systemctl restart mudro-api
fi

"${COMPOSE[@]}" up -d db api agent >/dev/null

if grep -Eq '^REPORT_BOT_TOKEN=.+$' "${REPORTER_ENV_FILE}" 2>/dev/null; then
  "${COMPOSE[@]}" up -d reporter >/dev/null
fi

for service in mudro-bot.service mudro-reporter.service; do
  if "${SUDO[@]}" systemctl list-unit-files "${service}" >/dev/null 2>&1; then
    if "${SUDO[@]}" systemctl is-active --quiet "${service}"; then
      "${SUDO[@]}" systemctl restart "${service}"
    fi
  fi
done

curl -fsS http://127.0.0.1:8080/healthz >/dev/null

echo "[ok] mudro-api DSN switched to ${APP_USER}"
echo "[ok] env/api.env, env/agent.env and env/reporter.env now use the dedicated app role"
echo "[ok] docker db port now follows compose bind on localhost only"
echo "[ok] inbound tcp/5433 is blocked outside loopback by systemd-managed iptables rule"
echo "[ok] mudro-api healthz passed"
