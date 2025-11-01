#!/bin/sh
set -e

# SQLite backup script for Fly.io
# Creates a backup of the database and stores it in a separate volume

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DB_PATH="${DATABASE_PATH:-/data/booklib.db}"
BACKUP_DIR="/backup"
BACKUP_FILE="$BACKUP_DIR/booklib_$TIMESTAMP.db"
RETENTION_DAYS=7

echo "Starting backup at $(date)"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Create backup using SQLite's backup command
sqlite3 "$DB_PATH" ".backup '$BACKUP_FILE'"

echo "Backup created: $BACKUP_FILE"

# Get backup file size
BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
echo "Backup size: $BACKUP_SIZE"

# Clean up old backups (keep only last 7 days)
find "$BACKUP_DIR" -name "booklib_*.db" -type f -mtime +$RETENTION_DAYS -delete
echo "Cleaned up backups older than $RETENTION_DAYS days"

# List current backups
echo "Current backups:"
ls -lh "$BACKUP_DIR"/booklib_*.db 2>/dev/null || echo "No backups found"

echo "Backup completed at $(date)"
