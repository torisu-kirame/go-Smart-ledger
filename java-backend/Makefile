.PHONY: up down logs build restart

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

build:
	docker compose build

restart: down up

offline-ai-up:
	docker compose --profile offline-ai up -d ollama ollama-init

help:
	@echo "Targets:"
	@echo "  make up              - 构建并启动 Java 全栈（含 Web + Agent）"
	@echo "  make down            - 停止并移除容器"
	@echo "  make logs            - 查看日志"
	@echo "  make offline-ai-up   - 额外启动 Ollama（本地模型）"
