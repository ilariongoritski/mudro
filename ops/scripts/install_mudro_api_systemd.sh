#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${PROJECT_DIR:-$(cd "${SCRIPT_DIR}/../.." && pwd)}"
RUNTIME_DIR="${MUDRO_SYSTEMD_RUNTIME_DIR:-/etc/mudro/runtime}"
RUNTIME_ENV_FILE="${RUNTIME_DIR}/mudro-api.env"
ENV_FILE="${PROJECT_DIR}/.env"
POSTGRES_PORT_DEFAULT="${POSTGRES_PORT_DEFAULT:-5433}"
MEDIA_ROOT_DEFAULT="${MEDIA_ROOT_DEFAULT:-/var/lib/mudro/media}"
NO_RESTART="${NO_RESTART:-0}"

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

if [[ ! -f "${RUNTIME_ENV_FILE}" ]]; then
  if [[ ! -f "${ENV_FILE}" ]]; then
    echo "[error] missing env file: ${ENV_FILE}" >&2
    exit 1
  fi

  EXISTING_DSN="$(read_env_value DSN)"
  POSTGRES_PASSWORD="$(read_env_value POSTGRES_PASSWORD)"
  POSTGRES_PORT="$(read_env_value POSTGRES_PORT)"

  if [[ -z "${EXISTING_DSN}" && -z "${POSTGRES_PASSWORD}" ]]; then
    echo "[error] POSTGRES_PASSWORD is missing in ${ENV_FILE}" >&2
    exit 1
  fi

  if [[ -z "${POSTGRES_PORT}" ]]; then
    POSTGRES_PORT="${POSTGRES_PORT_DEFAULT}"
  fi

  install -d -m 755 "${RUNTIME_DIR}"
  if [[ -n "${EXISTING_DSN}" ]]; then
    TARGET_DSN="${EXISTING_DSN}"
    TARGET_ENV="production"
  else
    TARGET_DSN="postgres://postgres:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/gallery?sslmode=disable"
    TARGET_ENV="development"
  fi
  cat > "${RUNTIME_ENV_FILE}" <<EOF
MUDRO_ENV=${TARGET_ENV}
MUDRO_ROOT=/opt/mudro/app
DSN=${TARGET_DSN}
API_ADDR=:8080
JWT_SECRET=change-me
CASINO_SERVICE_URL=http://127.0.0.1:8081
MEDIA_ROOT=${MEDIA_ROOT_DEFAULT}
EOF
  chmod 600 "${RUNTIME_ENV_FILE}"
  echo "[ok] created ${RUNTIME_ENV_FILE}"
else
  echo "[ok] keeping existing ${RUNTIME_ENV_FILE}"
fi

ONLY=api NO_START="${NO_RESTART}" bash "${SCRIPT_DIR}/install_mudro_systemd.sh"
