# Build Linux binaries, then docker compose (Windows)
param(
    [switch] $BuildOnly,
    [switch] $NoCache
)

$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent

& (Join-Path $PSScriptRoot "build-linux.ps1")

Push-Location $Root
try {
    $dockerArgs = @("compose", "build")
    if ($NoCache) { $dockerArgs += "--no-cache" }
    docker @dockerArgs
    if ($LASTEXITCODE -ne 0) { throw "docker compose build failed" }

    if (-not $BuildOnly) {
        docker compose up -d
        if ($LASTEXITCODE -ne 0) { throw "docker compose up failed" }
        Write-Host ""
        Write-Host "Gateway:    http://localhost:28080/api/v1/health"
        Write-Host "Web UI:     http://localhost:25173  (Nginx container)"
        Write-Host "Local dev:  make frontend-dev  (Vite, proxies to gateway)"
        Write-Host "MiniLedger: http://localhost:24441/dashboard"
        Write-Host "Login:      admin / admin123"
    }
}
finally {
    Pop-Location
}
