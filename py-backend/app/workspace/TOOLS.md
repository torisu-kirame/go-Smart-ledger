# Smart Ledger Agent Tools

Smart Ledger AI 助手由 py-backend（LangChain）驱动，可在对话中主动调用账本 API。

## 内置 Tools（`useTools: true`）

| Tool | 说明 | 输入示例 |
|------|------|----------|
| `list_ledgers` | 列出当前用户可访问的账本 | `{}` 或空 |
| `get_ledger_summary` | 账本元数据（序号、Merkle 根、锚定状态等） | `{"ledgerId":"..."}` |
| `search_ledger_rag` | 导出链上事件文本块，可按关键词过滤 | `{"ledgerId":"...","limit":40,"query":"报销"}` |
| `verify_ledger` | 校验 Merkle 完整性 | `{"ledgerId":"..."}` |
| `append_ledger_entry` | **记流水**（追加分录） | `{"amount":"200","note":"午餐-张总客户","category":"餐饮","entryType":"expense"}` |
| `get_financial_reports` | 财务报表（专业记账模式） | `{"period":"2026-05"}` |
| `get_ledger_budget` | 预算 / 预算执行分析 | `{"period":"2026-05","analysis":true}` |

助手页绑定账本后，`boundLedgerId` 会作为上述工具的默认账本 ID。

## 前端 Skills（豆包式芯片）

输入框上方可点击：记流水、对余额、贴标签、看趋势、做预算、账目总结、财务报表、导出 Excel。  
导出 Excel 由前端直接调用审计导出 API；其余技能注入引导话术并启用 Tool Calling。

## 知识来源

- 流式对话：`POST /api/v1/ai/chat`（JWT；`useTools` 启用 Tool Calling）
- 账本 RAG：`search_ledger_rag` 工具或 `GET /api/v1/ledgers/{id}/rag-export`
- 助手页绑定账本后，上下文仍会注入系统提示作为补充

## 行为约束

- 优先使用工具查询/写入真实账本数据；不足时明确说明
- 不编造交易、凭证号或链上哈希
- 记流水：时间、金额、用途三要素齐全再写入；用途尽量具体；转账勿记成支出
- 回复使用 Markdown，便于前端渲染
