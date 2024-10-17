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
  - [CLI Usage Examples](#cli-usage-examples)
    - [Linux Examples](#linux-examples)
    - [Windows Examples](#windows-examples)
    - [General Notes](#general-notes)
  - [Logging Usage](#logging-usage)
  - [Cloudflare R2 Configuration](#cloudflare-r2-configuration)
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
- `--provider`, `-p`: Storage provider (s3, b2, or r2) [default: "s3"]
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
- `--provider`, `-p`: Storage provider (s3, b2, or r2) [default: "s3"]
- `--region`, `-r`: AWS region (for S3 only) [default: "us-east-1"]
- `--bucket`, `-b`: Storage bucket name
- `--output`, `-o`: Output file name

### Prune

Remove old backup files from cloud storage:

```
baxfer prune [options]
```

Options:
- `--provider`, `-p`: Storage provider (s3, b2, or r2) [default: "s3"]
- `--region`, `-r`: AWS region (for S3 only) [default: "us-east-1"]
- `--bucket`, `-b`: Storage bucket name
- `--keyprefix`, `-k`: Prefix for storage keys
- `--age`, `-a`: Age of files to prune (e.g., 720h for 30 days)

## CLI Usage Examples

### Linux Examples

1. Default log file location:
   ```
   baxfer upload --bucket my-bucket /path/to/backups
   ```

2. Specifying a log file in the current directory:
   ```
   baxfer upload --logfile ./baxfer.log --bucket my-bucket /path/to/backups
   ```

3. Specifying a log file with an absolute path:
   ```
   baxfer upload --logfile /var/log/baxfer.log --bucket my-bucket /path/to/backups
   ```

4. Log file path with spaces:
   ```
   baxfer upload --logfile "/var/log/baxfer logs/app.log" --bucket my-bucket /path/to/backups
   ```

### Windows Examples

1. Default log file location:
   ```
   baxfer.exe upload --bucket my-bucket C:\path\to\backups
   ```

2. Specifying a log file in the current directory:
   ```
   baxfer.exe upload --logfile .\baxfer.log --bucket my-bucket C:\path\to\backups
   ```

3. Specifying a log file with an absolute path:
   ```
   baxfer.exe upload --logfile C:\Logs\baxfer.log --bucket my-bucket C:\path\to\backups
   ```

4. Log file path with spaces:
   ```
   baxfer.exe upload --logfile "C:\Program Files\Baxfer\logs\app.log" --bucket my-bucket C:\path\to\backups
   ```

Note: On Windows, you can use either forward slashes (/) or backslashes (\\) as path separators. Windows PowerShell and Command Prompt will understand both.

### General Notes

- Always enclose file paths with spaces in double quotes.
- For maximum portability, you can use forward slashes (/) as path separators on both Windows and Linux.
- When using environment variables for paths, remember to quote the variable expansion if it might contain spaces:
  ```
  baxfer upload --logfile "$LOG_FILE_PATH" --bucket my-bucket /path/to/backups
  ```

## Logging Usage

Baxfer now includes advanced logging options to help manage log file growth:

```
baxfer [global options] command [command options] [arguments...]
```

Global Options:
  --logfile value, -l value       Log file path (default: "baxfer.log")
  --log-max-size value            Maximum size of log file before rotation (in megabytes) (default: 10)
  --log-max-age value             Maximum number of days to retain old log files (default: 30)
  --log-max-backups value         Maximum number of old log files to retain (default: 5)
  --log-compress                  Compress rotated log files (default: true)
  --log-clear                     Clear log file on start (default: false)
  --quiet, -q                     Quiet mode (log only errors)

Example usage with logging options:

```
baxfer --logfile /var/log/baxfer.log --log-max-size 20 --log-max-age 7 --log-clear upload --bucket my-bucket /path/to/backups
```

This command will:
- Use `/var/log/baxfer.log` as the log file
- Rotate the log file when it reaches 20MB
- Keep rotated log files for 7 days
- Clear the log file before starting
- Compress old log files (default behavior)

## Cloudflare R2 Configuration

To use Cloudflare R2 as your storage provider, you need to set up the following environment variables:

- `CF_ACCOUNT_ID`: Your Cloudflare account ID
- `CF_ACCESS_KEY_ID`: Your R2 access key ID
- `CF_ACCESS_KEY_SECRET`: Your R2 access key secret

Example usage:

```
export CF_ACCOUNT_ID=your_account_id
export CF_ACCESS_KEY_ID=your_access_key_id
export CF_ACCESS_KEY_SECRET=your_access_key_secret

baxfer upload --provider r2 --bucket your-bucket-name /path/to/backups
```

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