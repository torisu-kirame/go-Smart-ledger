# Smart Ledger Java Backend

Spring Boot 微服务栈，与 Go 后端 API 契约对齐，可独立 `make up-java` 启动。

## 模块

| 模块 | 端口 | 说明 |
|------|------|------|
| `auth-service` | 28887 | 登录、注册、JWT |
| `ledger-service` | 28888 | 账本 API（桩 + Agent Tool 联调数据） |
| `storage-service` | 28890 | 存储 API |
| `agent-service` | 28891 | **LangChain4j ReAct Agent**、SSE 对话、工作区持久化 |
| `gateway-service` | 28080 | Spring Cloud Gateway、JWT、AI 流式代理 |

## Agent 能力

- LangChain4j ReAct Agent + 账本 Tool Calling（`list_ledgers` / `search_ledger_rag` 等）
- 云端 OpenAI 兼容 API 与本地 Ollama 双模式（由前端 `baseUrl` 自动识别）
- Spring Security + 网关注入 `X-User-Id`，Tool 调用成员级账本隔离
- 会话与工作区 Markdown 持久化（`deploy/config/agent` 卷）

## 启动

Docker 容器使用 **`smart-ledger-java-*`** 前缀（Compose 项目 `smart-ledger-java`），与 Go 栈 `smart-ledger-go-*` 区分。

**方式一 — java-backend 目录一键启动（推荐）**

```bash
cd java-backend
docker compose up -d --build
# 或
make up
```

**方式二 — 仓库根目录**

```bash
make up-java
```

控制台：http://localhost:25173  
默认账号：`admin / admin123`

离线 Ollama（在 `java-backend` 目录）：

```bash
make offline-ai-up
# baseUrl: http://host.docker.internal:11434/v1
```
