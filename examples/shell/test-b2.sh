#!/bin/bash
export RUN_INTEGRATION_TESTS=true
export B2_BUCKET=your-bucket
export B2_KEY_ID=your-key-id
export B2_APP_KEY=your-app-key

go test -v ./test/integration -run ".*B2[^S].*"