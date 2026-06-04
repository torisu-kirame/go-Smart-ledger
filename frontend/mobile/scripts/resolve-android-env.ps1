# Resolve JAVA_HOME and Android SDK for Gradle; write android/local.properties.
param(
    [string]$MobileDir = (Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path))
)

$ErrorActionPreference = "Stop"

function Test-JavaHome {
    param([string]$JavaHomePath)
    if (-not $JavaHomePath) { return $false }
    $exe = Join-Path $JavaHomePath "bin\java.exe"
    if (-not (Test-Path $exe)) { return $false }
    try {
        $p = Start-Process -FilePath $exe -ArgumentList @("-version") -Wait -PassThru -NoNewWindow `
            -RedirectStandardError "$env:TEMP\sl-java-err.txt" -RedirectStandardOutput "$env:TEMP\sl-java-out.txt"
        return $p.ExitCode -eq 0
    } catch {
        return $false
    }
}

function Find-JavaHome {
    $candidates = New-Object System.Collections.Generic.List[string]
    if ($env:JAVA_HOME) { [void]$candidates.Add($env:JAVA_HOME) }
    foreach ($p in @(
            "D:\work\API\Android\Android studio\jbr",
            "D:\work\API\DataGrip.2026.1.3\jbr",
            "D:\work\API\DevEco Studio\jbr",
            "$env:ProgramFiles\Android\Android Studio\jbr",
            "$env:LOCALAPPDATA\Programs\Android\Android Studio\jbr"
        )) {
        [void]$candidates.Add($p)
    }
    foreach ($glob in @(
            "$env:ProgramFiles\Eclipse Adoptium\jdk-*",
            "$env:ProgramFiles\Java\jdk-*",
            "$env:ProgramFiles\Microsoft\jdk-*"
        )) {
        Get-ChildItem $glob -ErrorAction SilentlyContinue | ForEach-Object { [void]$candidates.Add($_.FullName) }
    }
    foreach ($javaDir in $candidates) {
        if (Test-JavaHome $javaDir) { return $javaDir }
    }
    return $null
}

function Find-AndroidSdk {
    $candidates = New-Object System.Collections.Generic.List[string]
    if ($env:ANDROID_HOME) { [void]$candidates.Add($env:ANDROID_HOME) }
    if ($env:ANDROID_SDK_ROOT) { [void]$candidates.Add($env:ANDROID_SDK_ROOT) }
    foreach ($p in @(
            "D:\work\API\Android\Sdk",
            "$env:LOCALAPPDATA\Android\Sdk",
            "$env:USERPROFILE\AppData\Local\Android\Sdk"
        )) {
        [void]$candidates.Add($p)
    }
    foreach ($sdk in $candidates) {
        if (-not $sdk -or -not (Test-Path $sdk)) { continue }
        $hasAdb = Test-Path (Join-Path $sdk "platform-tools\adb.exe")
        $hasPlatforms = Test-Path (Join-Path $sdk "platforms")
        $hasBuildTools = Test-Path (Join-Path $sdk "build-tools")
        if ($hasAdb -or $hasPlatforms -or $hasBuildTools) {
            return (Resolve-Path $sdk).Path
        }
    }
    return $null
}

$AndroidDir = Join-Path $MobileDir "android"
if (-not (Test-Path $AndroidDir)) {
    Write-Error "Android 工程不存在，请先运行 npx cap add android"
}

$javaHome = Find-JavaHome
if (-not $javaHome) {
    throw @"
未找到可用的 JDK（java -version 无法运行）。

当前 JAVA_HOME=$($env:JAVA_HOME)
若该目录下 java.exe 损坏，请安装 JDK 17+ 或设置 Android Studio 自带 JBR，例如：
  `$env:JAVA_HOME = 'D:\work\API\Android\Android studio\jbr'

也可在 Android Studio 中打开 frontend/mobile/android，菜单 Build → Build APK。
"@
}

$sdk = Find-AndroidSdk
if (-not $sdk) {
    throw @"
未找到 Android SDK。请安装 Android Studio 并配置 SDK，或设置：
  `$env:ANDROID_HOME = 'D:\work\API\Android\Sdk'
"@
}

$escaped = $sdk -replace '\\', '\\'
Set-Content -Path (Join-Path $AndroidDir "local.properties") -Value "sdk.dir=$escaped`n" -Encoding UTF8
Write-Host "JAVA_HOME => $javaHome"
Write-Host "ANDROID_SDK => $sdk"

# 供 mobile-apk.ps1 读取
Write-Output "JAVA_HOME=$javaHome"
Write-Output "ANDROID_SDK=$sdk"
