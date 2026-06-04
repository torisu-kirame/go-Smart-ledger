# Smart Ledger 前端工作区

| 目录 | 说明 |
|------|------|
| **desktop/** | Vue 3 桌面控制台（:25173） |
| **mobile/** | Vue 3 + Vant 移动 Web / Capacitor Android APK（:25175） |

## 桌面开发

```bash
make frontend-dev
```

访问 http://localhost:25173 ，API 代理至网关 http://localhost:28080

## 移动端

```bash
make mobile-dev      # Vite :25175
make mobile-apk      # Android Debug APK（需 JDK 17 + Android SDK）
```

详见 [mobile/README.md](./mobile/README.md)
