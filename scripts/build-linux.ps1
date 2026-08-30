# Cross-compile Linux binaries for Docker COPY.
# Usage:
#   .\scripts\build-linux.ps1              # smart-ledger-api (Gin monolith)
#   .\scripts\build-linux.ps1 -Service api
#   .\scripts\build-linux.ps1 -Service all # monolith + legacy microservices

param(
    [ValidateSet("all", "api", "ledger", "storage", "gateway", "auth")]
    [string] $Service = "api"
)

$ErrorActionPreference = "Stop"
$Backend = Join-Path $PSScriptRoot "..\go-backend"
$BinDir  = Join-Path $Backend "bin\linux"

New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

$env:CGO_ENABLED = "0"
$env:GOOS        = "linux"
$env:GOARCH      = "amd64"

$builds = @{
    api     = @{ Out = "smart-ledger-api"; Pkg = "./services/api" }
    ledger  = @{ Out = "ledger-api";       Pkg = "./services/ledger" }
    storage = @{ Out = "storage-api";      Pkg = "./services/storage" }
    gateway = @{ Out = "gateway-api";      Pkg = "./services/gateway" }
    auth    = @{ Out = "auth-api";         Pkg = "./services/auth" }
}

Push-Location $Backend
try {
    $targets = if ($Service -eq "all") {
        @("api", "ledger", "storage", "gateway", "auth")
    } elseif ($Service -eq "api") {
        @("api")
    } else {
        @($Service)
    }
    foreach ($name in $targets) {
        $b = $builds[$name]
        $out = Join-Path $BinDir $b.Out
        Write-Host ">> building $out"
        go build -trimpath -ldflags "-s -w" -o $out $b.Pkg
        if ($LASTEXITCODE -ne 0) { throw "build failed: $name" }
    }
    Write-Host ">> done: $BinDir"
}
finally {
    Pop-Location
}
