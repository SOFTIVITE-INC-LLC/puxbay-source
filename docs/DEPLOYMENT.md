# POS System Deployment Guide

## Production Checklist

### 1. Security Configuration

```bash
# Generate new secret key
python -c "from django.core.management.utils import get_random_secret_key; print(get_random_secret_key())"

# Generate new Fernet key
python -c "from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())"
```

Update `.env`:
```env
DEBUG=False
SECRET_KEY=<new-secret-key>
FERNET_KEY=<new-fernet-key>
ALLOWED_HOSTS=yourdomain.com,www.yourdomain.com
CORS_ALLOWED_ORIGINS=https://yourdomain.com
```

### 2. Database

```bash
# Run migrations on production
python manage.py migrate_schemas --shared
python manage.py migrate_schemas

# Create initial tenant
python manage.py create_tenant --schema_name=public --name="Main Store"
```

### 3. Static Files

```bash
# Collect static files
python manage.py collectstatic --noinput

# Or use WhiteNoise (recommended)
pip install whitenoise
```

Add to `settings.py`:
```python
MIDDLEWARE = [
    'django.middleware.security.SecurityMiddleware',
    'whitenoise.middleware.WhiteNoiseMiddleware',  # Add after SecurityMiddleware
    ...
]

STATICFILES_STORAGE = 'whitenoise.storage.CompressedManifestStaticFilesStorage'
STATIC_ROOT = BASE_DIR / 'staticfiles'
```

### 4. HTTPS & Headers

```python
# settings.py
SECURE_SSL_REDIRECT = True
SECURE_HSTS_SECONDS = 31536000
SECURE_HSTS_INCLUDE_SUBDOMAINS = True
SECURE_HSTS_PRELOAD = True
SESSION_COOKIE_SECURE = True
CSRF_COOKIE_SECURE = True
```

### 5. Gunicorn Setup

```bash
pip install gunicorn

# Run with Gunicorn
gunicorn possystem.wsgi:application --bind 0.0.0.0:8000 --workers 4
```

### 6. Nginx Configuration

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl;
    server_name yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location /static/ {
        alias /path/to/possystem/staticfiles/;
    }

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 7. Systemd Service

```ini
# /etc/systemd/system/possystem.service
[Unit]
Description=POS System
After=network.target

[Service]
User=www-data
Group=www-data
WorkingDirectory=/path/to/possystem
Environment="PATH=/path/to/venv/bin"
ExecStart=/path/to/venv/bin/gunicorn --workers 4 --bind unix:/run/gunicorn.sock possystem.wsgi:application

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable possystem
sudo systemctl start possystem
```

---

## Docker Deployment (Optional)

```dockerfile
# Dockerfile
FROM python:3.11-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .
RUN python manage.py collectstatic --noinput

EXPOSE 8000
CMD ["gunicorn", "--bind", "0.0.0.0:8000", "possystem.wsgi:application"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  web:
    build: .
    ports:
      - "8000:8000"
    env_file:
      - .env
    depends_on:
      - db
  
  db:
    image: postgres:14
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      POSTGRES_DB: possystem
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD}

volumes:
  postgres_data:
```

---

## Monitoring

### Logging
Configure proper logging in production:

```python
LOGGING = {
    'version': 1,
    'disable_existing_loggers': False,
    'handlers': {
        'file': {
            'level': 'ERROR',
            'class': 'logging.FileHandler',
            'filename': '/var/log/possystem/error.log',
        },
    },
    'loggers': {
        'django': {
            'handlers': ['file'],
            'level': 'ERROR',
            'propagate': True,
        },
    },
}
```

### Health Check
Access `/api/v1/reports/` to verify API is running.
