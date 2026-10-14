#!/usr/bin/env bash
set -euo pipefail

# Creates local demo accounts through the registration API. Registration does
# not grant administrator privileges; never run this against production.
: "${MUDRO_DEMO_ADMIN_LOGIN_PASSWORD:?Set a unique MUDRO_DEMO_ADMIN_LOGIN_PASSWORD}"
: "${MUDRO_DEMO_USER_PASSWORD:?Set a unique MUDRO_DEMO_USER_PASSWORD}"
API_URL="${MUDRO_AUTH_API_URL:-http://127.0.0.1:8081/api/auth/register}"

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required to safely encode JSON request bodies." >&2
  exit 1
fi

for login in admin user; do
  if [[ "$login" == "admin" ]]; then
    password="$MUDRO_DEMO_ADMIN_LOGIN_PASSWORD"
  else
    password="$MUDRO_DEMO_USER_PASSWORD"
  fi

  payload=$(jq -n --arg login "$login" --arg password "$password" '{login: $login, password: $password}')
  curl --fail-with-body --silent --show-error \
    --request POST "$API_URL" \
    --header "Content-Type: application/json" \
    --data "$payload"
  printf '\n'
done
