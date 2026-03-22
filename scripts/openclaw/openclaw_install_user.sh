#!/usr/bin/env bash
set -euo pipefail

# Install OpenClaw for a non-root Linux user.
# Run as the target user (for example: openclaw).

if [[ "$(id -u)" -eq 0 ]]; then
  echo "Do not run as root"
  exit 1
fi

export NVM_DIR="$HOME/.nvm"

if [[ ! -s "$NVM_DIR/nvm.sh" ]]; then
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
fi

# shellcheck disable=SC1090
source "$NVM_DIR/nvm.sh"

nvm install 22
nvm alias default 22

curl -fsSL https://openclaw.ai/install.sh | bash

echo
echo "OpenClaw installed."
echo "Optional onboarding step:"
echo "  openclaw doctor --generate-gateway-token --non-interactive --yes"
echo "  openclaw onboard --install-daemon"
echo
echo "To install and start the gateway as a user service:"
echo "  bash scripts/openclaw/openclaw_gateway_user_service.sh"
echo
echo "After setup:"
echo "  bash scripts/openclaw/openclaw_post_install_checks.sh"