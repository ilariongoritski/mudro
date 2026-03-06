#!/usr/bin/env bash
set -euo pipefail

# Base host hardening for OpenClaw VPS.
# Run as root on Ubuntu 24.04.

OPENCLAW_USER="${OPENCLAW_USER:-openclaw}"
ADMIN_PUBKEY="${ADMIN_PUBKEY:-}"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root"
  exit 1
fi

export DEBIAN_FRONTEND=noninteractive

apt update && apt upgrade -y
apt install -y curl git jq fail2ban ufw ca-certificates

if ! id -u "$OPENCLAW_USER" >/dev/null 2>&1; then
  adduser --disabled-password --gecos "" "$OPENCLAW_USER"
fi
usermod -aG sudo "$OPENCLAW_USER"
echo "$OPENCLAW_USER ALL=(ALL) NOPASSWD:ALL" >/etc/sudoers.d/"$OPENCLAW_USER"
chmod 440 /etc/sudoers.d/"$OPENCLAW_USER"

install -d -m 700 -o "$OPENCLAW_USER" -g "$OPENCLAW_USER" "/home/$OPENCLAW_USER/.ssh"
if [[ -n "$ADMIN_PUBKEY" ]]; then
  echo "$ADMIN_PUBKEY" >"/home/$OPENCLAW_USER/.ssh/authorized_keys"
  chown "$OPENCLAW_USER:$OPENCLAW_USER" "/home/$OPENCLAW_USER/.ssh/authorized_keys"
  chmod 600 "/home/$OPENCLAW_USER/.ssh/authorized_keys"
fi

cp /etc/ssh/sshd_config /etc/ssh/sshd_config.bak.$(date +%s)
sed -i 's/^#\?PermitRootLogin .*/PermitRootLogin no/' /etc/ssh/sshd_config
sed -i 's/^#\?PasswordAuthentication .*/PasswordAuthentication no/' /etc/ssh/sshd_config
systemctl restart ssh

cp /etc/systemd/resolved.conf /etc/systemd/resolved.conf.bak.$(date +%s) || true
cat >/etc/systemd/resolved.conf <<'EOF'
[Resolve]
DNS=9.9.9.9 8.8.8.8 1.1.1.1
FallbackDNS=
EOF
systemctl restart systemd-resolved

sed -i 's/^#\?SystemMaxUse=.*/SystemMaxUse=100M/' /etc/systemd/journald.conf
systemctl restart systemd-journald
journalctl --vacuum-size=100M || true

ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

cat >/etc/fail2ban/jail.local <<'EOF'
[DEFAULT]
banaction = ufw
maxretry = 3
findtime = 3600
bantime = 86400

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
EOF
systemctl enable --now fail2ban

echo "Bootstrap done."
echo "Next: sudo -i -u $OPENCLAW_USER bash -lc 'scripts/openclaw/openclaw_install_user.sh'"
