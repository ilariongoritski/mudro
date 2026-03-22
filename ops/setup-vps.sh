#!/bin/bash
set -e

echo "=== MUDRO VPS Hardening & Setup ==="

# 1. UFW Firewall Setup
echo "Configuring UFW..."
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh   # Port 22
sudo ufw allow http  # Port 80
sudo ufw allow https # Port 443

# Enable UFW (if not already enabled)
sudo ufw --force enable
echo "UFW configured: Only 22, 80, 443 are open."

# 2. Setup Nginx
echo "Setting up Nginx..."
sudo apt-get update
sudo apt-get install -y nginx certbot python3-certbot-nginx

# Copy provided nginx config to sites-available
# Assuming this script is run from the project root on the VPS
sudo cp ops/nginx/mudro.conf /etc/nginx/sites-available/mudro
sudo ln -sf /etc/nginx/sites-available/mudro /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# Test and reload
sudo nginx -t
sudo systemctl reload nginx

# 3. SSL with Certbot
echo "Installing Let's Encrypt SSL..."
# replace 'mudro.yourdomain.com' and 'your_email@example.com' before running
# sudo certbot --nginx -d mudro.yourdomain.com --non-interactive --agree-tos -m your_email@example.com

echo "Setup Complete! Services should be protected and proxied via Nginx."
