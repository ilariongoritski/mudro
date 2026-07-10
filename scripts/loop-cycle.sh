#!/bin/bash
set -euo pipefail
YOLO=false
[[ "${1:-}" == "--yolo" ]] && YOLO=true
DATE=$(date +%Y-%m-%d)
SLUG=$(date +%H%M)
BRANCH="loop/${DATE}-${SLUG}"
ROOT_DIR="/opt/mudro"
AUDIT_DIR="${ROOT_DIR}/docs/audit"
AUDIT_FILE="${AUDIT_DIR}/${DATE}-loop-${SLUG}.md"
mkdir -p "${AUDIT_DIR}"
{
  echo "# Loop Cycle ${DATE}-${SLUG}"
  echo "## Branch: ${BRANCH}"
  echo ""
  echo "### SCAN"
  echo "- Date: $(date -Iseconds)"
  echo "- Services:"
  if services_output=$(docker ps --format "table {{.Names}}\t{{.Status}}" 2>/dev/null); then
    if [[ -n "${services_output}" ]]; then
      i=0
      while IFS= read -r line; do
        [[ -z "${line}" ]] && continue
        echo "${line}"
        i=$((i + 1))
        [[ ${i} -ge 15 ]] && break
      done <<< "${services_output}"
    else
      echo "  docker ps returned no containers"
    fi
  else
    echo "  docker not available"
  fi
  echo ""
  echo "- Healthz checks:"
  if docker_names=$(docker ps --format "{{.Names}}" 2>/dev/null); then
    for svc in mudro-api mudro-bff-web mudro-casino-api; do
      if grep -Eq "^${svc}(-[0-9]+)?$" <<< "${docker_names}"; then
        echo "  - ${svc}: UP"
      else
        echo "  - ${svc}: DOWN"
      fi
    done
  else
    echo "  - docker unavailable"
  fi
  echo ""
  echo "- Git status:"
  if git_status=$(git -C "${ROOT_DIR}" status --short 2>/dev/null); then
    if [[ -n "${git_status}" ]]; then
      i=0
      while IFS= read -r line; do
        [[ -z "${line}" ]] && continue
        echo "${line}"
        i=$((i + 1))
        [[ ${i} -ge 20 ]] && break
      done <<< "${git_status}"
    else
      echo "  clean"
    fi
  else
    echo "  git status unavailable"
  fi
  echo ""
  echo "### PLAN"
  echo "- [STUB] Requires Hermes LLM to identify top-3 improvements"
  echo "- Branch to create: ${BRANCH}"
  echo ""
} > "${AUDIT_FILE}"
if [[ -n "${REPORT_BOT_TOKEN:-}" && -n "${REPORT_CHAT_ID:-}" ]]; then
  curl -s -X POST "https://api.telegram.org/bot${REPORT_BOT_TOKEN}/sendMessage" \
    -d chat_id="${REPORT_CHAT_ID}" \
    -d text="Loop SCAN complete: ${DATE}-${SLUG} Branch: ${BRANCH}" > /dev/null || true
fi
echo "SCAN complete. See ${AUDIT_FILE}"
