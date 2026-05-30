import {
  buildSystemPromptFromWorkspace,
  defaultWorkspaceFiles,
} from './agentWorkspace'

const STORAGE_KEY = 'smart-ledger-ai-agents'

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
  for (const name of Object.keys(defaultWorkspaceFiles())) {
    if (agent.workspaceFiles[name] == null) {
      agent.workspaceFiles[name] = defaultWorkspaceFiles()[name]
    }
  }
  delete agent.systemPrompt
  if (agent.ledgerId == null) agent.ledgerId = ''
  if (agent.ledgerContextText == null) agent.ledgerContextText = ''
  return agent
}

export function createDefaultAgent() {
  return {
    id: newId(),
    name: '默认助手',
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
      const def = createDefaultAgent()
      return { agents: [def], activeId: def.id }
    }
    const parsed = JSON.parse(raw)
    const agents = Array.isArray(parsed.agents)
      ? parsed.agents.map(migrateAgent)
      : []
    if (!agents.length) {
      const def = createDefaultAgent()
      return { agents: [def], activeId: def.id }
    }
    const activeId = parsed.activeId && agents.some((a) => a.id === parsed.activeId)
      ? parsed.activeId
      : agents[0].id
    return { agents, activeId }
  } catch {
    const def = createDefaultAgent()
    return { agents: [def], activeId: def.id }
  }
}

export function saveAgentsState(state) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
  return state
}

export function getActiveAgent(state = loadAgentsState()) {
  return state.agents.find((a) => a.id === state.activeId) || state.agents[0]
}

export function addAgent(name, workspaceFiles = null) {
  const state = loadAgentsState()
  const agent = {
    id: newId(),
    name: name.trim() || '新助手',
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
  return state
}

export function updateAgentMessages(agentId, messages) {
  const state = loadAgentsState()
  const agent = state.agents.find((a) => a.id === agentId)
  if (agent) {
    agent.messages = messages
    saveAgentsState(state)
  }
  return state
}

export function updateAgentWorkspaceFiles(agentId, workspaceFiles) {
  const state = loadAgentsState()
  const agent = state.agents.find((a) => a.id === agentId)
  if (agent) {
    agent.workspaceFiles = { ...agent.workspaceFiles, ...workspaceFiles }
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
