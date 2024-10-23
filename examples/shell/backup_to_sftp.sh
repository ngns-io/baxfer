#!/bin/bash

# Set credentials
export SFTP_PRIVATE_KEY=/path/to/private_key
# or
export SFTP_PASSWORD=your_password

# Upload files
baxfer upload \
    --provider sftp \
    --sftp-host backup.example.com \
    --sftp-user backupuser \
    --sftp-path /backups \
    /path/to/local/backups