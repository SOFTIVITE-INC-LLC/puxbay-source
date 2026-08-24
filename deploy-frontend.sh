#!/bin/bash
# Ultimate Production Deployment Script for puxbay frontend (SSR)
# Usage: sudo bash deploy-frontend.sh

set -e

# Configuration
APP_NAME="puxbay-frontend"
APP_DIR="/opt/puxbay/frontend"
USER="softivite"
GROUP="www-data"

# Cloudflare Cache Purge (optional - fill in to auto-purge on deploy)
CF_ZONE_ID=""
CF_API_TOKEN=""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root (use sudo)"
    exit 1
fi

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE} puxbay Frontend Ultimate Deployment${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

echo -e "${YELLOW}Phase 1: System Setup${NC}"

# Ensure Node is installed
if ! command -v node &> /dev/null
then
    echo -e "${YELLOW}Node.js is not installed. Installing Node 20...${NC}"
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
    apt-get install -y nodejs
    echo -e "${GREEN}Node.js installed successfully.${NC}"
fi

# Ensure npm/node are in PATH when running under sudo
# sudo strips the user PATH, so we need to find node/npm manually
export PATH=$PATH:/usr/bin:/usr/local/bin:/snap/bin

# Check common nvm install locations for the invoking user
SUDO_USER_HOME=$(getent passwd "${SUDO_USER:-$USER}" | cut -d: -f6)
for NVM_DIR in "$SUDO_USER_HOME/.nvm" "/root/.nvm"; do
    if [ -d "$NVM_DIR" ]; then
        NVM_NODE=$(ls "$NVM_DIR/versions/node" 2>/dev/null | sort -V | tail -1)
        [ -n "$NVM_NODE" ] && export PATH="$NVM_DIR/versions/node/$NVM_NODE/bin:$PATH"
        break
    fi
done

# Verify npm is now accessible
if ! command -v npm &> /dev/null; then
    echo -e "${RED}npm not found. Node.js may be installed via nvm for a specific user.${NC}"
    echo -e "${YELLOW}Please install Node.js system-wide first:${NC}"
    echo "  curl -fsSL https://deb.nodesource.com/setup_20.x | bash -"
    echo "  apt-get install -y nodejs"
    exit 1
fi

echo -e "${GREEN}✓ Node $(node -v) / npm $(npm -v) found at $(which npm)${NC}"

echo -e "${YELLOW}Installing system dependencies (PM2, Nginx)...${NC}"
npm install -g pm2
apt-get update
apt-get install -y nginx

echo -e "${YELLOW}Phase 2: Application Setup${NC}"

mkdir -p "$APP_DIR"
cp -a frontend/. "$APP_DIR/"
chown -R $USER:$GROUP "$APP_DIR"

cd "$APP_DIR"

echo -e "${YELLOW}Phase 3: Build and Deploy${NC}"

# Capture the resolved PATH so sudo -u preserves it
RESOLVED_PATH="$PATH"

# Fix npm cache ownership in case previous root runs left root-owned files
if [ -d "/home/$USER/.npm" ]; then
    chown -R $USER:$USER "/home/$USER/.npm"
fi

echo -e "${YELLOW}Installing dependencies...${NC}"
sudo -u $USER env PATH="$RESOLVED_PATH" npm ci --legacy-peer-deps

echo -e "${YELLOW}Building the Angular SSR application...${NC}"
sudo -u $USER env PATH="$RESOLVED_PATH" npm run build -- --configuration production

echo -e "${YELLOW}Phase 4: PM2 Service Setup${NC}"

# Stop old PM2 process if exists
sudo -u $USER env PATH="$RESOLVED_PATH" pm2 stop $APP_NAME || true
sudo -u $USER env PATH="$RESOLVED_PATH" pm2 delete $APP_NAME || true

# Verify the build output exists before starting PM2
SSR_SERVER="$APP_DIR/dist/frontend/server/server.mjs"
if [ ! -f "$SSR_SERVER" ]; then
    echo -e "${RED}Build output not found at $SSR_SERVER${NC}"
    echo -e "${RED}The Angular build may have failed. Check the output above.${NC}"
    exit 1
fi

echo -e "${YELLOW}Starting SSR server with PM2...${NC}"
sudo -u $USER env PATH="$RESOLVED_PATH" PORT=4001 pm2 start "$SSR_SERVER" --name $APP_NAME

# Save PM2 process list and configure startup
sudo -u $USER env PATH="$RESOLVED_PATH" pm2 save
env PATH="$RESOLVED_PATH" pm2 startup systemd -u $USER --hp /home/$USER || true

echo -e "${YELLOW}Phase 5: Nginx Configuration${NC}"

read -p "Do you want to configure Nginx with Cloudflare SSL? (y/n) " SETUP_SSL
if [ "$SETUP_SSL" = "y" ]; then
    read -p "Enter your domain (e.g. yourdomain.com): " DOMAIN
    
    mkdir -p /etc/nginx/ssl/$DOMAIN
    
    if [ ! -f "/etc/nginx/ssl/$DOMAIN/cert.pem" ]; then
        echo "Please paste your Cloudflare Origin Certificate (Ctrl+D to save):"
        cat > /etc/nginx/ssl/$DOMAIN/cert.pem
    else
        echo -e "${GREEN}✓ Cloudflare Origin Certificate already exists${NC}"
    fi
    
    if [ ! -f "/etc/nginx/ssl/$DOMAIN/key.pem" ]; then
        echo "Please paste your Cloudflare Private Key (Ctrl+D to save):"
        cat > /etc/nginx/ssl/$DOMAIN/key.pem
    else
        echo -e "${GREEN}✓ Cloudflare Private Key already exists${NC}"
    fi
    
    cat << EOF > /etc/nginx/sites-available/$APP_NAME
# Block hq subdomain - reserved for admin-frontend app
server {
    listen 80;
    listen 443 ssl http2;
    server_name hq.$DOMAIN;

    ssl_certificate /etc/nginx/ssl/$DOMAIN/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/$DOMAIN/key.pem;

    return 444; # Drop connection - hq is reserved for admin-frontend
}

server {
    listen 80;
    server_name $DOMAIN *.$DOMAIN;
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name $DOMAIN *.$DOMAIN;

    ssl_certificate /etc/nginx/ssl/$DOMAIN/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/$DOMAIN/key.pem;

    location /api/ {
        proxy_pass http://localhost:5000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }

    location / {
        proxy_pass http://localhost:4001; # Default Angular SSR port
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }
}
EOF
else
    cat << EOF > /etc/nginx/sites-available/$APP_NAME
# Block hq subdomain - reserved for admin-frontend app
server {
    listen 80;
    server_name hq.*;

    return 444; # Drop connection - hq is reserved for admin-frontend
}

server {
    listen 80;
    server_name _;

    location /api/ {
        proxy_pass http://localhost:5000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }

    location / {
        proxy_pass http://localhost:4001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }
}
EOF
fi

ln -sf /etc/nginx/sites-available/$APP_NAME /etc/nginx/sites-enabled/
systemctl restart nginx

# Purge Cloudflare cache if credentials are set
if [ -n "$CF_ZONE_ID" ] && [ -n "$CF_API_TOKEN" ]; then
    echo -e "${YELLOW}Purging Cloudflare cache...${NC}"
    CF_RESPONSE=$(curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$CF_ZONE_ID/purge_cache" \
        -H "Authorization: Bearer $CF_API_TOKEN" \
        -H "Content-Type: application/json" \
        --data '{"purge_everything":true}')
    if echo "$CF_RESPONSE" | grep -q '"success":true'; then
        echo -e "${GREEN}✓ Cloudflare cache purged successfully${NC}"
    else
        echo -e "${RED}⚠ Cloudflare purge may have failed: $CF_RESPONSE${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Cloudflare credentials not set — skipping cache purge${NC}"
    echo -e "${YELLOW}  Set CF_ZONE_ID and CF_API_TOKEN in this script to enable auto-purge${NC}"
fi

echo -e "${GREEN}Frontend Deployment completed successfully!${NC}"
echo -e "${BLUE}Check status with: sudo -u $USER pm2 status${NC}"
