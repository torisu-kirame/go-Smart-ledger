# MiniLedger Raft 多节点集群（F22）

Smart Ledger 默认 `docker compose up` 为**单节点** MiniLedger。需要 Raft 共识与多副本时，使用 overlay 编排：

```bash
docker compose -f deploy/compose/docker-compose.yml -f deploy/compose/docker-compose.raft.yml --profile raft up -d
```

## 端口

| 节点 | P2P | HTTP API | 浏览器 |
|------|-----|----------|--------|
| node1（bootstrap） | 24440 | 24441 | http://localhost:24441/dashboard |
| node2 | 24442 | 24443 | http://localhost:24443/dashboard |
| node3 | 24444 | 24445 | http://localhost:24445/dashboard |

`ledger-api` 应连接 **leader 节点 API**（通常为 node1）：`MiniLedger.BaseURL: http://miniledger-1:24441`。

## 本地 CLI（非 Docker）

```bash
# Node 1
miniledger init -d ./node1
miniledger start -d ./node1 --consensus raft --p2p-port 24440 --api-port 24441

# Node 2
miniledger init -d ./node2
miniledger join ws://127.0.0.1:24440 -d ./node2 --p2p-port 24442 --api-port 24443

# Node 3
miniledger init -d ./node3
miniledger join ws://127.0.0.1:24440 -d ./node3 --p2p-port 24444 --api-port 24445
```

## 控制台内嵌浏览器

Vue 控制台 **链浏览器** 路由 `/chain` 通过 Nginx 反代 `/dashboard/` 嵌入官方 Dashboard（MiniLedger 静态资源使用绝对路径 `/dashboard/*`）。`docker-entrypoint.sh` 会将 `node_modules/miniledger/dashboard` 链接到 `/app/dashboard`，否则节点 API 不注册 `/dashboard` 会返回 404。

## 验证

```bash
curl http://localhost:24441/status
curl http://localhost:24441/consensus
curl http://localhost:24443/peers
```
