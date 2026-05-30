# Smart Ledger 全栈构建与启动（含 OpenClaw Gateway；Ollama 离线可选）
.PHONY: help build build-linux docker-build init-openclaw-config up down logs clean \
        frontend-install frontend-dev frontend-build dev-all start \
        offline-ai-up offline-ai-down offline-ai-logs \
        openclaw-gateway-up openclaw-logs openclaw-up openclaw-down

COMPOSE_ENV = --env-file .env.openclaw
COMPOSE = docker compose $(COMPOSE_ENV)
COMPOSE_OFFLINE = docker compose $(COMPOSE_ENV) --profile offline-ai

help:
	@echo "Targets:"
	@echo "  make up                - 账本 + Web + OpenClaw Gateway"
	@echo "  make offline-ai-up     - 额外启动 Ollama（离线模型，占磁盘）"
	@echo "  make build-linux       - 交叉编译全部 Go 服务"
	@echo "  make frontend-dev      - 本地 Vite 开发服（需后端已 up）"

build:
	$(MAKE) -C backend build-local

build-linux:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/build-linux.ps1
else
	$(MAKE) -C backend build-linux
endif

init-openclaw-config:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/init-openclaw-config.ps1
else
	@chmod +x scripts/init-openclaw-config.sh && ./scripts/init-openclaw-config.sh
endif

frontend-install:
	cd frontend/desktop && npm install --cache .npm-cache

frontend-dev: frontend-install
	cd frontend/desktop && npm run dev

frontend-build: frontend-install
	cd frontend/desktop && npm run build

docker-build: build-linux
	$(COMPOSE) build

up: docker-build init-openclaw-config
	$(COMPOSE) up -d
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/print-up-hints.ps1
else
	@echo ""
	@echo "Web UI:     http://localhost:25173"
	@echo "Gateway:    http://localhost:28080/api/v1/health"
	@echo "OpenClaw:   http://localhost:18789  (token: .env.openclaw)"
	@echo "MiniLedger: http://localhost:24441/dashboard"
	@echo "Login:      admin / admin123"
	@echo ""
	@echo "AI：设置 → AI → API Key + Token → 测试连接（Gateway 可填 http://127.0.0.1:18789）"
endif

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f

clean:
	$(MAKE) -C backend clean
	-$(COMPOSE) down --rmi local

dev-all: frontend-install
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/print-dev-hint.ps1
else
	@echo "请确保已执行: make up"
endif
	cd frontend/desktop && npm run dev

start: up

offline-ai-up: init-openclaw-config
	$(COMPOSE_OFFLINE) up -d ollama ollama-init
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/print-offline-ai-hints.ps1
else
	@echo ""
	@echo "Ollama: http://localhost:11434/v1"
	@echo "设置 → AI：选择 Ollama，API 地址 http://127.0.0.1:11434/v1"
endif

offline-ai-down:
	$(COMPOSE_OFFLINE) stop ollama

offline-ai-logs:
	$(COMPOSE_OFFLINE) logs -f ollama

openclaw-gateway-up: init-openclaw-config
	$(COMPOSE) up -d openclaw-gateway

openclaw-logs:
	$(COMPOSE) logs -f openclaw-gateway

# 兼容旧命令
openclaw-up: openclaw-gateway-up
openclaw-down: offline-ai-down
