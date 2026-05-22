# 外部交叉编译 + 增量构建（Go 构建缓存 + 按服务拆分目标）
# Docker 镜像仅 COPY bin/linux/*，不在容器内 go build

.PHONY: tidy test build build-linux build-linux-ledger build-linux-storage build-linux-gateway \
        build-local clean docker-build check-linux-bins

GO       ?= go
BIN_LINUX := bin/linux
BIN_LOCAL := bin
GOOS     ?= linux
GOARCH   ?= amd64
CGO      := 0
LDFLAGS  := -s -w
GOBUILD  := CGO_ENABLED=$(CGO) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -trimpath -ldflags "$(LDFLAGS)"

# 各服务依赖的源码目录（变更时仅重编对应服务）
LEDGER_SRC_DIRS   := pkg/domain pkg/ledgersvc pkg/miniledgerclient services/ledger
STORAGE_SRC_DIRS  := pkg/storage services/storage
GATEWAY_SRC_DIRS   := services/gateway

rwildcard = $(foreach d,$(wildcard $(1:=/*)),$(call rwildcard,$d,$2) $(filter $(subst *,%,$2),$d))
GOFILES_LEDGER  := $(foreach dir,$(LEDGER_SRC_DIRS),$(wildcard $(dir)/*.go) $(wildcard $(dir)/*/*.go) $(wildcard $(dir)/*/*/*.go))
GOFILES_STORAGE := $(foreach dir,$(STORAGE_SRC_DIRS),$(wildcard $(dir)/*.go) $(wildcard $(dir)/*/*.go) $(wildcard $(dir)/*/*/*.go))
GOFILES_GATEWAY  := $(foreach dir,$(GATEWAY_SRC_DIRS),$(wildcard $(dir)/*.go) $(wildcard $(dir)/*/*.go) $(wildcard $(dir)/*/*/*.go))

MOD_DEPS := go.mod go.sum

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...

# 本地开发（当前平台）
build-local: $(BIN_LOCAL)/ledger-api$(EXE) $(BIN_LOCAL)/storage-api$(EXE) $(BIN_LOCAL)/gateway-api$(EXE)

ifeq ($(OS),Windows_NT)
  EXE := .exe
else
  EXE :=
endif

$(BIN_LOCAL)/ledger-api$(EXE):
	@mkdir -p $(BIN_LOCAL)
	$(GO) build -trimpath -o $@ ./services/ledger

$(BIN_LOCAL)/storage-api$(EXE):
	@mkdir -p $(BIN_LOCAL)
	$(GO) build -trimpath -o $@ ./services/storage

$(BIN_LOCAL)/gateway-api$(EXE):
	@mkdir -p $(BIN_LOCAL)
	$(GO) build -trimpath -o $@ ./services/gateway

# Linux 二进制（供 Docker 使用）
build-linux: build-linux-ledger build-linux-storage build-linux-gateway

build-linux-ledger: $(BIN_LINUX)/ledger-api
build-linux-storage: $(BIN_LINUX)/storage-api
build-linux-gateway: $(BIN_LINUX)/gateway-api

$(BIN_LINUX)/ledger-api: $(MOD_DEPS) $(GOFILES_LEDGER)
	@echo ">> building $@"
	@mkdir -p $(BIN_LINUX)
	$(GOBUILD) -o $@ ./services/ledger

$(BIN_LINUX)/storage-api: $(MOD_DEPS) $(GOFILES_STORAGE)
	@echo ">> building $@"
	@mkdir -p $(BIN_LINUX)
	$(GOBUILD) -o $@ ./services/storage

$(BIN_LINUX)/gateway-api: $(MOD_DEPS) $(GOFILES_GATEWAY)
	@echo ">> building $@"
	@mkdir -p $(BIN_LINUX)
	$(GOBUILD) -o $@ ./services/gateway

check-linux-bins:
	@test -f $(BIN_LINUX)/ledger-api || (echo "missing $(BIN_LINUX)/ledger-api — run: make build-linux-ledger" && exit 1)
	@test -f $(BIN_LINUX)/storage-api || (echo "missing $(BIN_LINUX)/storage-api — run: make build-linux-storage" && exit 1)
	@test -f $(BIN_LINUX)/gateway-api || (echo "missing $(BIN_LINUX)/gateway-api — run: make build-linux-gateway" && exit 1)

docker-build: check-linux-bins
	@echo ">> binaries ready under $(BIN_LINUX)/"

clean:
	rm -rf $(BIN_LINUX) $(BIN_LOCAL)

# 兼容旧命令
build: build-local
