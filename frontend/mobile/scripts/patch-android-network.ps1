# Patches Capacitor Android project for local HTTP API (cleartext).
$ErrorActionPreference = "Stop"
$Mobile = Split-Path -Parent $MyInvocation.MyCommand.Path
$Android = Join-Path $Mobile "android"
$ResXml = Join-Path $Android "app\src\main\res\xml"
$Template = Join-Path $Mobile "android-templates\network_security_config.xml"
$Manifest = Join-Path $Android "app\src\main\AndroidManifest.xml"

if (-not (Test-Path $Android)) { exit 0 }

New-Item -ItemType Directory -Force -Path $ResXml | Out-Null
Copy-Item $Template (Join-Path $ResXml "network_security_config.xml") -Force

if (Test-Path $Manifest) {
  $xml = Get-Content $Manifest -Raw
  if ($xml -notmatch 'usesCleartextTraffic') {
    $xml = $xml -replace 'android:theme="@style/AppTheme"', 'android:theme="@style/AppTheme"`n        android:usesCleartextTraffic="true"`n        android:networkSecurityConfig="@xml/network_security_config"'
    Set-Content $Manifest $xml -NoNewline
  }
}
