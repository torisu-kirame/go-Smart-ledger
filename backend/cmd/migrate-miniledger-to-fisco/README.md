# migrate-miniledger-to-fisco

将 MiniLedger `world_state` 键值批量迁移至 FISCO BCOS 3.0 `LedgerRegistry` 合约（`putState`）。

## 用法

```bash
cd backend

# 1. 预览（不写链）
go run ./cmd/migrate-miniledger-to-fisco \
  -miniledger http://127.0.0.1:24441 \
  -dry-run -verbose

# 2. 正式迁移
go run ./cmd/migrate-miniledger-to-fisco \
  -miniledger http://127.0.0.1:24441 \
  -fisco-rpc http://127.0.0.1:20200 \
  -group group0 \
  -registry 0xYourLedgerRegistry \
  -private-key 0x... \
  -disable-ssl \
  -verify

# 3. 仅迁移前 N 条（调试）
go run ./cmd/migrate-miniledger-to-fisco ... -limit 100
```

## 行为说明

| 项 | 说明 |
|----|------|
| 数据源 | MiniLedger `SELECT ... WHERE key LIKE 'smartledger:%'` |
| 写入 | 每条调用 `FISCOStore.Submit` → `putState` + 自动重建链上索引 |
| 顺序 | 账本 meta 键优先，其余按 key 字典序 |
| 跳过 | FISCO 内部索引键（`__keys__`、`__index__`） |
| `-verify` | 迁移后逐键 `getState` 比对 JSON 语义 |

## 切换生产

1. 停止对 MiniLedger 的写入（只读窗口）。
2. 运行本工具并完成 `-verify`。
3. 将 `ledger-api` 的 `Chain.Backend` 改为 `fisco` 并重启。
4. 保留 MiniLedger 服务以便回滚。

详见 [docs/fisco-bcos-migration.md](../../../docs/fisco-bcos-migration.md)。
