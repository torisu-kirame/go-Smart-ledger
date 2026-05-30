#!/usr/bin/env node
/**
 * Repair openclaw.json: gateway.mode + migrate legacy agents.defaults.model schema.
 * Usage: node scripts/repair-openclaw-config.js [configPath] [templatePath]
 */
const fs = require('fs')
const path = require('path')

const root = path.join(__dirname, '..')
const configPath = path.resolve(process.argv[2] || path.join(root, 'data/openclaw/config/openclaw.json'))
const tplPath = path.resolve(process.argv[3] || path.join(root, 'integrations/openclaw/openclaw.docker.json'))

if (!fs.existsSync(configPath)) {
  console.log('>> OpenClaw: no config file, skipping repair')
  process.exit(0)
}

function readConfigFile(filePath) {
  let raw = fs.readFileSync(filePath)
  if (raw[0] === 0xef && raw[1] === 0xbb && raw[2] === 0xbf) raw = raw.subarray(3)
  if (raw.length >= 2 && raw[0] === 0xff && raw[1] === 0xfe) {
    return raw.subarray(2).toString('utf16le')
  }
  return raw.toString('utf8')
}

function salvageApiKey(text) {
  const m = text.match(/"apiKey"\s*:\s*"([^"]*)"/)
  return m?.[1] || ''
}

const tpl = JSON.parse(fs.readFileSync(tplPath, 'utf8'))
let cfg
let rawText = ''
try {
  rawText = readConfigFile(configPath)
  cfg = JSON.parse(rawText)
} catch {
  const salvagedKey = salvageApiKey(rawText || readConfigFile(configPath))
  fs.copyFileSync(tplPath, configPath)
  cfg = JSON.parse(fs.readFileSync(configPath, 'utf8'))
  if (salvagedKey) {
    const provider = cfg.models?.providers?.deepseek
    if (provider) provider.apiKey = salvagedKey
  }
  fs.writeFileSync(configPath, `${JSON.stringify(cfg, null, 2)}\n`)
  console.log('>> OpenClaw: reset invalid openclaw.json from template')
  process.exit(0)
}

let changed = false

if (!cfg.gateway?.mode) {
  cfg.gateway = { ...tpl.gateway, ...(cfg.gateway || {}), mode: 'local' }
  changed = true
  console.log('>> OpenClaw: added gateway.mode')
}

if (!cfg.gateway?.auth?.token) {
  cfg.gateway = {
    ...(cfg.gateway || {}),
    auth: {
      mode: 'token',
      token: '${OPENCLAW_GATEWAY_TOKEN}',
      ...(cfg.gateway?.auth || {}),
    },
  }
  if (!cfg.gateway.auth.mode) cfg.gateway.auth.mode = 'token'
  if (!cfg.gateway.auth.token) cfg.gateway.auth.token = '${OPENCLAW_GATEWAY_TOKEN}'
  changed = true
  console.log('>> OpenClaw: added gateway.auth.token')
}

const model = cfg.agents?.defaults?.model
const isLegacyModel =
  model &&
  typeof model === 'object' &&
  (typeof model.provider === 'string' || typeof model.id === 'string') &&
  typeof model.primary !== 'string'

if (isLegacyModel) {
  let providerKey = String(model.provider || 'deepseek')
  const baseUrlRaw = String(model.baseUrl || '')
  if (providerKey === 'openai' && baseUrlRaw.includes('deepseek')) {
    providerKey = 'deepseek'
  }
  if (providerKey === 'deepseek') {
    providerKey = 'deepseek'
  }
  const modelId = String(model.id || 'deepseek-chat')
  let baseUrl = baseUrlRaw.replace(/\/+$/, '')
  if (providerKey === 'deepseek' && baseUrl.endsWith('/v1')) {
    baseUrl = baseUrl.slice(0, -3)
  }
  if (!baseUrl) {
    baseUrl = providerKey === 'deepseek' ? 'https://api.deepseek.com' : 'https://api.openai.com/v1'
  }
  const apiKey = String(model.apiKey || '')
  const modelRef = `${providerKey}/${modelId}`

  cfg.models = {
    mode: 'merge',
    providers: {
      [providerKey]: {
        baseUrl,
        apiKey,
        api: 'openai-completions',
        models: [
          {
            id: modelId,
            name: modelId,
            reasoning: /reasoner|r1/i.test(modelId),
            input: ['text'],
            cost: { input: 0, output: 0, cacheRead: 0, cacheWrite: 0 },
            contextWindow: 128000,
            maxTokens: 8192,
          },
        ],
      },
    },
  }
  cfg.agents.defaults.model = { primary: modelRef }
  cfg.agents.defaults.models = { [modelRef]: {} }
  delete cfg.plugins
  changed = true
  console.log('>> OpenClaw: migrated agents.defaults.model to provider/model schema')
}

if (changed) {
  fs.writeFileSync(configPath, `${JSON.stringify(cfg, null, 2)}\n`)
}

const { syncOpenClawProviderAuth } = require('./openclaw-auth-sync')
if (syncOpenClawProviderAuth(cfg, path.dirname(configPath))) {
  changed = true
}

if (!cfg.agents?.defaults?.timeoutSeconds) {
  if (!cfg.agents) cfg.agents = {}
  if (!cfg.agents.defaults) cfg.agents.defaults = {}
  cfg.agents.defaults.timeoutSeconds = 180
  changed = true
  console.log('>> OpenClaw: set agents.defaults.timeoutSeconds=180')
}

if (changed) {
  fs.writeFileSync(configPath, `${JSON.stringify(cfg, null, 2)}\n`)
}
