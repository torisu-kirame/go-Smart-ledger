#!/usr/bin/env bash
# Clone OpenClaw into project root (not committed; see .gitignore).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$ROOT/openclaw"

if [[ -d "$DEST/.git" ]]; then
  echo "openclaw/ already exists — skip clone."
  exit 0
fi

echo "Cloning OpenClaw into openclaw/ ..."
git clone --depth 1 https://github.com/openclaw/openclaw.git "$DEST"

WS_SRC="$ROOT/integrations/openclaw/workspace-smart-ledger"
WS_DST="$DEST/workspace-smart-ledger"
if [[ ! -e "$WS_DST" ]]; then
  ln -sf "$WS_SRC" "$WS_DST"
fi

echo ""
echo "Next: cd openclaw && pnpm install && pnpm openclaw onboard --install-daemon"
echo "See docs/openclaw-integration.md"
