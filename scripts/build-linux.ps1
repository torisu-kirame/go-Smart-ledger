# 外部交叉编译 Linux 二进制（供 Docker COPY）
# 用法:
#   .\scripts\build-linux.ps1              # 编译全部
#   .\scripts\build-linux.ps1 -Service ledger   # 仅 ledger-api
#   .\scripts\build-linux.ps1 -Service storage
#   .\scripts\build-linux.ps1 -Service gateway

param(
    [ValidateSet("all", "ledger", "storage", "gateway")]
    [string] $Service = "all"
)

$ErrorActionPreference = "Stop"
$Backend = Join-Path $PSScriptRoot "..\backend"
$BinDir  = Join-Path $Backend "bin\linux"

New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

$env:CGO_ENABLED = "0"
$env:GOOS        = "linux"
$env:GOARCH      = "amd64"

$builds = @{
    ledger  = @{ Out = "ledger-api";  Pkg = "./services/ledger" }
    storage = @{ Out = "storage-api"; Pkg = "./services/storage" }
    gateway = @{ Out = "gateway-api"; Pkg = "./services/gateway" }
}

Push-Location $Backend
try {
    $targets = if ($Service -eq "all") { @("ledger", "storage", "gateway") } else { @($Service) }
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
