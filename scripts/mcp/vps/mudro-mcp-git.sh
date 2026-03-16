#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
. "$script_dir/common.sh"

mudro_mcp_require_command docker

host_repo="$(mudro_mcp_vps_repo_host)"
image="mudro-mcp-git:2026.1.14"
dockerfile="$host_repo/scripts/mcp/docker/git-server/Dockerfile"

if ! docker image inspect "$image" >/dev/null 2>&1; then
  docker build --tag "$image" --file "$dockerfile" "$host_repo"
fi

exec docker run --rm -i --network none \
  --mount "type=bind,src=$host_repo,dst=/workspace,readonly" \
  "$image" --repository /workspace
