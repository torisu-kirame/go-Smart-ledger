# Smart Ledger 全栈构建与启动
.PHONY: help build build-linux docker-build up down logs clean \
        frontend-install frontend-dev frontend-build dev-all

help:
	@echo "Targets:"
	@echo "  make build-linux       - 交叉编译全部 Go 服务"
	@echo "  make frontend-dev      - Vue3 桌面端开发服务器 :25173"
	@echo "  make up                - 编译 + Docker 启动"
	@echo "  make dev-all           - 提示后启动前端（需后端已运行）"

build:
	$(MAKE) -C backend build-local

build-linux:
	$(MAKE) -C backend build-linux

frontend-install:
	cd frontend/desktop && npm install

frontend-dev: frontend-install
	cd frontend/desktop && npm run dev

frontend-build: frontend-install
	cd frontend/desktop && npm run build

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

dev-all: frontend-install
	@echo "请确保: docker compose up 或本地已启动 gateway :28080"
	cd frontend/desktop && npm run dev
