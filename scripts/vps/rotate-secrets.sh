#!/bin/bash
# rotate-secrets.sh - Helper script to rotate Mudro secrets on VPS

ENV_FILE=".env"
BACKUP_FILE=".env.bak.$(date +%F_%T)"

if [ ! -f "$ENV_FILE" ]; then
    echo "Error: $ENV_FILE not found."
    exit 1
fi

echo "Backing up $ENV_FILE to $BACKUP_FILE..."
cp "$ENV_FILE" "$BACKUP_FILE"

# Function to update or add an env var
update_env() {
    local key=$1
    local val=$2
    if grep -q "^$key=" "$ENV_FILE"; then
        sed -i "s|^$key=.*|$key=$val|" "$ENV_FILE"
    else
        echo "$key=$val" >> "$ENV_FILE"
    fi
}

echo "Generating new secrets..."

# JWT Secret (Example)
JWT_SECRET=$(openssl rand -hex 32)
update_env "JWT_SECRET" "$JWT_SECRET"

# Add more rotations here as needed (e.g., API keys, etc.)
echo "Secrets updated in $ENV_FILE."
echo "Please restart your services to apply changes:"
echo "  docker compose restart"
echo "  systemctl restart mudro-api"
