#!/usr/bin/env bash
set -euo pipefail

RUNTIME_ENV="${OPENCLAW_RUNTIME_ENV:-/etc/openclaw/runtime/openclaw.env}"

if [[ -f "${RUNTIME_ENV}" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "${RUNTIME_ENV}"
  set +a
fi

export HOME="${HOME:-/home/openclaw}"
export NVM_DIR="${NVM_DIR:-${HOME}/.nvm}"

if [[ -s "${NVM_DIR}/nvm.sh" ]]; then
  # shellcheck disable=SC1090
  source "${NVM_DIR}/nvm.sh"
fi

OPENCLAW_BIN="${OPENCLAW_BIN:-$(command -v openclaw || true)}"
OPENCLAW_GATEWAY_PORT="${OPENCLAW_GATEWAY_PORT:-18789}"
OPENCLAW_GATEWAY_BIND="${OPENCLAW_GATEWAY_BIND:-loopback}"

if [[ -z "${OPENCLAW_BIN}" || ! -x "${OPENCLAW_BIN}" ]]; then
  echo "[error] openclaw binary not found. Set OPENCLAW_BIN in ${RUNTIME_ENV}." >&2
  exit 1
fi

if [[ -z "${OPENCLAW_GATEWAY_TOKEN:-}" ]]; then
  echo "[error] OPENCLAW_GATEWAY_TOKEN is required in ${RUNTIME_ENV}." >&2
  exit 1
fi

exec "${OPENCLAW_BIN}" gateway run \
  --port "${OPENCLAW_GATEWAY_PORT}" \
  --token "${OPENCLAW_GATEWAY_TOKEN}" \
  --bind "${OPENCLAW_GATEWAY_BIND}" \
  --auth token \
  --force
