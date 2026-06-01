#!/usr/bin/env node
/**
 * Generate OpenAPI 3.0 JSON and agent workspace API reference from gateway routes.
 * Outputs:
 *   - OpenAPI-swagger.json (full)
 *   - OpenAPI-swagger-user.json (user-facing only)
 *   - frontend/desktop/src/assets/agent-workspace/API-REFERENCE.md
 *   - integrations/openclaw/workspace-smart-ledger/API-REFERENCE.md
 */
import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '..')
const gatewayPath = path.join(root, 'backend/services/gateway/gateway.go')
const outFull = path.join(root, 'OpenAPI-swagger.json')
const outUser = path.join(root, 'OpenAPI-swagger-user.json')
const apiRefPaths = [
  path.join(root, 'frontend/desktop/src/assets/agent-workspace/API-REFERENCE.md'),
  path.join(root, 'integrations/openclaw/workspace-smart-ledger/API-REFERENCE.md'),
]

/** 非面向终端用户的运维 / 内部接口 */
const INTERNAL_PATH_PATTERNS = [
  /^\/api\/v1\/auth\/health$/,
  /^\/api\/v1\/storage\/health$/,
  /^\/api\/v1\/ai\/test$/,
  /^\/api\/v1\/ai\/agent\//,
  /^\/api\/v1\/chain\/queue/,
  /^\/api\/v1\/chain\/peers$/,
  /^\/api\/v1\/chain\/consensus$/,
]

function isUserFacing(routePath) {
  return !INTERNAL_PATH_PATTERNS.some((re) => re.test(routePath))
}

const src = fs.readFileSync(gatewayPath, 'utf8')
const routeRe = /\{Method:\s*http\.Method(\w+),\s*Path:\s*"([^"]+)"/g

/** @type {{ method: string, path: string }[]} */
const routes = []
let m
while ((m = routeRe.exec(src)) !== null) {
  routes.push({ method: m[1].toLowerCase(), path: m[2] })
}

function tagFor(p) {
  if (p.startsWith('/api/v1/auth')) return 'Auth'
  if (p.startsWith('/api/v1/users')) return 'Users'
  if (p.startsWith('/api/v1/friends')) return 'Friends'
  if (p.startsWith('/api/v1/teams')) return 'Teams'
  if (p.startsWith('/api/v1/entry-templates')) return 'EntryTemplates'
  if (p.startsWith('/api/v1/ai')) return 'AI'
  if (p.startsWith('/api/v1/storage')) return 'Storage'
  if (p.startsWith('/api/v1/chain')) return 'Chain'
  if (p.includes('/accounting')) return 'Accounting'
  if (p.includes('/import') || p === '/api/v1/import/template') return 'Import'
  if (p.startsWith('/api/v1/ledgers')) return 'Ledgers'
  return 'Gateway'
}

const TAG_DESCRIPTIONS = {
  Auth: '登录、注册、验证码、Token 刷新',
  Users: '用户资料、头像与端到端加密公钥',
  Friends: '好友列表与好友请求',
  Teams: '团队、团队聊天与团队账本',
  EntryTemplates: '简单流水记账模板（用户级，可跨账本同步）',
  Ledgers: '账本 CRUD、分录、成员、邀请、备份与 RAG 导出',
  Import: 'Excel 导入与模板下载',
  Accounting: '专业复式账：科目、凭证、报表、预算、税务等',
  AI: 'AI 对话（OpenClaw 代理）',
  Storage: '加密云备份存储',
  Chain: 'MiniLedger 链状态、区块与近期交易（查验锚定）',
}

