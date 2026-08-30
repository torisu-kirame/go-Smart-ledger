# 生产加固（F27）

## 环境变量

| 变量 | 说明 |
|------|------|
| `SL_ENV=production` | 启用生产模式：禁止默认 JWT 密钥、自动 `Cookie Secure` |
| `SL_ACCESS_SECRET` | JWT 访问令牌密钥（与 gateway 校验一致） |
| `SL_REFRESH_SECRET` | 刷新令牌密钥（仅 auth-api） |
| `SL_COOKIE_SECURE` | `true` 时 refresh cookie 带 `Secure` |
| `SL_COOKIE_DOMAIN` | 可选 Cookie `Domain` |

复制根目录 [`.env.example`](../.env.example) 为 `.env` 后填入随机密钥。

## 网关限流

`gateway-api` 配置 `Security.RateLimit`（见 `deploy/etc/gateway-api.docker.yaml`）：

- 全局限流：默认 600 请求/分钟/IP
- 登录/注册/验证码：30 请求/分钟/IP
- `/api/v1/health` 不计入限流

反向代理后请保持 `TrustForwardedProto: true`，以便识别 `X-Forwarded-For`。

## HTTPS

开发环境使用 HTTP；生产建议在 Nginx / 负载均衡终止 TLS：

1. 使用 [`deploy/nginx/ssl.conf.example`](../deploy/nginx/ssl.conf.example) 作为 `web` 容器配置参考。
2. 设置 `SL_COOKIE_SECURE=true` 或 `SL_ENV=production`。
3. 将前端与 API 统一在 `https://` 域名下，避免混合内容。

```bash
# 可选：带自签名证书的 HTTPS 前端（profile）
docker compose -f deploy/compose/docker-compose.yml -f deploy/compose/docker-compose.https.yml --profile https up -d
```

## 密钥勿入库

- 勿将 `.env`、私钥、`HDWallet.Mnemonic` 提交到 Git。
- Docker 通过 `env_file` 注入（见 `deploy/compose/docker-compose.yml`）。
