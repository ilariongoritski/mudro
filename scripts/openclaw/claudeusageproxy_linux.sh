#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${MUDRO_CLAUDE_PROXY_ENV_FILE:-/etc/openclaw/runtime/claudeusageproxy.env}"
if [[ -f "${ENV_FILE}" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  set +a
fi

: "${MUDRO_ROOT:=/opt/mudro/app}"
: "${GO_BIN:=go}"
: "${CLAUDEUSAGEPROXY_ADDR:=127.0.0.1:8788}"
: "${MUDRO_CLAUDE_USAGE_LOG:=/var/lib/openclaw/claude-orch/ledger/usage_log.jsonl}"
: "${MUDRO_CLAUDE_TOKEN_USAGE:=/var/lib/openclaw/claude-orch/ledger/token_usage.yaml}"
: "${MUDRO_CLAUDE_ROLE_USAGE:=/var/lib/openclaw/claude-orch/ledger/role_usage.yaml}"

mkdir -p \
  "$(dirname "${MUDRO_CLAUDE_USAGE_LOG}")" \
  "$(dirname "${MUDRO_CLAUDE_TOKEN_USAGE}")" \
  "$(dirname "${MUDRO_CLAUDE_ROLE_USAGE}")"

cd "${MUDRO_ROOT}"
exec "${GO_BIN}" run ./cmd/claudeusageproxy
