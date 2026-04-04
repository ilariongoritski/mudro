#!/usr/bin/env bash
# railway-setup-vars.sh
# Sets required environment variables for the mudro service on Railway.
# Usage: RAILWAY_TOKEN=<token> bash scripts/railway-setup-vars.sh

set -euo pipefail

TOKEN="${RAILWAY_TOKEN:-177bf586-4304-46fb-89b4-8e86eb61666d}"
PROJ="f49a785f-629e-4b0c-8983-6b50411db455"
SVC="77ffd7a1-bf9d-429e-8504-16cb19f16326"
ENV="841aef1c-c311-4d19-a019-ecfca8d78114"
API="https://backboard.railway.app/graphql/v2"

gql() {
  curl -sf "$API" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$1"
}

echo "==> Setting env vars for mudro service (diligent-heart / production)"

# Build the JSON payload for variableCollectionUpsert
PAYLOAD=$(cat <<'PAYLOAD'
{
  "query": "mutation UpsertVars($input: VariableCollectionUpsertInput!) { variableCollectionUpsert(input: $input) }",
  "variables": {
    "input": {
      "projectId":     "f49a785f-629e-4b0c-8983-6b50411db455",
      "serviceId":     "77ffd7a1-bf9d-429e-8504-16cb19f16326",
      "environmentId": "841aef1c-c311-4d19-a019-ecfca8d78114",
      "variables": {
        "PORT":         "8080",
        "ADDR":         "0.0.0.0:8080",
        "DATABASE_URL": "${{Postgres.DATABASE_URL}}",
        "GIN_MODE":     "release",
        "LOG_LEVEL":    "info"
      }
    }
  }
}
PAYLOAD
)

RESULT=$(gql "$PAYLOAD")
echo "$RESULT" | grep -q '"variableCollectionUpsert":true' \
  && echo "    OK: variables upserted" \
  || { echo "    FAIL: $RESULT"; exit 1; }

echo ""
echo "==> Verifying variables"
VERIFY=$(cat <<'VERIFY'
{
  "query": "query GetVars($projectId: String!, $serviceId: String!, $environmentId: String!) { variables(projectId: $projectId, serviceId: $serviceId, environmentId: $environmentId) }",
  "variables": {
    "projectId":     "f49a785f-629e-4b0c-8983-6b50411db455",
    "serviceId":     "77ffd7a1-bf9d-429e-8504-16cb19f16326",
    "environmentId": "841aef1c-c311-4d19-a019-ecfca8d78114"
  }
}
VERIFY
)

gql "$VERIFY" | tr ',' '\n' | grep -E '"PORT|ADDR|DATABASE_URL|GIN_MODE|LOG_LEVEL'

echo ""
echo "==> Triggering redeploy"
REDEPLOY=$(cat <<'REDEPLOY'
{
  "query": "mutation Redeploy($id: String!) { serviceInstanceRedeploy(serviceId: $id) }",
  "variables": {
    "id": "77ffd7a1-bf9d-429e-8504-16cb19f16326"
  }
}
REDEPLOY
)

RDRESULT=$(gql "$REDEPLOY")
echo "    Response: $RDRESULT"

echo ""
echo "Done. Check deployment status at:"
echo "  https://railway.app/project/f49a785f-629e-4b0c-8983-6b50411db455"
