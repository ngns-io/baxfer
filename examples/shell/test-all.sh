#!/bin/bash
echo "Running all provider tests..."

for script in test-s3.sh test-r2.sh test-b2.sh test-b2s3.sh test-sftp.sh; do
    echo "Running $script..."
    ./$script
    echo "----------------------------------------"
done