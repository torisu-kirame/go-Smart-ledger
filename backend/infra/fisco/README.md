# FISCO BCOS 3.x 本地联盟链（Smart Ledger）

目标版本：**FISCO BCOS 3.0 产品线**（Air 单节点，官方 `build_chain.sh` v3.6.0）。

## 方式 A：Docker Compose（推荐）

无需在宿主机安装 FISCO 二进制，链节点已容器化：

```bash
# 构建镜像并启动 FISCO 链（**首次启动**会下载 build_chain 与二进制，需联网，约 3–8 分钟）
make fisco-up

# 或连同 ledger-api（FISCO 后端）一起启
make fisco-dev-up
```

| 端点 | 地址 |
|------|------|
| JSON-RPC（宿主机） | http://localhost:20200 |
| JSON-RPC（Compose 内） | http://fisco-node:20200 |
| 群组 / 链 ID | `group0` / `chain0` |

查看链状态：

```bash
curl -s -X POST http://127.0.0.1:20200 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getBlockNumber","params":[],"id":1}'
```

链数据持久化在 Docker volume `fisco-data`。查看建链 Admin 提示：

```bash
docker logs smart-ledger-fisco-node 2>&1 | head -30
cat admin.txt   # 在容器内 /data/admin.txt，或 docker exec 查看
```

### 对接 ledger-api

1. 部署 [`LedgerRegistry.sol`](../../contracts/fisco/LedgerRegistry.sol)（Console / WeBASE）。
2. 设置 `SL_FISCO_REGISTRY_CONTRACT` 与 `SL_FISCO_PRIVATE_KEY`（或编辑 [`deploy/etc/ledger-api.fisco.docker.yaml`](../../deploy/etc/ledger-api.fisco.docker.yaml) 后自行挂载）。
3. `make fisco-dev-up` 会通过环境变量切换 `ledger-api` 至 FISCO 后端。

## 方式 B：宿主机建链（高级）

适合多节点或需 Console 本地调试：

```bash
curl -#LO https://github.com/FISCO-BCOS/FISCO-BCOS/releases/download/v3.6.0/build_chain.sh
chmod +x build_chain.sh
bash build_chain.sh -l 127.0.0.1:4 -p 30300,20200
bash nodes/127.0.0.1/start_all.sh
```

Compose 仍可用 overlay，将 `ledger-api` 的 `SL_FISCO_JSONRPC` 改为 `http://host.docker.internal:20200`。

## 镜像说明

| 文件 | 说明 |
|------|------|
| [`Dockerfile`](Dockerfile) | Ubuntu 基础镜像；依赖在**首次容器启动**时由 `build_chain.sh` 拉取 |
| [`docker-entrypoint.sh`](docker-entrypoint.sh) | 首次运行建链 + 启动 `fisco-bcos`；数据写入 volume |
| RPC | `disable_ssl=true`，监听 `0.0.0.0:20200`（容器端口映射） |

构建参数（`docker compose build --build-arg`）：

| 环境变量 / ARG | 默认 |
|----------------|------|
| `FISCO_BUILD_CHAIN_URL` | GitHub `build_chain.sh` v3.6.0 |
| `FISCO_BUILD_CHAIN_MIRROR` | 腾讯云镜像（entrypoint 回退） |

## 故障排除

若首次启动时 `build_chain` 下载 **tassl** 或二进制返回 403/超时：

1. 在 WSL2 / Linux 用「方式 B」建链，得到 `nodes/` 目录。
2. 挂载到容器（跳过容器内 init）：

```yaml
volumes:
  - fisco-data:/data
  - ./backend/infra/fisco/nodes:/data/nodes
```

3. 删除旧 volume 后重试：`docker volume rm smart-ledger_fisco-data`

## 目录约定

| 路径 | 说明 |
|------|------|
| Docker volume `fisco-data` | 运行时链数据 |
| `nodes/` | 仅镜像内 `/opt/fisco/nodes`，不提交 Git |

详见 [`docs/fisco-bcos-migration.md`](../../../docs/fisco-bcos-migration.md)。
