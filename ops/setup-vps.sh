#!/usr/bin/env bash
set -euo pipefail

echo "=== MUDRO VPS Security-First Setup ==="

sudo apt-get update
sudo apt-get install -y ca-certificates curl git make rsync nginx certbot python3-certbot-nginx ufw fail2ban

echo "Configuring UFW..."
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow http
sudo ufw allow https
sudo ufw --force enable

echo "Installing canonical nginx config..."
sudo install -D -m 644 scripts/ops/mudro.nginx.conf /etc/nginx/sites-available/mudro
sudo ln -sfn /etc/nginx/sites-available/mudro /etc/nginx/sites-enabled/mudro
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl enable nginx >/dev/null 2>&1
sudo systemctl reload nginx

cat <<'EOF'
Next steps:
1. Clone the repository to /srv/mudro and ensure admin owns the tree.
2. Copy env/*.env.example to env/*.env and fill real secrets locally on the VPS.
3. Start the runtime with: docker compose -f docker-compose.prod.yml up -d
4. Verify healthz on 127.0.0.1:8080 and through nginx on 127.0.0.1/healthz
5. Only after SSH key login works, disable root login and password authentication.
6. Run Certbot only after a real domain points to the VPS.
EOF
