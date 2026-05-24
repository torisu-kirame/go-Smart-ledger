const STORAGE_KEY = 'smart-ledger-ai-config'

const defaults = {
  enabled: false,
  provider: 'ollama',
  baseUrl: 'http://127.0.0.1:11434/v1',
  chatModel: 'llama3.2',
  embedModel: 'nomic-embed-text',
  apiKey: 'ollama',
  openclawGateway: 'http://127.0.0.1:18789',
  workspacePath: 'integrations/openclaw/workspace-smart-ledger',
}

export function loadAiConfig() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return { ...defaults }
    return { ...defaults, ...JSON.parse(raw) }
  } catch {
    return { ...defaults }
  }
}

export function saveAiConfig(cfg) {
  const next = { ...loadAiConfig(), ...cfg }
  localStorage.setItem(STORAGE_KEY, JSON.stringify(next))
  return next
}

export function exportOpenClawSnippet(cfg = loadAiConfig()) {
  return JSON.stringify(
    {
      agents: {
        defaults: {
          model: {
            provider: cfg.provider,
            baseUrl: cfg.baseUrl,
            id: cfg.chatModel,
            apiKey: cfg.apiKey || 'ollama',
          },
        },
      },
      plugins: {
        slots: { memory: 'memory-lancedb' },
        entries: {
          'memory-lancedb': {
            enabled: true,
            config: {
              embedding: {
                baseUrl: cfg.baseUrl,
                model: cfg.embedModel,
                apiKey: cfg.apiKey || 'ollama',
              },
              dbPath: 'data/openclaw/lancedb',
            },
          },
        },
      },
      gateway: { port: 18789 },
    },
    null,
    2
  )
}
