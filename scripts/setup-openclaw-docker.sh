#!/usr/bin/env bash
# @deprecated Use `make up` — OpenClaw is bundled in docker-compose.yml
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
echo "Note: OpenClaw is now included in the main stack (make up)."
chmod +x scripts/init-openclaw-config.sh
./scripts/init-openclaw-config.sh
docker compose --env-file .env.openclaw up -d ollama ollama-init openclaw-gateway
echo ""
echo "OpenClaw (integrated stack)"
echo "  Gateway UI:  http://127.0.0.1:${OPENCLAW_GATEWAY_PORT:-18789}"
echo "  Ollama:      http://127.0.0.1:${OLLAMA_HOST_PORT:-11434}/v1"
echo "  Token:       .env.openclaw → OPENCLAW_GATEWAY_TOKEN"
echo "  Full stack:  make up"
