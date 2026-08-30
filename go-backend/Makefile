# 外部交叉编译 + 增量构建（Go 构建缓存 + 按服务拆分目标）
# Docker 镜像仅 COPY bin/linux/*，不在容器内 go build
# 默认产物：Gin 单体 smart-ledger-api（./services/api）

.PHONY: tidy test build build-linux build-linux-api build-linux-ledger build-linux-storage build-linux-gateway build-linux-auth \
        build-local clean docker-build check-linux-bins

GO       ?= go
BIN_LINUX := bin/linux
BIN_LOCAL := bin
GOOS     ?= linux
GOARCH   ?= amd64
CGO      := 0
LDFLAGS  := -s -w

ifeq ($(OS),Windows_NT)
  EXE := .exe
  GOBUILD = cmd /C "set CGO_ENABLED=$(CGO)&& set GOOS=$(GOOS)&& set GOARCH=$(GOARCH)&& $(GO) build -trimpath -ldflags \"-s -w\""
  MKDIR_BIN = if not exist $(subst /,\,$1) mkdir $(subst /,\,$1)
else
  EXE :=
  GOBUILD = CGO_ENABLED=$(CGO) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -trimpath -ldflags "$(LDFLAGS)"
  MKDIR_BIN = mkdir -p $1
endif

API_SRC_DIRS     := pkg services/api services/auth/mount services/ledger/mount services/storage/mount services/gateway/mount
LEDGER_SRC_DIRS  := pkg/accounting pkg/auditexport pkg/domain pkg/evmanchor pkg/importfile \
	pkg/chainstore \
	pkg/importxlsx pkg/ipfsclient pkg/ledgerhd pkg/ledgersvc pkg/miniledgerclient pkg/snowflake \
	pkg/storage pkg/mq pkg/txqueue pkg/registry pkg/grpchsrv services/ledger
STORAGE_SRC_DIRS := pkg/storage services/storage
GATEWAY_SRC_DIRS := pkg/authjwt pkg/registry services/gateway
AUTH_SRC_DIRS    := pkg/authjwt pkg/captcha pkg/userstore services/auth

GOFILES_API     := $(foreach dir,$(API_SRC_DIRS),$(wildcard $(dir)/*.go) $(wildcard $(dir)/*/*.go) $(wildcard $(dir)/*/*/*.go))
GOFILES_LEDGER  := $(foreach dir,$(LEDGER_SRC_DIRS),$(wildcard $(dir)/*.go) $(wildcard $(dir)/*/*.go) $(wildcard $(dir)/*/*/*.go))
GOFILES_STORAGE := $(foreach dir,$(STORAGE_SRC_DIRS),$(wildcard $(dir)/*.go) $(wildcard $(dir)/*/*.go) $(wildcard $(dir)/*/*/*.go))
GOFILES_GATEWAY  := $(foreach dir,$(GATEWAY_SRC_DIRS),$(wildcard $(dir)/*.go) $(wildcard $(dir)/*/*.go) $(wildcard $(dir)/*/*/*.go))
GOFILES_AUTH     := $(foreach dir,$(AUTH_SRC_DIRS),$(wildcard $(dir)/*.go) $(wildcard $(dir)/*/*.go) $(wildcard $(dir)/*/*/*.go))

MOD_DEPS := go.mod go.sum

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...

build-local: $(BIN_LOCAL)/smart-ledger-api$(EXE)

$(BIN_LOCAL)/smart-ledger-api$(EXE):
	@$(call MKDIR_BIN,$(BIN_LOCAL))
	$(GO) build -trimpath -o $@ ./services/api

build-linux:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File ../scripts/build-linux.ps1 -Service api
else
	$(MAKE) build-linux-api
endif

build-linux-api: $(BIN_LINUX)/smart-ledger-api
build-linux-ledger: $(BIN_LINUX)/ledger-api
build-linux-storage: $(BIN_LINUX)/storage-api
build-linux-gateway: $(BIN_LINUX)/gateway-api
build-linux-auth: $(BIN_LINUX)/auth-api

$(BIN_LINUX)/smart-ledger-api: $(MOD_DEPS) $(GOFILES_API)
	@echo ">> building $@"
	@$(call MKDIR_BIN,$(BIN_LINUX))
	$(GOBUILD) -o $@ ./services/api

$(BIN_LINUX)/ledger-api: $(MOD_DEPS) $(GOFILES_LEDGER)
	@echo ">> building $@"
	@$(call MKDIR_BIN,$(BIN_LINUX))
	$(GOBUILD) -o $@ ./services/ledger

$(BIN_LINUX)/storage-api: $(MOD_DEPS) $(GOFILES_STORAGE)
	@echo ">> building $@"
	@$(call MKDIR_BIN,$(BIN_LINUX))
	$(GOBUILD) -o $@ ./services/storage

$(BIN_LINUX)/gateway-api: $(MOD_DEPS) $(GOFILES_GATEWAY)
	@echo ">> building $@"
	@$(call MKDIR_BIN,$(BIN_LINUX))
	$(GOBUILD) -o $@ ./services/gateway

$(BIN_LINUX)/auth-api: $(MOD_DEPS) $(GOFILES_AUTH)
	@echo ">> building $@"
	@$(call MKDIR_BIN,$(BIN_LINUX))
	$(GOBUILD) -o $@ ./services/auth

check-linux-bins:
	@test -f $(BIN_LINUX)/smart-ledger-api || (echo "missing smart-ledger-api — run: make build-linux-api" && exit 1)

docker-build: check-linux-bins
	@echo ">> binaries ready under $(BIN_LINUX)/"

clean:
	rm -rf $(BIN_LINUX) $(BIN_LOCAL)

build: build-local
