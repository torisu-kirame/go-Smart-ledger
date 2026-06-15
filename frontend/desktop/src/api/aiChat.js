import { authHeaders } from './http'
import {
  isAiConfigReady,
  loadAiConfig,
  markConnectionVerified,
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

/** Test LangChain LLM connection via ledger-api. */
export async function testAiConnection(cfg = loadAiConfig()) {
  const res = await fetch(`${BASE}/ai/test`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(),
    },
    body: JSON.stringify({
      baseUrl: cfg.baseUrl,
      apiKey: cfg.apiKey,
      model: cfg.chatModel,
    }),
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(parseApiError(text, res.statusText))
  }
  saveAiConfig(markConnectionVerified({ ...cfg, enabled: true }))
  return true
}

function extractStreamDelta(chunk) {
  const choice = chunk?.choices?.[0]
  if (choice) {
    const fromDelta = choice.delta?.content
    if (typeof fromDelta === 'string' && fromDelta) return fromDelta
    const fromMsg = choice.message?.content
    if (typeof fromMsg === 'string' && fromMsg) return fromMsg
  }
  if (typeof chunk?.content === 'string' && chunk.content) return chunk.content
  if (typeof chunk?.text === 'string' && chunk.text) return chunk.text
  if (typeof chunk?.delta === 'string' && chunk.delta) return chunk.delta
  const msg = chunk?.message
  if (typeof msg?.content === 'string' && msg.content) return msg.content
  if (Array.isArray(msg?.content)) {
    return msg.content.map((c) => c?.text || c?.content || '').join('')
  }
  return ''
}

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
    const delta = extractStreamDelta(chunk)
    if (delta && onDelta) onDelta(delta)
    return { done: false, delta }
  } catch (e) {
    if (e instanceof SyntaxError) return { done: false }
    throw e
  }
}

function parseApiError(text, fallback = '') {
  if (!text?.trim()) return fallback
  try {
    const j = JSON.parse(text)
    return j.msg || j.message || fallback
  } catch {
    const m = text.match(/msg:\s*(.+)/s)
    if (m?.[1]) return m[1].trim()
    return text.trim() || fallback
  }
}

function tryParseJsonCompletion(text) {
  try {
    const data = JSON.parse(text)
    return extractStreamDelta(data) || data?.choices?.[0]?.message?.content || ''
  } catch {
    return ''
  }
}

/** Stream chat via LangChain Agent backend (ledger-api). */
export async function streamChat({ messages, signal, onDelta, useTools = false, boundLedgerId = '' }) {
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
      baseUrl: cfg.baseUrl,
      apiKey: cfg.apiKey,
      model: cfg.chatModel,
      messages,
      stream: true,
      useTools: !!useTools,
      boundLedgerId: boundLedgerId || '',
    }),
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(parseApiError(text, res.statusText))
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
    const parts = buffer.split(/\r?\n/)
    buffer = parts.pop() || ''
    for (const line of parts) {
      const { done: streamDone } = parseSseDataLine(line, (delta) => {
        full += delta
        if (onDelta) onDelta(delta)
      })
      if (streamDone) break
    }
  }
  if (buffer.trim()) {
    for (const line of buffer.split(/\r?\n/)) {
      const { done: streamDone } = parseSseDataLine(line, (delta) => {
        full += delta
        if (onDelta) onDelta(delta)
      })
      if (streamDone) break
    }
    if (!full.trim()) {
      const fallback =
        tryParseJsonCompletion(buffer) ||
        tryParseJsonCompletion(
          buffer
            .split(/\r?\n/)
            .map((line) => line.trim().replace(/^data:\s*/, ''))
            .filter(Boolean)
            .join('')
        )
      if (fallback) {
        full = fallback
        if (onDelta) onDelta(fallback)
      }
    }
  }
  if (!full.trim()) {
    const plainErr = parseApiError(buffer, '')
    if (plainErr) throw new Error(plainErr)
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
