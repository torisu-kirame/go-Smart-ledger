# Tools

Smart Ledger 控制台 AI 助手当前通过网关代理 OpenClaw，能力与知识来源如下。

## 对话与记忆

- 流式对话：`POST /api/v1/ai/chat`（经 Smart Ledger 网关转发 OpenClaw Gateway）
- Agent 设定：本工作区 MD 文件（AGENTS、SOUL、API-REFERENCE 等）合并为 system prompt
- 可选 OpenClaw memory-lancedb（离线栈 `make offline-ai-up`）

## 账本数据

- 助手页绑定账本 → 注入同步摘要或 RAG 导出到上下文
- 全量导出：`GET /api/v1/ledgers/{id}/rag-export`
- 增量事件：`GET /api/v1/ledgers/{id}/sync`

## API 参考

- **API-REFERENCE.md**：用户向 Smart Ledger REST API（由 OpenAPI 自动生成）
- 完整机器可读规范：仓库根目录 `OpenAPI-swagger-user.json`
- 勿向用户推荐内部接口（健康检查、Agent 磁盘读写、链队列 retry 等，见 API-REFERENCE 末尾说明）

高级 OpenClaw 原生工具以 Gateway 配置为准；账本业务以 Smart Ledger API 为准。
