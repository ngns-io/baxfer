@echo off
setlocal

set AWS_ACCESS_KEY_ID=your_access_key_id
set AWS_SECRET_ACCESS_KEY=your_secret_access_key
set AWS_REGION=your_s3_bucket_region

C:\path\to\baxfer.exe --non-interactive --logfile C:\path\to\baxfer.log upload --provider s3 --bucket your-s3-bucket-name C:\path\to\backups