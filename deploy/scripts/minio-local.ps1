param(
    [ValidateSet("start", "stop", "logs", "smoke", "console")]
    [string]$Command = "console"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
Push-Location $repoRoot

try {
    $endpoint = if ($env:EDU_S3_ENDPOINT) { $env:EDU_S3_ENDPOINT } else { "http://localhost:9000" }
    $accessKey = if ($env:EDU_S3_ACCESS_KEY) { $env:EDU_S3_ACCESS_KEY } else { "minioadmin" }
    $secretKey = if ($env:EDU_S3_SECRET_KEY) { $env:EDU_S3_SECRET_KEY } else { "minioadmin" }
    $bucket = if ($env:EDU_S3_BUCKET) { $env:EDU_S3_BUCKET } else { "edu-assets" }

    switch ($Command) {
        "start" {
            docker compose up -d minio minio-init
        }
        "stop" {
            docker compose stop minio minio-init
        }
        "logs" {
            docker compose logs -f minio minio-init
        }
        "smoke" {
            Invoke-WebRequest -UseBasicParsing -Uri "$endpoint/minio/health/live" | Out-Null
            docker compose run --rm minio-init /bin/sh -c "mc alias set local http://minio:9000 $accessKey $secretKey >/dev/null && mc ls local/$bucket"
            Write-Host "MinIO local smoke test passed." -ForegroundColor Green
        }
        "console" {
            Write-Host "MinIO API:     $endpoint"
            Write-Host "MinIO console: http://localhost:9001"
            Write-Host "Access key:    $accessKey"
            Write-Host "Secret key:    $secretKey"
            Write-Host ""
            Write-Host "Windows shortcut examples:"
            Write-Host "  powershell -File deploy/scripts/minio-local.ps1 start"
            Write-Host "  powershell -File deploy/scripts/minio-local.ps1 smoke"
        }
    }
}
finally {
    Pop-Location
}
