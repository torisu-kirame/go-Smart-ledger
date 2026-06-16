# Smart Ledger 全栈构建与启动（Go / Java 双后端可选）
.PHONY: help build build-linux build-java docker-build docker-build-go docker-build-java \
        init-openclaw-config up up-go up-java down logs clean \
        frontend-install frontend-dev frontend-build \
        mobile-install mobile-dev mobile-build mobile-apk \
        dev-all start \
        offline-ai-up offline-ai-down offline-ai-logs \
        openclaw-gateway-up openclaw-logs openclaw-up openclaw-down \
        fisco-up fisco-dev-up fisco-logs fisco-down

COMPOSE_ROOT = --project-directory .
COMPOSE_MAIN = -f deploy/compose/docker-compose.yml
COMPOSE_ENV = --env-file deploy/env/stack.env
COMPOSE_GO = docker compose $(COMPOSE_ROOT) $(COMPOSE_ENV) $(COMPOSE_MAIN)
COMPOSE_JAVA = docker compose $(COMPOSE_ROOT) $(COMPOSE_ENV) $(COMPOSE_MAIN) -f deploy/compose/docker-compose.java.yml
COMPOSE_OFFLINE = docker compose $(COMPOSE_ROOT) $(COMPOSE_ENV) $(COMPOSE_MAIN) --profile offline-ai

help:
	@echo "Targets:"
	@echo "  make up-go             - Go 后端 + py-backend AI + Web（默认）"
	@echo "  make up-java           - Java 后端 + py-backend AI + Web"
	@echo "  make up                - 同 make up-go"
	@echo "  make offline-ai-up     - 额外启动 Ollama（离线模型，占磁盘）"
	@echo "  make build-linux       - 交叉编译 Go 服务"
	@echo "  make build-java        - Maven 打包 Java 服务"
	@echo "  make frontend-dev      - 本地 Vite 开发服（需后端已 up）"
	@echo "  make fisco-dev-up      - FISCO 链 + ledger-api（Go）"

COMPOSE_FISCO = docker compose $(COMPOSE_ROOT) $(COMPOSE_ENV) --env-file deploy/env/fisco.env $(COMPOSE_MAIN) -f deploy/compose/docker-compose.fisco.yml --profile fisco

fisco-up:
	$(COMPOSE_FISCO) up -d --build fisco-node

fisco-dev-up:
	$(COMPOSE_FISCO) up -d --build

fisco-logs:
	$(COMPOSE_FISCO) logs -f fisco-node

fisco-down:
	$(COMPOSE_FISCO) down

build:
	$(MAKE) -C go-backend build-local

build-linux:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/build-linux.ps1
else
	$(MAKE) -C go-backend build-linux
endif

build-java:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/build-java.ps1
else
	cd java-backend && mvn -q package -DskipTests
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

mobile-install:
	cd frontend/mobile && npm install --cache .npm-cache

mobile-dev: mobile-install
	cd frontend/mobile && npm run dev

mobile-build: mobile-install
	cd frontend/mobile && npm run build

mobile-apk: mobile-install
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/mobile-apk.ps1
else
	chmod +x scripts/mobile-apk.sh && ./scripts/mobile-apk.sh
endif

docker-build-go: build-linux init-openclaw-config
	$(COMPOSE_GO) build

docker-build-java: init-openclaw-config
	$(COMPOSE_JAVA) build auth-api ledger-api storage-api gateway-api ai-api

docker-build: docker-build-go

up-go: docker-build-go
	$(COMPOSE_GO) up -d
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/print-up-hints.ps1 -Backend go
else
	@echo ""
	@echo "Backend:    Go (go-backend/) + AI (py-backend/)"
	@echo "Web UI:     http://localhost:25173"
	@echo "Gateway:    http://localhost:28080/api/v1/health"
	@echo "MiniLedger: http://localhost:24441/dashboard"
	@echo "Login:      admin / admin123"
endif

up-java: docker-build-java
	$(COMPOSE_JAVA) up -d
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/print-up-hints.ps1 -Backend java
else
	@echo ""
	@echo "Backend:    Java (java-backend/) + AI (py-backend/)"
	@echo "Web UI:     http://localhost:25173"
	@echo "Gateway:    http://localhost:28080/api/v1/health"
	@echo "MiniLedger: http://localhost:24441/dashboard"
	@echo "Login:      admin / admin123"
endif

up: up-go

down:
	-$(COMPOSE_JAVA) down
	$(COMPOSE_GO) down

logs:
	$(COMPOSE_GO) logs -f

clean:
	$(MAKE) -C go-backend clean
	-$(COMPOSE_JAVA) down --rmi local
	-$(COMPOSE_GO) down --rmi local

dev-all: frontend-install
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/print-dev-hint.ps1
else
	@echo "请确保已执行: make up-go 或 make up-java"
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
endif

offline-ai-down:
	$(COMPOSE_OFFLINE) stop ollama

offline-ai-logs:
	$(COMPOSE_OFFLINE) logs -f ollama

openclaw-gateway-up: init-openclaw-config
	$(COMPOSE_GO) up -d openclaw-gateway

openclaw-logs:
	$(COMPOSE_GO) logs -f openclaw-gateway

openclaw-up: openclaw-gateway-up
openclaw-down: offline-ai-down
