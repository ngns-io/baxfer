@echo off
setlocal

set AWS_ACCESS_KEY_ID=your_access_key_id
set AWS_SECRET_ACCESS_KEY=your_secret_access_key
set AWS_REGION=your_s3_bucket_region

REM Logging flags can be placed before OR after the subcommand - both styles work:
REM Style 1: baxfer.exe --logfile C:\path\to\baxfer.log upload --bucket ...
REM Style 2: baxfer.exe upload --bucket ... --logfile C:\path\to\baxfer.log

C:\path\to\baxfer.exe upload --provider s3 --bucket your-s3-bucket-name --non-interactive --logfile C:\path\to\baxfer.log C:\path\to\backups