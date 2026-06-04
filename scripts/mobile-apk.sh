#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MOBILE="$ROOT/frontend/mobile"

cd "$MOBILE"
npm install --cache .npm-cache
npm run cap:sync
if [[ ! -d android ]]; then
  npx cap add android
  npm run cap:sync
fi
cd android
./gradlew assembleDebug
APK="$MOBILE/android/app/build/outputs/apk/debug/app-debug.apk"
echo ""
echo "APK: $APK"
