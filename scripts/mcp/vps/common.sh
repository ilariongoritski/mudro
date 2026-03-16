#!/usr/bin/env bash
set -euo pipefail

mudro_mcp_vps_repo_container() {
  printf '%s\n' "${MUDRO_MCP_REPO_PATH_CONTAINER:-/home/node/work/mudro11}"
}

mudro_mcp_vps_repo_host() {
  printf '%s\n' "${MUDRO_MCP_REPO_PATH_HOST:-/root/projects/mudro}"
}

mudro_mcp_vps_openclaw_home() {
  printf '%s\n' "${MUDRO_MCP_OPENCLAW_HOME:-/home/node/.openclaw}"
}

mudro_mcp_vps_secrets_dir() {
  printf '%s\n' "$(mudro_mcp_vps_openclaw_home)/mcp-secrets"
}

mudro_mcp_load_env_file() {
  local file="$1"
  if [ ! -f "$file" ]; then
    return 0
  fi

  set -a
  # shellcheck disable=SC1090
  . "$file"
  set +a
}

mudro_mcp_require_command() {
  local name="$1"
  if ! command -v "$name" >/dev/null 2>&1; then
    echo "Required command not found in PATH: $name" >&2
    exit 1
  fi
}

mudro_mcp_init_secrets_dir() {
  mkdir -p "$(mudro_mcp_vps_secrets_dir)"
}
