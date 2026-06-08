# FISCO BCOS 迁移至权威账本

Smart Ledger 自 **v0.14** 起将 **FISCO BCOS** 作为目标权威顶层链，替代 Chainscore MiniLedger。  
**v0.1.0–v0.13.0-miniledger** 标签线仍以 MiniLedger 为权威链（见 Git 标签 `miniledger-era`）。

## 版本提交计划（按工程量）

| 版本标签 | 状态 | 内容 |
|----------|------|------|
| `v0.13.0-miniledger` | ✅ | MiniLedger 时代线终点（含移动 Web/APK） |
| `v0.14.0` | ✅ | README：移动端 APK、版本标签说明、FISCO 路线图 |
| `v0.14.1-fisco.1` | ✅ | `pkg/chainstore` 抽象 + MiniLedger 适配器；`ledgersvc` 解耦 |
| `v0.14.2-fisco.2` | ✅ | FISCO 合约骨架、`docker-compose.fisco.yml`、FISCO Store 桩、配置项 |
| `v0.15.0-fisco.3` | ✅ | `putState`/`getState` KV 镜像 MiniLedger；链上键索引；`PrivateKeyHex` 写交易 |
| `v0.15.1-fisco.4` | ✅ | FISCO BCOS **3.x 原生 JSON-RPC**（`call`/`sendTransaction`/`getBlockNumber`）；纯 Go 交易签名；`/chain` 浏览器数据 |
| `v0.16.0-fisco.5` | ⬜ | `txqueue` 适配 FISCO 交易回执；NSQ 重试语义 |
| `v0.16.1-fisco.6` | ⬜ | 前端 `/chain` 页 FISCO 展示优化 |
| `v0.17.0-fisco.7` | ⬜ | `cmd/migrate-miniledger-to-fisco` 历史数据迁移 |
| `v0.18.0-fisco` | ⬜ | 默认 Compose 切换 FISCO；MiniLedger 移入 `legacy` profile |

## 架构变化

```text
ledger-api
    └── pkg/chainstore.Store
            ├── miniledger.Adapter   （legacy，profile miniledger）
            └── fisco.Store          （目标默认，FISCO BCOS 3.x 原生 JSON-RPC）
                    └── LedgerRegistry.sol putState/getState
```

## 合约与表设计

| 域 | MiniLedger（旧） | FISCO BCOS（新） |
|----|------------------|------------------|
| 账本元数据 / 事件 / 协作 / 复式 | `world_state` KV | `putState` / `getState`（同一键空间 `smartledger:…`） |
| 合约高级 API | — | `createLedger` / `appendEvent`（可选，v0.15 未用） |
| 封账锚定 | MiniLedger seal + 可选 EVM | 链上 `latestRoot` + 可选保留 EVM 双锚 |

合约源码：[`backend/contracts/fisco/LedgerRegistry.sol`](../backend/contracts/fisco/LedgerRegistry.sol)

## 部署 FISCO 开发链

**推荐（Docker）**：`make fisco-up` 启动内置 `fisco-node` 容器（FISCO BCOS 3.6 / Air 单节点，RPC `:20200`）。  
详见 [`backend/infra/fisco/README.md`](../backend/infra/fisco/README.md)。

1. （可选）宿主机建链见 infra README「方式 B」。
2. 部署合约（Foundry / WeBASE / Console）并记录地址。
3. 启动 overlay：

```bash
# 仅 FISCO 链
make fisco-up

# FISCO 链 + ledger-api（需先 make build-linux）
make fisco-dev-up
```

4. 切换 ledger-api 配置（示例 [`deploy/etc/ledger-api.fisco.docker.yaml`](../deploy/etc/ledger-api.fisco.yaml)）：

```yaml
Chain:
  Backend: fisco
  FISCO:
    JSONRPCURL: http://127.0.0.1:20200
    GroupID: group0
    ChainID: chain0
    RegistryContract: "0x..."
    PrivateKeyHex: "0x..."   # 或环境变量 SL_FISCO_PRIVATE_KEY
    ExplorerURL: http://localhost:25002
MiniLedger:
  BaseURL: http://miniledger:24441   # legacy profile 仍可用
```

## 数据迁移（规划）

工具目录：`backend/cmd/migrate-miniledger-to-fisco/`（v0.17 实现）

1. 从 MiniLedger SQL 查询导出全部 `world_state` 键值。
2. 按账本 ID 分组，映射为合约调用批次。
3. 写入 FISCO 后校验 Merkle 根与事件序号一致。
4. 只读窗口内禁止双写；切换 `Chain.Backend` 后重启 ledger-api。

## 前端影响

| 模块 | 变更 |
|------|------|
| `/chain` 页 | 改接 FISCO 区块/交易 API 或 WeBASE 浏览器 iframe |
| 仪表盘链状态 | 已返回 `backend: fisco` |
| 移动 / 桌面 API | 无变更（仍经 gateway JWT） |

## 回滚

保留 `docker-compose` 中 `miniledger` 服务与 `Chain.Backend: miniledger` 配置，可随时回退至 `v0.13.0-miniledger` 行为。
