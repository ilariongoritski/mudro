#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${PROJECT_DIR:-$(cd "${SCRIPT_DIR}/../.." && pwd)}"
DEPLOY_ROOT="${MUDRO_DEPLOY_ROOT:-/opt/mudro}"
APP_DIR="${DEPLOY_ROOT}/app"
BIN_DIR="${DEPLOY_ROOT}/bin"
RUNTIME_DIR="${MUDRO_SYSTEMD_RUNTIME_DIR:-/etc/mudro/runtime}"
STATE_DIR="${MUDRO_STATE_DIR:-/var/lib/mudro}"
LOG_DIR="${MUDRO_LOG_DIR:-/var/log/mudro}"
SYSTEM_USER="${MUDRO_SYSTEM_USER:-mudro}"
ONLY="${ONLY:-all}"
NO_START="${NO_START:-0}"

ensure_root() {
  if [[ "$(id -u)" -ne 0 ]]; then
    echo "[error] run as root" >&2
    exit 1
  fi
}

has_service() {
  case "${ONLY}" in
    all) return 0 ;;
    api) [[ "$1" == "api" ]] ;;
    bot) [[ "$1" == "bot" ]] ;;
    agent) [[ "$1" == "agent" ]] ;;
    *) return 1 ;;
  esac
}

ensure_system_user() {
  if ! id -u "${SYSTEM_USER}" >/dev/null 2>&1; then
    useradd --system --home-dir "${STATE_DIR}" --create-home --shell /usr/sbin/nologin "${SYSTEM_USER}"
  fi
  if getent group docker >/dev/null 2>&1; then
    usermod -aG docker "${SYSTEM_USER}" >/dev/null 2>&1 || true
  fi
}

ensure_directories() {
  install -d -m 755 "${DEPLOY_ROOT}" "${APP_DIR}" "${BIN_DIR}"
  install -d -m 755 "${RUNTIME_DIR}"
  install -d -o "${SYSTEM_USER}" -g "${SYSTEM_USER}" -m 755 "${STATE_DIR}" "${LOG_DIR}"
}

sync_repo_tree() {
  local tmp_dir
  tmp_dir="$(mktemp -d)"
  tar \
    --exclude=".git" \
    --exclude="data" \
    --exclude="out" \
    --exclude="var/log" \
    --exclude=".env" \
    --exclude="frontend/node_modules" \
    --exclude="node_modules" \
    -C "${PROJECT_DIR}" -cf - . | tar -C "${tmp_dir}" -xf -
  shopt -s dotglob nullglob
  rm -rf "${APP_DIR:?}/"*
  shopt -u dotglob nullglob
  cp -a "${tmp_dir}/." "${APP_DIR}/"
  rm -rf "${tmp_dir}"
  chown -R root:root "${APP_DIR}"
  if [[ -d "${APP_DIR}/.codex" ]]; then
    chown -R "${SYSTEM_USER}:${SYSTEM_USER}" "${APP_DIR}/.codex"
  fi
}

ensure_executable_scripts() {
  local script
  for script in \
    "${APP_DIR}/ops/scripts/install_mudro_systemd.sh" \
    "${APP_DIR}/ops/scripts/install_openclaw_systemd.sh" \
    "${APP_DIR}/ops/scripts/install_mudro_api_systemd.sh" \
    "${APP_DIR}/ops/scripts/harden_vps_db_auth.sh" \
    "${APP_DIR}/scripts/openclaw/openclaw_gateway_systemd.sh" \
    "${APP_DIR}/scripts/skaro/skaro_ui_linux.sh" \
    "${APP_DIR}/scripts/openclaw/openclaw_install_user.sh" \
    "${APP_DIR}/scripts/openclaw/openclaw_gateway_user_service.sh" \
    "${APP_DIR}/scripts/openclaw/openclaw_post_install_checks.sh" \
    "${APP_DIR}/scripts/openclaw/server_bootstrap_root.sh"; do
    if [[ -f "${script}" ]]; then
      chmod 755 "${script}"
    fi
  done
}

build_binary() {
  local name="$1"
  local target="$2"
  local out="${BIN_DIR}/${name}"
  (
    cd "${APP_DIR}"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${out}" "${target}"
  )
  chmod 755 "${out}"
}

