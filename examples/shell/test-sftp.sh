#!/bin/bash
export RUN_INTEGRATION_TESTS=true
export SFTP_HOST=your-sftp-server
export SFTP_USER=your-username
export SFTP_PATH=/path/on/server
export SFTP_PORT=22
# Choose one authentication method:
export SFTP_PRIVATE_KEY=/path/to/private/key
# OR
# export SFTP_PASSWORD=your-password

go test -v ./test/integration -run ".*SFTP.*"