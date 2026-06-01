import {
  AUTO_GENERATED_WORKSPACE_FILES,
  buildSystemPromptFromWorkspace,
  defaultWorkspaceFiles,
} from './agentWorkspace'

const STORAGE_KEY = 'smart-ledger-ai-agents'
const CHAT_KEY_PREFIX = 'smart-ledger-ai-chat-'
const WS_KEY_PREFIX = 'smart-ledger-ai-ws-'
const CTX_KEY_PREFIX = 'smart-ledger-ai-ctx-'

/** OpenClaw 默认主 Agent 目录（相对 config 根） */
export const DEFAULT_AGENT_PATH = 'agents/main'

export function defaultChatHistoryPath(agentPath = DEFAULT_AGENT_PATH) {
  const base = (agentPath || DEFAULT_AGENT_PATH).replace(/\\/g, '/').replace(/\/+$/, '')
  return `${base}/ChatMessages`
}

export function agentPathForNewAgent(name, id) {
  const slug = (name || 'agent')
    .trim()
    .replace(/\s+/g, '-')
    .replace(/[^\w\u4e00-\u9fff-]/g, '')
    .slice(0, 24) || 'agent'
  const short = (id || '').split('-').pop()?.slice(0, 6) || Date.now().toString(36).slice(-6)
  return `agents/${slug}-${short}`
}

export function resolveAgentPaths(agent) {
  const agentPath = (agent?.agentPath || DEFAULT_AGENT_PATH).trim() || DEFAULT_AGENT_PATH
  const chatHistoryPath =
    (agent?.chatHistoryPath || '').trim() || defaultChatHistoryPath(agentPath)
  return { agentPath, chatHistoryPath }
}

function chatStorageKey(agentId) {
  return CHAT_KEY_PREFIX + agentId
}

export function loadAgentMessages(agentId) {
  if (!agentId) return []
  try {
    const raw = localStorage.getItem(chatStorageKey(agentId))
    if (!raw) return []
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export function saveAgentMessages(agentId, messages) {
  if (!agentId) return false
  try {
    if (!messages?.length) {
      localStorage.removeItem(chatStorageKey(agentId))
      return true
    }
    localStorage.setItem(chatStorageKey(agentId), JSON.stringify(messages))
    return true
  } catch (err) {
    console.warn('[aiAgents] chat save failed:', err)
    return false
  }
}

function loadAgentWorkspaceFiles(agentId) {
  if (!agentId) return null
  try {
    const raw = localStorage.getItem(WS_KEY_PREFIX + agentId)
    if (!raw) return null
    const parsed = JSON.parse(raw)
    return parsed && typeof parsed === 'object' ? parsed : null
  } catch {
    return null
  }
}

function saveAgentWorkspaceFiles(agentId, workspaceFiles) {
  if (!agentId || !workspaceFiles) return false
  try {
    localStorage.setItem(WS_KEY_PREFIX + agentId, JSON.stringify(workspaceFiles))
    return true
  } catch (err) {
    console.warn('[aiAgents] workspace save failed:', err)
    return false
  }
}

function loadAgentContext(agentId) {
  if (!agentId) return null
  try {
    const raw = localStorage.getItem(CTX_KEY_PREFIX + agentId)
    if (!raw) return null
    return JSON.parse(raw)
  } catch {
    return null
  }
}

function saveAgentContext(agentId, ledgerId, ledgerContextText) {
  if (!agentId) return false
  try {
    if (!ledgerId && !ledgerContextText) {
      localStorage.removeItem(CTX_KEY_PREFIX + agentId)
      return true
    }
    localStorage.setItem(
      CTX_KEY_PREFIX + agentId,
      JSON.stringify({ ledgerId: ledgerId || '', ledgerContextText: ledgerContextText || '' })
    )
    return true
  } catch (err) {
    console.warn('[aiAgents] context save failed:', err)
    return false
  }
}

function attachAgentData(agent) {
  const base = migrateAgent({ ...agent })
  const storedMessages = loadAgentMessages(base.id)
  if (storedMessages.length) {
    base.messages = storedMessages
  } else if (Array.isArray(base.messages) && base.messages.length) {
    saveAgentMessages(base.id, base.messages)
  } else {
    base.messages = []
  }
  const storedWs = loadAgentWorkspaceFiles(base.id)
  if (storedWs) base.workspaceFiles = storedWs
  const storedCtx = loadAgentContext(base.id)
  if (storedCtx) {
    base.ledgerId = storedCtx.ledgerId ?? base.ledgerId
    base.ledgerContextText = storedCtx.ledgerContextText ?? ''
  }
  return base
}

/** 从孤儿 chat key 恢复 Agent 列表（主存储缺失或损坏时） */
function recoverAgentsFromChatKeys() {
  if (typeof localStorage === 'undefined') return null
  const ids = []
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i)
    if (!key?.startsWith(CHAT_KEY_PREFIX)) continue
    const agentId = key.slice(CHAT_KEY_PREFIX.length)
    if (agentId && loadAgentMessages(agentId).length) ids.push(agentId)
  }
  if (!ids.length) return null
  ids.sort()
  const agents = ids.map((id, index) =>
    attachAgentData({
      id,
      name: index === 0 ? '默认助手' : `助手 ${index + 1}`,
      agentPath: index === 0 ? DEFAULT_AGENT_PATH : agentPathForNewAgent(`助手-${index + 1}`, id),
      chatHistoryPath: '',
      workspaceFiles: defaultWorkspaceFiles(),
      messages: [],
      ledgerId: '',
      ledgerContextText: '',
      createdAt: new Date().toISOString(),
    })
  )
  return { agents, activeId: agents[0].id }
}

