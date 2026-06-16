# Smart Ledger — Java 后端（Spring Boot 3）

Smart Ledger 微服务后端，Spring Boot 3 + Spring Cloud Gateway，与 Vue 桌面端配套使用。

## 模块

| 模块 | 端口 | 说明 |
|------|------|------|
| `gateway-service` | 28080 | API 网关、JWT 校验、路由转发 |
| `auth-service` | 28887 | 登录、验证码、双令牌认证 |
| `ledger-service` | 28888 | 账本与链上 API |
| `storage-service` | 28890 | 备份与对象存储 |
| `common` | — | JWT、验证码、统一错误响应 |

## 启动

```bash
make up-java
```

## 本地开发

```bash
cd java-backend
mvn -pl auth-service -am spring-boot:run
# 另开终端分别启动 ledger / storage / gateway
```

## 当前进度

- 已完成：认证、网关、健康检查、MiniLedger 探活
- 开发中：账本 CRUD、链上写入、AI 助手接口
