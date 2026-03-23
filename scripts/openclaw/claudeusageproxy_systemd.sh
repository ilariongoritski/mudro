#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${PROJECT_DIR:-$(cd "${SCRIPT_DIR}/../.." && pwd)}"
RUNTIME_ENV="${CLAUDEUSAGEPROXY_ENV_FILE:-/etc/openclaw/runtime/claudeusageproxy.env}"

if [[ -f "${RUNTIME_ENV}" ]]; then
  # shellcheck disable=SC1090
  source "${RUNTIME_ENV}"
fi

export HOME="${HOME:-/home/openclaw}"
export MUDRO_CLAUDE_ACCOUNTING_ROOT="${MUDRO_CLAUDE_ACCOUNTING_ROOT:-/var/lib/openclaw/claude-orch}"
export MUDRO_CLAUDE_USAGE_LOG="${MUDRO_CLAUDE_USAGE_LOG:-${MUDRO_CLAUDE_ACCOUNTING_ROOT}/ledger/usage_log.jsonl}"
export MUDRO_CLAUDE_TOKEN_USAGE="${MUDRO_CLAUDE_TOKEN_USAGE:-${MUDRO_CLAUDE_ACCOUNTING_ROOT}/ledger/token_usage.yaml}"
export MUDRO_CLAUDE_ROLE_USAGE="${MUDRO_CLAUDE_ROLE_USAGE:-${MUDRO_CLAUDE_ACCOUNTING_ROOT}/ledger/role_usage.yaml}"

mkdir -p \
  "${MUDRO_CLAUDE_ACCOUNTING_ROOT}" \
  "$(dirname "${MUDRO_CLAUDE_USAGE_LOG}")" \
  "$(dirname "${MUDRO_CLAUDE_TOKEN_USAGE}")" \
  "$(dirname "${MUDRO_CLAUDE_ROLE_USAGE}")"

PROXY_BIN="${CLAUDEUSAGEPROXY_BIN:-${CLAUDE_USAGE_PROXY_BIN:-/opt/mudro/bin/claudeusageproxy}}"
if [[ -x "${PROXY_BIN}" ]]; then
  exec "${PROXY_BIN}"
fi

if command -v go >/dev/null 2>&1; then
  cd "${PROJECT_DIR}"
  exec go run ./cmd/claudeusageproxy
fi

echo "[error] claudeusageproxy binary not found at ${PROXY_BIN} and Go toolchain is unavailable" >&2
exit 1
