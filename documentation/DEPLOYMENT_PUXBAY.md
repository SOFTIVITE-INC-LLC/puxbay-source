# Puxbay Production Deployment Guide

## Domain Configuration: www.puxbay.com

Your deployment is now configured for **www.puxbay.com** with:
- ✅ Main domain: www.puxbay.com
- ✅ Apex domain: puxbay.com
- ✅ Wildcard subdomains: *.puxbay.com (for multi-tenancy)
- ✅ WebSocket support: wss://www.puxbay.com/ws/

---

## Quick Deployment

```bash
# Run deployment script
sudo bash deploy.sh

# The script will:
# 1. Install dependencies
# 2. Set up PostgreSQL (database: puxbay, user: puxbay)
# 3. Configure 3 application replicas
# 4. Set up Nginx with puxbay.com domain
# 5. Configure WebSocket support
```

---

## SSL Certificate Setup

### Using Let's Encrypt (Free)

```bash
# Install Certbot
sudo apt-get install certbot python3-certbot-nginx

# Get wildcard certificate (requires DNS validation)
sudo certbot certonly --manual --preferred-challenges dns -d puxbay.com -d *.puxbay.com

# Follow prompts to add TXT records to your DNS:
# _acme-challenge.puxbay.com TXT "random-string-here"

# Or get certificates for specific domains
sudo certbot --nginx -d puxbay.com -d www.puxbay.com

# Auto-renewal is configured automatically
sudo certbot renew --dry-run
```

---

## DNS Configuration

### Required DNS Records

```
# A Records (point to your server IP)
puxbay.com          A    YOUR_SERVER_IP
www.puxbay.com      A    YOUR_SERVER_IP

# Wildcard for subdomains (multi-tenancy)
*.puxbay.com        A    YOUR_SERVER_IP

# Optional: IPv6
puxbay.com          AAAA YOUR_SERVER_IPv6
www.puxbay.com      AAAA YOUR_SERVER_IPv6
*.puxbay.com        AAAA YOUR_SERVER_IPv6
```

### Example Subdomains

With wildcard DNS, these will work automatically:
- `tenant1.puxbay.com`
- `tenant2.puxbay.com`
- `api.puxbay.com`
- `admin.puxbay.com`

---

## WebSocket Configuration

### Nginx WebSocket Proxy

Already configured in deploy.sh:

```nginx
location /ws/ {
    proxy_pass http://django_backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 86400;  # 24 hours
}
```

### Django Channels Setup

If using Django Channels for WebSockets:

**1. Install Channels:**
```bash
pip install channels channels-redis
```

**2. Update settings.py:**
```python
INSTALLED_APPS = [
    'daphne',  # Add at top
    # ... other apps
]

ASGI_APPLICATION = 'possystem.asgi.application'

CHANNEL_LAYERS = {
    'default': {
        'BACKEND': 'channels_redis.core.RedisChannelLayer',
        'CONFIG': {
            "hosts": [('127.0.0.1', 6379)],
        },
    },
}
```

**3. Update Supervisor to use Daphne:**
```ini
[program:puxbay_replica_1]
command=/opt/puxbay/venv/bin/daphne -b 127.0.0.1 -p 8001 possystem.asgi:application
```

### Test WebSocket Connection

```javascript
// In browser console
const ws = new WebSocket('wss://www.puxbay.com/ws/notifications/');

ws.onopen = () => console.log('Connected!');
ws.onmessage = (e) => console.log('Message:', e.data);
ws.onerror = (e) => console.error('Error:', e);
```

---

## Environment Configuration

### Update .env

```bash
# Domain settings
ALLOWED_HOSTS=puxbay.com,www.puxbay.com,.puxbay.com
DOMAIN=puxbay.com

# Database
DB_NAME=puxbay
DB_USER=puxbay
DB_PASSWORD=Thinkce@softivitepuxbay

# Security
SECRET_KEY=your-generated-secret-key
FERNET_KEY=your-generated-fernet-key

# Email
EMAIL_HOST=smtp.puxbay.com
DEFAULT_FROM_EMAIL=Puxbay <noreply@puxbay.com>
```

---

## Subdomain Routing

### Django Tenants Configuration

Your system uses django-tenants for multi-tenancy:

```python
# settings.py
TENANT_MODEL = "accounts.Tenant"
TENANT_DOMAIN_MODEL = "accounts.Domain"

# Each tenant gets a subdomain
# tenant1.puxbay.com → Tenant 1
# tenant2.puxbay.com → Tenant 2
```

### Add New Tenant

```python
from accounts.models import Tenant, Domain

# Create tenant
tenant = Tenant.objects.create(
    schema_name='tenant1',
    name='Tenant 1'
)

# Add domain
Domain.objects.create(
    domain='tenant1.puxbay.com',
    tenant=tenant,
    is_primary=True
)
```

---

## Security Configuration

### Firewall Rules

```bash
# Allow HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Allow SSH
sudo ufw allow 22/tcp

# Enable firewall
sudo ufw enable
```

### SSL Security Headers

Already configured in Nginx:
- TLS 1.2 and 1.3 only
- Strong cipher suites
- HTTP/2 enabled

### Additional Security

Add to Nginx config:
```nginx
# Security headers
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
```

---

## Monitoring

### Check Application Status

```bash
# All replicas
sudo supervisorctl status puxbay_replicas:*

# Nginx status
sudo systemctl status nginx

# SSL certificate expiry
sudo certbot certificates
```

### View Logs

```bash
# Application logs
tail -f /opt/puxbay/logs/gunicorn-replica-1.log

# Nginx access logs
tail -f /var/log/nginx/access.log

# Nginx error logs
tail -f /var/log/nginx/error.log
```

---

## Troubleshooting

### WebSocket Connection Failed

**Check Nginx configuration:**
```bash
sudo nginx -t
sudo systemctl reload nginx
```

**Verify WebSocket headers:**
```bash
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" https://www.puxbay.com/ws/
```

### Subdomain Not Working

**Check DNS:**
```bash
nslookup tenant1.puxbay.com
dig tenant1.puxbay.com
```

**Verify Nginx wildcard:**
```bash
grep "server_name" /etc/nginx/sites-available/puxbay
# Should show: *.puxbay.com
```

### SSL Certificate Issues

**Check certificate:**
```bash
sudo certbot certificates
```

**Renew manually:**
```bash
sudo certbot renew --force-renewal
```

---

## Performance Optimization

### Enable HTTP/2

Already configured:
```nginx
listen 443 ssl http2;
```

### Enable Gzip Compression

Add to Nginx:
```nginx
gzip on;
gzip_types text/plain text/css application/json application/javascript text/xml application/xml;
gzip_min_length 1000;
```

### CDN Integration

For static files, use CloudFlare or similar:
1. Point DNS to CloudFlare
2. Enable caching for /static/ and /media/
3. Enable WebSocket support in CloudFlare

---

## Quick Reference

```bash
# Deploy
sudo bash deploy.sh

# Get SSL certificate
sudo certbot --nginx -d puxbay.com -d www.puxbay.com

# Restart application
sudo supervisorctl restart puxbay_replicas:*

# Restart Nginx
sudo systemctl restart nginx

# View logs
tail -f /opt/puxbay/logs/gunicorn-replica-1.log

# Test WebSocket
wscat -c wss://www.puxbay.com/ws/test/
```

Your Puxbay system is ready for production! 🚀
