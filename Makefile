# 根目录：先外部编译，再打包 Docker（增量编译见 backend/Makefile）
.PHONY: build build-linux docker-build up down logs clean

build:
	$(MAKE) -C backend build-local

build-linux:
	$(MAKE) -C backend build-linux

# 仅重编有变更的服务 + 构建镜像 + 启动
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
