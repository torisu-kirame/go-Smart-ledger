# Smart Ledger Agent Tools

Smart Ledger AI 助手由 py-backend（LangChain）驱动，可在对话中主动调用账本 API。

## 内置 Tools（`useTools: true`）

| Tool | 说明 | 输入示例 |
|------|------|----------|
| `list_ledgers` | 列出当前用户可访问的账本 | `{}` 或空 |
| `get_ledger_summary` | 账本元数据（序号、Merkle 根、锚定状态等） | `{"ledgerId":"..."}` |
| `search_ledger_rag` | 导出链上事件文本块，可按关键词过滤 | `{"ledgerId":"...","limit":40,"query":"报销"}` |
| `verify_ledger` | 校验 Merkle 完整性 | `{"ledgerId":"..."}` |

助手页绑定账本后，`boundLedgerId` 会作为上述工具的默认账本 ID。

## 知识来源

- 流式对话：`POST /api/v1/ai/chat`（JWT；`useTools` 启用 Tool Calling）
- 账本 RAG：`search_ledger_rag` 工具或 `GET /api/v1/ledgers/{id}/rag-export`
- 助手页绑定账本后，上下文仍会注入系统提示作为补充

## 行为约束

- 优先使用工具查询真实账本数据；不足时明确说明
- 不编造交易、凭证号或链上哈希
- 涉及写操作指引用户在控制台完成，不假装已代用户调用写 API
