#!/bin/bash
set -euo pipefail

# ═══════════════════════════════════════════════
# MUDRO VPS Deploy Script
# Run as root on fresh Ubuntu 24.04
# Usage: bash deploy.sh
# ═══════════════════════════════════════════════

REPO_DIR="/opt/mudro"
ENV_FILE="/etc/mudro/.env"
DOMAIN=""

echo "🪨 MUDRO VPS Bootstrap"
echo ""

# ── 1. Install dependencies ──
echo "==> Installing Docker, Nginx, Certbot..."
apt update -qq
apt install -y -qq docker.io docker-compose-plugin nginx certbot python3-certbot-nginx ufw git > /dev/null
systemctl enable --now docker
systemctl enable --now nginx

# ── 2. Firewall ──
echo "==> Configuring firewall..."
ufw --force reset
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# ── 3. Clone repo ──
if [ ! -d "$REPO_DIR" ]; then
    echo "==> Cloning repo to $REPO_DIR..."
    git clone https://github.com/goritskimihail/mudro "$REPO_DIR" || {
        echo "ERROR: Cannot clone. Copy repo manually to $REPO_DIR"
        exit 1
    }
fi
cd "$REPO_DIR"

# ── 4. Environment ──
mkdir -p /etc/mudro
if [ ! -f "$ENV_FILE" ]; then
    echo "==> Copying .env.production template..."
    cp .env.production "$ENV_FILE"
    echo "ERROR: Edit $ENV_FILE with real values, then re-run this script."
    echo "  nano $ENV_FILE"
    exit 1
fi

echo "==> Loading environment..."
set -a; . "$ENV_FILE"; set +a

# ── 5. Docker Compose ──
echo "==> Starting Docker Compose stack..."
docker compose -f docker-compose.prod.yml --env-file "$ENV_FILE" up -d

# ── 6. Migrations ──
echo "==> Applying migrations..."
# Main DB
for f in migrations/0*.sql; do
    [ "$f" = "*.down.sql" ] && continue
    [[ "$f" == *down* ]] && continue
    echo "  → $f"
    cat "$f" | docker compose -f docker-compose.prod.yml exec -T db psql -U postgres -d gallery -v ON_ERROR_STOP=1 2>/dev/null || true
done

# Casino DB
for f in services/casino/migrations/0*.sql; do
    [[ "$f" == *down* ]] && continue
    echo "  → $f"
    cat "$f" | docker compose -f docker-compose.prod.yml exec -T casino-db psql -U postgres -d mudro_casino -v ON_ERROR_STOP=1 2>/dev/null || true
done

# ── 7. Build frontend ──
echo "==> Building frontend..."
cd frontend
docker run --rm -v "$REPO_DIR/frontend:/app" -w /app node:24-alpine sh -c "npm ci --silent && npm run build" 2>/dev/null || {
    echo "  (frontend build failed, using pre-built if available)"
}
cd "$REPO_DIR"

# ── 8. Nginx ──
echo "==> Configuring Nginx..."
cp ops/nginx/mudro.conf /etc/nginx/sites-available/mudro.conf
# Replace domain placeholders
if [ -n "$DOMAIN" ]; then
    sed -i "s/yourdomain.com/$DOMAIN/g" /etc/nginx/sites-available/mudro.conf
fi
ln -sf /etc/nginx/sites-available/mudro.conf /etc/nginx/sites-enabled/mudro.conf
rm -f /etc/nginx/sites-enabled/default
nginx -t && systemctl reload nginx

# ── 9. SSL (if domain set) ──
if [ -n "$DOMAIN" ]; then
    echo "==> Requesting SSL certificate..."
    certbot --nginx -d "$DOMAIN" -d "api.$DOMAIN" --non-interactive --agree-tos -m "admin@$DOMAIN" || {
        echo "  (SSL failed — run manually: certbot --nginx -d $DOMAIN)"
    }
fi

# ── 10. Health check ──
echo ""
echo "==> Health checks..."
sleep 5
curl -s http://127.0.0.1:8080/healthz && echo " ✓ feed-api" || echo " ✗ feed-api"
curl -s http://127.0.0.1:8082/healthz && echo " ✓ casino-api" || echo " ✗ casino-api"

echo ""
echo "🪨 MUDRO deployed!"
echo "  Frontend:  http://$(hostname -I | awk '{print $1}')/"
echo "  API:       http://$(hostname -I | awk '{print $1}')/api/"
echo "  Casino TMA: http://$(hostname -I | awk '{print $1}')/tma/casino"
echo ""
echo "Next steps:"
echo "  1. Set domain DNS A record → $(curl -s ifconfig.me)"
echo "  2. Run: certbot --nginx -d yourdomain.com"
echo "  3. Configure BotFather: /newapp → https://yourdomain.com/tma/casino"
echo "  4. (Optional) Install OpenClaw: bash ops/scripts/openclaw_gateway_systemd.sh"
