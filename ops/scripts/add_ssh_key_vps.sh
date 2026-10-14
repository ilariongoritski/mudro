#!/usr/bin/env bash
set -euo pipefail

# Use SSH key authentication for the connection. Never embed passwords in this script.
# Verify the server fingerprint and populate ~/.ssh/known_hosts before running.
: "${VPS_HOST:?Set VPS_HOST to the server hostname or IP}"
: "${SSH_PUBLIC_KEY_FILE:?Set SSH_PUBLIC_KEY_FILE to the public-key file path}"
VPS_USER="${VPS_USER:-root}"

if [[ ! -r "$SSH_PUBLIC_KEY_FILE" ]]; then
  echo "Cannot read public key: $SSH_PUBLIC_KEY_FILE" >&2
  exit 1
fi

ssh -o StrictHostKeyChecking=yes "$VPS_USER@$VPS_HOST" \
  'mkdir -p ~/.ssh && chmod 700 ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo "SSH public key installed."' \
  < "$SSH_PUBLIC_KEY_FILE"
