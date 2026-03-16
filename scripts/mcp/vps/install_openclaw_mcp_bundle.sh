#!/usr/bin/env bash
set -euo pipefail

repo_root="${1:-/root/projects/mudro}"
openclaw_home="${2:-/root/.openclaw}"
secrets_dir="$openclaw_home/mcp-secrets"
git_image="mudro-mcp-git:2026.1.14"
runtime_uid="${MUDRO_MCP_OPENCLAW_UID:-1000}"
runtime_gid="${MUDRO_MCP_OPENCLAW_GID:-1000}"

mkdir -p "$secrets_dir"
chmod 700 "$openclaw_home" "$secrets_dir" || true

postgres_env="$secrets_dir/postgres-readonly.env"
if [ ! -f "$postgres_env" ]; then
  cat > "$postgres_env" <<'EOF'
MUDRO_MCP_VPS_DSN=postgres://postgres:postgres@91.218.113.247:5433/gallery?sslmode=disable
EOF
  chmod 600 "$postgres_env"
fi

github_example="$secrets_dir/github-readonly.env.example"
if [ ! -f "$github_example" ]; then
  cat > "$github_example" <<'EOF'
# Fine-grained read-only PAT limited to mudro
MUDRO_MCP_VPS_GITHUB_PAT=
EOF
fi

chmod 600 "$postgres_env" "$github_example"
chown -R "$runtime_uid:$runtime_gid" "$secrets_dir"

docker build --tag "$git_image" --file "$repo_root/scripts/mcp/docker/git-server/Dockerfile" "$repo_root"

cat <<'EOF'
Mudro OpenClaw MCP bundle is prepared.

Available wrapper scripts inside the repo:
- scripts/mcp/vps/mudro-mcp-filesystem.sh
- scripts/mcp/vps/mudro-mcp-git.sh
- scripts/mcp/vps/mudro-mcp-postgres.sh
- scripts/mcp/vps/mudro-mcp-github.sh

Next step:
1. Add OpenClaw runtime entries that call these wrappers.
2. Fill /root/.openclaw/mcp-secrets/github-readonly.env if GitHub read-only MCP is needed.
EOF
