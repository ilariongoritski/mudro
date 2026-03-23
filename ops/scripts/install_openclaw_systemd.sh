#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${PROJECT_DIR:-$(cd "${SCRIPT_DIR}/../.." && pwd)}"
MUDRO_APP_ROOT="${MUDRO_APP_ROOT:-/opt/mudro/app}"
RUNTIME_DIR="${OPENCLAW_RUNTIME_DIR:-/etc/openclaw/runtime}"
STATE_DIR="${OPENCLAW_STATE_DIR:-/var/lib/openclaw}"
LOG_DIR="${OPENCLAW_LOG_DIR:-/var/log/openclaw}"
OPENCLAW_USER="${OPENCLAW_USER:-openclaw}"
NO_START="${NO_START:-0}"
INSTALL_OPENCLAW="${INSTALL_OPENCLAW:-0}"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "[error] run as root" >&2
  exit 1
fi

if [[ ! -d "${MUDRO_APP_ROOT}" ]]; then
  echo "[error] missing deployed MUDRO app tree: ${MUDRO_APP_ROOT}" >&2
  exit 1
fi

if ! id -u "${OPENCLAW_USER}" >/dev/null 2>&1; then
  useradd --system --create-home --home-dir "/home/${OPENCLAW_USER}" --shell /usr/sbin/nologin "${OPENCLAW_USER}"
fi

install -d -m 755 "${RUNTIME_DIR}"
install -d -o "${OPENCLAW_USER}" -g "${OPENCLAW_USER}" -m 755 "${STATE_DIR}" "${LOG_DIR}"

install -m 644 "${PROJECT_DIR}/ops/systemd/openclaw.service" /etc/systemd/system/openclaw.service
install -m 644 "${PROJECT_DIR}/ops/systemd/skaro.service" /etc/systemd/system/skaro.service

for env_name in openclaw.env skaro.env; do
  target="${RUNTIME_DIR}/${env_name}"
  source="${PROJECT_DIR}/ops/systemd/${env_name}.example"
  if [[ ! -f "${target}" ]]; then
    install -m 640 "${source}" "${target}"
    echo "[ok] created ${target}"
  else
    echo "[ok] keeping existing ${target}"
  fi
  chown root:"${OPENCLAW_USER}" "${target}"
  chmod 640 "${target}"
done

if [[ "${INSTALL_OPENCLAW}" == "1" ]]; then
  sudo -u "${OPENCLAW_USER}" HOME="/home/${OPENCLAW_USER}" bash "${MUDRO_APP_ROOT}/scripts/openclaw/openclaw_install_user.sh"
fi

if [[ "${NO_START}" != "1" ]]; then
  grep -q '^OPENCLAW_GATEWAY_TOKEN=$' "${RUNTIME_DIR}/openclaw.env" && {
    echo "[error] set OPENCLAW_GATEWAY_TOKEN in ${RUNTIME_DIR}/openclaw.env before starting openclaw" >&2
    exit 1
  }
  if grep -q '^ANTHROPIC_API_KEY=$' "${RUNTIME_DIR}/skaro.env" && grep -q '^CLAUDE_API_KEY=$' "${RUNTIME_DIR}/skaro.env"; then
    echo "[error] set ANTHROPIC_API_KEY or CLAUDE_API_KEY in ${RUNTIME_DIR}/skaro.env before starting skaro" >&2
    exit 1
  fi
fi

systemctl daemon-reload
systemctl enable openclaw.service skaro.service >/dev/null 2>&1 || true

if [[ "${NO_START}" != "1" ]]; then
  if ! systemctl restart openclaw.service; then
    systemctl disable openclaw.service >/dev/null 2>&1 || true
    exit 1
  fi
  if ! systemctl restart skaro.service; then
    systemctl disable skaro.service >/dev/null 2>&1 || true
    exit 1
  fi
fi

echo "[ok] tracked OpenClaw/Skaro systemd install completed"
