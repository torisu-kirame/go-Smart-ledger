# Clone OpenClaw into project root (not committed; see .gitignore).
# Requires: git, Node 22+, pnpm (https://pnpm.io)
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$Dest = Join-Path $Root "openclaw"

if (Test-Path (Join-Path $Dest ".git")) {
    Write-Host "openclaw/ already exists — skip clone. Run: cd openclaw && pnpm install"
    exit 0
}

Write-Host "Cloning OpenClaw into openclaw/ ..."
git clone --depth 1 https://github.com/openclaw/openclaw.git $Dest

Write-Host "Linking Smart Ledger workspace skill..."
$WsSrc = Join-Path $Root "integrations\openclaw\workspace-smart-ledger"
$WsDst = Join-Path $Dest "workspace-smart-ledger"
if (-not (Test-Path $WsDst)) {
    if ($IsWindows -or $env:OS -match "Windows") {
        cmd /c mklink /J `"$WsDst`" `"$WsSrc`"
    } else {
        New-Item -ItemType SymbolicLink -Path $WsDst -Target $WsSrc | Out-Null
    }
}

Write-Host ""
Write-Host "Next steps:"
Write-Host "  cd openclaw"
Write-Host "  pnpm install"
Write-Host "  pnpm openclaw onboard --install-daemon"
Write-Host "  Copy integrations/openclaw/openclaw.example.json -> openclaw/openclaw.json (merge with onboard output)"
Write-Host "  See docs/openclaw-integration.md"
