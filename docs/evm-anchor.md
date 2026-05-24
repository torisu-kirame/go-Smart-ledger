# 公链 / L2 合约锚定（F29）

封账时除写入 MiniLedger 外，可将 Merkle 根锚定到 EVM 合约 `LedgerAnchor`（见 [`backend/contracts/LedgerAnchor.sol`](../backend/contracts/LedgerAnchor.sol)）。

## 启用步骤

1. 启动本地链（可选）：

```bash
docker compose -f docker-compose.evm.yml up -d
```

2. 部署合约（需 Foundry `forge`）：

```bash
cd backend/contracts
forge create LedgerAnchor --rpc-url http://127.0.0.1:8545 --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
```

3. 配置 ledger-api（`deploy/etc/ledger-api.docker.yaml` 或环境变量）：

```yaml
ExternalAnchor:
  Enabled: true
  RPCURL: http://anvil:8545
  ChainID: 31337
  ChainName: anvil
  Contract: "0x..."   # forge create 输出的地址
  ExplorerURLTemplate: ""
```

环境变量（优先级更高）：

- `SL_EVM_ANCHOR_ENABLED=true`
- `SL_EVM_RPC_URL`
- `SL_EVM_CHAIN_ID`
- `SL_EVM_CONTRACT`
- `SL_EVM_ANCHOR_PRIVATE_KEY`（**仅环境变量**，勿写入 YAML）

4. 重建 ledger-api：`docker compose build ledger-api && docker compose up -d ledger-api`

## 行为

- 账本「封账并锚定」成功后，若 EVM 已启用，会调用 `anchor(ledgerId, merkleRoot, seqFrom, seqTo)`。
- 链上交易哈希写入账本元数据 `externalAnchor`，并追加 `ExternalAnchored` 事件。
- 未启用 EVM 时行为与原先一致（仅 MiniLedger）。

## 前端

账本详情页展示「链外锚定」交易哈希；若配置了 `ExplorerURLTemplate`，提供区块浏览器链接。
