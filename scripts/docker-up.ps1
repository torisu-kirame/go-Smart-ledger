# 先外部编译，再 docker compose（Windows 推荐）
param(
    [switch] $BuildOnly,
    [switch] $NoCache
)

$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent

& (Join-Path $PSScriptRoot "build-linux.ps1")

Push-Location $Root
try {
    $args = @("compose", "build")
    if ($NoCache) { $args += "--no-cache" }
    docker @args
    if ($LASTEXITCODE -ne 0) { throw "docker compose build failed" }

    if (-not $BuildOnly) {
        docker compose up -d
        if ($LASTEXITCODE -ne 0) { throw "docker compose up failed" }
        Write-Host ""
        Write-Host "Gateway:    http://localhost:8080/api/v1/health"
        Write-Host "MiniLedger: http://localhost:4441/dashboard"
    }
}
finally {
    Pop-Location
}
