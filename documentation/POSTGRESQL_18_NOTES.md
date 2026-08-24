# PostgreSQL 18 Installation & Configuration Notes

## What's New in PostgreSQL 18

PostgreSQL 18 includes several improvements over version 15:
- Enhanced performance for parallel queries
- Improved JSON/JSONB operations
- Better replication performance
- Enhanced security features
- Improved monitoring capabilities

## Installation

The deployment script automatically:
1. Adds PostgreSQL official repository
2. Installs PostgreSQL 18
3. Configures for production use

## Paths for PostgreSQL 18

```bash
# Configuration files
/etc/postgresql/18/main/postgresql.conf
/etc/postgresql/18/main/pg_hba.conf

# Data directories
/var/lib/postgresql/18/main/          # Primary
/var/lib/postgresql/18/standby1/      # Replica 1
/var/lib/postgresql/18/standby2/      # Replica 2
/var/lib/postgresql/18/standby3/      # Replica 3
/var/lib/postgresql/18/standby4/      # Replica 4
/var/lib/postgresql/18/standby5/      # Replica 5

# Binaries
/usr/lib/postgresql/18/bin/

# Logs
/var/log/postgresql/postgresql-18-main.log
```

## Manual Installation (if needed)

```bash
# Add PostgreSQL repository
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" | sudo tee /etc/apt/sources.list.d/pgdg.list

# Update and install
sudo apt-get update
sudo apt-get install -y postgresql-18 postgresql-contrib-18

# Check version
psql --version
# Output: psql (PostgreSQL) 18.x
```

## Configuration Differences from Version 15

### 1. Default Settings
PostgreSQL 18 has optimized defaults for:
- `shared_buffers` - Better auto-tuning
- `work_mem` - Improved memory management
- `max_connections` - Enhanced connection handling

### 2. Replication Improvements
- Faster streaming replication
- Better WAL compression
- Improved standby feedback

### 3. Performance Enhancements
- Parallel query improvements
- Better index performance
- Enhanced query planner

## Verification

```bash
# Check PostgreSQL version
sudo -u postgres psql -c "SELECT version();"

# Check running instances
sudo systemctl status postgresql

# Check all replicas
for i in 1 2 3 4 5; do
    sudo systemctl status postgresql-standby$i
done

# Verify replication
sudo -u postgres psql -c "SELECT * FROM pg_stat_replication;"
```

## Upgrading from PostgreSQL 15 to 18

If you have existing PostgreSQL 15 installation:

```bash
# 1. Backup existing data
sudo -u postgres pg_dumpall > /tmp/pg15_backup.sql

# 2. Stop PostgreSQL 15
sudo systemctl stop postgresql@15-main

# 3. Install PostgreSQL 18
sudo apt-get install postgresql-18

# 4. Initialize new cluster
sudo -u postgres /usr/lib/postgresql/18/bin/initdb -D /var/lib/postgresql/18/main

# 5. Restore data
sudo -u postgres psql -f /tmp/pg15_backup.sql

# 6. Update configuration
sudo cp /etc/postgresql/15/main/postgresql.conf /etc/postgresql/18/main/
sudo cp /etc/postgresql/15/main/pg_hba.conf /etc/postgresql/18/main/

# 7. Start PostgreSQL 18
sudo systemctl start postgresql@18-main
```

## Common Commands

```bash
# Connect to database
sudo -u postgres psql -d puxbay

# Check configuration
sudo -u postgres psql -c "SHOW config_file;"

# Reload configuration
sudo systemctl reload postgresql

# View logs
sudo tail -f /var/log/postgresql/postgresql-18-main.log

# Check replication lag
sudo -u postgres psql -c "
SELECT client_addr, state, sync_state, 
       pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) AS lag_bytes
FROM pg_stat_replication;
"
```

## Troubleshooting

### Port Already in Use

```bash
# Check what's using port 5432
sudo lsof -i :5432

# Stop old PostgreSQL version
sudo systemctl stop postgresql@15-main
sudo systemctl disable postgresql@15-main
```

### Replication Not Working

```bash
# Check replication user
sudo -u postgres psql -c "\du replicator"

# Check pg_hba.conf
sudo cat /etc/postgresql/18/main/pg_hba.conf | grep replication

# Check standby status
sudo -u postgres psql -p 5433 -c "SELECT pg_is_in_recovery();"
```

### Performance Issues

```bash
# Check current settings
sudo -u postgres psql -c "SHOW ALL;"

# Analyze slow queries
sudo -u postgres psql -c "
SELECT query, calls, total_exec_time, mean_exec_time
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 10;
"
```

## Best Practices for PostgreSQL 18

1. **Use Connection Pooling**
   - Install PgBouncer for better connection management
   - Recommended for high-traffic applications

2. **Enable Query Statistics**
   ```sql
   CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
   ```

3. **Regular Maintenance**
   ```bash
   # Auto-vacuum is enabled by default
   # Manual vacuum if needed
   sudo -u postgres psql -d puxbay -c "VACUUM ANALYZE;"
   ```

4. **Monitor Replication Lag**
   ```bash
   # Add to cron for monitoring
   */5 * * * * sudo -u postgres psql -c "SELECT * FROM pg_stat_replication;" >> /var/log/replication-status.log
   ```

5. **Optimize for Your Workload**
   ```bash
   # For read-heavy workloads
   shared_buffers = 4GB
   effective_cache_size = 12GB
   
   # For write-heavy workloads
   wal_buffers = 16MB
   checkpoint_completion_target = 0.9
   ```

Your Puxbay system now uses PostgreSQL 18! 🚀
