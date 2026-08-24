# POS System - Developer Documentation

## Quick Start

### Prerequisites
- Python 3.11+
- PostgreSQL 14+
- Node.js 18+ (for TailwindCSS)

### Setup

1. **Clone and create virtual environment**
```bash
git clone <repository>
cd possystem
python -m venv venv
venv\Scripts\activate  # Windows
source venv/bin/activate  # Linux/Mac
```

2. **Install dependencies**
```bash
pip install -r requirements.txt
npm install
```

3. **Configure environment**
```bash
copy .env.example .env
# Edit .env with your database credentials and secret keys
```

4. **Setup database**
```bash
# Create PostgreSQL database
createdb possystem

# Run migrations
python manage.py migrate_schemas --shared
python manage.py migrate_schemas
```

5. **Create superuser and tenant**
```bash
python manage.py create_tenant --schema_name=public
python manage.py createsuperuser
```

6. **Build frontend assets**
```bash
npm run build
```

7. **Run development server**
```bash
python manage.py runserver
```

---

## Project Structure

```
possystem/
├── accounts/          # User & Tenant management
├── api/               # REST API endpoints
├── billing/           # Subscription management
├── branches/          # Branch operations & inventory
├── main/              # Core POS models (Product, Order, Customer)
├── notifications/     # System notifications
├── storefront/        # Public-facing storefront
├── static/            # CSS, JS, images
├── templates/         # Django templates
│   └── partials/      # Reusable template components
└── possystem/         # Project settings
```

---

## API Reference

Base URL: `/api/v1/`

### Authentication
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/auth/login/` | POST | Get JWT tokens |
| `/api/v1/auth/refresh/` | POST | Refresh access token |
| `/api/v1/auth/logout/` | POST | Blacklist refresh token |
| `/api/v1/auth/user/` | GET | Get current user details |

### Products
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/products/` | GET | List products |
| `/api/v1/products/` | POST | Create product |
| `/api/v1/products/{id}/` | GET | Get product |
| `/api/v1/products/{id}/` | PUT | Update product |
| `/api/v1/products/{id}/` | DELETE | Delete product |

### Orders
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/orders/` | GET | List orders |
| `/api/v1/orders/` | POST | Create order |
| `/api/v1/orders/{id}/` | GET | Get order details |

### Customers
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/customers/` | GET | List customers |
| `/api/v1/customers/` | POST | Create customer |
| `/api/v1/customers/{id}/` | GET | Get customer |

### Reports
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/reports/financial_summary/` | GET | Financial summary |
| `/api/v1/reports/daily_sales/` | GET | Daily sales report |
| `/api/v1/reports/top_products/` | GET | Top selling products |
| `/api/v1/reports/low_stock/` | GET | Low stock alerts |

---

## Running Tests

```bash
# Run all tests
python manage.py test

# Run specific app tests
python manage.py test main
python manage.py test api

# Run with coverage
pip install coverage
coverage run manage.py test
coverage report
```

---

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `SECRET_KEY` | Django secret key | `your-secret-key` |
| `DEBUG` | Debug mode | `True` or `False` |
| `ALLOWED_HOSTS` | Allowed hosts | `localhost,127.0.0.1` |
| `DB_NAME` | Database name | `possystem` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `your-password` |
| `STRIPE_SECRET_KEY` | Stripe API key | `sk_test_...` |
| `FERNET_KEY` | Encryption key | `base64-encoded-key` |

---

## Multi-Tenancy

This project uses `django-tenants` for schema-based multi-tenancy.

- **Public schema**: Shared tables (User, Tenant, Billing)
- **Tenant schemas**: Isolated tables (Product, Order, Customer)

Each tenant has their own subdomain: `store1.yourdomain.com`
