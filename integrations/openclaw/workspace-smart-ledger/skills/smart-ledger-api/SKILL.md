---
name: smart-ledger-api
description: Create ledgers/sheets/entries via gateway tools (create_ledger, create_ledger_sheet, append_ledger_entry) or call_ledger_api. Prefer specialized tools.
---

# Smart Ledger API（OpenClaw）

控制台 → **gateway-api** → OpenClaw；账本写操作由 gateway 工具完成（用户 JWT）。

## 首选工具（按场景）

| 场景 | 工具 |
|------|------|
| 新建账本 | **`create_ledger`**（name；可选 type=private\|multi） |
| 导入 Markdown 表 | **`import_markdown_table`** → 转 CSV → `POST .../import/sheet-csv` |
| 导入 CSV | **`import_sheet_csv`** 或 `POST /api/v1/ledgers/{id}/import/sheet-csv` |
| 新建 Sheet | **`create_ledger_sheet`**（ledgerId, name, fields[{key,label,type,required}]） |
| 写一行 | **`append_ledger_entry`**（ledgerId, tableId, data） |
| 批量写 | **`append_ledger_entries_batch`**（最多 30 行） |
| 查询 | `list_ledgers` / `get_ledger_summary` / `search_ledger_rag` |
| 其它 REST | `call_ledger_api`（白名单） |

## 写分录格式（重要）

API 要求：

```json
{ "entry": { "tableId": "...", "signerId": "...", "data": { "field_key": "value" } } }
```

`append_ledger_entry` / `call_ledger_api`（POST `.../entries`）会自动包成上述结构。  
**禁止**把 `tableId` 放在 body 顶层而不包 `entry`。

建 Sheet 时不要创建「序号」列；多表账本写分录必须带真实 `tableId`（来自 create_ledger_sheet 或 get_ledger_summary）。

完整路径见 **API-REFERENCE.md**。破坏性操作先确认用户意图。
