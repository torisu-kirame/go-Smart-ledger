# 构建 Smart Ledger Android Debug APK
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Mobile = Join-Path $Root "frontend\mobile"

Write-Host "==> npm install (mobile)"
Push-Location $Mobile
try {
  npm install --cache .npm-cache
  Write-Host "==> vite build + cap sync"
  if (-not (Test-Path "capacitor.config.json")) {
    npx cap init "Smart Ledger" com.smartledger.mobile --web-dir dist
  }
  npm run cap:sync
  if (-not (Test-Path "android")) {
    Write-Host "==> cap add android"
    npx cap add android
  }
  Write-Host "==> patch android network (cleartext HTTP)"
  powershell -NoProfile -ExecutionPolicy Bypass -File scripts\patch-android-network.ps1
  npm run cap:sync
  Write-Host "==> gradlew assembleDebug"
  Push-Location android
  if (Test-Path ".\gradlew.bat") {
    .\gradlew.bat assembleDebug
  } else {
    .\gradlew assembleDebug
  }
  Pop-Location
  $apk = Join-Path $Mobile "android\app\build\outputs\apk\debug\app-debug.apk"
  if (Test-Path $apk) {
    Write-Host ""
    Write-Host "APK: $apk" -ForegroundColor Green
  }
} finally {
  Pop-Location
}
