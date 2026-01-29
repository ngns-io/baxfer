<#
.SYNOPSIS
    Runs baxfer to upload backup files to S3.

.DESCRIPTION
    This script uploads backup files to Amazon S3 using baxfer. It supports both
    explicit AWS credentials (via environment variables) and IAM roles when running
    on EC2 instances.

    AWS Credential Resolution:
    The AWS SDK for Go v2 (used by baxfer) automatically resolves credentials in this order:
    1. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
    2. Shared credentials file (~/.aws/credentials)
    3. IAM role for Amazon EC2 (via instance metadata service IMDS)

    On EC2 instances with an attached IAM role, you do NOT need to set AWS credentials.
    The SDK will automatically retrieve temporary credentials from the instance metadata
    service. Simply leave the credential parameters empty or don't pass them.

.PARAMETER BaxferPath
    Path to the baxfer executable. Default: C:\Program Files\baxfer\baxfer.exe

.PARAMETER LogFile
    Path to the log file. Default: C:\ProgramData\baxfer\baxfer.log

.PARAMETER Bucket
    S3 bucket name (required).

.PARAMETER BackupPath
    Path to the directory containing backup files (required).

.PARAMETER Region
    AWS region. If not specified, uses AWS_REGION environment variable or SDK default.

.PARAMETER KeyPrefix
    Optional prefix for S3 keys.

.PARAMETER BackupExtension
    File extension filter for backup files. Default: .bak

.PARAMETER Compress
    Compress files before uploading.

.PARAMETER AwsAccessKeyId
    AWS Access Key ID. Leave empty to use IAM role or shared credentials.

.PARAMETER AwsSecretAccessKey
    AWS Secret Access Key. Leave empty to use IAM role or shared credentials.

.EXAMPLE
    # On EC2 with IAM role (no credentials needed):
    .\baxfer_backup.ps1 -Bucket "my-backup-bucket" -BackupPath "D:\Backups"

.EXAMPLE
    # With explicit credentials:
    .\baxfer_backup.ps1 -Bucket "my-backup-bucket" -BackupPath "D:\Backups" `
        -AwsAccessKeyId "AKIA..." -AwsSecretAccessKey "secret..."

.EXAMPLE
    # With all options:
    .\baxfer_backup.ps1 -Bucket "my-backup-bucket" -BackupPath "D:\SQLBackups" `
        -LogFile "D:\Logs\baxfer.log" -Region "us-west-2" -KeyPrefix "sqlserver/" `
        -BackupExtension ".bak" -Compress
#>

[CmdletBinding()]
param(
    [Parameter()]
    [string]$BaxferPath = "C:\Program Files\baxfer\baxfer.exe",

    [Parameter()]
    [string]$LogFile = "C:\ProgramData\baxfer\baxfer.log",

    [Parameter(Mandatory = $true)]
    [string]$Bucket,

    [Parameter(Mandatory = $true)]
    [string]$BackupPath,

    [Parameter()]
    [string]$Region,

    [Parameter()]
    [string]$KeyPrefix,

    [Parameter()]
    [string]$BackupExtension = ".bak",

    [Parameter()]
    [switch]$Compress,

    [Parameter()]
    [string]$AwsAccessKeyId,

    [Parameter()]
    [string]$AwsSecretAccessKey
)

$ErrorActionPreference = "Stop"
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

# Ensure log directory exists
$logDir = Split-Path -Parent $LogFile
if (-not (Test-Path $logDir)) {
    New-Item -ItemType Directory -Path $logDir -Force | Out-Null
}

function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $logEntry = "$timestamp [$Level] $Message"
    Add-Content -Path $LogFile -Value $logEntry
    if ($Level -eq "ERROR") {
        Write-Error $Message
    } else {
        Write-Host $logEntry
    }
}

# Validate baxfer executable exists
if (-not (Test-Path $BaxferPath)) {
    Write-Log "Baxfer executable not found at: $BaxferPath" -Level "ERROR"
    exit 1
}

# Validate backup path exists
if (-not (Test-Path $BackupPath)) {
    Write-Log "Backup path not found: $BackupPath" -Level "ERROR"
    exit 1
}

# Set AWS credentials if provided (otherwise SDK will use IAM role or shared credentials)
if ($AwsAccessKeyId -and $AwsSecretAccessKey) {
    $env:AWS_ACCESS_KEY_ID = $AwsAccessKeyId
    $env:AWS_SECRET_ACCESS_KEY = $AwsSecretAccessKey
    Write-Log "Using provided AWS credentials"
} else {
    Write-Log "No explicit credentials provided - using IAM role or shared credentials"
}

if ($Region) {
    $env:AWS_REGION = $Region
    Write-Log "Using AWS region: $Region"
}

# Build baxfer arguments
# Note: Logging flags can be placed before OR after the subcommand - both work.
$baxferArgs = @(
    "upload"
    "--provider", "s3"
    "--bucket", $Bucket
    "--backupext", $BackupExtension
    "--non-interactive"
    "--logfile", $LogFile
)

if ($KeyPrefix) {
    $baxferArgs += "--keyprefix", $KeyPrefix
}

if ($Compress) {
    $baxferArgs += "--compress"
}

if ($Region) {
    $baxferArgs += "--region", $Region
}

$baxferArgs += $BackupPath

Write-Log "Starting baxfer backup to bucket: $Bucket"
Write-Log "Backup source: $BackupPath"
Write-Log "Command: $BaxferPath $($baxferArgs -join ' ')"

try {
    $process = Start-Process -FilePath $BaxferPath -ArgumentList $baxferArgs -Wait -PassThru -NoNewWindow

    if ($process.ExitCode -eq 0) {
        Write-Log "Baxfer backup completed successfully"
        exit 0
    } else {
        Write-Log "Baxfer exited with code: $($process.ExitCode)" -Level "ERROR"
        exit $process.ExitCode
    }
} catch {
    Write-Log "Baxfer execution failed: $_" -Level "ERROR"
    exit 1
}
