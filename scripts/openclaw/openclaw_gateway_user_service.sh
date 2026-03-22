#!/usr/bin/env bash
set -euo pipefail

if [[ "$(id -u)" -eq 0 ]]; then
  echo "Do not run as root"
  exit 1
fi

OPENCLAW_BIN="${OPENCLAW_BIN:-$HOME/.nvm/versions/node/v22.22.1/bin/openclaw}"
NODE_BIN_DIR="${NODE_BIN_DIR:-$HOME/.nvm/versions/node/v22.22.1/bin}"
OPENCLAW_DIR="$HOME/.openclaw"
SERVICE_DIR="$HOME/.config/systemd/user"
SERVICE_FILE="$SERVICE_DIR/openclaw-gateway.service"
LAUNCHER="$HOME/.local/bin/openclaw-gateway-launch.sh"
RUNTIME_DIR="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}"

if [[ ! -x "$OPENCLAW_BIN" ]]; then
  echo "openclaw binary not found: $OPENCLAW_BIN"
  exit 1
fi

install -d -m 700 "$OPENCLAW_DIR" "$SERVICE_DIR" "$HOME/.local/bin"
install -d -m 700 "$OPENCLAW_DIR/agents/main/sessions"

"$OPENCLAW_BIN" config set gateway.mode local >/dev/null
"$OPENCLAW_BIN" config set gateway.bind loopback >/dev/null

TOKEN=$(python3 - <<'PY'
import json
import pathlib
import sys

path = pathlib.Path.home() / '.openclaw' / 'openclaw.json'
if not path.exists():
    sys.exit('openclaw.json not found')

data = json.loads(path.read_text())
token = data.get('gateway', {}).get('auth', {}).get('token')
if not token:
    sys.exit('gateway token not found in openclaw.json')

print(token)
PY
)

cat > "$OPENCLAW_DIR/gateway.env" <<EOF
OPENCLAW_GATEWAY_TOKEN=$TOKEN
EOF
chmod 600 "$OPENCLAW_DIR/gateway.env"

cat > "$LAUNCHER" <<EOF
#!/usr/bin/env bash
set -euo pipefail
export PATH="$NODE_BIN_DIR:$PATH"
source "$OPENCLAW_DIR/gateway.env"
exec "$OPENCLAW_BIN" gateway run --port 18789 --token "$OPENCLAW_GATEWAY_TOKEN" --bind loopback --auth token --force
EOF
chmod 700 "$LAUNCHER"

cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=OpenClaw Gateway
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$LAUNCHER
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
EOF
chmod 600 "$SERVICE_FILE"

export XDG_RUNTIME_DIR="$RUNTIME_DIR"
if systemctl --user daemon-reload && systemctl --user enable --now openclaw-gateway; then
  echo "OpenClaw gateway service enabled"
else
  echo "systemctl --user unavailable; falling back to nohup"
  nohup "$LAUNCHER" >/tmp/openclaw-gateway.log 2>&1 &
  sleep 2
fi

"$OPENCLAW_BIN" gateway status