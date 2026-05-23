# One-command full stack: Docker backend + Nginx frontend (Windows PowerShell)
$ErrorActionPreference = "Stop"

& (Join-Path $PSScriptRoot "docker-up.ps1")
exit $LASTEXITCODE
