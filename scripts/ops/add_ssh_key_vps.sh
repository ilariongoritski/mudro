#!/usr/bin/env bash
set -euo pipefail

VPS_HOST="${MUDRO_VPS_HOST:?MUDRO_VPS_HOST is required}"
VPS_USER="${MUDRO_VPS_USER:-root}"
SSH_PASSWORD="${MUDRO_SSH_PASSWORD:?MUDRO_SSH_PASSWORD is required}"
PUBLIC_KEY="${MUDRO_VPS_PUBLIC_KEY:?MUDRO_VPS_PUBLIC_KEY is required}"

if ! command -v sshpass >/dev/null 2>&1; then
  echo "sshpass is required" >&2
  exit 1
fi

export SSHPASS="${SSH_PASSWORD}"
printf '%s\n' "${PUBLIC_KEY}" | sshpass -e ssh -o StrictHostKeyChecking=accept-new "${VPS_USER}@${VPS_HOST}" \
  'mkdir -p ~/.ssh && chmod 700 ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo "KEY ADDED SUCCESSFULLY"'

echo "[next] verify SSH key login before disabling root login and password authentication"
