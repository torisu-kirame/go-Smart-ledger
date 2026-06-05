# FISCO BCOS 迁移至权威账本

Smart Ledger 自 **v0.14** 起将 **FISCO BCOS** 作为目标权威顶层链，替代 Chainscore MiniLedger。  
**v0.18+** 默认 `docker compose` / `make up` 已对接 FISCO BCOS 3.0。  
**v0.1.0–v0.13.0-miniledger** 标签线仍以 MiniLedger 为权威链（见 Git 标签 `miniledger-era`）。

## 版本提交计划（按工程量）

| 版本标签 | 状态 | 内容 |
|----------|------|------|
| `v0.13.0-miniledger` | ✅ | MiniLedger 时代线终点（含移动 Web/APK） |
| `v0.14.0` | ✅ | README：移动端 APK、版本标签说明、FISCO 路线图 |
| `v0.14.1-fisco.1` | ✅ | `pkg/chainstore` 抽象 + MiniLedger 适配器；`ledgersvc` 解耦 |
| `v0.14.2-fisco.2` | ✅ | FISCO 合约骨架、`docker-compose.fisco.yml`、FISCO Store 桩、配置项 |
| `v0.15.0-fisco.3` | ✅ | `putState`/`getState` KV 镜像 MiniLedger；链上键索引；`PrivateKeyHex` 写交易 |
| `v0.15.1-fisco.4` | ✅ | FISCO BCOS **3.0** 原生 JSON-RPC（`group0`/`20200`/`call`/`sendTransaction`）；跨平台纯 Go 客户端 |
| `v0.16.0-fisco.5` | ✅ | `txqueue` 可重试分类（`IsRetryable`/`ErrPermanent`）；队列项 `backend` 字段 |
| `v0.16.1-fisco.6` | ✅ | FISCO 3.0 链浏览器：`GetRaw` 映射 `/blocks`、`/consensus`、`/peers`、`/tx/recent` |
| `v0.17.0-fisco.7` | ✅ | `cmd/migrate-miniledger-to-fisco` 历史 KV 迁移 + 校验 |
| `v0.18.0-fisco.8` | ✅ | 默认 Compose 切 FISCO；MiniLedger 移入 `legacy` profile |

## 架构变化

```text
ledger-api
    └── pkg/chainstore.Store
            ├── miniledger.Adapter   （legacy，profile legacy）
            └── fisco.Store          （默认）
                    └── LedgerRegistry.sol + FISCO BCOS 3.0 JSON-RPC
```

## 合约与表设计

| 域 | MiniLedger（legacy） | FISCO BCOS（默认） |
|----|----------------------|-------------------|
| 账本元数据 / 事件 / 协作 / 复式 | `world_state` KV | `putState` / `getState`（同一键空间 `smartledger:…`） |
| 合约高级 API | — | `createLedger` / `appendEvent`（可选，v0.15 未用） |
| 封账锚定 | MiniLedger seal + 可选 EVM | 链上 `latestRoot` + 可选保留 EVM 双锚 |

合约源码：[`backend/contracts/fisco/LedgerRegistry.sol`](../backend/contracts/fisco/LedgerRegistry.sol)

## 部署 FISCO 开发链

1. 按 [`backend/infra/fisco/README.md`](../infra/fisco/README.md) 生成本地联盟链（官方 `build_chain.sh`）。
2. 部署合约（Foundry / WeBASE）并记录地址。
3. 在 `.env` 设置 `SL_FISCO_REGISTRY`、`SL_FISCO_PRIVATE_KEY`（见 `.env.example`）。
4. 启动全栈：

```bash
make up
```

`ledger-api` 默认使用 [`deploy/etc/ledger-api.docker.yaml`](../deploy/etc/ledger-api.docker.yaml)（`Chain.Backend: fisco`）。

## MiniLedger 回退（legacy）

```bash
make legacy-up
# 或
docker compose -f docker-compose.yml -f docker-compose.legacy.yml --profile legacy up -d
```

将 `SL_CHAIN_BACKEND=miniledger` 注入 `ledger-api` 并启动 `miniledger` 容器。

## 配置示例

```yaml
Chain:
  Backend: fisco
  FISCO:
    JSONRPCURL: http://127.0.0.1:20200
    GroupID: group0
    ChainID: chain0
    DisableSsl: true              # 本地 disable_ssl 建链
    RegistryContract: "0x..."
    PrivateKeyHex: "0x..."         # 或 SL_FISCO_PRIVATE_KEY / SL_FISCO_REGISTRY
    ExplorerURL: http://localhost:25002
MiniLedger:
  BaseURL: http://miniledger:24441   # legacy profile 仍可用
```

## 数据迁移

工具：[`backend/cmd/migrate-miniledger-to-fisco/`](../backend/cmd/migrate-miniledger-to-fisco/)

```bash
cd backend
go run ./cmd/migrate-miniledger-to-fisco \
  -miniledger http://127.0.0.1:24441 \
  -fisco-rpc http://127.0.0.1:20200 \
  -registry 0x... \
  -private-key 0x... \
  -verify
```

1. 从 MiniLedger 导出 `smartledger:%` 键值。
2. 按 meta 优先顺序调用 FISCO `putState`（索引自动重建）。
3. `-verify` 逐键比对 FISCO 读回值。
4. 只读窗口内禁止双写；切换 `Chain.Backend: fisco` 后重启 ledger-api。

## 前端影响

| 模块 | 变更 |
|------|------|
| `/chain` 页 | FISCO 3.0 RPC 区块/共识/节点（`GetRaw`）；官方 Tab 仍可用 WeBASE iframe |
| 仪表盘链状态 | 已返回 `backend: fisco` |
| 移动 / 桌面 API | 无变更（仍经 gateway JWT） |

## 回滚

使用 `make legacy-up` 或 `docker-compose.legacy.yml` + `--profile legacy`，将 `Chain.Backend` 切回 `miniledger`，可随时回退至 `v0.13.0-miniledger` 行为。
