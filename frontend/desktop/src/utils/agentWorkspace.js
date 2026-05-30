import AGENTS from '../assets/agent-workspace/AGENTS.md?raw'
import BOOTSTRAP from '../assets/agent-workspace/BOOTSTRAP.md?raw'
import HEARTBEAT from '../assets/agent-workspace/HEARTBEAT.md?raw'
import IDENTITY from '../assets/agent-workspace/IDENTITY.md?raw'
import SOUL from '../assets/agent-workspace/SOUL.md?raw'
import TOOLS from '../assets/agent-workspace/TOOLS.md?raw'
import USER from '../assets/agent-workspace/USER.md?raw'

/** OpenClaw 工作区标准设定文件（官方默认版本） */
export const WORKSPACE_FILE_NAMES = [
  'AGENTS.md',
  'BOOTSTRAP.md',
  'HEARTBEAT.md',
  'IDENTITY.md',
  'SOUL.md',
  'TOOLS.md',
  'USER.md',
]

const DEFAULTS = {
  'AGENTS.md': AGENTS,
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
  const order = ['IDENTITY.md', 'SOUL.md', 'AGENTS.md', 'USER.md', 'TOOLS.md']
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
