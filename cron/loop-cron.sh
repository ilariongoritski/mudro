#!/bin/bash
set -euo pipefail
export PATH="/usr/local/bin:/usr/bin:/bin"
ENV_FILE="/opt/mudro/.env.loop"
if [[ -f "${ENV_FILE}" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  set +a
fi
export REPORT_BOT_TOKEN="${REPORT_BOT_TOKEN:-}"
export REPORT_CHAT_ID="${REPORT_CHAT_ID:-}"
exec /opt/mudro/scripts/loop-cycle.sh
