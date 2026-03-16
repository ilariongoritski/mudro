#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
. "$script_dir/common.sh"

mudro_mcp_require_command npx
mudro_mcp_init_secrets_dir

env_file="$(mudro_mcp_vps_secrets_dir)/github-readonly.env"
mudro_mcp_load_env_file "$env_file"

if [ -z "${MUDRO_MCP_VPS_GITHUB_PAT:-}" ]; then
  cat >&2 <<'EOF'
MUDRO_MCP_VPS_GITHUB_PAT is not set.
Create /home/node/.openclaw/mcp-secrets/github-readonly.env with a fine-grained read-only PAT limited to mudro.
EOF
  exit 1
fi

export GITHUB_PERSONAL_ACCESS_TOKEN="$MUDRO_MCP_VPS_GITHUB_PAT"
exec npx -y @modelcontextprotocol/server-github@2025.4.8