function describeRoute(method, p) {
  const key = `${method.toUpperCase()} ${p}`
  /** @type {Record<string, string>} */
  const exact = {
    'GET /api/v1/auth/captcha': '获取图形验证码',
    'POST /api/v1/auth/login': '登录，返回 accessToken',
    'POST /api/v1/auth/register': '注册新用户',
    'POST /api/v1/auth/refresh': '刷新 accessToken',
    'POST /api/v1/auth/logout': '登出',
    'GET /api/v1/users/:userId/avatar': '获取指定用户头像',
    'GET /api/v1/users/me': '获取当前用户资料',
    'PATCH /api/v1/users/me': '更新当前用户资料',
    'POST /api/v1/users/me/avatar': '上传当前用户头像',
    'POST /api/v1/users/me/delete-account': '注销账号',
    'PUT /api/v1/users/me/public-key': '设置端到端加密公钥',
    'GET /api/v1/users/me/public-key': '获取当前用户公钥',
    'GET /api/v1/users/search': '按用户名搜索用户',
    'POST /api/v1/users/me/verify-password': '验证当前用户密码',
    'GET /api/v1/friends/requests/incoming': '收到的待处理好友请求',
    'GET /api/v1/friends/requests/outgoing': '发出的待处理好友请求',
    'POST /api/v1/friends/requests/:fromUserId/accept': '接受好友请求',
    'POST /api/v1/friends/requests/:fromUserId/reject': '拒绝好友请求',
    'DELETE /api/v1/friends/requests/:toUserId': '取消发出的好友请求',
    'GET /api/v1/friends': '好友列表',
    'POST /api/v1/friends': '发送好友请求',
    'DELETE /api/v1/friends/:friendId': '删除好友',
    'GET /api/v1/teams': '团队列表',
    'POST /api/v1/teams': '创建团队',
    'POST /api/v1/teams/read-all': '将全部团队消息标为已读',
    'POST /api/v1/teams/:teamId/read': '将指定团队消息标为已读',
    'GET /api/v1/teams/:teamId/messages': '团队聊天消息',
    'POST /api/v1/teams/:teamId/messages': '发送团队文本消息',
    'POST /api/v1/teams/:teamId/messages/file': '发送团队文件消息',
    'GET /api/v1/teams/:teamId/chat/files/:messageId': '下载团队聊天文件',
    'POST /api/v1/teams/:teamId/ledgers': '将账本关联到团队',
    'DELETE /api/v1/teams/:teamId/ledgers/:ledgerId': '从团队移除账本',
    'GET /api/v1/teams/:teamId': '团队详情',
    'GET /api/v1/entry-templates': '列出用户级流水模板',
    'POST /api/v1/entry-templates': '创建流水模板',
    'GET /api/v1/entry-templates/:templateId': '获取模板详情',
    'PUT /api/v1/entry-templates/:templateId': '更新模板',
    'DELETE /api/v1/entry-templates/:templateId': '删除模板',
    'GET /api/v1/ledgers': '列出可访问账本',
    'POST /api/v1/ledgers': '创建账本（simple 或 professional）',
    'POST /api/v1/ledgers/sync-entry-template': '将模板同步到其他简单流水账本',
    'GET /api/v1/ledgers/:id': '账本详情',
    'PATCH /api/v1/ledgers/:id': '更新账本名称等设置',
    'DELETE /api/v1/ledgers/:id': '删除账本',
    'POST /api/v1/ledgers/:id/entries': '追加简单流水分录',
    'GET /api/v1/ledgers/:id/events': '账本事件流（含链上序列）',
    'POST /api/v1/ledgers/:id/anchor': '触发封账锚定',
    'GET /api/v1/ledgers/:id/verify': '校验账本 Merkle 根与链上锚定',
    'GET /api/v1/entry-schema/templates': '内置分录 schema 模板列表',
    'GET /api/v1/import/template': '下载 Excel 导入模板',
    'POST /api/v1/ledgers/:id/import/preview': '标准 Excel 导入预览',
    'POST /api/v1/ledgers/:id/import/commit': '提交标准 Excel 导入',
    'POST /api/v1/ledgers/:id/import/adaptive/preview': '自适应 Excel 导入预览',
    'POST /api/v1/ledgers/:id/import/adaptive/commit': '提交自适应 Excel 导入',
    'POST /api/v1/ledgers/:id/backup': '导出账本加密备份',
    'POST /api/v1/ledgers/:id/restore/preview': '恢复备份预览',
    'POST /api/v1/ledgers/:id/restore/commit': '提交恢复备份',
    'GET /api/v1/ledgers/invites/mine': '我收到的账本邀请',
    'GET /api/v1/ledgers/:id/pending': '待审批分录列表',
    'POST /api/v1/ledgers/:id/entries/propose': '提议分录（多人协作）',
    'POST /api/v1/ledgers/:id/pending/:pendingId/approve': '批准待审批分录',
    'POST /api/v1/ledgers/:id/pending/:pendingId/reject': '拒绝待审批分录',
    'POST /api/v1/ledgers/:id/members/invite': '邀请成员加入账本',
    'GET /api/v1/ledgers/:id/invites': '账本发出的邀请',
    'POST /api/v1/ledgers/:id/invites/accept': '接受账本邀请',
    'GET /api/v1/ledgers/:id/sync': '同步账本增量事件',
    'GET /api/v1/ledgers/:id/rag-export': '导出账本文本供 AI 分析（RAG）',
    'POST /api/v1/ledgers/:id/encryption/rotate': '轮换账本加密密钥',
    'PATCH /api/v1/ledgers/:id/storage-location': '设置账本备份存储位置',
    'PATCH /api/v1/ledgers/:id/approval-policy': '设置多人审批策略',
    'PATCH /api/v1/ledgers/:id/multi-table': '启用/配置多表流水',
    'POST /api/v1/ledgers/:id/tables': '创建子表',
    'PATCH /api/v1/ledgers/:id/tables/:tableId': '更新子表',
    'DELETE /api/v1/ledgers/:id/tables/:tableId': '删除子表',
    'POST /api/v1/ledgers/:id/encryption/enable': '启用账本端到端加密',
    'PATCH /api/v1/ledgers/:id/encryption/passphrase-view-policy': '设置口令查看策略',
    'PUT /api/v1/ledgers/:id/encryption/passphrase-view-wrap': '更新口令查看包装',
    'GET /api/v1/ledgers/:id/audit-export': '导出审计包',
    'POST /api/v1/ai/chat': '流式 AI 对话（OpenClaw Gateway）',
    'GET /api/v1/ledgers/:id/accounting/chart': '会计科目表',
    'PUT /api/v1/ledgers/:id/accounting/chart': '更新科目表',
    'GET /api/v1/ledgers/:id/accounting/journals': '凭证列表',
    'POST /api/v1/ledgers/:id/accounting/journals': '新建凭证',
    'GET /api/v1/ledgers/:id/accounting/periods': '会计期间列表',
    'POST /api/v1/ledgers/:id/accounting/periods/:period/close': '关账',
    'POST /api/v1/ledgers/:id/accounting/periods/:period/reopen': '反关账',
    'GET /api/v1/ledgers/:id/accounting/reports': '财务报表',
    'GET /api/v1/ledgers/:id/accounting/attachments': '凭证附件列表',
    'POST /api/v1/ledgers/:id/accounting/attachments': '上传凭证附件',
    'PATCH /api/v1/ledgers/:id/accounting/attachments/:attachId': '更新附件元数据',
    'GET /api/v1/ledgers/:id/accounting/bank-statements': '银行对账单列表',
    'POST /api/v1/ledgers/:id/accounting/bank-statements/import': '导入银行对账单',
    'POST /api/v1/ledgers/:id/accounting/bank-statements/:stmtId/match': '对账单与凭证匹配',
    'GET /api/v1/ledgers/:id/accounting/budget': '预算配置',
    'PUT /api/v1/ledgers/:id/accounting/budget': '更新预算',
    'GET /api/v1/ledgers/:id/accounting/budget/analysis': '预算执行分析',
    'GET /api/v1/ledgers/:id/accounting/aging': '账龄分析',
    'GET /api/v1/ledgers/:id/accounting/currency': '多币种设置',
    'PUT /api/v1/ledgers/:id/accounting/currency': '更新多币种设置',
    'GET /api/v1/ledgers/:id/accounting/currency/fx-rates': '汇率列表',
    'PUT /api/v1/ledgers/:id/accounting/currency/fx-rates': '更新汇率',
    'GET /api/v1/ledgers/:id/accounting/currency/balances': '外币余额',
    'GET /api/v1/ledgers/:id/accounting/currency/revaluation': '汇兑重估',
    'GET /api/v1/ledgers/:id/accounting/tax/presets': '税务预设',
    'GET /api/v1/ledgers/:id/accounting/tax': '税务配置',
    'PUT /api/v1/ledgers/:id/accounting/tax': '更新税务配置',
    'POST /api/v1/ledgers/:id/accounting/tax/apply-preset': '应用税务预设',
    'GET /api/v1/ledgers/:id/accounting/tax/report': '税务报表',
    'POST /api/v1/storage/backup': '上传加密备份到存储服务',
    'POST /api/v1/storage/backup/fetch': '从存储服务拉取备份',
    'GET /api/v1/chain/status': '链与锚定队列状态摘要',
    'GET /api/v1/chain/blocks': '区块列表',
    'GET /api/v1/chain/blocks/latest': '最新区块',
    'GET /api/v1/chain/blocks/:height': '按高度查询区块',
    'GET /api/v1/chain/tx/recent': '近期链上交易',
  }
  if (exact[key]) return exact[key]
  const tail = p.split('/').filter(Boolean).slice(-2).join(' ')
  return `${method.toUpperCase()} ${tail || p}`
}

