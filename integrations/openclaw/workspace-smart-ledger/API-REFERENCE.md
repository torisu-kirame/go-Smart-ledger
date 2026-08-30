# Smart Ledger 用户 API 参考

> 本文件由 `node scripts/generate-swagger.mjs` 自动生成，对应 `OpenAPI-swagger-user.json`。
> 网关基址：`http://127.0.0.1:28080`。除登录/注册/验证码等外，请求需 Header：`Authorization: Bearer <accessToken>`。

## 产品概念（回答用户问题时请结合）

- **简单流水**（`bookkeepingMode: simple`）：基于用户级 **entry-templates** 自定义字段记账；支持 Excel 导入、多人协作审批、链上锚定。
- **专业复式**（`bookkeepingMode: professional`）：科目表、凭证、期间关账、报表、预算、银行对账、多币种、税务等；路径前缀 `/api/v1/ledgers/{id}/accounting/...`。
- **模板同步**：`POST /api/v1/ledgers/sync-entry-template` 可将某简单流水账本的模板字段同步到用户其他简单流水账本。
- **AI 上下文**：助手页可绑定账本；也可通过 `GET /api/v1/ledgers/{id}/rag-export` 拉取导出文本。
- **链上锚定**：`POST .../anchor` 封账锚定；`GET .../verify` 与 `GET /api/v1/chain/*` 用于查验，勿编造区块高度或哈希。

## 接口列表

### Auth

_登录、注册、验证码、Token 刷新_

- `GET /api/v1/auth/captcha` — 获取图形验证码
- `POST /api/v1/auth/login` — 登录，返回 accessToken
- `POST /api/v1/auth/logout` — 登出
- `POST /api/v1/auth/refresh` — 刷新 accessToken
- `POST /api/v1/auth/register` — 注册新用户

### Users

_用户资料、头像与端到端加密公钥_

- `GET /api/v1/users/{userId}/avatar` — 获取指定用户头像
- `GET /api/v1/users/me` — 获取当前用户资料
- `PATCH /api/v1/users/me` — 更新当前用户资料
- `POST /api/v1/users/me/avatar` — 上传当前用户头像
- `POST /api/v1/users/me/delete-account` — 注销账号
- `GET /api/v1/users/me/public-key` — 获取当前用户公钥
- `PUT /api/v1/users/me/public-key` — 设置端到端加密公钥
- `POST /api/v1/users/me/verify-password` — 验证当前用户密码
- `GET /api/v1/users/search` — 按用户名搜索用户

### Friends

_好友列表与好友请求_

- `GET /api/v1/friends` — 好友列表
- `POST /api/v1/friends` — 发送好友请求
- `DELETE /api/v1/friends/{friendId}` — 删除好友
- `POST /api/v1/friends/requests/{fromUserId}/accept` — 接受好友请求
- `POST /api/v1/friends/requests/{fromUserId}/reject` — 拒绝好友请求
- `DELETE /api/v1/friends/requests/{toUserId}` — 取消发出的好友请求
- `GET /api/v1/friends/requests/incoming` — 收到的待处理好友请求
- `GET /api/v1/friends/requests/outgoing` — 发出的待处理好友请求

### Teams

_团队、团队聊天与团队账本_

- `GET /api/v1/teams` — 团队列表
- `POST /api/v1/teams` — 创建团队
- `GET /api/v1/teams/{teamId}` — 团队详情
- `GET /api/v1/teams/{teamId}/chat/files/{messageId}` — 下载团队聊天文件
- `POST /api/v1/teams/{teamId}/ledgers` — 将账本关联到团队
- `DELETE /api/v1/teams/{teamId}/ledgers/{ledgerId}` — 从团队移除账本
- `GET /api/v1/teams/{teamId}/messages` — 团队聊天消息
- `POST /api/v1/teams/{teamId}/messages` — 发送团队文本消息
- `POST /api/v1/teams/{teamId}/messages/file` — 发送团队文件消息
- `POST /api/v1/teams/{teamId}/read` — 将指定团队消息标为已读
- `POST /api/v1/teams/read-all` — 将全部团队消息标为已读

### EntryTemplates

_简单流水记账模板（用户级，可跨账本同步）_

