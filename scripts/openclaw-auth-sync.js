/**
 * Sync OpenClaw provider API keys into env/auth metadata and auth-profiles.json.
 * @param {Record<string, unknown>} cfg
 * @param {string} configDir absolute path to OpenClaw state dir (parent of openclaw.json)
 */
const fs = require('fs')
const path = require('path')

const PROVIDER_ENV = {
  deepseek: 'DEEPSEEK_API_KEY',
  openai: 'OPENAI_API_KEY',
  qwen: 'DASHSCOPE_API_KEY',
  moonshot: 'MOONSHOT_API_KEY',
  groq: 'GROQ_API_KEY',
  ollama: 'OLLAMA_API_KEY',
  lmstudio: 'LMSTUDIO_API_KEY',
}

function extractProviderKeys(cfg) {
  const providers = cfg?.models?.providers
  if (!providers || typeof providers !== 'object') return {}
  const out = {}
  for (const [name, block] of Object.entries(providers)) {
    if (!block || typeof block !== 'object') continue
    const key = String(block.apiKey || '').trim()
    if (key) out[name] = key
  }
  return out
}

function applyAuthToConfig(cfg, providerKeys) {
  if (!Object.keys(providerKeys).length) return false
  let changed = false
  if (!cfg.env || typeof cfg.env !== 'object') {
    cfg.env = {}
    changed = true
  }
  for (const [provider, key] of Object.entries(providerKeys)) {
    const envName = PROVIDER_ENV[provider]
    if (envName && cfg.env[envName] !== key) {
      cfg.env[envName] = key
      changed = true
    }
  }
  if (!cfg.auth || typeof cfg.auth !== 'object') {
    cfg.auth = {}
    changed = true
  }
  if (!cfg.auth.profiles || typeof cfg.auth.profiles !== 'object') {
    cfg.auth.profiles = {}
    changed = true
  }
  if (!cfg.auth.order || typeof cfg.auth.order !== 'object') {
    cfg.auth.order = {}
    changed = true
  }
  for (const provider of Object.keys(providerKeys)) {
    const profileId = `${provider}:default`
    const existing = cfg.auth.profiles[profileId]
    if (!existing || existing.provider !== provider || existing.mode !== 'api_key') {
      cfg.auth.profiles[profileId] = { provider, mode: 'api_key' }
      changed = true
    }
    const order = cfg.auth.order[provider]
    if (!Array.isArray(order) || order[0] !== profileId) {
      cfg.auth.order[provider] = [profileId]
      changed = true
    }
  }
  return changed
}

function writeAuthProfiles(configDir, providerKeys) {
  if (!Object.keys(providerKeys).length) return false
  const agentDir = path.join(configDir, 'agents', 'main', 'agent')
  fs.mkdirSync(agentDir, { recursive: true })
  const profiles = { version: 1, profiles: {} }
  for (const [provider, key] of Object.entries(providerKeys)) {
    profiles.profiles[`${provider}:default`] = {
      type: 'api_key',
      provider,
      key,
    }
  }
  const authPath = path.join(agentDir, 'auth-profiles.json')
  const next = `${JSON.stringify(profiles, null, 2)}\n`
  const prev = fs.existsSync(authPath) ? fs.readFileSync(authPath, 'utf8') : ''
  if (prev !== next) {
    fs.writeFileSync(authPath, next, { mode: 0o600 })
    return true
  }
  return false
}

function syncOpenClawProviderAuth(cfg, configDir) {
  const providerKeys = extractProviderKeys(cfg)
  const configChanged = applyAuthToConfig(cfg, providerKeys)
  const authFileChanged = writeAuthProfiles(configDir, providerKeys)
  return configChanged || authFileChanged
}

module.exports = {
  PROVIDER_ENV,
  extractProviderKeys,
  applyAuthToConfig,
  writeAuthProfiles,
  syncOpenClawProviderAuth,
}
