# Prepare stack env and OpenClaw legacy config dirs (deploy/compose/docker-compose.yml).
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$EnvFile = Join-Path $Root "deploy\env\stack.env"
$EnvExample = Join-Path $Root "deploy\env\stack.env.example"
$ConfigDir = Join-Path $Root "deploy\config\openclaw"
$ConfigFile = Join-Path $ConfigDir "openclaw.json"
$AgentDir = Join-Path $Root "deploy\config\agent"
$DockerJson = Join-Path $Root "integrations\openclaw\openclaw.docker.json"

New-Item -ItemType Directory -Force -Path $ConfigDir, (Join-Path $Root "deploy\config\openclaw\lancedb"), (Join-Path $Root "deploy\config\openclaw\auth-secrets"), $AgentDir | Out-Null

if (-not (Test-Path $ConfigFile)) {
    Write-Host ">> OpenClaw: writing openclaw.json"
    Copy-Item $DockerJson $ConfigFile
}

node (Join-Path $Root "scripts\repair-openclaw-config.js") $ConfigFile $DockerJson

if (-not (Test-Path $EnvFile)) {
    Write-Host ">> stack: creating deploy/env/stack.env from example"
    Copy-Item $EnvExample $EnvFile
}

$content = Get-Content $EnvFile -Raw -ErrorAction SilentlyContinue
if ($content -notmatch 'OPENCLAW_GATEWAY_TOKEN=\S+') {
    $bytes = New-Object byte[] 24
    [Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
    $token = [BitConverter]::ToString($bytes).Replace("-", "").ToLower()
    Add-Content -Path $EnvFile -Value "OPENCLAW_GATEWAY_TOKEN=$token"
    Write-Host ">> stack: generated OPENCLAW_GATEWAY_TOKEN in deploy/env/stack.env"
}
