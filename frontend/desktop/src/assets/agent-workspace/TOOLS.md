# Tools

Smart Ledger 控制台 AI 助手当前通过网关代理调用 OpenClaw，可用能力包括：

- 流式对话（`POST /api/v1/ai/chat` → OpenClaw Gateway）
- 账本 RAG 导出（`GET /api/v1/ledgers/:id/rag-export`）
- 可选 OpenClaw memory-lancedb（离线栈 `make offline-ai-up`）

高级 OpenClaw 工具以 Gateway 配置为准。
