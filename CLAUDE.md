# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Build the binary
go build -o baxfer ./cmd/baxfer

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -v ./pkg/storage -run TestUpload

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...

# Install the binary to GOPATH
go install ./cmd/baxfer
```

## Architecture

Baxfer is a CLI tool for managing backup file storage across multiple cloud providers (S3, Backblaze B2, Cloudflare R2, SFTP).

### Package Structure

- **cmd/baxfer/main.go** - Entry point, creates CLI app and runs it
- **internal/cli/** - CLI command definitions using urfave/cli/v2
  - Defines upload, download, and prune commands
  - `getUploader()` factory function creates provider-specific uploaders based on `--provider` flag
- **pkg/storage/** - Storage abstraction layer
  - `Uploader` interface defines the contract all providers implement (Upload, Download, List, Delete, FileExists, GetFileInfo)
  - Provider implementations: `s3_uploader.go`, `b2_uploader.go`, `b2s3_uploader.go`, `r2_uploader.go`, `sftp_uploader.go`
  - `storage.go` contains core logic for Upload/Download/Prune operations that work with any Uploader
- **pkg/logger/** - Logging abstraction using zap with lumberjack for rotation

### Key Design Patterns

1. **Provider Abstraction**: The `Uploader` interface allows storage operations to work identically across S3, B2, R2, and SFTP
2. **Streaming Compression**: Files can be compressed (zip) during upload via `streamingZipCompress`
3. **Incremental Upload**: `fileUploadEligible()` skips files already uploaded unless modified (compares mod time; also compares size for uncompressed uploads)
4. **Upload Integrity Verification**: S3-compatible providers (S3, B2S3, R2) use CRC32 checksums to verify data integrity during upload; the SDK computes checksums and the server rejects corrupted uploads

### Environment Variables by Provider

- **S3**: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`
- **B2**: `B2_KEY_ID`, `B2_APP_KEY`
- **R2**: `CF_ACCOUNT_ID`, `CF_ACCESS_KEY_ID`, `CF_ACCESS_KEY_SECRET`
- **SFTP**: `SFTP_HOST`, `SFTP_PORT`, `SFTP_USER`, `SFTP_PATH`, `SFTP_PRIVATE_KEY` or `SFTP_PASSWORD`
