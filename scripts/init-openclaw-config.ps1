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

# If deploy config has empty provider keys, seed from legacy data/openclaw/config when present
$LegacyConfig = Join-Path $Root "data\openclaw\config\openclaw.json"
if (Test-Path $LegacyConfig) {
    node -e @"
const fs=require('fs'); const path=require('path');
const { syncOpenClawProviderAuth } = require('./scripts/openclaw-auth-sync');
const legacy=JSON.parse(fs.readFileSync(process.argv[1],'utf8'));
const dstPath=process.argv[2];
const dst=JSON.parse(fs.readFileSync(dstPath,'utf8'));
const providers=legacy.models?.providers||{};
let seeded=false;
for (const [name, block] of Object.entries(providers)) {
  const key=String(block?.apiKey||'').trim();
  if (!key) continue;
  if (!dst.models) dst.models={mode:'merge',providers:{}};
  if (!dst.models.providers) dst.models.providers={};
  if (!dst.models.providers[name]) dst.models.providers[name]={...block};
  else if (!String(dst.models.providers[name].apiKey||'').trim()) {
    dst.models.providers[name].apiKey=key;
  } else continue;
  seeded=true;
}
if (seeded) {
  syncOpenClawProviderAuth(dst, path.dirname(dstPath));
  fs.writeFileSync(dstPath, JSON.stringify(dst,null,2)+'\n');
  console.log('>> OpenClaw: seeded provider API keys from data/openclaw/config');
} else {
  syncOpenClawProviderAuth(dst, path.dirname(dstPath));
  fs.writeFileSync(dstPath, JSON.stringify(dst,null,2)+'\n');
}
"@ $LegacyConfig $ConfigFile
} else {
    node -e "const fs=require('fs');const path=require('path');const {syncOpenClawProviderAuth}=require('./scripts/openclaw-auth-sync');const p=process.argv[1];const c=JSON.parse(fs.readFileSync(p,'utf8'));if(syncOpenClawProviderAuth(c,path.dirname(p))){fs.writeFileSync(p,JSON.stringify(c,null,2)+'\n');console.log('>> OpenClaw: synced auth-profiles.json')}" $ConfigFile
}

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
