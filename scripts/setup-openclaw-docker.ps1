# @deprecated Use `make up` — OpenClaw is bundled in docker-compose.yml
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root
Write-Host "Note: OpenClaw is now included in the main stack (make up)."
& (Join-Path $Root "scripts\init-openclaw-config.ps1")
docker compose --env-file .env.openclaw up -d ollama ollama-init openclaw-gateway
Write-Host ""
Write-Host "OpenClaw (integrated stack)"
Write-Host "  Gateway UI:  http://127.0.0.1:18789"
Write-Host "  Ollama:      http://127.0.0.1:11434/v1"
Write-Host "  Token:       .env.openclaw -> OPENCLAW_GATEWAY_TOKEN"
Write-Host "  Full stack:  make up"