function toOpenApiPath(p) {
  return p.replace(/:([A-Za-z0-9_]+)/g, '{$1}')
}

function buildPaths(routeList) {
  const paths = {}
  for (const { method, path: p } of routeList) {
    const oaPath = toOpenApiPath(p)
    if (!paths[oaPath]) paths[oaPath] = {}
    const op = {
      tags: [tagFor(p)],
      summary: describeRoute(method, p),
      operationId: `${method}_${oaPath.replace(/[^\w]+/g, '_')}`,
      responses: {
        '200': { description: 'Success', content: { 'application/json': { schema: { type: 'object' } } } },
        '400': { description: 'Bad request', content: { 'application/json': { schema: { $ref: '#/components/schemas/Error' } } } },
        '401': { description: 'Unauthorized', content: { 'application/json': { schema: { $ref: '#/components/schemas/Error' } } } },
        '404': { description: 'Not found', content: { 'application/json': { schema: { $ref: '#/components/schemas/Error' } } } },
      },
    }
    const publicAuth =
      p.startsWith('/api/v1/auth/captcha') ||
      p.startsWith('/api/v1/auth/login') ||
      p.startsWith('/api/v1/auth/register') ||
      p.startsWith('/api/v1/auth/refresh') ||
      p.startsWith('/api/v1/auth/logout') ||
      p.startsWith('/api/v1/auth/health') ||
      (p.startsWith('/api/v1/users/') && p.endsWith('/avatar') && method === 'get')
    if (!publicAuth) {
      op.security = [{ bearerAuth: [] }]
    }
    if (['post', 'put', 'patch'].includes(method)) {
      op.requestBody = {
        content: {
          'application/json': { schema: { type: 'object' } },
          'multipart/form-data': { schema: { type: 'object' } },
        },
      }
    }
    if (method === 'get' && p.includes(':')) {
      op.parameters = [...p.matchAll(/:([A-Za-z0-9_]+)/g)].map((x) => ({
        name: x[1],
        in: 'path',
        required: true,
        schema: { type: 'string' },
      }))
    }
    paths[oaPath][method] = op
  }
  return paths
}

