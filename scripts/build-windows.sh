#!/bin/bash
#
# Build baxfer for Windows amd64
#
# Usage: ./scripts/build-windows.sh [version]
#
# Examples:
#   ./scripts/build-windows.sh           # builds with version "dev"
#   ./scripts/build-windows.sh v0.1.0    # builds with version "v0.1.0"

set -e

VERSION="${1:-dev}"
OUTPUT_DIR="bin"
OUTPUT_FILE="${OUTPUT_DIR}/baxfer.exe"

echo "Building baxfer for Windows amd64..."
echo "  Version: ${VERSION}"
echo "  Output:  ${OUTPUT_FILE}"

mkdir -p "${OUTPUT_DIR}"

GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="-s -w -X github.com/ngns-io/baxfer/internal/cli.version=${VERSION}" \
    -o "${OUTPUT_FILE}" \
    ./cmd/baxfer

echo "Done. Built ${OUTPUT_FILE} ($(du -h "${OUTPUT_FILE}" | cut -f1))"
