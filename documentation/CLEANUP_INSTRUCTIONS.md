# Database Cleanup Instructions

## Quick Reset Command

Since you have the server running, you'll need to:

1. **Stop the server** (press `Ctrl+C` in the terminal)

2. **Run the cleanup script**:
   ```bash
   python3 manage.py shell < clean_database.py
   ```

3. **Restart the server**:
   ```bash
   python3 manage.py runserver
   ```

## What the Script Does

The `clean_database.py` script will:
- ✅ Delete all users (except superusers)
- ✅ Delete all tenants
- ✅ Delete all subscriptions
- ✅ Delete all branches
- ✅ Delete all orders, products, customers
- ✅ Delete all inventory data
- ✅ **PRESERVE** all pricing plans
- ✅ **PRESERVE** superuser accounts

## After Cleanup - Testing Steps

1. **Register a new account** at `/accounts/register/`
2. **Do NOT subscribe** to any plan
3. **Try to access** `/dashboard/`
4. **Expected result**: You should be redirected to `/billing/subscription-required/`

## Alternative: Manual Database Reset

If you prefer to completely reset the database:

```bash
# Stop the server first (Ctrl+C)

# Delete the database file
rm db.sqlite3

# Recreate the database
python3 manage.py migrate

# Create a superuser
python3 manage.py createsuperuser

# Populate pricing plans
python3 manage.py shell < populate_pricing.py

# Restart the server
python3 manage.py runserver
```

## Verifying the Cleanup

After running the cleanup, you can verify by checking:
- Go to `/admin/` and login as superuser
- Check that there are no tenants, users (except superuser), or subscriptions
- Plans should still be there

Then test the subscription enforcement by registering a new account without subscribing.
