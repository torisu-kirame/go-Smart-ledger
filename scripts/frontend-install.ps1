# Install frontend/desktop deps using project-local npm cache
$ErrorActionPreference = "Stop"
$Desktop = Join-Path (Split-Path $PSScriptRoot -Parent) "frontend\desktop"
$Cache = Join-Path $Desktop ".npm-cache"

Push-Location $Desktop
try {
    New-Item -ItemType Directory -Force -Path $Cache | Out-Null
    $env:NPM_CONFIG_CACHE = $Cache
    npm install
    if ($LASTEXITCODE -ne 0) {
        throw "npm install failed (cache: $Cache). Try: run terminal as Administrator, or delete global cache and retry."
    }
}
finally {
    Pop-Location
}
