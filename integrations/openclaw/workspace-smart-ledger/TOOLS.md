# Tools

Smart Ledger AI：**控制台 → gateway-api → OpenClaw Gateway**。

## 账本操作

由 gateway 工具循环执行（用户 JWT）：

| 工具 | 用途 |
|------|------|
| **`import_markdown_table`** | MD→CSV→`/import/sheet-csv` 批量导入（首选） |
| **`import_sheet_csv`** | 直接提交 CSV 文本到 sheet-csv API |
| **`create_ledger`** | 创建账本（simple + 多表默认开） |
| **`create_ledger_sheet`** | 开多表并创建 Sheet（fields） |
| **`append_ledger_entry`** | 追加一条（自动 `{entry:{...}}`） |
| **`append_ledger_entries_batch`** | 批量追加（表格场景请用 sheet-csv） |
| `list_ledgers` / `get_ledger_summary` / `search_ledger_rag` | 查询 |
| `call_ledger_api` | 白名单 REST（entries 会自动 wrap） |

技能：`skills/smart-ledger-api/SKILL.md`

## 对话

- `POST /api/v1/ai/chat`（网关编排；推理经 OpenClaw）
