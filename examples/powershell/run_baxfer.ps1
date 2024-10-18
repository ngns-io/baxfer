$env:AWS_ACCESS_KEY_ID = "your_access_key_id"
$env:AWS_SECRET_ACCESS_KEY = "your_secret_access_key"
$env:AWS_REGION = "your_s3_bucket_region"

$logFile = "C:\path\to\baxfer.log"
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

try {
    $output = & C:\path\to\baxfer.exe --non-interactive --logfile $logFile upload --provider s3 --bucket your-s3-bucket-name C:\path\to\backups
    Add-Content $logFile "$timestamp - Baxfer executed successfully"
} catch {
    Add-Content $logFile "$timestamp - Error: $_"
}