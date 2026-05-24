# Index one ledger export into JSONL for OpenClaw memory_store (manual / skill hook).
param(
    [Parameter(Mandatory)][string]$LedgerId,
    [Parameter(Mandatory)][string]$AccessToken,
    [string]$ApiBase = "http://127.0.0.1:28080/api/v1",
    [string]$OutFile = ""
)
$ErrorActionPreference = "Stop"
if (-not $OutFile) { $OutFile = "data/openclaw/index-$LedgerId.jsonl" }
New-Item -ItemType Directory -Force -Path (Split-Path $OutFile) | Out-Null

$uri = "$ApiBase/ledgers/$LedgerId/rag-export"
$headers = @{ Authorization = "Bearer $AccessToken" }
$export = Invoke-RestMethod -Uri $uri -Headers $headers -Method Get

$lines = @()
foreach ($ch in $export.chunks) {
    $lines += (@{ id = $ch.id; text = $ch.text; ledgerId = $ch.ledgerId; seq = $ch.seq; type = $ch.type } | ConvertTo-Json -Compress)
}
$lines | Set-Content -Encoding utf8 $OutFile
Write-Host "Wrote $($lines.Count) chunks to $OutFile"