function bootstrapDefaultState() {
  const def = createDefaultAgent()
  const state = { agents: [attachAgentData(def)], activeId: def.id }
  saveAgentsState(state)
  return state
}

function newId() {
  return `agent-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

function migrateAgent(agent) {
  if (!agent.workspaceFiles || typeof agent.workspaceFiles !== 'object') {
    const files = defaultWorkspaceFiles()
    if (agent.systemPrompt?.trim()) {
      files['AGENTS.md'] = agent.systemPrompt.trim()
    }
    agent.workspaceFiles = files
  }
  const defaults = defaultWorkspaceFiles()
  for (const name of Object.keys(defaults)) {
    if (agent.workspaceFiles[name] == null) {
      agent.workspaceFiles[name] = defaults[name]
    }
  }
  for (const name of AUTO_GENERATED_WORKSPACE_FILES) {
    agent.workspaceFiles[name] = defaults[name]
  }
  delete agent.systemPrompt
  if (agent.ledgerId == null) agent.ledgerId = ''
  if (agent.ledgerContextText == null) agent.ledgerContextText = ''
  if (!agent.agentPath?.trim()) agent.agentPath = DEFAULT_AGENT_PATH
  if (!agent.chatHistoryPath?.trim()) {
    agent.chatHistoryPath = defaultChatHistoryPath(agent.agentPath)
  }
  if (!Array.isArray(agent.messages)) agent.messages = []
  return agent
}

export function createDefaultAgent() {
  const id = newId()
  const agentPath = DEFAULT_AGENT_PATH
  return {
    id,
    name: '默认助手',
    agentPath,
    chatHistoryPath: defaultChatHistoryPath(agentPath),
    workspaceFiles: defaultWorkspaceFiles(),
    messages: [],
    ledgerId: '',
    ledgerContextText: '',
    createdAt: new Date().toISOString(),
  }
}

export function agentSystemPrompt(agent) {
  if (!agent) return ''
  return buildSystemPromptFromWorkspace(agent.workspaceFiles || defaultWorkspaceFiles())
}

export function loadAgentsState() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      const recovered = recoverAgentsFromChatKeys()
      if (recovered) {
        saveAgentsState(recovered)
        return recovered
      }
      return bootstrapDefaultState()
    }
    const parsed = JSON.parse(raw)
    const agents = Array.isArray(parsed.agents)
      ? parsed.agents.map((a) => attachAgentData(a))
      : []
    if (!agents.length) {
      const recovered = recoverAgentsFromChatKeys()
      if (recovered) {
        saveAgentsState(recovered)
        return recovered
      }
      return bootstrapDefaultState()
    }
    const activeId = parsed.activeId && agents.some((a) => a.id === parsed.activeId)
      ? parsed.activeId
      : agents[0].id
    return { agents, activeId }
  } catch (err) {
    console.warn('[aiAgents] load failed, attempting recovery:', err)
    const recovered = recoverAgentsFromChatKeys()
    if (recovered) {
      saveAgentsState(recovered)
      return recovered
    }
    return bootstrapDefaultState()
  }
}

export function saveAgentsState(state) {
  const slim = {
    activeId: state.activeId,
    agents: state.agents.map(({ messages, workspaceFiles, ledgerContextText, ...rest }) => rest),
  }
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(slim))
  } catch (err) {
    console.warn('[aiAgents] save failed:', err)
  }
  for (const agent of state.agents) {
    if (agent.workspaceFiles) saveAgentWorkspaceFiles(agent.id, agent.workspaceFiles)
    saveAgentContext(agent.id, agent.ledgerId, agent.ledgerContextText)
    if (Array.isArray(agent.messages) && agent.messages.length) {
      saveAgentMessages(agent.id, agent.messages)
    }
  }
  return state
}

/** @returns {typeof state | null} */
export function patchAgentMessages(agentId, messages) {
  saveAgentMessages(agentId, messages)
  const state = loadAgentsState()
  const agent = state.agents.find((a) => a.id === agentId)
  if (!agent) return null
  return {
    ...state,
    agents: state.agents.map((a) =>
      a.id === agentId ? { ...a, messages } : a
    ),
  }
}

export function getActiveAgent(state = loadAgentsState()) {
  return state.agents.find((a) => a.id === state.activeId) || state.agents[0]
}

export function addAgent(name, workspaceFiles = null) {
  const state = loadAgentsState()
  const id = newId()
  const agentPath = agentPathForNewAgent(name, id)
  const agent = {
    id,
    name: name.trim() || '新助手',
    agentPath,
    chatHistoryPath: defaultChatHistoryPath(agentPath),
    workspaceFiles: workspaceFiles || defaultWorkspaceFiles(),
    messages: [],
    ledgerId: '',
    ledgerContextText: '',
    createdAt: new Date().toISOString(),
  }
  state.agents.push(agent)
  state.activeId = agent.id
  saveAgentsState(state)
  return agent
}

export function removeAgent(agentId) {
  const state = loadAgentsState()
  if (state.agents.length <= 1) return state
  state.agents = state.agents.filter((a) => a.id !== agentId)
  if (state.activeId === agentId) {
    state.activeId = state.agents[0].id
  }
  saveAgentsState(state)
  try {
    localStorage.removeItem(chatStorageKey(agentId))
    localStorage.removeItem(WS_KEY_PREFIX + agentId)
    localStorage.removeItem(CTX_KEY_PREFIX + agentId)
  } catch {
    /* ignore */
  }
  return state
}

export function updateAgentMessages(agentId, messages) {
  return patchAgentMessages(agentId, messages)
}

export function updateAgentWorkspaceFiles(agentId, workspaceFiles) {
  const state = loadAgentsState()
  const agent = state.agents.find((a) => a.id === agentId)
  if (agent) {
    agent.workspaceFiles = { ...agent.workspaceFiles, ...workspaceFiles }
    saveAgentWorkspaceFiles(agentId, agent.workspaceFiles)
    saveAgentsState(state)
  }
  return state
}

export function resetAgentWorkspaceFiles(agentId) {
  const state = loadAgentsState()
  const agent = state.agents.find((a) => a.id === agentId)
  if (agent) {
    agent.workspaceFiles = defaultWorkspaceFiles()
    saveAgentsState(state)
  }
  return state
}

export function updateAgentLedgerContext(agentId, ledgerId, ledgerContextText) {
  const state = loadAgentsState()
  const agent = state.agents.find((a) => a.id === agentId)
  if (agent) {
    agent.ledgerId = ledgerId || ''
    agent.ledgerContextText = ledgerContextText || ''
    saveAgentContext(agentId, agent.ledgerId, agent.ledgerContextText)
    saveAgentsState(state)
  }
  return state
}

export function setActiveAgent(agentId) {
  const state = loadAgentsState()
  if (state.agents.some((a) => a.id === agentId)) {
    state.activeId = agentId
    saveAgentsState(state)
  }
  return state
}

export function updateAgentPaths(agentId, agentPath, chatHistoryPath) {
  const state = loadAgentsState()
  const agent = state.agents.find((a) => a.id === agentId)
  if (!agent) return state
  const nextAgentPath = (agentPath || DEFAULT_AGENT_PATH).trim() || DEFAULT_AGENT_PATH
  agent.agentPath = nextAgentPath
  agent.chatHistoryPath =
    (chatHistoryPath || '').trim() || defaultChatHistoryPath(nextAgentPath)
  saveAgentsState(state)
  return state
}
