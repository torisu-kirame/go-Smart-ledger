import { authHeaders } from './http'
import { loadAiConfig } from '../utils/aiConfig'

const BASE = '/api/v1'

const SYSTEM_PROMPT = `你是 Smart Ledger 智能账本助手。你帮助用户理解账本流水、封账锚定、复式记账与导入数据。
回答应简洁、准确；涉及金额与日期时引用上下文中的具体数字。若上下文不足，请明确说明并建议用户同步账本或选择其他账本。
不要编造不存在的交易或链上哈希。默认使用简体中文回答。`

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
 * Stream chat completion via gateway proxy (avoids browser CORS to Ollama).
 * @param {object} params
 * @param {Array<{role:string,content:string}>} params.messages
 * @param {AbortSignal} [params.signal]
 * @param {(delta: string) => void} [params.onDelta]
 */
export async function streamChat({ messages, signal, onDelta }) {
  const cfg = loadAiConfig()
  if (!cfg.enabled) {
    throw new Error('AI_DISABLED')
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
      apiKey: cfg.apiKey || 'ollama',
      model: cfg.chatModel,
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
      const trimmed = line.trim()
      if (!trimmed.startsWith('data:')) continue
      const payload = trimmed.slice(5).trim()
      if (payload === '[DONE]') continue
      try {
        const chunk = JSON.parse(payload)
        const delta = chunk.choices?.[0]?.delta?.content || ''
        if (delta) {
          full += delta
          if (onDelta) onDelta(delta)
        }
      } catch {
        /* skip malformed sse line */
      }
    }
  }
  return full
}

export function defaultSystemMessages(ledgerContext = '') {
  const sys = ledgerContext
    ? `${SYSTEM_PROMPT}\n\n---\n账本上下文\n---\n${ledgerContext}`
    : SYSTEM_PROMPT
  return [{ role: 'system', content: sys }]
}

export { SYSTEM_PROMPT }
