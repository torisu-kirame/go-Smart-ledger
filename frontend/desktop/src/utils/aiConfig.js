const STORAGE_KEY = 'smart-ledger-ai-config'
const PROFILES_KEY = 'smart-ledger-ai-profiles'
export const AI_CONFIG_CHANGED_EVENT = 'smart-ledger-ai-config-changed'

function notifyAiConfigChanged() {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new Event(AI_CONFIG_CHANGED_EVENT))
  }
}

/** 本地离线提供方（需自行部署 Ollama / LM Studio 等） */
export const OFFLINE_PROVIDERS = ['ollama', 'lmstudio']

/** @typedef {{ label: string, baseUrl: string, apiKeyDefault: string, models: string[], embedModel: string }} ProviderPreset */

/** @type {Record<string, ProviderPreset>} */
export const AI_PROVIDERS = {
  deepseek: {
    label: 'DeepSeek',
    baseUrl: 'https://api.deepseek.com/v1',
    apiKeyDefault: '',
    models: ['deepseek-chat', 'deepseek-reasoner'],
    embedModel: 'text-embedding-3-small',
  },
  openai: {
    label: 'OpenAI',
    baseUrl: 'https://api.openai.com/v1',
    apiKeyDefault: '',
    models: ['gpt-4o-mini', 'gpt-4o', 'gpt-3.5-turbo'],
    embedModel: 'text-embedding-3-small',
  },
  qwen: {
    label: '通义千问',
    baseUrl: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
    apiKeyDefault: '',
    models: ['qwen-plus', 'qwen-turbo', 'qwen-max'],
    embedModel: 'text-embedding-v3',
  },
  moonshot: {
    label: 'Moonshot（Kimi）',
    baseUrl: 'https://api.moonshot.cn/v1',
    apiKeyDefault: '',
    models: ['moonshot-v1-8k', 'moonshot-v1-32k', 'moonshot-v1-128k'],
    embedModel: 'text-embedding-3-small',
  },
  groq: {
    label: 'Groq',
    baseUrl: 'https://api.groq.com/openai/v1',
    apiKeyDefault: '',
    models: ['llama-3.3-70b-versatile', 'llama-3.1-8b-instant', 'mixtral-8x7b-32768'],
    embedModel: 'text-embedding-3-small',
  },
  ollama: {
    label: 'Ollama（本地离线）',
    baseUrl: 'http://127.0.0.1:11434/v1',
    apiKeyDefault: 'ollama',
    models: ['llama3.2', 'qwen2.5:7b', 'mistral', 'gemma2:9b', 'deepseek-r1:7b'],
    embedModel: 'nomic-embed-text',
  },
  lmstudio: {
    label: 'LM Studio（本地离线）',
    baseUrl: 'http://127.0.0.1:1234/v1',
    apiKeyDefault: 'lmstudio',
    models: ['local-model'],
    embedModel: 'nomic-embed-text',
  },
}

const defaults = {
  enabled: false,
  provider: 'deepseek',
  baseUrl: AI_PROVIDERS.deepseek.baseUrl,
  chatModel: AI_PROVIDERS.deepseek.models[0],
  apiKey: '',
  connectionVerified: false,
  connectionVerifiedAt: '',
  workspacePath: 'agents/main/workspace',
}

