#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
. "$script_dir/common.sh"

mudro_mcp_require_command npx

repo_path="$(mudro_mcp_vps_repo_container)"
exec npx -y @modelcontextprotocol/server-filesystem@2026.1.14 "$repo_path"