install_unit() {
  local file="$1"
  install -m 644 "${PROJECT_DIR}/ops/systemd/${file}" "/etc/systemd/system/${file}"
}

ensure_runtime_env() {
  local env_name="$1"
  local example="${PROJECT_DIR}/ops/systemd/${env_name}.example"
  local target="${RUNTIME_DIR}/${env_name%.example}"
  if [[ ! -f "${target}" ]]; then
    install -m 640 "${example}" "${target}"
    chown root:"${SYSTEM_USER}" "${target}"
    echo "[ok] created ${target}"
  else
    chown root:"${SYSTEM_USER}" "${target}"
    chmod 640 "${target}"
    echo "[ok] keeping existing ${target}"
  fi
}

validate_runtime_env() {
  local service="$1"
  local target
  case "${service}" in
    api)
      target="${RUNTIME_DIR}/mudro-api.env"
      grep -q '<APP_PASSWORD>' "${target}" && {
        echo "[error] replace <APP_PASSWORD> in ${target} before starting api" >&2
        return 1
      }
      grep -q '^JWT_SECRET=change-me$' "${target}" && {
        echo "[error] replace JWT_SECRET in ${target} before starting api" >&2
        return 1
      }
      ;;
    bot)
      target="${RUNTIME_DIR}/mudro-bot.env"
      grep -q '<APP_PASSWORD>' "${target}" && {
        echo "[error] replace <APP_PASSWORD> in ${target} before starting bot" >&2
        return 1
      }
      grep -q '^TELEGRAM_BOT_TOKEN=$' "${target}" && {
        echo "[error] set TELEGRAM_BOT_TOKEN in ${target} before starting bot" >&2
        return 1
      }
      grep -q '^TELEGRAM_ALLOWED_USERNAME=$' "${target}" && {
        echo "[error] set TELEGRAM_ALLOWED_USERNAME in ${target} before starting bot" >&2
        return 1
      }
      ;;
    agent)
      target="${RUNTIME_DIR}/mudro-agent.env"
      grep -q '<APP_PASSWORD>' "${target}" && {
        echo "[error] replace <APP_PASSWORD> in ${target} before starting agent" >&2
        return 1
      }
      ;;
  esac
}

enable_start() {
  local unit="$1"
  systemctl enable "${unit}" >/dev/null 2>&1 || true
  if [[ "${NO_START}" != "1" ]]; then
    if ! systemctl restart "${unit}"; then
      systemctl disable "${unit}" >/dev/null 2>&1 || true
      return 1
    fi
  fi
}

ensure_root
ensure_system_user
ensure_directories
sync_repo_tree
ensure_executable_scripts

if has_service api; then
  build_binary "mudro-api" "./services/feed-api/cmd"
  install_unit "mudro-api.service"
  ensure_runtime_env "mudro-api.env.example"
fi

if has_service bot; then
  build_binary "mudro-bot" "./services/bot/cmd"
  install_unit "mudro-bot.service"
  ensure_runtime_env "mudro-bot.env.example"
fi

if has_service agent; then
  build_binary "mudro-agent" "./services/agent/cmd"
  install_unit "mudro-agent-worker.service"
  install_unit "mudro-agent-planner.service"
  install_unit "mudro-agent-planner.timer"
  ensure_runtime_env "mudro-agent.env.example"
fi

systemctl daemon-reload

if has_service api; then
  validate_runtime_env api
  enable_start "mudro-api.service"
fi

if has_service bot; then
  validate_runtime_env bot
  enable_start "mudro-bot.service"
fi

if has_service agent; then
  validate_runtime_env agent
  systemctl enable mudro-agent-planner.timer >/dev/null 2>&1 || true
  if [[ "${NO_START}" != "1" ]]; then
    if ! systemctl restart mudro-agent-worker.service; then
      systemctl disable mudro-agent-worker.service >/dev/null 2>&1 || true
      exit 1
    fi
    if ! systemctl restart mudro-agent-planner.timer; then
      systemctl disable mudro-agent-planner.timer >/dev/null 2>&1 || true
      exit 1
    fi
    systemctl start mudro-agent-planner.service || true
  fi
fi

echo "[ok] tracked MUDRO systemd install completed (only=${ONLY}, no_start=${NO_START})"
