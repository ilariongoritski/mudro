#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

APP_ROOT="${MUDRO_APP_ROOT:-/opt/mudro/app}"
SYSTEMD_DIR="${SYSTEMD_DIR:-/etc/systemd/system}"
RUNTIME_DIR="${OPENCLAW_RUNTIME_DIR:-/etc/openclaw/runtime}"
SERVICE_NAME="claudeusageproxy.service"
SERVICE_USER="${OPENCLAW_SERVICE_USER:-openclaw}"
SERVICE_GROUP="${OPENCLAW_SERVICE_GROUP:-openclaw}"

install -d -m 0755 "${RUNTIME_DIR}" /var/lib/openclaw/claude-orch/ledger /var/log/openclaw

if ! id -u "${SERVICE_USER}" >/dev/null 2>&1; then
  useradd --system --create-home --home-dir /home/openclaw --shell /usr/sbin/nologin "${SERVICE_USER}"
fi

chown -R "${SERVICE_USER}:${SERVICE_GROUP}" /var/lib/openclaw /var/log/openclaw

install -m 0644 "${REPO_ROOT}/ops/systemd/${SERVICE_NAME}" "${SYSTEMD_DIR}/${SERVICE_NAME}"
install -m 0755 "${REPO_ROOT}/scripts/openclaw/claudeusageproxy_linux.sh" "${APP_ROOT}/scripts/openclaw/claudeusageproxy_linux.sh"

if [[ ! -f "${RUNTIME_DIR}/claudeusageproxy.env" ]]; then
  install -m 0640 "${REPO_ROOT}/ops/systemd/claudeusageproxy.env.example" "${RUNTIME_DIR}/claudeusageproxy.env"
fi

systemctl daemon-reload
systemctl enable "${SERVICE_NAME}"

if ! grep -q "__SET_REAL_KEY__" "${RUNTIME_DIR}/claudeusageproxy.env"; then
  systemctl restart "${SERVICE_NAME}"
else
  echo "claudeusageproxy.env still contains placeholder values; service enabled but not started."
fi
