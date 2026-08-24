# Cloudflare SSL Configuration Guide

## Overview

When using Cloudflare as a proxy, you need SSL certificates on both:
1. **Cloudflare → Visitor** (handled by Cloudflare)
2. **Cloudflare → Your Server** (you need to configure)

This guide covers the **recommended Full (Strict) SSL mode**.

---

## Cloudflare SSL Modes

### 1. Flexible SSL ❌ (Not Recommended)
- Visitor → Cloudflare: HTTPS ✅
- Cloudflare → Server: HTTP ❌
- **Problem**: Traffic between Cloudflare and your server is unencrypted

### 2. Full SSL ⚠️ (Acceptable)
- Visitor → Cloudflare: HTTPS ✅
- Cloudflare → Server: HTTPS ✅ (self-signed OK)
- **Issue**: Doesn't validate certificate

### 3. Full (Strict) SSL ✅ (Recommended)
- Visitor → Cloudflare: HTTPS ✅
- Cloudflare → Server: HTTPS ✅ (valid certificate required)
- **Best**: End-to-end encryption with validation

---

## Option 1: Cloudflare Origin Certificate (Recommended)

### Step 1: Generate Origin Certificate in Cloudflare

1. Log in to Cloudflare Dashboard
2. Select your domain (puxbay.com)
3. Go to **SSL/TLS** → **Origin Server**
4. Click **Create Certificate**
5. Select:
   - **Private key type**: RSA (2048)
   - **Hostnames**: `puxbay.com` and `*.puxbay.com`
   - **Certificate Validity**: 15 years
6. Click **Create**
7. Copy both:
   - **Origin Certificate** (save as `cloudflare-origin.pem`)
   - **Private Key** (save as `cloudflare-origin-key.pem`)

### Step 2: Install Certificate on Server

```bash
# Create SSL directory
sudo mkdir -p /etc/ssl/cloudflare

# Save origin certificate
sudo nano /etc/ssl/cloudflare/origin.pem
# Paste the Origin Certificate

# Save private key
sudo nano /etc/ssl/cloudflare/origin-key.pem
# Paste the Private Key

# Set permissions
sudo chmod 644 /etc/ssl/cloudflare/origin.pem
sudo chmod 600 /etc/ssl/cloudflare/origin-key.pem
```

### Step 3: Update Nginx Configuration

Edit `/etc/nginx/sites-available/puxbay`:

```nginx
# Main HTTPS server
server {
    listen 443 ssl http2;
    server_name www.puxbay.com puxbay.com *.puxbay.com;

    # Cloudflare Origin Certificate
    ssl_certificate /etc/ssl/cloudflare/origin.pem;
    ssl_certificate_key /etc/ssl/cloudflare/origin-key.pem;
    
    # SSL Configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Cloudflare Real IP (important!)
    set_real_ip_from 173.245.48.0/20;
    set_real_ip_from 103.21.244.0/22;
    set_real_ip_from 103.22.200.0/22;
    set_real_ip_from 103.31.4.0/22;
    set_real_ip_from 141.101.64.0/18;
    set_real_ip_from 108.162.192.0/18;
    set_real_ip_from 190.93.240.0/20;
    set_real_ip_from 188.114.96.0/20;
    set_real_ip_from 197.234.240.0/22;
    set_real_ip_from 198.41.128.0/17;
    set_real_ip_from 162.158.0.0/15;
    set_real_ip_from 104.16.0.0/13;
    set_real_ip_from 104.24.0.0/14;
    set_real_ip_from 172.64.0.0/13;
    set_real_ip_from 131.0.72.0/22;
    set_real_ip_from 2400:cb00::/32;
    set_real_ip_from 2606:4700::/32;
    set_real_ip_from 2803:f800::/32;
    set_real_ip_from 2405:b500::/32;
    set_real_ip_from 2405:8100::/32;
    set_real_ip_from 2a06:98c0::/29;
    set_real_ip_from 2c0f:f248::/32;
    real_ip_header CF-Connecting-IP;

    # Rest of your configuration...
    client_max_body_size 100M;
    
    location /static/ {
        alias /opt/puxbay/staticfiles/;
        expires 30d;
    }
    
    location /media/ {
        alias /opt/puxbay/media/;
        expires 7d;
    }
    
    location /ws/ {
        proxy_pass http://django_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
    }
    
    location / {
        proxy_pass http://django_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Step 4: Test and Reload Nginx

```bash
# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

### Step 5: Configure Cloudflare

1. Go to **SSL/TLS** → **Overview**
2. Set SSL/TLS encryption mode to **Full (strict)**
3. Enable **Always Use HTTPS**
4. Enable **Automatic HTTPS Rewrites**

---

## Option 2: Let's Encrypt with Cloudflare (Alternative)

If you prefer Let's Encrypt certificates:

### Step 1: Install Certbot with Cloudflare Plugin

```bash
sudo apt-get install certbot python3-certbot-dns-cloudflare
```

### Step 2: Create Cloudflare API Token

1. Go to Cloudflare Dashboard
2. My Profile → API Tokens
3. Create Token → Edit zone DNS template
4. Zone Resources: Include → Specific zone → puxbay.com
5. Create Token and copy it

### Step 3: Configure Cloudflare Credentials

```bash
# Create credentials file
sudo mkdir -p /root/.secrets
sudo nano /root/.secrets/cloudflare.ini
```

Add:
```ini
dns_cloudflare_api_token = YOUR_API_TOKEN_HERE
```

Set permissions:
```bash
sudo chmod 600 /root/.secrets/cloudflare.ini
```

### Step 4: Get Certificate

```bash
sudo certbot certonly \
  --dns-cloudflare \
  --dns-cloudflare-credentials /root/.secrets/cloudflare.ini \
  -d puxbay.com \
  -d *.puxbay.com \
  --preferred-challenges dns-01
```

