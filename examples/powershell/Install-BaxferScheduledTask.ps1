<#
.SYNOPSIS
    Installs a Windows Scheduled Task to run baxfer backups daily.

.DESCRIPTION
    This script creates a Windows Scheduled Task that runs the baxfer_backup.ps1 script
    daily at a specified time (default: 11:30 PM).

    The task is configured to:
    - Run whether user is logged on or not
    - Run with highest privileges
    - Start even if on battery power
    - Wake the computer to run the task (optional)

    AWS Credentials on EC2:
    When running on an EC2 instance with an attached IAM role, the AWS SDK
    automatically retrieves credentials from the instance metadata service (IMDS).
    You do NOT need to provide AWS credentials - simply ensure the IAM role has
    the necessary S3 permissions (s3:PutObject, s3:GetObject, s3:ListBucket).

    For non-EC2 environments, you can either:
    1. Pass credentials via -AwsAccessKeyId and -AwsSecretAccessKey parameters
    2. Configure the AWS CLI credentials file (~/.aws/credentials)
    3. Set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables

.PARAMETER TaskName
    Name of the scheduled task. Default: BaxferDailyBackup

.PARAMETER Bucket
    S3 bucket name for backups (required).

.PARAMETER BackupPath
    Path to the directory containing backup files (required).

.PARAMETER LogFile
    Path to the log file. Default: C:\ProgramData\baxfer\baxfer.log

.PARAMETER BaxferPath
    Path to the baxfer executable. Default: C:\Program Files\baxfer\baxfer.exe

.PARAMETER ScriptPath
    Path to baxfer_backup.ps1 script. Default: Same directory as this installer.

.PARAMETER TriggerTime
    Time to run the backup daily (24-hour format). Default: 23:30

.PARAMETER Region
    AWS region for S3 bucket.

.PARAMETER KeyPrefix
    Optional prefix for S3 keys.

.PARAMETER BackupExtension
    File extension filter for backup files. Default: .bak

.PARAMETER Compress
    Compress files before uploading.

.PARAMETER AwsAccessKeyId
    AWS Access Key ID. Leave empty on EC2 with IAM role.

.PARAMETER AwsSecretAccessKey
    AWS Secret Access Key. Leave empty on EC2 with IAM role.

.PARAMETER RunAsSystem
    Run the task as SYSTEM account. Default: $true
    Set to $false to run as a specific user (will prompt for credentials).

.PARAMETER WakeToRun
    Wake the computer to run the task if it's sleeping. Default: $false

.PARAMETER Uninstall
    Remove the scheduled task instead of installing it.

.EXAMPLE
    # Install on EC2 with IAM role (recommended):
    .\Install-BaxferScheduledTask.ps1 -Bucket "my-backup-bucket" -BackupPath "D:\Backups"

.EXAMPLE
    # Install with explicit credentials (non-EC2):
    .\Install-BaxferScheduledTask.ps1 -Bucket "my-backup-bucket" -BackupPath "D:\Backups" `
        -AwsAccessKeyId "AKIA..." -AwsSecretAccessKey "secret..."

.EXAMPLE
    # Install with custom time and options:
    .\Install-BaxferScheduledTask.ps1 -Bucket "my-backup-bucket" -BackupPath "D:\SQLBackups" `
        -TriggerTime "02:00" -Region "us-west-2" -Compress -WakeToRun

.EXAMPLE
    # Uninstall the scheduled task:
    .\Install-BaxferScheduledTask.ps1 -Uninstall -TaskName "BaxferDailyBackup"

.NOTES
    Requires Administrator privileges to create scheduled tasks.
    On EC2, ensure the instance has an IAM role with appropriate S3 permissions.
#>

#Requires -RunAsAdministrator