function buildOpenApiDoc(routeList, { userFacing = false } = {}) {
  const tagsUsed = new Set(routeList.map((r) => tagFor(r.path)))
  const tags = Object.entries(TAG_DESCRIPTIONS)
    .filter(([name]) => tagsUsed.has(name))
    .map(([name, description]) => ({ name, description }))

  return {
    openapi: '3.0.3',
    info: {
      title: userFacing ? 'Smart Ledger API（用户向）' : 'Smart Ledger API',
      description: userFacing
        ? '面向终端用户与 AI 助手的 Smart Ledger 网关 API 子集（已排除健康检查、Agent 磁盘存储、链队列运维等内部接口）。默认基址：http://127.0.0.1:28080。除认证相关接口外，需在 Header 携带 `Authorization: Bearer <accessToken>`。'
        : 'Smart Ledger 网关统一 API（OpenAPI 3.0）。默认基址：http://127.0.0.1:28080。除认证相关接口外，需在 Header 携带 `Authorization: Bearer <accessToken>`。',
      version: '1.0.0',
      contact: { name: 'Smart Ledger' },
    },
    servers: [
      { url: 'http://127.0.0.1:28080', description: 'Local gateway (docker compose)' },
      { url: 'http://localhost:28080', description: 'Local gateway' },
    ],
    tags,
    paths: buildPaths(routeList),
    components: {
      securitySchemes: {
        bearerAuth: {
          type: 'http',
          scheme: 'bearer',
          bearerFormat: 'JWT',
          description: '登录后获得的短期 access token',
        },
      },
      schemas: {
        Error: {
          type: 'object',
          properties: {
            code: { type: 'integer' },
            msg: { type: 'string' },
          },
        },
        LoginRequest: {
          type: 'object',
          required: ['username', 'password', 'captchaId', 'captchaCode'],
          properties: {
            username: { type: 'string' },
            password: { type: 'string' },
            captchaId: { type: 'string' },
            captchaCode: { type: 'string' },
          },
        },
        LoginResponse: {
          type: 'object',
          properties: {
            accessToken: { type: 'string' },
            expiresIn: { type: 'integer' },
            tokenType: { type: 'string', example: 'Bearer' },
            user: { $ref: '#/components/schemas/UserInfo' },
          },
        },
        UserInfo: {
          type: 'object',
          properties: {
            id: { type: 'string' },
            username: { type: 'string' },
            nickname: { type: 'string' },
          },
        },
        EntryTemplate: {
          type: 'object',
          properties: {
            templateId: { type: 'string' },
            name: { type: 'string' },
            builtin: { type: 'boolean' },
            fields: {
              type: 'array',
              items: {
                type: 'object',
                properties: {
                  key: { type: 'string' },
                  label: { type: 'string' },
                  type: { type: 'string', enum: ['text', 'number', 'date', 'user'] },
                  required: { type: 'boolean' },
                },
              },
            },
          },
        },
        Ledger: {
          type: 'object',
          properties: {
            id: { type: 'string' },
            type: { type: 'string', enum: ['private', 'multi'] },
            name: { type: 'string' },
            bookkeepingMode: { type: 'string', enum: ['simple', 'professional'] },
            latestSeq: { type: 'integer' },
            latestRoot: { type: 'string' },
            anchorStatus: { type: 'string' },
          },
        },
      },
    },
  }
}

