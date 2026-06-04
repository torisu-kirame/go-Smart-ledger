# migrate-miniledger-to-fisco

将 MiniLedger `world_state` 键值迁移至 FISCO BCOS `LedgerRegistry` 合约。

**状态**：v0.17.0-fisco.7 实现（当前为占位）。

## 规划用法

```bash
go run ./cmd/migrate-miniledger-to-fisco \
  -miniledger http://127.0.0.1:24441 \
  -fisco-rpc http://127.0.0.1:8545 \
  -registry 0x... \
  -dry-run
```

见 [docs/fisco-bcos-migration.md](../../../docs/fisco-bcos-migration.md)。
