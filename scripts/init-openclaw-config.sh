#!/usr/bin/env bash
# Prepare stack env and OpenClaw legacy config dirs (deploy/compose/docker-compose.yml).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ENV_FILE="$ROOT/deploy/env/stack.env"
ENV_EXAMPLE="$ROOT/deploy/env/stack.env.example"
CONFIG_DIR="$ROOT/deploy/config/openclaw"
CONFIG_FILE="$CONFIG_DIR/openclaw.json"
AGENT_DIR="$ROOT/deploy/config/agent"
DOCKER_JSON="$ROOT/integrations/openclaw/openclaw.docker.json"

mkdir -p "$CONFIG_DIR" "$ROOT/deploy/config/openclaw/lancedb" "$ROOT/deploy/config/openclaw/auth-secrets" "$AGENT_DIR"

if [[ ! -f "$CONFIG_FILE" ]]; then
  echo ">> OpenClaw: writing $CONFIG_FILE"
  cp "$DOCKER_JSON" "$CONFIG_FILE"
fi

node "$ROOT/scripts/repair-openclaw-config.js" "$CONFIG_FILE" "$DOCKER_JSON"

if [[ ! -f "$ENV_FILE" ]]; then
  echo ">> stack: creating deploy/env/stack.env from example"
  cp "$ENV_EXAMPLE" "$ENV_FILE"
fi

# shellcheck disable=SC1090
source "$ENV_FILE" 2>/dev/null || true

if [[ -z "${OPENCLAW_GATEWAY_TOKEN:-}" ]]; then
  if command -v openssl >/dev/null 2>&1; then
    TOKEN="$(openssl rand -hex 24)"
  else
    TOKEN="$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')"
  fi
  if grep -q '^OPENCLAW_GATEWAY_TOKEN=' "$ENV_FILE" 2>/dev/null; then
    sed -i.bak "s/^OPENCLAW_GATEWAY_TOKEN=.*/OPENCLAW_GATEWAY_TOKEN=$TOKEN/" "$ENV_FILE" && rm -f "$ENV_FILE.bak"
  else
    echo "OPENCLAW_GATEWAY_TOKEN=$TOKEN" >>"$ENV_FILE"
  fi
  echo ">> stack: generated OPENCLAW_GATEWAY_TOKEN in deploy/env/stack.env"
fi
