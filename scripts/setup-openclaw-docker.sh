#!/usr/bin/env bash
# Initialize config and start OpenClaw + Ollama via Docker Compose.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE=(docker compose --env-file .env.openclaw -f docker-compose.openclaw.yml)
ENV_FILE="$ROOT/.env.openclaw"
ENV_EXAMPLE="$ROOT/.env.openclaw.example"
CONFIG_DIR="$ROOT/data/openclaw/config"
CONFIG_FILE="$CONFIG_DIR/openclaw.json"
DOCKER_JSON="$ROOT/integrations/openclaw/openclaw.docker.json"

mkdir -p "$CONFIG_DIR" "$ROOT/data/openclaw/lancedb" "$ROOT/data/openclaw/auth-secrets"

if [[ ! -f "$CONFIG_FILE" ]]; then
  echo "Writing $CONFIG_FILE from openclaw.docker.json"
  cp "$DOCKER_JSON" "$CONFIG_FILE"
fi

if [[ ! -f "$ENV_FILE" ]]; then
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
  if grep -q '^OPENCLAW_GATEWAY_TOKEN=' "$ENV_FILE"; then
    sed -i.bak "s/^OPENCLAW_GATEWAY_TOKEN=.*/OPENCLAW_GATEWAY_TOKEN=$TOKEN/" "$ENV_FILE" && rm -f "$ENV_FILE.bak"
  else
    echo "OPENCLAW_GATEWAY_TOKEN=$TOKEN" >>"$ENV_FILE"
  fi
  echo "Generated OPENCLAW_GATEWAY_TOKEN in .env.openclaw"
fi

echo "Pulling images..."
"${COMPOSE[@]}" pull ollama openclaw-gateway 2>/dev/null || true

echo "Starting Ollama..."
"${COMPOSE[@]}" up -d ollama

CHAT_MODEL="${OLLAMA_CHAT_MODEL:-llama3.2}"
EMBED_MODEL="${OLLAMA_EMBED_MODEL:-nomic-embed-text}"

echo "Waiting for Ollama..."
for i in $(seq 1 60); do
  if "${COMPOSE[@]}" exec -T ollama ollama list >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo "Pulling chat model: $CHAT_MODEL"
"${COMPOSE[@]}" exec -T ollama ollama pull "$CHAT_MODEL" || echo "warn: pull $CHAT_MODEL failed (retry manually)"

echo "Pulling embed model: $EMBED_MODEL"
"${COMPOSE[@]}" exec -T ollama ollama pull "$EMBED_MODEL" || echo "warn: pull $EMBED_MODEL failed (retry manually)"

echo "Starting OpenClaw Gateway..."
"${COMPOSE[@]}" up -d openclaw-gateway

echo ""
echo "OpenClaw Docker stack is up."
echo "  Gateway UI:  http://127.0.0.1:${OPENCLAW_GATEWAY_PORT:-18789}"
echo "  Ollama:      http://127.0.0.1:${OLLAMA_HOST_PORT:-11434}"
echo "  Config:      data/openclaw/config/openclaw.json"
echo "  Token:       see OPENCLAW_GATEWAY_TOKEN in .env.openclaw (paste into Gateway Control UI)"
echo ""
echo "Smart Ledger console → Settings → AI:"
echo "  Ollama base URL:  http://127.0.0.1:11434/v1"
echo "  OpenClaw Gateway: http://127.0.0.1:18789"
echo ""
echo "CLI (optional): docker compose -f docker-compose.openclaw.yml --profile cli run --rm openclaw-cli status"
echo "Docs: docs/openclaw-integration.md"