[CmdletBinding()]
param(
    [Parameter()]
    [string]$TaskName = "BaxferDailyBackup",

    [Parameter(Mandatory = $true, ParameterSetName = "Install")]
    [string]$Bucket,

    [Parameter(Mandatory = $true, ParameterSetName = "Install")]
    [string]$BackupPath,

    [Parameter()]
    [string]$LogFile = "C:\ProgramData\baxfer\baxfer.log",

    [Parameter()]
    [string]$BaxferPath = "C:\Program Files\baxfer\baxfer.exe",

    [Parameter()]
    [string]$ScriptPath,

    [Parameter()]
    [string]$TriggerTime = "23:30",

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
    [string]$AwsSecretAccessKey,

    [Parameter()]
    [bool]$RunAsSystem = $true,

    [Parameter()]
    [switch]$WakeToRun,

    [Parameter(ParameterSetName = "Uninstall")]
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"

# Handle uninstall
if ($Uninstall) {
    Write-Host "Removing scheduled task: $TaskName"
    try {
        Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
        Write-Host "Scheduled task '$TaskName' removed successfully." -ForegroundColor Green
    } catch {
        Write-Error "Failed to remove scheduled task: $_"
        exit 1
    }
    exit 0
}

# Determine script path
if (-not $ScriptPath) {
    $ScriptPath = Join-Path (Split-Path -Parent $MyInvocation.MyCommand.Path) "baxfer_backup.ps1"
}

# Validate paths
if (-not (Test-Path $ScriptPath)) {
    Write-Error "Backup script not found at: $ScriptPath"
    exit 1
}

if (-not (Test-Path $BaxferPath)) {
    Write-Warning "Baxfer executable not found at: $BaxferPath"
    Write-Warning "Ensure baxfer is installed before the scheduled task runs."
}

if (-not (Test-Path $BackupPath)) {
    Write-Warning "Backup path not found: $BackupPath"
    Write-Warning "Ensure the backup directory exists before the scheduled task runs."
}

# Ensure log directory exists
$logDir = Split-Path -Parent $LogFile
if (-not (Test-Path $logDir)) {
    New-Item -ItemType Directory -Path $logDir -Force | Out-Null
    Write-Host "Created log directory: $logDir"
}

# Build the PowerShell arguments for the backup script
$scriptArgs = @(
    "-BaxferPath", "`"$BaxferPath`""
    "-LogFile", "`"$LogFile`""
    "-Bucket", "`"$Bucket`""
    "-BackupPath", "`"$BackupPath`""
    "-BackupExtension", "`"$BackupExtension`""
)

if ($Region) {
    $scriptArgs += "-Region", "`"$Region`""
}

if ($KeyPrefix) {
    $scriptArgs += "-KeyPrefix", "`"$KeyPrefix`""
}

if ($Compress) {
    $scriptArgs += "-Compress"
}

if ($AwsAccessKeyId -and $AwsSecretAccessKey) {
    $scriptArgs += "-AwsAccessKeyId", "`"$AwsAccessKeyId`""
    $scriptArgs += "-AwsSecretAccessKey", "`"$AwsSecretAccessKey`""
}

$scriptArgsString = $scriptArgs -join " "

# Build the action
$actionArgs = "-NoProfile -ExecutionPolicy Bypass -File `"$ScriptPath`" $scriptArgsString"

$action = New-ScheduledTaskAction `
    -Execute "powershell.exe" `
    -Argument $actionArgs

# Parse trigger time
try {
    $triggerDateTime = [DateTime]::ParseExact($TriggerTime, "HH:mm", $null)
} catch {
    Write-Error "Invalid time format. Use 24-hour format (e.g., 23:30)"
    exit 1
}

# Create daily trigger
$trigger = New-ScheduledTaskTrigger -Daily -At $triggerDateTime

# Configure task settings
$settingsParams = @{
    AllowStartIfOnBatteries = $true
    DontStopIfGoingOnBatteries = $true
    StartWhenAvailable = $true
    ExecutionTimeLimit = New-TimeSpan -Hours 4
    RestartCount = 3
    RestartInterval = New-TimeSpan -Minutes 5
}

if ($WakeToRun) {
    $settingsParams.WakeToRun = $true
}

$settings = New-ScheduledTaskSettingsSet @settingsParams

# Determine principal (who runs the task)
if ($RunAsSystem) {
    $principal = New-ScheduledTaskPrincipal `
        -UserId "SYSTEM" `
        -LogonType ServiceAccount `
        -RunLevel Highest
    Write-Host "Task will run as SYSTEM account"
} else {
    # Will prompt for credentials when registering
    $principal = New-ScheduledTaskPrincipal `
        -UserId "$env:USERDOMAIN\$env:USERNAME" `
        -LogonType Password `
        -RunLevel Highest
    Write-Host "Task will run as: $env:USERDOMAIN\$env:USERNAME"
}

# Check if task already exists
$existingTask = Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue
if ($existingTask) {
    Write-Host "Scheduled task '$TaskName' already exists. Updating..."
    Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
}

# Register the scheduled task
Write-Host ""
Write-Host "Creating scheduled task with the following configuration:"
Write-Host "  Task Name:    $TaskName"
Write-Host "  Trigger:      Daily at $TriggerTime"
Write-Host "  Bucket:       $Bucket"
Write-Host "  Backup Path:  $BackupPath"
Write-Host "  Log File:     $LogFile"
Write-Host "  Baxfer Path:  $BaxferPath"
Write-Host "  Region:       $(if ($Region) { $Region } else { '(SDK default)' })"
Write-Host "  Key Prefix:   $(if ($KeyPrefix) { $KeyPrefix } else { '(none)' })"
Write-Host "  Compress:     $Compress"
Write-Host "  Wake to Run:  $WakeToRun"
Write-Host "  Credentials:  $(if ($AwsAccessKeyId) { 'Explicit' } else { 'IAM Role / Shared Credentials' })"
Write-Host ""

try {
    if ($RunAsSystem) {
        $task = Register-ScheduledTask `
            -TaskName $TaskName `
            -Action $action `
            -Trigger $trigger `
            -Settings $settings `
            -Principal $principal `
            -Description "Daily baxfer backup to S3 bucket: $Bucket"
    } else {
        $task = Register-ScheduledTask `
            -TaskName $TaskName `
            -Action $action `
            -Trigger $trigger `
            -Settings $settings `
            -Principal $principal `
            -Description "Daily baxfer backup to S3 bucket: $Bucket" `
            -User "$env:USERDOMAIN\$env:USERNAME" `
            -Password (Read-Host "Enter password for $env:USERDOMAIN\$env:USERNAME" -AsSecureString | ConvertFrom-SecureString -AsPlainText)
    }

    Write-Host ""
    Write-Host "Scheduled task '$TaskName' created successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "To test the task immediately, run:"
    Write-Host "  Start-ScheduledTask -TaskName `"$TaskName`""
    Write-Host ""
    Write-Host "To view task status:"
    Write-Host "  Get-ScheduledTask -TaskName `"$TaskName`" | Get-ScheduledTaskInfo"
    Write-Host ""

    # EC2 IAM Role reminder
    if (-not $AwsAccessKeyId) {
        Write-Host "NOTE: No AWS credentials were provided." -ForegroundColor Yellow
        Write-Host "If running on EC2, ensure the instance has an IAM role with S3 permissions:" -ForegroundColor Yellow
        Write-Host "  - s3:PutObject" -ForegroundColor Yellow
        Write-Host "  - s3:GetObject" -ForegroundColor Yellow
        Write-Host "  - s3:ListBucket" -ForegroundColor Yellow
        Write-Host "  - s3:GetBucketLocation" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "If running outside EC2, configure AWS credentials via:" -ForegroundColor Yellow
        Write-Host "  - AWS CLI: aws configure" -ForegroundColor Yellow
        Write-Host "  - Or re-run this script with -AwsAccessKeyId and -AwsSecretAccessKey" -ForegroundColor Yellow
        Write-Host ""
    }

} catch {
    Write-Error "Failed to create scheduled task: $_"
    exit 1
}
