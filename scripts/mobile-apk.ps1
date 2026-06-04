# 构建 Smart Ledger Android Debug APK
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Mobile = Join-Path $Root "frontend\mobile"

Write-Host "==> npm install (mobile)"
Push-Location $Mobile
try {
  npm install --cache .npm-cache

  if (-not (Test-Path "capacitor.config.json")) {
    Write-Host "==> cap init"
    npx cap init "Smart Ledger" com.smartledger.mobile --web-dir dist
  }

  if (-not (Test-Path "android")) {
    Write-Host "==> cap add android"
    npx cap add android
  }

  Write-Host "==> patch android network (cleartext HTTP)"
  & powershell -NoProfile -ExecutionPolicy Bypass -File scripts\patch-android-network.ps1

  Write-Host "==> cap sync"
  npx cap sync android

  Write-Host "==> resolve JDK / Android SDK"
  $resolveOut = & powershell -NoProfile -ExecutionPolicy Bypass -File scripts\resolve-android-env.ps1 -MobileDir (Get-Location).Path 2>&1
  $resolveOut | ForEach-Object { Write-Host $_ }
  if ($LASTEXITCODE -ne 0) {
    throw "JDK / Android SDK 解析失败"
  }
  foreach ($line in $resolveOut) {
    if ($line -match '^JAVA_HOME=(.+)$') { $env:JAVA_HOME = $Matches[1].Trim() }
    if ($line -match '^ANDROID_SDK=(.+)$') { $env:ANDROID_HOME = $Matches[1].Trim() }
  }
  if (-not $env:JAVA_HOME) {
    throw "无法解析 JAVA_HOME，见上方错误说明"
  }
  $env:Path = "$env:JAVA_HOME\bin;$env:ANDROID_HOME\platform-tools;$env:Path"

  Write-Host "==> gradlew assembleDebug (JAVA_HOME=$env:JAVA_HOME)"
  Push-Location android
  try {
    cmd /c "gradlew.bat assembleDebug"
    if ($LASTEXITCODE -ne 0) {
      throw "Gradle 构建失败 (exit $LASTEXITCODE)"
    }
  } finally {
    Pop-Location
  }

  $apk = Join-Path $Mobile "android\app\build\outputs\apk\debug\app-debug.apk"
  if (Test-Path $apk) {
    Write-Host ""
    Write-Host "APK: $apk" -ForegroundColor Green
  } else {
    Write-Warning "构建完成但未找到 APK，请检查 android/app/build/outputs/apk/"
  }
} finally {
  Pop-Location
}
