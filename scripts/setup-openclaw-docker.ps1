# Initialize config and start OpenClaw + Ollama via Docker Compose.
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$Compose = @("docker", "compose", "--env-file", ".env.openclaw", "-f", "docker-compose.openclaw.yml")
$EnvFile = Join-Path $Root ".env.openclaw"
$EnvExample = Join-Path $Root ".env.openclaw.example"
$ConfigDir = Join-Path $Root "data\openclaw\config"
$ConfigFile = Join-Path $ConfigDir "openclaw.json"
$DockerJson = Join-Path $Root "integrations\openclaw\openclaw.docker.json"

New-Item -ItemType Directory -Force -Path $ConfigDir, (Join-Path $Root "data\openclaw\lancedb"), (Join-Path $Root "data\openclaw\auth-secrets") | Out-Null

if (-not (Test-Path $ConfigFile)) {
    Write-Host "Writing openclaw.json from openclaw.docker.json"
    Copy-Item $DockerJson $ConfigFile
}

if (-not (Test-Path $EnvFile)) {
    Copy-Item $EnvExample $EnvFile
}

$envContent = Get-Content $EnvFile -Raw
if ($envContent -notmatch 'OPENCLAW_GATEWAY_TOKEN=\S+') {
    $bytes = New-Object byte[] 24
    [Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
    $token = [BitConverter]::ToString($bytes).Replace("-", "").ToLower()
    Add-Content -Path $EnvFile -Value "OPENCLAW_GATEWAY_TOKEN=$token"
    Write-Host "Generated OPENCLAW_GATEWAY_TOKEN in .env.openclaw"
}

Write-Host "Pulling images..."
& @Compose pull ollama openclaw-gateway 2>$null

Write-Host "Starting Ollama..."
& @Compose up -d ollama

$chatModel = if ($env:OLLAMA_CHAT_MODEL) { $env:OLLAMA_CHAT_MODEL } else { "llama3.2" }
$embedModel = if ($env:OLLAMA_EMBED_MODEL) { $env:OLLAMA_EMBED_MODEL } else { "nomic-embed-text" }

Write-Host "Waiting for Ollama..."
for ($i = 0; $i -lt 60; $i++) {
    & @Compose exec -T ollama ollama list 2>$null | Out-Null
    if ($LASTEXITCODE -eq 0) { break }
    Start-Sleep -Seconds 2
}

Write-Host "Pulling chat model: $chatModel"
& @Compose exec -T ollama ollama pull $chatModel

Write-Host "Pulling embed model: $embedModel"
& @Compose exec -T ollama ollama pull $embedModel

Write-Host "Starting OpenClaw Gateway..."
& @Compose up -d openclaw-gateway

Write-Host ""
Write-Host "OpenClaw Docker stack is up."
Write-Host "  Gateway UI:  http://127.0.0.1:18789"
Write-Host "  Ollama:      http://127.0.0.1:11434"
Write-Host "  Config:      data/openclaw/config/openclaw.json"
Write-Host "  Token:       OPENCLAW_GATEWAY_TOKEN in .env.openclaw"
Write-Host ""
Write-Host "Smart Ledger -> Settings -> AI:"
Write-Host "  Ollama:   http://127.0.0.1:11434/v1"
Write-Host "  Gateway:  http://127.0.0.1:18789"
Write-Host ""
Write-Host "Docs: docs/openclaw-integration.md"
