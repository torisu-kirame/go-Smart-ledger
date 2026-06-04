# Smart Ledger 移动端

Vue 3 + Vant 4 + Capacitor 7，支持 **移动 Web** 与 **Android APK**。

## 功能

- 底部 Tab：首页 / 账本 / 协作 / 我的
- 登录注册、账本列表与创建、账本详情与简单记一笔
- 好友与团队（协作 Tab）
- 个人资料、服务器地址配置（APK 跨域 / 局域网）

## 开发（移动 Web）

```bash
# 根目录
make mobile-dev

# 或
cd frontend/mobile && npm install && npm run dev
```

浏览器访问 http://localhost:25175（需后端 `make up`，Vite 代理 `/api` → 28080）。

## Docker 部署

全栈启动后移动端 Web：

```bash
make up   # 含 web-mobile 服务
```

访问 http://localhost:25175

## 打包 APK

**环境：** Node.js 22+、Android Studio（含 SDK）、JDK 17+

```powershell
# Windows
.\scripts\mobile-apk.ps1
```

```bash
# Linux / macOS
./scripts/mobile-apk.sh
```

或手动：

```bash
cd frontend/mobile
npm install
npm run cap:sync
cd android && ./gradlew assembleDebug
```

APK 输出：`frontend/mobile/android/app/build/outputs/apk/debug/app-debug.apk`

### APK 连接服务器

1. 安装 APK 后打开 **我的 → 服务器**
2. 填写 API 基址，例如：
   - 局域网 Docker：`http://192.168.x.x:25175/api/v1`
   - Android 模拟器访问本机：`http://10.0.2.2:25175/api/v1`
3. 保存后重新登录

> Cookie 刷新需网关 CORS 允许 Capacitor 来源；推荐通过 Nginx 同源部署（25175 反代 `/api`）。

## 与 desktop 的关系

- API 约定与 `frontend/desktop/src/api/http.js` 一致（移动端精简子集）
- 专业复式、AI 助手、链浏览器等复杂页面请使用桌面控制台（25173）
