# Baxfer PowerShell Examples

This folder contains PowerShell scripts for running baxfer on Windows systems.

> **Note:** Logging flags (`--logfile`, `--log-max-size`, `--quiet`, etc.) can be placed either before or after the subcommand. Both styles are valid:
> ```powershell
> # Before subcommand
> baxfer.exe --logfile C:\Logs\baxfer.log upload --bucket my-bucket C:\Backups
> # After subcommand
> baxfer.exe upload --bucket my-bucket --logfile C:\Logs\baxfer.log C:\Backups
> ```

## Scripts

### run_baxfer.ps1

A simple example script showing basic baxfer usage with explicit AWS credentials.

### baxfer_backup.ps1

A production-ready backup script with full parameterization and logging. Supports:
- Explicit AWS credentials OR IAM role authentication (for EC2)
- Configurable paths, bucket, region, and backup options
- Detailed logging
- Error handling

### Install-BaxferScheduledTask.ps1

Installs a Windows Scheduled Task to run daily backups automatically.

## Quick Start

### On EC2 with IAM Role (Recommended)

When running on an EC2 instance with an attached IAM role, AWS credentials are automatically retrieved from the instance metadata service. No credential configuration is needed.

1. Ensure your EC2 instance has an IAM role with these S3 permissions:
   - `s3:PutObject`
   - `s3:GetObject`
   - `s3:ListBucket`
   - `s3:GetBucketLocation`

2. Install the scheduled task (run as Administrator):

```powershell
.\Install-BaxferScheduledTask.ps1 -Bucket "my-backup-bucket" -BackupPath "D:\Backups"
```

That's it! The task will run daily at 11:30 PM.

### On Non-EC2 Systems

For systems outside EC2, provide AWS credentials:

```powershell
.\Install-BaxferScheduledTask.ps1 `
    -Bucket "my-backup-bucket" `
    -BackupPath "D:\Backups" `
    -AwsAccessKeyId "AKIAIOSFODNN7EXAMPLE" `
    -AwsSecretAccessKey "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
```

Alternatively, configure AWS CLI credentials (`aws configure`) and the SDK will use them automatically.

## Installation Examples

### Basic Installation (EC2 with IAM role)

```powershell
.\Install-BaxferScheduledTask.ps1 -Bucket "my-backup-bucket" -BackupPath "D:\SQLBackups"
```

### Custom Time (2:00 AM)

```powershell
.\Install-BaxferScheduledTask.ps1 `
    -Bucket "my-backup-bucket" `
    -BackupPath "D:\SQLBackups" `
    -TriggerTime "02:00"
```

### With Compression and Key Prefix

```powershell
.\Install-BaxferScheduledTask.ps1 `
    -Bucket "my-backup-bucket" `
    -BackupPath "D:\SQLBackups" `
    -KeyPrefix "server01/sqlbackups/" `
    -Compress
```

### Full Configuration

```powershell
.\Install-BaxferScheduledTask.ps1 `
    -TaskName "SQLServerBackup" `
    -Bucket "company-backups" `
    -BackupPath "D:\MSSQL\Backup" `
    -LogFile "D:\Logs\baxfer.log" `
    -BaxferPath "C:\Tools\baxfer.exe" `
    -TriggerTime "01:00" `
    -Region "us-west-2" `
    -KeyPrefix "sqlserver/prod/" `
    -BackupExtension ".bak" `
    -Compress `
    -WakeToRun
```

### Uninstall

```powershell
.\Install-BaxferScheduledTask.ps1 -Uninstall -TaskName "BaxferDailyBackup"
```

## Manual Execution

To run the backup script manually:

```powershell
# On EC2 with IAM role:
.\baxfer_backup.ps1 -Bucket "my-backup-bucket" -BackupPath "D:\Backups"

# With explicit credentials:
.\baxfer_backup.ps1 `
    -Bucket "my-backup-bucket" `
    -BackupPath "D:\Backups" `
    -AwsAccessKeyId "AKIA..." `
    -AwsSecretAccessKey "secret..."
```

## Testing the Scheduled Task

After installation, test the task:

```powershell
# Run the task immediately
Start-ScheduledTask -TaskName "BaxferDailyBackup"

# Check task status
Get-ScheduledTask -TaskName "BaxferDailyBackup" | Get-ScheduledTaskInfo

# View task history (requires Task Scheduler UI or Event Viewer)
```

## AWS Credential Resolution

The AWS SDK for Go v2 (used by baxfer) resolves credentials in this order:

1. **Environment variables**: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
2. **Shared credentials file**: `~/.aws/credentials`
3. **IAM role for EC2**: Instance Metadata Service (IMDS)

On EC2 instances with an attached IAM role, credentials are automatically retrieved from IMDS. This is the most secure approach as credentials are temporary and automatically rotated.

## Troubleshooting

### Check the log file

```powershell
Get-Content "C:\ProgramData\baxfer\baxfer.log" -Tail 50
```

### Verify baxfer is working

```powershell
& "C:\Program Files\baxfer\baxfer.exe" --help
```

### Test AWS connectivity on EC2

```powershell
# Check if IMDS is accessible (should return instance metadata)
Invoke-RestMethod -Uri "http://169.254.169.254/latest/meta-data/iam/security-credentials/"
```

### Common Issues

1. **"Access Denied" errors**: Verify IAM role/credentials have correct S3 permissions
2. **"No credentials" errors**: On EC2, ensure an IAM role is attached to the instance
3. **Task doesn't run**: Check Task Scheduler history and ensure baxfer.exe path is correct
