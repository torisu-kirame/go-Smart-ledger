# Smart Ledger 全栈构建与启动
.PHONY: help build build-linux docker-build up down logs clean \
        frontend-install frontend-dev frontend-build dev-all start \
        openclaw-up openclaw-down openclaw-logs

help:
	@echo "Targets:"
	@echo "  make start             - 一键全栈：Docker（含 Nginx 前端 :25173）"
	@echo "  make up                - 同 start"
	@echo "  make build-linux       - 交叉编译全部 Go 服务"
	@echo "  make frontend-dev      - 本地 Vite 开发服（需后端已 up）"
	@echo "  make frontend-build    - 仅构建前端 dist"
	@echo "  make dev-all           - 本地前端开发（需后端已 up）"
	@echo "  make openclaw-up       - Docker 启动 OpenClaw + Ollama（见 docs/openclaw-integration.md）"
	@echo "  make openclaw-down     - 停止 OpenClaw 栈"

build:
	$(MAKE) -C backend build-local

build-linux:
	$(MAKE) -C backend build-linux

frontend-install:
	cd frontend/desktop && npm install --cache .npm-cache

frontend-dev: frontend-install
	cd frontend/desktop && npm run dev

frontend-build: frontend-install
	cd frontend/desktop && npm run build

docker-build: build-linux
	docker compose build

up: docker-build
	docker compose up -d
	@echo ""
	@echo "Web UI:     http://localhost:25173"
	@echo "Gateway:    http://localhost:28080/api/v1/health"
	@echo "MiniLedger: http://localhost:24441/dashboard"
	@echo "Login:      admin / admin123"

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	$(MAKE) -C backend clean
	docker compose down --rmi local 2>/dev/null || true

dev-all: frontend-install
	@echo "请确保已执行: make up"
	cd frontend/desktop && npm run dev

# 全栈 Docker（后端 + Nginx 前端）
start: up

openclaw-up:
	@if [ -f scripts/setup-openclaw-docker.sh ]; then chmod +x scripts/setup-openclaw-docker.sh && ./scripts/setup-openclaw-docker.sh; else echo "Run scripts/setup-openclaw-docker.ps1 on Windows"; exit 1; fi

openclaw-down:
	docker compose --env-file .env.openclaw -f docker-compose.openclaw.yml down

openclaw-logs:
	docker compose --env-file .env.openclaw -f docker-compose.openclaw.yml logs -f openclaw-gateway ollama
