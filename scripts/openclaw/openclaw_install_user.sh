#!/usr/bin/env bash
set -euo pipefail

# Install OpenClaw for a non-root Linux user.
# Run as target user (for example: openclaw).

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
echo "Run interactive onboarding now:"
echo "  openclaw onboard --install-daemon"
echo
echo "After onboarding:"
echo "  openclaw status"
echo "  systemctl --user status openclaw-gateway"
