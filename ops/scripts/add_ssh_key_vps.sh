#!/usr/bin/env bash
set -euo pipefail

# Set these in your local shell; never store the password in this script:
#   VPS_HOST=your-server.example
#   VPS_USER=root
#   SSHPASS=...                 # only if password-based SSH is required
#   SSH_PUBLIC_KEY_FILE=~/.ssh/id_ed25519.pub
#
# Verify the host fingerprint and populate ~/.ssh/known_hosts before running.

: "${VPS_HOST:?Set VPS_HOST to the server hostname or IP}"
: "${SSHPASS:?Set SSHPASS in the environment (or use SSH key authentication instead)}"
: "${SSH_PUBLIC_KEY_FILE:?Set SSH_PUBLIC_KEY_FILE to the public-key file path}"

VPS_USER="${VPS_USER:-root}"

if ! command -v sshpass >/dev/null 2>&1; then
  echo "sshpass is required for password-based SSH; install it or use SSH key authentication." >&2
  exit 1
fi

if [[ ! -r "$SSH_PUBLIC_KEY_FILE" ]]; then
  echo "Cannot read public key: $SSH_PUBLIC_KEY_FILE" >&2
  exit 1
fi

sshpass -e ssh -o StrictHostKeyChecking=yes "$VPS_USER@$VPS_HOST" \
  'mkdir -p ~/.ssh && chmod 700 ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo "SSH public key installed."' \
  < "$SSH_PUBLIC_KEY_FILE"
