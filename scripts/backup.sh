#!/bin/bash
set -e

# Load environment variables
source ../.env

# Configuration
BACKUP_DIR="/tmp/puxbay_backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
FILENAME="puxbay_db_backup_${TIMESTAMP}.sql.gz"
FILEPATH="${BACKUP_DIR}/${FILENAME}"
S3_BUCKET="s3://puxbay-database-backups"

# Ensure backup directory exists
mkdir -p "${BACKUP_DIR}"

echo "Starting database backup at ${TIMESTAMP}..."

# Dump database and compress
export PGPASSWORD="${DB_PASSWORD}"
pg_dump -h "${DB_HOST}" -U "${DB_USER}" -p "${DB_PORT:-5432}" -d "${DB_NAME}" | gzip > "${FILEPATH}"

echo "Backup created: ${FILEPATH}"

# Upload to S3 if aws-cli is available
if command -v aws &> /dev/null; then
    echo "Uploading to S3..."
    aws s3 cp "${FILEPATH}" "${S3_BUCKET}/${FILENAME}"
    
    echo "Upload complete."
    
    # Optional: Clean up local file after successful upload
    # rm "${FILEPATH}"
else
    echo "Warning: aws-cli not found, skipping S3 upload."
fi

echo "Backup process finished."
