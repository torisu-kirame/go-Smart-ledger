# Prepare OpenClaw config dirs and .env.openclaw (main docker-compose.yml).
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$EnvFile = Join-Path $Root ".env.openclaw"
$EnvExample = Join-Path $Root ".env.openclaw.example"
$ConfigDir = Join-Path $Root "data\openclaw\config"
$ConfigFile = Join-Path $ConfigDir "openclaw.json"
$DockerJson = Join-Path $Root "integrations\openclaw\openclaw.docker.json"

New-Item -ItemType Directory -Force -Path $ConfigDir, (Join-Path $Root "data\openclaw\lancedb"), (Join-Path $Root "data\openclaw\auth-secrets") | Out-Null

if (-not (Test-Path $ConfigFile)) {
    Write-Host ">> OpenClaw: writing openclaw.json"
    Copy-Item $DockerJson $ConfigFile
}

node (Join-Path $Root "scripts\repair-openclaw-config.js") $ConfigFile $DockerJson

if (-not (Test-Path $EnvFile)) {
    Write-Host ">> OpenClaw: creating .env.openclaw from example"
    Copy-Item $EnvExample $EnvFile
}

$content = Get-Content $EnvFile -Raw -ErrorAction SilentlyContinue
if ($content -notmatch 'OPENCLAW_GATEWAY_TOKEN=\S+') {
    $bytes = New-Object byte[] 24
    [Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
    $token = [BitConverter]::ToString($bytes).Replace("-", "").ToLower()
    Add-Content -Path $EnvFile -Value "OPENCLAW_GATEWAY_TOKEN=$token"
    Write-Host ">> OpenClaw: generated OPENCLAW_GATEWAY_TOKEN in .env.openclaw"
}