### Step 5: Update Nginx

```nginx
server {
    listen 443 ssl http2;
    server_name www.puxbay.com puxbay.com *.puxbay.com;

    ssl_certificate /etc/letsencrypt/live/puxbay.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/puxbay.com/privkey.pem;
    
    # Rest of configuration...
}
```

---

## Cloudflare DNS Configuration

### Required DNS Records

```
Type    Name    Content             Proxy Status
A       @       YOUR_SERVER_IP      Proxied (orange cloud)
A       www     YOUR_SERVER_IP      Proxied (orange cloud)
A       *       YOUR_SERVER_IP      Proxied (orange cloud)
```

**Important**: Enable the orange cloud (Proxied) for all records.

---

## WebSocket Configuration with Cloudflare

Cloudflare supports WebSockets by default, but you need to ensure:

### 1. Enable WebSocket in Cloudflare

1. Go to **Network** tab
2. Ensure **WebSockets** is ON (it's on by default)

### 2. Nginx WebSocket Configuration

Already configured in your deploy.sh:

```nginx
location /ws/ {
    proxy_pass http://django_backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 86400;
}
```

### 3. Test WebSocket

```javascript
const ws = new WebSocket('wss://www.puxbay.com/ws/test/');
ws.onopen = () => console.log('Connected via Cloudflare!');
```

---

## Security Best Practices

### 1. Restrict Direct IP Access

Add to Nginx:

```nginx
# Block direct IP access
server {
    listen 443 ssl http2 default_server;
    server_name _;
    ssl_certificate /etc/ssl/cloudflare/origin.pem;
    ssl_certificate_key /etc/ssl/cloudflare/origin-key.pem;
    return 444;
}
```

### 2. Verify Cloudflare Requests

Add to Nginx:

```nginx
# Only allow Cloudflare IPs
allow 173.245.48.0/20;
allow 103.21.244.0/22;
allow 103.22.200.0/22;
allow 103.31.4.0/22;
allow 141.101.64.0/18;
allow 108.162.192.0/18;
allow 190.93.240.0/20;
allow 188.114.96.0/20;
allow 197.234.240.0/22;
allow 198.41.128.0/17;
allow 162.158.0.0/15;
allow 104.16.0.0/13;
allow 104.24.0.0/14;
allow 172.64.0.0/13;
allow 131.0.72.0/22;
deny all;
```

### 3. Enable Authenticated Origin Pulls (Advanced)

1. Cloudflare Dashboard → SSL/TLS → Origin Server
2. Enable **Authenticated Origin Pulls**
3. Download Cloudflare CA certificate
4. Configure Nginx to verify client certificates

---

## Troubleshooting

### SSL Handshake Failed

**Check:**
```bash
# Test SSL
openssl s_client -connect puxbay.com:443 -servername puxbay.com

# Check Nginx logs
sudo tail -f /var/log/nginx/error.log
```

### WebSocket Connection Failed

**Check:**
1. Cloudflare Network → WebSockets is ON
2. Nginx WebSocket configuration is correct
3. No firewall blocking port 443

### Mixed Content Warnings

**Fix:**
1. Cloudflare → SSL/TLS → Edge Certificates
2. Enable **Automatic HTTPS Rewrites**

---

## Performance Optimization

### 1. Enable Cloudflare Caching

```nginx
# Add cache headers for static files
location /static/ {
    alias /opt/puxbay/staticfiles/;
    expires 30d;
    add_header Cache-Control "public, immutable";
    add_header X-Cache-Status $upstream_cache_status;
}
```

### 2. Enable Cloudflare Argo

For faster WebSocket connections:
1. Cloudflare Dashboard → Traffic → Argo
2. Enable Argo Smart Routing

### 3. Enable HTTP/3 (QUIC)

1. Cloudflare Dashboard → Network
2. Enable **HTTP/3 (with QUIC)**

---

## Quick Setup Script

```bash
#!/bin/bash
# Quick Cloudflare SSL setup

# 1. Create SSL directory
sudo mkdir -p /etc/ssl/cloudflare

# 2. Save certificates (paste when prompted)
echo "Paste Origin Certificate and press Ctrl+D:"
sudo tee /etc/ssl/cloudflare/origin.pem > /dev/null

echo "Paste Private Key and press Ctrl+D:"
sudo tee /etc/ssl/cloudflare/origin-key.pem > /dev/null

# 3. Set permissions
sudo chmod 644 /etc/ssl/cloudflare/origin.pem
sudo chmod 600 /etc/ssl/cloudflare/origin-key.pem

# 4. Test Nginx
sudo nginx -t

# 5. Reload Nginx
sudo systemctl reload nginx

echo "✓ Cloudflare SSL configured!"
```

---

## Verification

### Check SSL Grade

Visit: https://www.ssllabs.com/ssltest/analyze.html?d=puxbay.com

### Check Cloudflare Status

```bash
curl -I https://www.puxbay.com
# Look for: cf-ray header (confirms Cloudflare is active)
```

### Test WebSocket

```bash
wscat -c wss://www.puxbay.com/ws/test/
```

---

## Summary

**Recommended Setup:**
1. ✅ Use Cloudflare Origin Certificate (free, easy, 15-year validity)
2. ✅ Set Cloudflare to Full (Strict) SSL mode
3. ✅ Enable Cloudflare proxy (orange cloud)
4. ✅ Configure real IP forwarding in Nginx
5. ✅ Test WebSocket connections

Your Puxbay system will have:
- End-to-end encryption
- DDoS protection
- CDN acceleration
- WebSocket support
- Free SSL certificates

🚀 Production-ready with Cloudflare!
