# Baxfer

Baxfer is a CLI tool designed to help manage storage for files in a folder hierarchy such as database backup files. It supports uploading, downloading, and pruning files from cloud storage providers such as Amazon S3, Backblaze B2, and Cloudflare R2.

# ⚠️ Important Notice

This project is currently in development and is not considered production-ready. While effort has been made to ensure reliability, the software should be thoroughly tested in your specific environment before any production use. Use at your own risk.

No warranty is provided, either expressed or implied. The authors and contributors are not liable for any data loss or other damages arising from the use of this software.

## Table of Contents

- [Baxfer](#baxfer)
- [⚠️ Important Notice](#️-important-notice)
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
  - [Amazon S3 Configuration](#amazon-s3-configuration)
  - [Backblaze B2 Configuration](#backblaze-b2-configuration)
  - [Cloudflare R2 Configuration](#cloudflare-r2-configuration)
  - [SFTP Configuration](#sftp-configuration)
    - [Environment Variables](#environment-variables)
  - [Running baxfer as a Background Process](#running-baxfer-as-a-background-process)
    - [Windows Task Scheduler Setup](#windows-task-scheduler-setup)
    - [Linux Cron Setup](#linux-cron-setup)
  - [Building from Source](#building-from-source)
  - [Testing](#testing)
  - [Contributing](#contributing)
  - [License](#license)

## Features

- Upload backup files to Amazon S3, Backblaze B2, Cloudflare R2, or SFTP servers
- Download backup files from cloud storage or SFTP
- Prune old backup files from storage
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
- `--provider`, `-p`: Storage provider (s3, b2, b2s3, r2, or sftp) [default: "s3"]
- `--region`, `-r`: AWS region (for S3 and b2s3 only) [default: "us-east-1" for s3 and "us-west-002" for b2s3]
- `--bucket`, `-b`: Storage bucket name
- `--keyprefix`, `-k`: Prefix for storage keys
- `--backupext`, `-x`: File extension for backup files [default: ".bak"]
- `--compress`, `-c`: Compress files before uploading
- `--non-interactive`: Run in non-interactive mode (no progress bars)

SFTP-specific options:
- `--sftp-host`: SFTP server hostname (env: SFTP_HOST)
- `--sftp-port`: SFTP server port [default: 22] (env: SFTP_PORT)
- `--sftp-user`: SFTP username (env: SFTP_USER)
- `--sftp-path`: Base path on SFTP server (env: SFTP_PATH)

### Download

Download a backup file from cloud storage:

```
baxfer download [options] <key>
```

Options:
- `--provider`, `-p`: Storage provider (s3, b2, b2s3, r2, or sftp) [default: "s3"]
- `--region`, `-r`: AWS region (for S3 and b2s3 only) [default: "us-east-1" for s3 and "us-west-002" for b2s3]
- `--bucket`, `-b`: Storage bucket name
- `--output`, `-o`: Output file name

SFTP-specific options:
- `--sftp-host`: SFTP server hostname (env: SFTP_HOST)
- `--sftp-port`: SFTP server port [default: 22] (env: SFTP_PORT)
- `--sftp-user`: SFTP username (env: SFTP_USER)
- `--sftp-path`: Base path on SFTP server (env: SFTP_PATH)

### Prune

Remove old backup files from cloud storage:

```
baxfer prune [options]
```

Options:
- `--provider`, `-p`: Storage provider (s3, b2, b2s3, r2, or sftp) [default: "s3"]
- `--region`, `-r`: AWS region (for S3 and b2s3 only) [default: "us-east-1" for s3 and "us-west-002" for b2s3]
- `--bucket`, `-b`: Storage bucket name
- `--keyprefix`, `-k`: Prefix for storage keys
- `--age`, `-a`: Age of files to prune (e.g., 720h for 30 days)

SFTP-specific options:
- `--sftp-host`: SFTP server hostname (env: SFTP_HOST)
- `--sftp-port`: SFTP server port [default: 22] (env: SFTP_PORT)
- `--sftp-user`: SFTP username (env: SFTP_USER)
- `--sftp-path`: Base path on SFTP server (env: SFTP_PATH)

## CLI Usage Examples

### Linux Examples

1. Default log file location:
   ```
   baxfer upload --bucket my-bucket /path/to/backups
   ```

2. Specifying a log file in the current directory:
   ```
   baxfer --logfile ./baxfer.log upload --bucket my-bucket /path/to/backups
   ```

3. Specifying a log file with an absolute path:
   ```
   baxfer --logfile /var/log/baxfer.log upload --bucket my-bucket /path/to/backups
   ```

4. Log file path with spaces:
   ```
   baxfer --logfile "/var/log/baxfer logs/app.log" upload --bucket my-bucket /path/to/backups
   ```

### Windows Examples

1. Default log file location:
   ```
   baxfer.exe upload --bucket my-bucket C:\path\to\backups
   ```

2. Specifying a log file in the current directory:
   ```
   baxfer.exe --logfile .\baxfer.log upload --bucket my-bucket C:\path\to\backups
   ```

3. Specifying a log file with an absolute path:
   ```
   baxfer.exe --logfile C:\Logs\baxfer.log upload --bucket my-bucket C:\path\to\backups
   ```

4. Log file path with spaces:
   ```
   baxfer.exe --logfile "C:\Program Files\Baxfer\logs\app.log" upload --bucket my-bucket C:\path\to\backups
   ```

Note: On Windows, you can use either forward slashes (/) or backslashes (\\) as path separators. Windows PowerShell and Command Prompt will understand both.

### General Notes

- Always enclose file paths with spaces in double quotes.
- For maximum portability, you can use forward slashes (/) as path separators on both Windows and Linux.
- When using environment variables for paths, remember to quote the variable expansion if it might contain spaces:
  ```
  baxfer --logfile "$LOG_FILE_PATH" upload --bucket my-bucket /path/to/backups
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

## Amazon S3 Configuration

To use Amazon S3 as your storage provider with baxfer, you need to set up the following environment variables:

- `AWS_ACCESS_KEY_ID`: Your AWS access key ID
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret access key
- `AWS_REGION`: (Optional) The AWS region for your S3 bucket

You can obtain these credentials by following these steps:

1. Log in to your AWS Management Console.
2. Go to the IAM (Identity and Access Management) dashboard.
3. Create a new IAM user or select an existing one.
4. Attach the `AmazonS3FullAccess` policy to the user (or a more restrictive custom policy if desired).
5. Generate new access keys for the user and securely store the Access Key ID and Secret Access Key.

Please note that the AWS region can be specified in three ways, in order of precedence:
1. Command line flag: `--region` or `-r`
2. Environment variable: `AWS_REGION`
3. Default value: "us-east-1" for s3 and "us-west-002" for b2s3

Example usage with region specified by command line flag:
```bash
export AWS_ACCESS_KEY_ID=your_access_key_id
export AWS_SECRET_ACCESS_KEY=your_secret_access_key

baxfer upload --provider s3 --region us-west-2 --bucket your-s3-bucket-name /path/to/backups
```

Example usage with region specified by environment variable:
```bash
export AWS_ACCESS_KEY_ID=your_access_key_id
export AWS_SECRET_ACCESS_KEY=your_secret_access_key
export AWS_REGION=us-west-2

baxfer upload --provider s3 --bucket your-s3-bucket-name /path/to/backups
```

## Backblaze B2 Configuration

To use Backblaze B2 as your storage provider with baxfer, you need to set up the following environment variables:

- `B2_KEY_ID`: Your Backblaze B2 application key ID
- `B2_APP_KEY`: Your Backblaze B2 application key

You can obtain these credentials by following these steps:

1. Log in to your Backblaze account.
2. Go to the "App Keys" section in your account settings.
3. Click "Add a New Application Key".
4. Set the permissions for the key (ensure it has read and write access to the bucket you'll be using).
5. Copy the "applicationKeyId" and "applicationKey" values.

Example usage:

```bash
export B2_KEY_ID=your_b2_key_id
export B2_APP_KEY=your_b2_app_key

baxfer upload --provider b2 --bucket your-b2-bucket-name /path/to/backups
```

Note: Make sure your B2 bucket is created before running baxfer. You can create a bucket through the Backblaze B2 web interface or using their CLI tool.

When using Backblaze B2 with the b2s3 provider, baxfer will automatically use the correct endpoint for B2's S3-compatible API. Please note that the region can be specified in three ways, in order of precedence:
1. Command line flag: `--region` or `-r`
2. Environment variable: `AWS_REGION`
3. Default value: "us-west-002" for b2s3

Remember to keep your B2 credentials secure and never commit them to version control. Using environment variables as shown above is a safe way to provide these credentials to baxfer.

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

## SFTP Configuration

To use SFTP as your storage provider, you need to set up either password or private key authentication:

### Environment Variables
- `SFTP_HOST`: SFTP server hostname (can be set via --sftp-host flag)
- `SFTP_PORT`: SFTP server port (can be set via --sftp-port flag, defaults to 22)
- `SFTP_USER`: SFTP username (can be set via --sftp-user flag)
- `SFTP_PATH`: Base path on SFTP server (can be set via --sftp-path flag)
- `SFTP_PRIVATE_KEY`: Path to SSH private key file
- `SFTP_PASSWORD`: Password for authentication (only used if SFTP_PRIVATE_KEY is not set)

Example usage with private key authentication:
```bash
export SFTP_PRIVATE_KEY=/path/to/private_key
export SFTP_USER=backup_user

baxfer upload \
    --provider sftp \
    --sftp-host backup.example.com \
    --sftp-path /backup/files \
    /path/to/backups
```

Example usage with password authentication:
```bash
export SFTP_PASSWORD=your_secure_password
export SFTP_USER=backup_user

baxfer upload \
    --provider sftp \
    --sftp-host backup.example.com \
    --sftp-port 2222 \
    --sftp-path /backup/files \
    /path/to/backups
```

Configuration options can be provided either through environment variables or command-line flags. Command-line flags take precedence over environment variables.

Note: For security reasons, it's recommended to use private key authentication rather than password authentication. Ensure that your private key file has appropriate permissions (600 or more restrictive).

## Running baxfer as a Background Process

Sample scripts for running baxfer as a background process can be found in the `examples` directory of the repository:
- Windows batch files: `examples/batch/`
- PowerShell scripts: `examples/powershell/`
- Linux/Unix shell scripts: `examples/shell/`

These scripts demonstrate how to:
- Set cloud storage provider credentials
- Configure logging
- Set up scheduled backups
- Handle errors and notifications

### Windows Task Scheduler Setup

1. Copy the appropriate script from the examples directory and modify it for your environment.

2. Set up a Windows Task Scheduler task:
   - Open Task Scheduler and create a new task
   - Set the trigger to run on your desired schedule
   - For the action, choose "Start a program" and point to either:
     - Your modified .cmd file
     - `powershell.exe` with argument `-File "path/to/your/script.ps1"`

### Linux Cron Setup

1. Copy and modify an example bash script from `examples/shell/`.

2. Make the script executable:
   ```bash
   chmod +x /path/to/your/backup_script.sh
   ```

3. Open your crontab for editing:
   ```bash
   crontab -e
   ```

4. Add a cron entry (example for daily backup at 2 AM):
   ```
   0 2 * * * /path/to/your/backup_script.sh
   ```

Remember:
- Use `--non-interactive` flag for background processes
- Keep credentials secure and never commit them to version control
- Use environment variables or secure script files with restricted access
- Test thoroughly in your environment before scheduling

Reference the example scripts for best practices in error handling, logging, and security.

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