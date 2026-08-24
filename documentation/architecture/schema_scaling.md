# Schema Scaling & Migration Strategy

As Puxbay scales to thousands of tenants, schema management becomes a critical operational concern. This document outlines our strategy for maintaining performance and reliability during database evolution.

## 1. Schema-Level Isolation
We utilize `django-tenants` for schema-based multi-tenancy. Each tenant's data resides in a dedicated PostgreSQL schema.

### Benefits
- **Strict Data Isolation**: Guaranteed by database schemas.
- **Customization**: Ease of per-tenant logic if needed.
- **Backup/Restore**: Granular tenant-level recovery.

## 2. Migration Management
`migrate_schemas` runs migrations sequentially by default, which can be slow for many tenants.

### Best Practices
- **Atomic Migrations**: Keep migrations small and focused.
- **Zero-Downtime**: Avoid blocking operations (e.g., adding a default value to a large existing table).
- **Phased Rollouts**: For critical migrations, use `tenant_command` to migrate a "Canary" group before the whole fleet.

## 3. Optimizing Shared Data
To reduce schema bloat, metadata that is identical across all tenants (e.g., Global Product Catalog, FAQ, Pricing Plans) is stored in the `public` schema.

### Strategy
- Use `SHARED_APPS` for global data.
- Use `TENANT_APPS` for tenant-specific data (Orders, Inventory, Customers).

## 4. Monitoring Scaling
As the number of schemas grows:
- **Connection Pooling**: Use **PgBouncer** to manage the increased connection overhead.
- **Migration Logs**: Always monitor `migration_error.log` (now in `maintenance/diagnostics/`) after a deploy.
