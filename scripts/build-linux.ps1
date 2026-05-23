# 澶栭儴浜ゅ弶缂栬瘧 Linux 浜岃繘鍒讹紙渚?Docker COPY锛?# 鐢ㄦ硶:
#   .\scripts\build-linux.ps1              # 缂栬瘧鍏ㄩ儴
#   .\scripts\build-linux.ps1 -Service ledger   # 浠?ledger-api
#   .\scripts\build-linux.ps1 -Service storage
#   .\scripts\build-linux.ps1 -Service gateway

param(
    [ValidateSet("all", "ledger", "storage", "gateway", "auth")]
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
    auth    = @{ Out = "auth-api";    Pkg = "./services/auth" }
}

Push-Location $Backend
try {
    $targets = if ($Service -eq "all") { @("ledger", "storage", "gateway", "auth") } else { @($Service) }
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
