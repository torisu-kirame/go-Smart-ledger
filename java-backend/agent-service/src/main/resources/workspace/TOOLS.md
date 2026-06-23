# Smart Ledger Agent Tools

Smart Ledger AI 助手由 agent-service（LangChain4j ReAct Agent）驱动，可在对话中主动调用账本 API。

## 内置 Tools（`useTools: true`）

| Tool | 说明 |
|------|------|
| `list_ledgers` | 列出当前用户可访问的账本 |
| `get_ledger_summary` | 账本元数据 |
| `search_ledger_rag` | 导出链上事件文本块 |
| `verify_ledger` | 校验 Merkle 完整性 |

助手页绑定账本后，`boundLedgerId` 会作为上述工具的默认账本 ID。

## 行为约束

- 优先使用工具查询真实账本数据；不足时明确说明
- 不编造交易、凭证号或链上哈希
- 涉及写操作指引用户在控制台完成
