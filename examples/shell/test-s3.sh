#!/bin/bash
export RUN_INTEGRATION_TESTS=true
export AWS_REGION=us-east-1
export AWS_BUCKET=your-bucket
export AWS_ACCESS_KEY_ID=your-key-id
export AWS_SECRET_ACCESS_KEY=your-secret-key

go test -v ./test/integration -run ".*AWS.*"