- `GET /api/v1/entry-templates` — 列出用户级流水模板
- `POST /api/v1/entry-templates` — 创建流水模板
- `DELETE /api/v1/entry-templates/{templateId}` — 删除模板
- `GET /api/v1/entry-templates/{templateId}` — 获取模板详情
- `PUT /api/v1/entry-templates/{templateId}` — 更新模板

### Ledgers

_账本 CRUD、分录、成员、邀请、备份与 RAG 导出_

- `GET /api/v1/ledgers` — 列出可访问账本
- `POST /api/v1/ledgers` — 创建账本（simple 或 professional）
- `DELETE /api/v1/ledgers/{id}` — 删除账本
- `GET /api/v1/ledgers/{id}` — 账本详情
- `PATCH /api/v1/ledgers/{id}` — 更新账本名称等设置
- `POST /api/v1/ledgers/{id}/anchor` — 触发封账锚定
- `PATCH /api/v1/ledgers/{id}/approval-policy` — 设置多人审批策略
- `GET /api/v1/ledgers/{id}/audit-export` — 导出审计包
- `POST /api/v1/ledgers/{id}/backup` — 导出账本加密备份
- `POST /api/v1/ledgers/{id}/encryption/enable` — 启用账本端到端加密
- `PATCH /api/v1/ledgers/{id}/encryption/passphrase-view-policy` — 设置口令查看策略
- `PUT /api/v1/ledgers/{id}/encryption/passphrase-view-wrap` — 更新口令查看包装
- `POST /api/v1/ledgers/{id}/encryption/rotate` — 轮换账本加密密钥
- `POST /api/v1/ledgers/{id}/entries` — 追加分录；body 必须为 `{"entry":{"tableId":"...","data":{...},"signerId":"..."}}`（优先用工具 `append_ledger_entry`）
- `POST /api/v1/ledgers/{id}/entries/propose` — 提议分录（多人协作）
- `POST /api/v1/ledgers` — 创建账本（优先 `create_ledger`）
- `POST /api/v1/ledgers/{id}/tables` — 创建 Sheet；body `{"name":"...","entrySchema":{"fields":[{"key","label","type","required"}]}}`（优先 `create_ledger_sheet`）
- `GET /api/v1/ledgers/{id}/events` — 账本事件流（含链上序列）
- `GET /api/v1/ledgers/{id}/invites` — 账本发出的邀请
- `POST /api/v1/ledgers/{id}/invites/accept` — 接受账本邀请
- `POST /api/v1/ledgers/{id}/members/invite` — 邀请成员加入账本
- `PATCH /api/v1/ledgers/{id}/multi-table` — 启用/配置多表流水
- `GET /api/v1/ledgers/{id}/pending` — 待审批分录列表
- `POST /api/v1/ledgers/{id}/pending/{pendingId}/approve` — 批准待审批分录
- `POST /api/v1/ledgers/{id}/pending/{pendingId}/reject` — 拒绝待审批分录
- `GET /api/v1/ledgers/{id}/rag-export` — 导出账本文本供 AI 分析（RAG）
- `POST /api/v1/ledgers/{id}/restore/commit` — 提交恢复备份
- `POST /api/v1/ledgers/{id}/restore/preview` — 恢复备份预览
- `PATCH /api/v1/ledgers/{id}/storage-location` — 设置账本备份存储位置
- `GET /api/v1/ledgers/{id}/sync` — 同步账本增量事件
- `DELETE /api/v1/ledgers/{id}/tables/{tableId}` — 删除子表
- `PATCH /api/v1/ledgers/{id}/tables/{tableId}` — 更新子表
- `GET /api/v1/ledgers/{id}/verify` — 校验账本 Merkle 根与链上锚定
- `GET /api/v1/ledgers/invites/mine` — 我收到的账本邀请
- `POST /api/v1/ledgers/sync-entry-template` — 将模板同步到其他简单流水账本

### Import

_CSV/Excel 导入_

- `POST /api/v1/ledgers/{id}/import/sheet-csv` — **CSV→Sheet 一站式导入**：JSON `{csv, tableId?, sheetName?}` 或 multipart `file`；无 `tableId` 自动新建 Sheet，有则追加到底部
- `POST /api/v1/ledgers/{id}/import/adaptive/preview` — 自适应导入预览
- `POST /api/v1/ledgers/{id}/import/adaptive/commit` — 自适应导入提交
- `POST /api/v1/ledgers/{id}/import/preview` — 按已有 schema 预览（xlsx/csv）
- `POST /api/v1/ledgers/{id}/import/commit` — 按已有 schema 提交批量导入
- `GET /api/v1/import/template` — 下载 Excel 导入模板