function buildApiReferenceMarkdown(userRoutes) {
  const byTag = /** @type {Record<string, { method: string, path: string, desc: string }[]>} */ ({})
  for (const { method, path: p } of userRoutes) {
    const tag = tagFor(p)
    if (!byTag[tag]) byTag[tag] = []
    byTag[tag].push({
      method: method.toUpperCase(),
      path: toOpenApiPath(p),
      desc: describeRoute(method, p),
    })
  }

  const tagOrder = Object.keys(TAG_DESCRIPTIONS).filter((t) => byTag[t])
  const lines = [
    '# Smart Ledger 用户 API 参考',
    '',
    '> 本文件由 `node scripts/generate-swagger.mjs` 自动生成，对应 `OpenAPI-swagger-user.json`。',
    '> 网关基址：`http://127.0.0.1:28080`。除登录/注册/验证码等外，请求需 Header：`Authorization: Bearer <accessToken>`。',
    '',
    '## 产品概念（回答用户问题时请结合）',
    '',
    '- **简单流水**（`bookkeepingMode: simple`）：基于用户级 **entry-templates** 自定义字段记账；支持 Excel 导入、多人协作审批、链上锚定。',
    '- **专业复式**（`bookkeepingMode: professional`）：科目表、凭证、期间关账、报表、预算、银行对账、多币种、税务等；路径前缀 `/api/v1/ledgers/{id}/accounting/...`。',
    '- **模板同步**：`POST /api/v1/ledgers/sync-entry-template` 可将某简单流水账本的模板字段同步到用户其他简单流水账本。',
    '- **AI 上下文**：助手页可绑定账本；也可通过 `GET /api/v1/ledgers/{id}/rag-export` 拉取导出文本。',
    '- **链上锚定**：`POST .../anchor` 封账锚定；`GET .../verify` 与 `GET /api/v1/chain/*` 用于查验，勿编造区块高度或哈希。',
    '',
    '## 接口列表',
    '',
  ]

  for (const tag of tagOrder) {
    lines.push(`### ${tag}`, '')
    lines.push(`_${TAG_DESCRIPTIONS[tag]}_`, '')
    const items = byTag[tag].sort((a, b) => a.path.localeCompare(b.path) || a.method.localeCompare(b.method))
    for (const { method, path: apiPath, desc } of items) {
      lines.push(`- \`${method} ${apiPath}\` — ${desc}`)
    }
    lines.push('')
  }

  lines.push(
    '## 未收录（内部 / 运维）',
    '',
    '以下接口不对终端用户开放，回答时不要主动推荐：',
    '- 健康检查：`GET /api/v1/auth/health`、`GET /api/v1/storage/health`',
    '- AI 连接测试与 Agent 磁盘读写：`POST /api/v1/ai/test`、`/api/v1/ai/agent/load`、`/api/v1/ai/agent/save`',
    '- 链队列运维：`GET /api/v1/chain/queue`、`POST /api/v1/chain/queue/{id}/retry`',
    '- 节点共识细节：`GET /api/v1/chain/peers`、`GET /api/v1/chain/consensus`',
    ''
  )

  return `${lines.join('\n')}`
}

const userRoutes = routes.filter((r) => isUserFacing(r.path))
const fullDoc = buildOpenApiDoc(routes)
const userDoc = buildOpenApiDoc(userRoutes, { userFacing: true })
const apiRef = buildApiReferenceMarkdown(userRoutes)

fs.writeFileSync(outFull, `${JSON.stringify(fullDoc, null, 2)}\n`, 'utf8')
fs.writeFileSync(outUser, `${JSON.stringify(userDoc, null, 2)}\n`, 'utf8')
for (const p of apiRefPaths) {
  fs.mkdirSync(path.dirname(p), { recursive: true })
  fs.writeFileSync(p, apiRef, 'utf8')
}

console.log(`Wrote ${routes.length} routes to ${path.basename(outFull)}`)
console.log(`Wrote ${userRoutes.length} user-facing routes to ${path.basename(outUser)}`)
console.log(`Wrote API-REFERENCE.md (${userRoutes.length} endpoints)`)
