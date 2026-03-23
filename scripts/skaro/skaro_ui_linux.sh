#!/usr/bin/env bash
set -euo pipefail

RUNTIME_ENV="${SKARO_RUNTIME_ENV:-/etc/openclaw/runtime/skaro.env}"

if [[ -f "${RUNTIME_ENV}" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "${RUNTIME_ENV}"
  set +a
fi

SKARO_BIN="${SKARO_BIN:-$(command -v skaro || true)}"
SKARO_PORT="${SKARO_PORT:-4700}"
export HOME="${HOME:-/home/openclaw}"
export PYTHONUTF8=1
export PYTHONIOENCODING=utf-8
export TZ="${TZ:-Europe/Moscow}"

if [[ -z "${SKARO_BIN}" || ! -x "${SKARO_BIN}" ]]; then
  echo "[error] skaro binary not found. Set SKARO_BIN in ${RUNTIME_ENV}." >&2
  exit 1
fi

if [[ -n "${ANTHROPIC_API_KEY:-}" && -z "${CLAUDE_API_KEY:-}" ]]; then
  export CLAUDE_API_KEY="${ANTHROPIC_API_KEY}"
fi

mkdir -p "${SKARO_LOCAL_ROOT:-/var/lib/openclaw/skaro}"
mkdir -p "${MUDRO_CLAUDE_ACCOUNTING_ROOT:-/var/lib/openclaw/claude-orch}"
mkdir -p "${MUDRO_CLAUDE_RUNS_DIR:-/var/lib/openclaw/claude-orch/runs}"
mkdir -p "${MUDRO_CLAUDE_STATE_DIR:-/var/lib/openclaw/claude-orch/state}"
mkdir -p "$(dirname "${MUDRO_CLAUDE_USAGE_LOG:-/var/lib/openclaw/claude-orch/ledger/usage_log.jsonl}")"

exec "${SKARO_BIN}" ui --port "${SKARO_PORT}" --no-browser
