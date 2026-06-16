# Maven package for all Java backend modules (local dev / CI)
$ErrorActionPreference = "Stop"
$Root = Join-Path $PSScriptRoot "..\java-backend"
Push-Location $Root
try {
    mvn -q package -DskipTests
    Write-Host ">> done: java-backend modules packaged"
}
finally {
    Pop-Location
}
