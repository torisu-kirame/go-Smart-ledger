# FISCO BCOS 本地联盟链（Smart Ledger）

Smart Ledger 不内置 FISCO 二进制，请使用 [FISCO BCOS 官方文档](https://fisco-bcos-doc.readthedocs.io/) 生成开发链。

## 推荐步骤

1. 克隆 FISCO BCOS 建链脚本（示例）：

```bash
curl -LO https://github.com/FISCO-BCOS/FISCO-BCOS/releases/download/v3.6.0/build_chain.sh
chmod +x build_chain.sh
bash build_chain.sh -l 127.0.0.1:4 -p 30300,20200,8545 -e 8545
```

2. 启动节点后 JSON-RPC 通常为 `http://127.0.0.1:8545`（group 1）。

3. 部署 [`../../contracts/fisco/LedgerRegistry.sol`](../../contracts/fisco/LedgerRegistry.sol)（Solidity 0.8，FISCO 支持 EVM）。

4. 将合约地址写入 `ledger-api` 的 `Chain.FISCO.RegistryContract`。

5. 使用 Compose overlay：

```bash
docker compose -f docker-compose.yml -f docker-compose.fisco.yml --profile fisco up -d
```

若节点跑在宿主机，将 `Chain.FISCO.JSONRPCURL` 设为 `http://host.docker.internal:8545`（Windows/Mac Docker Desktop）。

## 目录约定

| 路径 | 说明 |
|------|------|
| `nodes/` | 建链输出（gitignore，本地生成） |
| `certs/` | SDK 连接证书（从 nodes 复制） |

详见 [`docs/fisco-bcos-migration.md`](../../../docs/fisco-bcos-migration.md)。
