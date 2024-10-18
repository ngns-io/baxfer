#!/bin/bash

# Set error handling
set -e

# Configuration
BACKUP_DIR="/path/to/backups"
LOG_DIR="/var/log/baxfer"
LOG_FILE="$LOG_DIR/baxfer.log"
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")

# AWS credentials
export AWS_ACCESS_KEY_ID="your_access_key_id"
export AWS_SECRET_ACCESS_KEY="your_secret_access_key"
export AWS_REGION="us-east-1"

# Ensure log directory exists
mkdir -p "$LOG_DIR"

# Run backup with logging
{
    echo "[$TIMESTAMP] Starting backup..."
    
    baxfer --non-interactive \
           --logfile "$LOG_FILE" \
           --log-max-size 10 \
           upload \
           --provider s3 \
           --bucket your-bucket-name \
           --keyprefix "daily-backup/$TIMESTAMP" \
           "$BACKUP_DIR"

    echo "[$TIMESTAMP] Backup completed successfully"
} >> "$LOG_FILE" 2>&1 || {
    echo "[$TIMESTAMP] Backup failed with exit code $?" >> "$LOG_FILE"
    # Optionally add notification command here (e.g., email)
    exit 1
}