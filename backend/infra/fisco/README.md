# FISCO BCOS 3.0 本地联盟链（Smart Ledger）

Smart Ledger 目标链为 **FISCO BCOS 3.x**（Air 建链），请使用 [官方文档](https://fisco-bcos-doc.readthedocs.io/) 生成开发链。

## 推荐步骤（FISCO BCOS 3.0 Air）

1. 下载建链脚本（示例 v3.6.0，属 3.0 产品线）：

```bash
curl -LO https://github.com/FISCO-BCOS/FISCO-BCOS/releases/download/v3.6.0/build_chain.sh
chmod +x build_chain.sh
bash build_chain.sh -l 127.0.0.1:4 -p 30300,20200
```

2. 启动后 JSON-RPC 默认为 **`http://127.0.0.1:20200`**，群组 **`group0`**。

3. 开发环境可在节点 `config.ini` 中设置 `[rpc] disable_ssl=true`，并在 ledger-api 配置 `DisableSsl: true`（生产请使用 TLS 证书）。

4. 用控制台或 WeBASE 部署 [`../../contracts/fisco/LedgerRegistry.sol`](../../contracts/fisco/LedgerRegistry.sol)（Solidity 0.8）。

5. 将合约地址与链上写账户私钥写入 `ledger-api`：

```yaml
Chain:
  Backend: fisco
  FISCO:
    JSONRPCURL: http://127.0.0.1:20200
    GroupID: group0
    ChainID: chain0
    DisableSsl: true          # 仅本地 disable_ssl 建链
    RegistryContract: "0x..."
    PrivateKeyHex: "0x..."      # 或 SL_FISCO_PRIVATE_KEY
```

6. 若从 MiniLedger 迁移历史数据，先只读 MiniLedger，再运行迁移工具（见 [`cmd/migrate-miniledger-to-fisco/README.md`](../../cmd/migrate-miniledger-to-fisco/README.md)）：

```bash
cd backend
go run ./cmd/migrate-miniledger-to-fisco \
  -miniledger http://127.0.0.1:24441 \
  -fisco-rpc http://127.0.0.1:20200 \
  -registry 0x... \
  -private-key 0x... \
  -disable-ssl \
  -verify
```

7. 启动全栈（v0.18+ 默认 FISCO，无需 fisco overlay）：

```bash
make up
```

MiniLedger 回退：`make legacy-up`（见 [`docker-compose.legacy.yml`](../../../docker-compose.legacy.yml)）。

Docker 内访问宿主机节点：`http://host.docker.internal:20200`（Windows/Mac Docker Desktop）。

## RPC 说明

| 项 | FISCO BCOS 3.0 |
|----|----------------|
| 块高 | `getBlockNumber`，params `[]` |
| 读合约 | `call`，transaction `{to, data}` |
| 写合约 | `sendTransaction`，`[groupID, node, signedTxHex, withProof]` |

`ledger-api` 通过 `pkg/chainstore` 纯 Go HTTP 客户端对接上述接口（跨平台编译，无需 bcos-c-sdk）。

## 目录约定

| 路径 | 说明 |
|------|------|
| `nodes/` | 建链输出（gitignore，本地生成） |
| `certs/` | SDK 连接证书（TLS 建链时从 nodes 复制） |

详见 [`docs/fisco-bcos-migration.md`](../../../docs/fisco-bcos-migration.md)。
