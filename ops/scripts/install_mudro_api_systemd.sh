#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/projects/mudro}"
UNIT_SOURCE="${PROJECT_DIR}/ops/systemd/mudro-api.service"
UNIT_DEST="/etc/systemd/system/mudro-api.service"
RUNTIME_DIR="${MUDRO_SYSTEMD_RUNTIME_DIR:-/etc/mudro/runtime}"
RUNTIME_ENV_FILE="${RUNTIME_DIR}/mudro-api.env"
ENV_FILE="${PROJECT_DIR}/.env"
POSTGRES_PORT_DEFAULT="${POSTGRES_PORT_DEFAULT:-5433}"
MEDIA_ROOT_DEFAULT="${MEDIA_ROOT_DEFAULT:-${PROJECT_DIR}/data/nu}"
NO_RESTART="${NO_RESTART:-0}"

if [[ ! -f "${UNIT_SOURCE}" ]]; then
  echo "[error] missing tracked unit: ${UNIT_SOURCE}" >&2
  exit 1
fi

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "[error] missing env file: ${ENV_FILE}" >&2
  exit 1
fi

read_env_value() {
  local key="$1"
  python3 - "$ENV_FILE" "$key" <<'PY'
from pathlib import Path
import sys

env_path = Path(sys.argv[1])
key = sys.argv[2]
prefix = key + "="

for line in env_path.read_text(encoding="utf-8").splitlines():
    if line.startswith(prefix):
        print(line[len(prefix):].strip())
        break
PY
}

POSTGRES_PASSWORD="$(read_env_value POSTGRES_PASSWORD)"
POSTGRES_PORT="$(read_env_value POSTGRES_PORT)"

if [[ -z "${POSTGRES_PASSWORD}" ]]; then
  echo "[error] POSTGRES_PASSWORD is missing in ${ENV_FILE}" >&2
  exit 1
fi

if [[ -z "${POSTGRES_PORT}" ]]; then
  POSTGRES_PORT="${POSTGRES_PORT_DEFAULT}"
fi

install -d -m 755 "${RUNTIME_DIR}"
install -m 644 "${UNIT_SOURCE}" "${UNIT_DEST}"

if [[ ! -f "${RUNTIME_ENV_FILE}" ]]; then
  cat > "${RUNTIME_ENV_FILE}" <<EOF
DSN=postgres://postgres:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/gallery?sslmode=disable
API_ADDR=:8080
MEDIA_ROOT=${MEDIA_ROOT_DEFAULT}
EOF
  chmod 600 "${RUNTIME_ENV_FILE}"
  echo "[ok] created ${RUNTIME_ENV_FILE}"
else
  echo "[ok] keeping existing ${RUNTIME_ENV_FILE}"
fi

if [[ -f /etc/systemd/system/mudro-api.service.d/10-dsn.conf ]]; then
  rm -f /etc/systemd/system/mudro-api.service.d/10-dsn.conf
  rmdir /etc/systemd/system/mudro-api.service.d 2>/dev/null || true
  echo "[ok] removed legacy DSN override"
fi

systemctl daemon-reload
systemctl enable mudro-api.service >/dev/null 2>&1 || true

if [[ "${NO_RESTART}" != "1" ]]; then
  systemctl restart mudro-api.service
  curl -fsS http://127.0.0.1:8080/healthz >/dev/null
  echo "[ok] mudro-api restarted and healthz passed"
else
  echo "[ok] systemd reloaded without restart"
fi
