import { authHeaders } from './http'
import {
  isAiConfigReady,
  loadAiConfig,
  markConnectionVerified,
  openClawConfigObject,
  saveAiConfig,
} from '../utils/aiConfig'

const BASE = '/api/v1'

export function buildLedgerContextPrompt(exportData, ledgerName) {
  if (!exportData?.chunks?.length) {
    return ''
  }
  const lines = exportData.chunks
    .slice(0, 80)
    .map((c) => c.text)
    .filter(Boolean)
  const header = ledgerName
    ? `以下为账本「${ledgerName}」的链上事件摘要（RAG 导出，共 ${exportData.chunks.length} 条，展示前 ${lines.length} 条）：`
    : `以下为账本链上事件摘要（共 ${exportData.chunks.length} 条）：`
  return `${header}\n\n${lines.join('\n')}`
}

/**
 * Test OpenClaw Gateway connection; on success marks config as verified.
 */
export async function testAiConnection(cfg = loadAiConfig()) {
  let openclawConfig
  try {
    openclawConfig = openClawConfigObject(cfg)
  } catch {
    throw new Error('OPENCLAW_CONFIG_INVALID')
  }
  const res = await fetch(`${BASE}/ai/test`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(),
    },
    body: JSON.stringify({
      gatewayUrl: cfg.openclawGateway,
      gatewayToken: cfg.openclawGatewayToken,
      openclawConfig,
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
  saveAiConfig(markConnectionVerified({ ...cfg, enabled: true }))
  return true
}

/**
 * Parse one SSE data line from OpenClaw / OpenAI-compatible stream.
 */
function parseSseDataLine(line, onDelta) {
  const trimmed = line.trim()
  if (!trimmed.startsWith('data:')) return { done: false }
  const payload = trimmed.slice(5).trim()
  if (payload === '[DONE]') return { done: true }
  try {
    const chunk = JSON.parse(payload)
    if (chunk.error?.message) {
      throw new Error(chunk.error.message)
    }
    const choice = chunk.choices?.[0]
    const delta = choice?.delta?.content || choice?.message?.content || ''
    if (delta && onDelta) onDelta(delta)
    return { done: false, delta }
  } catch (e) {
    if (e instanceof Error && e.message && !e.message.includes('JSON')) throw e
    return { done: false }
  }
}

/**
 * Stream chat via OpenClaw Gateway (backend proxy).
 */
export async function streamChat({ messages, signal, onDelta, agentUser }) {
  const cfg = loadAiConfig()
  if (!cfg.enabled) {
    throw new Error('AI_DISABLED')
  }
  if (!isAiConfigReady(cfg)) {
    throw new Error('CONNECTION_NOT_VERIFIED')
  }
  const res = await fetch(`${BASE}/ai/chat`, {
    method: 'POST',
    credentials: 'include',
    signal,
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(),
    },
    body: JSON.stringify({
      gatewayUrl: cfg.openclawGateway,
      gatewayToken: cfg.openclawGatewayToken,
      openclawModel: cfg.openclawModel || 'openclaw/default',
      agentUser: agentUser || 'smart-ledger-default',
      messages,
      stream: true,
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
  const reader = res.body?.getReader()
  if (!reader) {
    const data = await res.json()
    const content = data?.choices?.[0]?.message?.content || ''
    if (content && onDelta) onDelta(content)
    return content
  }
  const decoder = new TextDecoder()
  let full = ''
  let buffer = ''
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const parts = buffer.split('\n')
    buffer = parts.pop() || ''
    for (const line of parts) {
      const { done } = parseSseDataLine(line, (delta) => {
        full += delta
        if (onDelta) onDelta(delta)
      })
      if (done) break
    }
  }
  if (!full.trim()) {
    throw new Error('AI_EMPTY_RESPONSE')
  }
  return full
}

export function defaultSystemMessages(ledgerContext = '', systemPrompt = '') {
  const base = systemPrompt?.trim() || '你是 Smart Ledger 智能账本助手。'
  const sys = ledgerContext
    ? `${base}\n\n---\n账本上下文\n---\n${ledgerContext}`
    : base
  return [{ role: 'system', content: sys }]
}
