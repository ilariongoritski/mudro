#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
. "$script_dir/common.sh"

mudro_mcp_require_command npx
mudro_mcp_init_secrets_dir

env_file="$(mudro_mcp_vps_secrets_dir)/postgres-readonly.env"
mudro_mcp_load_env_file "$env_file"

if [ -z "${MUDRO_MCP_VPS_DSN:-}" ]; then
  cat >&2 <<'EOF'
MUDRO_MCP_VPS_DSN is not set.
Create /home/node/.openclaw/mcp-secrets/postgres-readonly.env with:
MUDRO_MCP_VPS_DSN=postgres://user:pass@host:5432/db?sslmode=disable
EOF
  exit 1
fi

exec npx -y @modelcontextprotocol/server-postgres@0.6.2 "$MUDRO_MCP_VPS_DSN"