### Accounting

_专业复式账：科目、凭证、报表、预算、税务等_

- `GET /api/v1/ledgers/{id}/accounting/aging` — 账龄分析
- `GET /api/v1/ledgers/{id}/accounting/attachments` — 凭证附件列表
- `POST /api/v1/ledgers/{id}/accounting/attachments` — 上传凭证附件
- `PATCH /api/v1/ledgers/{id}/accounting/attachments/{attachId}` — 更新附件元数据
- `GET /api/v1/ledgers/{id}/accounting/bank-statements` — 银行对账单列表
- `POST /api/v1/ledgers/{id}/accounting/bank-statements/{stmtId}/match` — 对账单与凭证匹配
- `POST /api/v1/ledgers/{id}/accounting/bank-statements/import` — 导入银行对账单
- `GET /api/v1/ledgers/{id}/accounting/budget` — 预算配置
- `PUT /api/v1/ledgers/{id}/accounting/budget` — 更新预算
- `GET /api/v1/ledgers/{id}/accounting/budget/analysis` — 预算执行分析
- `GET /api/v1/ledgers/{id}/accounting/chart` — 会计科目表
- `PUT /api/v1/ledgers/{id}/accounting/chart` — 更新科目表
- `GET /api/v1/ledgers/{id}/accounting/currency` — 多币种设置
- `PUT /api/v1/ledgers/{id}/accounting/currency` — 更新多币种设置
- `GET /api/v1/ledgers/{id}/accounting/currency/balances` — 外币余额
- `GET /api/v1/ledgers/{id}/accounting/currency/fx-rates` — 汇率列表
- `PUT /api/v1/ledgers/{id}/accounting/currency/fx-rates` — 更新汇率
- `GET /api/v1/ledgers/{id}/accounting/currency/revaluation` — 汇兑重估
- `GET /api/v1/ledgers/{id}/accounting/journals` — 凭证列表
- `POST /api/v1/ledgers/{id}/accounting/journals` — 新建凭证
- `GET /api/v1/ledgers/{id}/accounting/periods` — 会计期间列表
- `POST /api/v1/ledgers/{id}/accounting/periods/{period}/close` — 关账
- `POST /api/v1/ledgers/{id}/accounting/periods/{period}/reopen` — 反关账
- `GET /api/v1/ledgers/{id}/accounting/reports` — 财务报表
- `GET /api/v1/ledgers/{id}/accounting/tax` — 税务配置
- `PUT /api/v1/ledgers/{id}/accounting/tax` — 更新税务配置
- `POST /api/v1/ledgers/{id}/accounting/tax/apply-preset` — 应用税务预设
- `GET /api/v1/ledgers/{id}/accounting/tax/presets` — 税务预设
- `GET /api/v1/ledgers/{id}/accounting/tax/report` — 税务报表

### AI

_AI 对话（OpenClaw 代理）_

- `POST /api/v1/ai/chat` — 流式 AI 对话（OpenClaw Gateway）

### Storage

_加密云备份存储_

- `POST /api/v1/storage/backup` — 上传加密备份到存储服务
- `POST /api/v1/storage/backup/fetch` — 从存储服务拉取备份

### Chain

_MiniLedger 链状态、区块与近期交易（查验锚定）_

- `GET /api/v1/chain/blocks` — 区块列表
- `GET /api/v1/chain/blocks/{height}` — 按高度查询区块
- `GET /api/v1/chain/blocks/latest` — 最新区块
- `GET /api/v1/chain/status` — 链与锚定队列状态摘要
- `GET /api/v1/chain/tx/recent` — 近期链上交易

## 未收录（内部 / 运维）

以下接口不对终端用户开放，回答时不要主动推荐：
- 健康检查：`GET /api/v1/auth/health`、`GET /api/v1/storage/health`
- AI 连接测试与 Agent 磁盘读写：`POST /api/v1/ai/test`、`/api/v1/ai/agent/load`、`/api/v1/ai/agent/save`
- 链队列运维：`GET /api/v1/chain/queue`、`POST /api/v1/chain/queue/{id}/retry`
- 节点共识细节：`GET /api/v1/chain/peers`、`GET /api/v1/chain/consensus`