function newProfileId() {
  return `cfg-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

function normalizeConfig(raw) {
  const merged = { ...defaults, ...raw }
  delete merged.embedModel
  delete merged.openclawGateway
  delete merged.openclawGatewayToken
  delete merged.openclawModel
  delete merged.openclawJson
  if (merged.connectionVerified) {
    merged.enabled = true
  }
  return merged
}

function readLegacyConfig() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    return normalizeConfig(JSON.parse(raw))
  } catch {
    return null
  }
}

function writeLegacyConfig(cfg) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(normalizeConfig(cfg)))
}

export function isOfflineProvider(provider) {
  return OFFLINE_PROVIDERS.includes(provider)
}

export function needsApiKey(cfg) {
  return !isOfflineProvider(cfg.provider)
}

export function isAiConfigReady(cfg) {
  if (!cfg.connectionVerified) return false
  if (!cfg.enabled) return false
  if (needsApiKey(cfg) && !cfg.apiKey?.trim()) return false
  if (!cfg.baseUrl?.trim()) return false
  if (!cfg.chatModel?.trim()) return false
  return true
}

/** @returns {'connectionNotVerified'|'notEnabled'|'apiKeyRequired'|'baseUrlRequired'|'disabledHint'|null} */
export function aiConfigBlockReason(cfg) {
  if (!cfg.connectionVerified) return 'connectionNotVerified'
  if (!cfg.enabled) return 'notEnabled'
  if (needsApiKey(cfg) && !cfg.apiKey?.trim()) return 'apiKeyRequired'
  if (!cfg.baseUrl?.trim()) return 'baseUrlRequired'
  if (!cfg.chatModel?.trim()) return 'disabledHint'
  return null
}

export function markConnectionUnverified(cfg) {
  return { ...cfg, connectionVerified: false, connectionVerifiedAt: '' }
}

export function markConnectionVerified(cfg) {
  return {
    ...cfg,
    connectionVerified: true,
    connectionVerifiedAt: new Date().toISOString(),
  }
}

export function offlineDockerConfig() {
  return normalizeConfig(
    applyProviderDefaults(
      {
        enabled: true,
        baseUrl: 'http://host.docker.internal:11434/v1',
      },
      'ollama'
    )
  )
}

/** @deprecated use offlineDockerConfig */
export function dockerBuiltinConfig() {
  return offlineDockerConfig()
}

function createProfile(name, description, config) {
  const now = new Date().toISOString()
  return {
    id: newProfileId(),
    name: name.trim() || '未命名配置',
    description: (description || '').trim(),
    config: normalizeConfig(config),
    createdAt: now,
    updatedAt: now,
  }
}

function initProfilesState() {
  const legacy = readLegacyConfig()
  const profiles = []

  if (legacy) {
    profiles.push(createProfile('默认配置', '云端 API 或本地离线', legacy))
  } else {
    profiles.push(
      createProfile(
        '默认配置',
        'DeepSeek 云端；填写 API Key 后启用',
        { ...defaults }
      )
    )
  }

  return { profiles, activeProfileId: profiles[0].id }
}

export function loadProfilesState() {
  try {
    const raw = localStorage.getItem(PROFILES_KEY)
    if (!raw) return initProfilesState()
    const parsed = JSON.parse(raw)
    const profiles = Array.isArray(parsed.profiles)
      ? parsed.profiles.map((p) => ({
          ...p,
          config: normalizeConfig(p.config || {}),
        }))
      : []
    if (!profiles.length) return initProfilesState()
    const activeProfileId =
      parsed.activeProfileId && profiles.some((p) => p.id === parsed.activeProfileId)
        ? parsed.activeProfileId
        : profiles[0].id
    return { profiles, activeProfileId }
  } catch {
    return initProfilesState()
  }
}

export function saveProfilesState(state) {
  localStorage.setItem(PROFILES_KEY, JSON.stringify(state))
  const active = state.profiles.find((p) => p.id === state.activeProfileId)
  if (active) writeLegacyConfig(active.config)
  notifyAiConfigChanged()
  return state
}

export function getActiveProfile(state = loadProfilesState()) {
  return state.profiles.find((p) => p.id === state.activeProfileId) || state.profiles[0]
}

export function profileOptions(state = loadProfilesState()) {
  return state.profiles.map((p) => ({
    value: p.id,
    label: p.description ? `${p.name} — ${p.description}` : p.name,
  }))
}

export function applyProfile(profileId) {
  const state = loadProfilesState()
  if (!state.profiles.some((p) => p.id === profileId)) return state
  state.activeProfileId = profileId
  saveProfilesState(state)
  return state
}

export function saveProfileAs(name, description, config) {
  const state = loadProfilesState()
  const profile = createProfile(name, description, config)
  state.profiles.push(profile)
  state.activeProfileId = profile.id
  saveProfilesState(state)
  return profile
}

export function updateActiveProfileConfig(cfg) {
  const state = loadProfilesState()
  const profile = state.profiles.find((p) => p.id === state.activeProfileId)
  if (!profile) return state
  profile.config = normalizeConfig({ ...profile.config, ...cfg })
  profile.updatedAt = new Date().toISOString()
  saveProfilesState(state)
  return profile.config
}

export function renameProfile(profileId, name, description) {
  const state = loadProfilesState()
  const profile = state.profiles.find((p) => p.id === profileId)
  if (!profile) return state
  profile.name = name.trim() || profile.name
  profile.description = (description ?? profile.description).trim()
  profile.updatedAt = new Date().toISOString()
  saveProfilesState(state)
  return profile
}

export function deleteProfile(profileId) {
  const state = loadProfilesState()
  if (state.profiles.length <= 1) return state
  state.profiles = state.profiles.filter((p) => p.id !== profileId)
  if (state.activeProfileId === profileId) {
    state.activeProfileId = state.profiles[0].id
  }
  saveProfilesState(state)
  return state
}

export function getProviderPreset(provider) {
  return AI_PROVIDERS[provider] || AI_PROVIDERS.deepseek
}

export function providerModelOptions(provider) {
  const preset = getProviderPreset(provider)
  return preset.models.map((m) => ({ value: m, label: m }))
}

export function applyProviderDefaults(cfg, provider) {
  const preset = getProviderPreset(provider)
  const switching = cfg.provider !== provider
  return {
    ...cfg,
    provider,
    baseUrl: preset.baseUrl,
    chatModel: preset.models[0] || cfg.chatModel,
    apiKey: switching ? preset.apiKeyDefault || '' : cfg.apiKey || preset.apiKeyDefault || '',
  }
}

export function embedModelForProvider(provider) {
  return getProviderPreset(provider).embedModel
}

export function loadAiConfig() {
  const profile = getActiveProfile()
  return profile ? { ...profile.config } : { ...defaults }
}

export function saveAiConfig(cfg) {
  return updateActiveProfileConfig(cfg)
}

function openclawProviderKey(provider) {
  return provider === 'deepseek' ? 'deepseek' : provider
}

function openclawBaseUrl(cfg, preset) {
  let url = (cfg.baseUrl || preset.baseUrl || '').replace(/\/+$/, '')
  if (cfg.provider === 'deepseek' && url.endsWith('/v1')) {
    url = url.slice(0, -3)
  }
  return url
}

function openclawModelEntry(modelId, preset) {
  return {
    id: modelId,
    name: `${preset.label} ${modelId}`,
    reasoning: /reasoner|r1/i.test(modelId),
    input: ['text'],
    cost: { input: 0, output: 0, cacheRead: 0, cacheWrite: 0 },
    contextWindow: 128000,
    maxTokens: 8192,
  }
}

const OPENCLAW_PROVIDER_ENV = {
  deepseek: 'DEEPSEEK_API_KEY',
  openai: 'OPENAI_API_KEY',
  qwen: 'DASHSCOPE_API_KEY',
  moonshot: 'MOONSHOT_API_KEY',
  groq: 'GROQ_API_KEY',
  ollama: 'OLLAMA_API_KEY',
  lmstudio: 'LMSTUDIO_API_KEY',
}

function openclawAuthBlocks(providerKey, apiKey) {
  if (!apiKey) return {}
  const profileId = `${providerKey}:default`
  const envName = OPENCLAW_PROVIDER_ENV[providerKey]
  return {
    env: envName ? { [envName]: apiKey } : {},
    auth: {
      profiles: {
        [profileId]: { provider: providerKey, mode: 'api_key' },
      },
      order: {
        [providerKey]: [profileId],
      },
    },
  }
}

export function buildOpenClawConfig(cfg = loadAiConfig()) {
  const preset = getProviderPreset(cfg.provider)
  const providerKey = openclawProviderKey(cfg.provider)
  const modelId = cfg.chatModel || preset.models[0]
  const modelRef = `${providerKey}/${modelId}`
  const apiKey = cfg.apiKey || preset.apiKeyDefault || ''
  const baseUrl = openclawBaseUrl(cfg, preset)

  const config = {
    gateway: {
      mode: 'local',
      port: 18789,
      bind: 'lan',
      controlUi: {
        allowedOrigins: [
          'http://localhost:18789',
          'http://127.0.0.1:18789',
          'http://localhost:25173',
          'http://127.0.0.1:25173',
        ],
      },
      http: {
        endpoints: {
          chatCompletions: { enabled: true },
        },
      },
    },
    models: {
      mode: 'merge',
      providers: {
        [providerKey]: {
          baseUrl,
          apiKey,
          api: 'openai-completions',
          models: [openclawModelEntry(modelId, preset)],
        },
      },
    },
    agents: {
      defaults: {
        workspace: 'workspace-smart-ledger',
        timeoutSeconds: 180,
        model: { primary: modelRef },
        models: { [modelRef]: {} },
      },
    },
    ...openclawAuthBlocks(providerKey, apiKey),
  }

  if (isOfflineProvider(cfg.provider)) {
    const embedModel = preset.embedModel
    config.plugins = {
      slots: { memory: 'memory-lancedb' },
      entries: {
        'memory-lancedb': {
          enabled: true,
          config: {
            embedding: {
              baseUrl: cfg.baseUrl,
              model: embedModel,
              apiKey,
            },
            dbPath: '/data/openclaw/lancedb',
          },
        },
      },
    }
  }

  return config
}

export function exportOpenClawSnippet(cfg = loadAiConfig()) {
  const custom = (cfg.openclawJson || '').trim()
  if (custom) {
    try {
      return JSON.stringify(JSON.parse(custom), null, 2)
    } catch {
      return custom
    }
  }
  return JSON.stringify(buildOpenClawConfig(cfg), null, 2)
}

export function syncOpenClawJsonFromFields(cfg) {
  return exportOpenClawSnippet({ ...cfg, openclawJson: '' })
}

export function parseOpenClawJson(text) {
  return JSON.parse(text)
}
