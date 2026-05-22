# Smart Ledger 全栈构建与启动
.PHONY: help build build-linux docker-build up down logs clean \
        frontend-install frontend-dev frontend-build dev-all

help:
	@echo "Targets:"
	@echo "  make build-linux     - 交叉编译全部 Go 服务 (linux)"
	@echo "  make frontend-build  - 构建前端静态资源"
	@echo "  make docker-build    - 编译后端 + docker compose build"
	@echo "  make up              - 编译 + 启动 Docker 全栈"
	@echo "  make dev-all         - 本地开发（需先启动后端各服务）"
	@echo "  make frontend-dev    - 仅前端 Vite :25173"

build:
	$(MAKE) -C backend build-local

build-linux:
	$(MAKE) -C backend build-linux

frontend-install:
	cd frontend && npm install

frontend-dev: frontend-install
	cd frontend && npm run dev

frontend-build: frontend-install
	cd frontend && npm run build

# Docker：先后端二进制，再镜像
docker-build: build-linux
	docker compose build

up: docker-build
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	$(MAKE) -C backend clean
	docker compose down --rmi local 2>/dev/null || true

# 本地联调提示（后端需另开终端手动 go run 各服务，或使用 docker up）
dev-all: frontend-install
	@echo "请先确保网关 http://localhost:28080 可用（docker compose up 或本地启动各 api）"
	cd frontend && npm run dev
