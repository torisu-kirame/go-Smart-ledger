import AGENTS from '../assets/agent-workspace/AGENTS.md?raw'
import API_REFERENCE from '../assets/agent-workspace/API-REFERENCE.md?raw'
import BOOTSTRAP from '../assets/agent-workspace/BOOTSTRAP.md?raw'
import HEARTBEAT from '../assets/agent-workspace/HEARTBEAT.md?raw'
import IDENTITY from '../assets/agent-workspace/IDENTITY.md?raw'
import SOUL from '../assets/agent-workspace/SOUL.md?raw'
import TOOLS from '../assets/agent-workspace/TOOLS.md?raw'
import USER from '../assets/agent-workspace/USER.md?raw'

/** OpenClaw 工作区标准设定文件（官方默认版本） */
export const WORKSPACE_FILE_NAMES = [
  'AGENTS.md',
  'API-REFERENCE.md',
  'BOOTSTRAP.md',
  'HEARTBEAT.md',
  'IDENTITY.md',
  'SOUL.md',
  'TOOLS.md',
  'USER.md',
]

/** 随 OpenAPI 重新生成，加载 Agent 时用最新默认覆盖 */
export const AUTO_GENERATED_WORKSPACE_FILES = ['API-REFERENCE.md']

const DEFAULTS = {
  'AGENTS.md': AGENTS,
  'API-REFERENCE.md': API_REFERENCE,
  'BOOTSTRAP.md': BOOTSTRAP,
  'HEARTBEAT.md': HEARTBEAT,
  'IDENTITY.md': IDENTITY,
  'SOUL.md': SOUL,
  'TOOLS.md': TOOLS,
  'USER.md': USER,
}

export function defaultWorkspaceFiles() {
  return { ...DEFAULTS }
}

export function buildSystemPromptFromWorkspace(files = {}) {
  const order = ['IDENTITY.md', 'SOUL.md', 'AGENTS.md', 'API-REFERENCE.md', 'USER.md', 'TOOLS.md']
  const parts = []
  for (const name of order) {
    const text = (files[name] ?? DEFAULTS[name] ?? '').trim()
    if (text) parts.push(`## ${name}\n${text}`)
  }
  return parts.join('\n\n')
}

export function workspaceFileLabel(name) {
  return name.replace(/\.md$/, '')
}
