# Baxfer

Baxfer is a CLI tool designed to help manage storage for database backups. It supports uploading, downloading, and pruning backup files from cloud storage providers such as Amazon S3 and Backblaze B2.

## Table of Contents

- [Baxfer](#baxfer)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
    - [Pre-built Binaries](#pre-built-binaries)
    - [Using Go](#using-go)
  - [Usage](#usage)
    - [Upload](#upload)
    - [Download](#download)
    - [Prune](#prune)
  - [Building from Source](#building-from-source)
  - [Testing](#testing)
  - [Contributing](#contributing)
  - [License](#license)

## Features

- Upload backup files to Amazon S3 or Backblaze B2
- Download backup files from cloud storage
- Prune old backup files from cloud storage
- Supports both interactive and non-interactive modes
- Progress bar for file transfers in interactive mode
- Configurable file extension filtering
- Optional file compression before upload

## Installation

### Pre-built Binaries

Download the latest pre-built binary for your operating system from the [Releases](https://github.com/ngns-io/baxfer/releases) page.

### Using Go

If you have Go installed, you can install Baxfer using:

```
go install github.com/ngns-io/baxfer@latest
```

## Usage

### Upload

Upload backup files to cloud storage:

```
baxfer upload [options] <root directory>
```

Options:
- `--provider`, `-p`: Storage provider (s3 or b2) [default: "s3"]
- `--region`, `-r`: AWS region (for S3 only) [default: "us-east-1"]
- `--bucket`, `-b`: Storage bucket name
- `--keyprefix`, `-k`: Prefix for storage keys
- `--backupext`, `-x`: File extension for backup files [default: ".bak"]
- `--compress`, `-c`: Compress files before uploading
- `--non-interactive`: Run in non-interactive mode (no progress bars)

### Download

Download a backup file from cloud storage:

```
baxfer download [options] <key>
```

Options:
- `--provider`, `-p`: Storage provider (s3 or b2) [default: "s3"]
- `--region`, `-r`: AWS region (for S3 only) [default: "us-east-1"]
- `--bucket`, `-b`: Storage bucket name
- `--output`, `-o`: Output file name

### Prune

Remove old backup files from cloud storage:

```
baxfer prune [options]
```

Options:
- `--provider`, `-p`: Storage provider (s3 or b2) [default: "s3"]
- `--region`, `-r`: AWS region (for S3 only) [default: "us-east-1"]
- `--bucket`, `-b`: Storage bucket name
- `--keyprefix`, `-k`: Prefix for storage keys
- `--age`, `-a`: Age of files to prune (e.g., 720h for 30 days)

## Building from Source

1. Ensure you have Go 1.16 or later installed.

2. Clone the repository:
   ```
   git clone https://github.com/ngns-io/baxfer.git
   cd baxfer
   ```

3. Install dependencies:
   ```
   go mod download
   ```

4. Build the project:
   ```
   go build -o baxfer ./cmd/baxfer
   ```

5. (Optional) Install the binary to your GOPATH:
   ```
   go install ./cmd/baxfer
   ```

## Testing

To run the test suite:

```
go test ./...
```

To run tests with verbose output:

```
go test -v ./...
```

To run a specific test:

```
go test -v ./pkg/storage -run TestUpload
```

To run tests with coverage:

```
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.