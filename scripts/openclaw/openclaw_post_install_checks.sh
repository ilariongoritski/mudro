#!/usr/bin/env bash
set -euo pipefail

# Post-install quick checks for OpenClaw host setup.
# Run as the OpenClaw user.

export PATH="$HOME/.nvm/versions/node/v22.22.1/bin:$PATH"
export XDG_RUNTIME_DIR="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}"

echo "== Versions =="
node -v || true
npm -v || true
openclaw --version || true

echo

echo "== User service =="
systemctl --user daemon-reload || true
systemctl --user status openclaw-gateway --no-pager || true

echo

echo "== OpenClaw status =="
openclaw status || true
openclaw gateway status || true
openclaw gateway probe || true

echo

echo "== Local gateway check =="
curl -fsS http://127.0.0.1:18789/ >/dev/null && echo "Gateway UI: OK" || echo "Gateway UI: not reachable"

echo
echo "If UI is local-only, create tunnel from your PC:"
echo "ssh -N -L 18789:127.0.0.1:18789 <user>@<server-ip>"