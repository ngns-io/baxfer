#!/bin/bash
export RUN_INTEGRATION_TESTS=true
export R2_BUCKET=your-bucket
export CF_ACCOUNT_ID=your-account-id
export CF_ACCESS_KEY_ID=your-key-id
export CF_ACCESS_KEY_SECRET=your-secret-key

go test -v ./test/integration -run ".*R2.*"