#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

OPENCLAW_USER="${OPENCLAW_USER:-openclaw}"
OPENCLAW_GROUP="${OPENCLAW_GROUP:-openclaw}"
MUDRO_APP_ROOT="${MUDRO_APP_ROOT:-/opt/mudro/app}"
SYSTEMD_DIR="${SYSTEMD_DIR:-/etc/systemd/system}"
RUNTIME_DIR="${RUNTIME_DIR:-/etc/openclaw/runtime}"
STATE_DIR="${STATE_DIR:-/var/lib/openclaw}"
LOG_DIR="${LOG_DIR:-/var/log/openclaw}"
NO_START="${NO_START:-0}"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "run as root" >&2
  exit 1
fi

ensure_user() {
  if ! id -u "${OPENCLAW_USER}" >/dev/null 2>&1; then
    useradd --system --create-home --home-dir "/home/${OPENCLAW_USER}" --shell /bin/bash "${OPENCLAW_USER}"
  fi
}

ensure_directories() {
  install -d -m 0755 "${RUNTIME_DIR}"
  install -d -o "${OPENCLAW_USER}" -g "${OPENCLAW_GROUP}" -m 0755 \
    "${STATE_DIR}" \
    "${STATE_DIR}/claude-orch" \
    "${STATE_DIR}/claude-orch/ledger" \
    "${STATE_DIR}/claude-orch/runs" \
    "${STATE_DIR}/claude-orch/state" \
    "${STATE_DIR}/skaro" \
    "${LOG_DIR}"
}

ensure_executable_scripts() {
  local script
  for script in \
    "${MUDRO_APP_ROOT}/scripts/openclaw/openclaw_gateway_systemd.sh" \
    "${MUDRO_APP_ROOT}/scripts/openclaw/openclaw_install_user.sh" \
    "${MUDRO_APP_ROOT}/scripts/openclaw/openclaw_gateway_user_service.sh" \
    "${MUDRO_APP_ROOT}/scripts/openclaw/openclaw_post_install_checks.sh" \
    "${MUDRO_APP_ROOT}/scripts/openclaw/server_bootstrap_root.sh" \
    "${MUDRO_APP_ROOT}/scripts/openclaw/claudeusageproxy_systemd.sh" \
    "${MUDRO_APP_ROOT}/scripts/skaro/skaro_ui_linux.sh"; do
    if [[ -f "${script}" ]]; then
      chmod 0755 "${script}"
    fi
  done
}

install_unit() {
  local src="$1"
  local dst="$2"
  install -m 0644 "${src}" "${dst}"
}

install_env_if_missing() {
  local example_name="$1"
  local env_name="$2"
  local src="${PROJECT_DIR}/ops/systemd/${example_name}"
  local dst="${RUNTIME_DIR}/${env_name}"

  if [[ ! -f "${dst}" ]]; then
    install -m 0640 -o root -g "${OPENCLAW_GROUP}" "${src}" "${dst}"
  fi
}

fail_if_placeholder() {
  local env_file="$1"
  local key="$2"
  if grep -Eq "^${key}=([[:space:]]*)$" "${env_file}"; then
    echo "missing required value ${key} in ${env_file}" >&2
    exit 1
  fi
}

main() {
  ensure_user
  ensure_directories
  ensure_executable_scripts

  install_unit "${PROJECT_DIR}/ops/systemd/claudeusageproxy.service" "${SYSTEMD_DIR}/claudeusageproxy.service"
  install_unit "${PROJECT_DIR}/ops/systemd/openclaw.service" "${SYSTEMD_DIR}/openclaw.service"
  install_unit "${PROJECT_DIR}/ops/systemd/skaro.service" "${SYSTEMD_DIR}/skaro.service"

  install_env_if_missing "claudeusageproxy.env.example" "claudeusageproxy.env"
  install_env_if_missing "openclaw.env.example" "openclaw.env"
  install_env_if_missing "skaro.env.example" "skaro.env"

  systemctl daemon-reload

  if [[ "${NO_START}" == "1" ]]; then
    echo "installed openclaw/skaro/claudeusageproxy units without starting"
    exit 0
  fi

  fail_if_placeholder "${RUNTIME_DIR}/claudeusageproxy.env" "MUDRO_CLAUDE_API_KEY"
  fail_if_placeholder "${RUNTIME_DIR}/openclaw.env" "OPENCLAW_GATEWAY_TOKEN"
  fail_if_placeholder "${RUNTIME_DIR}/skaro.env" "CLAUDE_API_KEY"

  systemctl enable claudeusageproxy.service openclaw.service skaro.service
  systemctl restart claudeusageproxy.service
  systemctl restart openclaw.service
  systemctl restart skaro.service
}

main "$@"
