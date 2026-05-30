import { authHeaders } from './http'
import { resolveAgentPaths } from '../utils/aiAgents'

const BASE = '/api/v1'

/**
 * Load chat history and optional workspace from OpenClaw agent directories on disk.
 */
export async function loadAgentFromDisk(agent, { loadWorkspace = false } = {}) {
  const { agentPath, chatHistoryPath } = resolveAgentPaths(agent)
  const res = await fetch(`${BASE}/ai/agent/load`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(),
    },
    body: JSON.stringify({
      agentPath,
      chatHistoryPath,
      loadWorkspace,
    }),
  })
  if (!res.ok) {
    const text = await res.text()
    let msg = res.statusText
    try {
      const j = JSON.parse(text)
      msg = j.msg || j.message || msg
    } catch {
      if (text) msg = text
    }
    throw new Error(msg)
  }
  return res.json()
}

/**
 * Persist chat history and/or workspace markdown to agent directories on disk.
 */
export async function saveAgentToDisk(agent, { messages, workspaceFiles } = {}) {
  const { agentPath, chatHistoryPath } = resolveAgentPaths(agent)
  const body = { agentPath, chatHistoryPath }
  if (messages?.length) body.messages = messages
  if (workspaceFiles && Object.keys(workspaceFiles).length) {
    body.workspaceFiles = workspaceFiles
  }
  if (!body.messages && !body.workspaceFiles) return
  const res = await fetch(`${BASE}/ai/agent/save`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(),
    },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    const text = await res.text()
    let msg = res.statusText
    try {
      const j = JSON.parse(text)
      msg = j.msg || j.message || msg
    } catch {
      if (text) msg = text
    }
    throw new Error(msg)
  }
  return res.json()
}
