import { decryptEntryData } from './e2eCrypto'
import { cellLabel, cellValue } from './entrySchema'

export const ENTRY_EVENT_TYPE = 'EntryAdded'
export const ENTRY_VOIDED_TYPE = 'EntryVoided'

const VIEW_MODE_KEY = 'ledger-content-view'

export function getContentViewMode() {
  const v = localStorage.getItem(VIEW_MODE_KEY)
  return v === 'rows' ? 'rows' : 'table'
}

export function setContentViewMode(mode) {
  localStorage.setItem(VIEW_MODE_KEY, mode === 'rows' ? 'rows' : 'table')
}

function parsePayload(raw) {
  if (!raw) return null
  if (typeof raw === 'object') return raw
  if (typeof raw === 'string') {
    try {
      return JSON.parse(raw)
    } catch {
      return null
    }
  }
  return null
}

async function resolveEntryData(payload, groupKey) {
  const entry = parsePayload(payload)
  if (!entry) return null
  let data = entry.data
  if (!data && typeof entry === 'object') {
    const { schemaId: _s, signerId: _u, data: _d, ...rest } = entry
    if (Object.keys(rest).length) data = rest
  }
  if (!data) return null
  if (data.__encrypted) {
    if (!groupKey) return { __locked: true }
    try {
      return await decryptEntryData(groupKey, data)
    } catch {
      return { __locked: true }
    }
  }
  return data
}

function payloadTableId(ev) {
  try {
    const p = typeof ev.payload === 'string' ? JSON.parse(ev.payload) : ev.payload
    return p?.tableId || 'default'
  } catch {
    return 'default'
  }
}

/** Collect voided entry seqs from EntryVoided events. */
export function collectVoidedSeqs(events) {
  const out = new Set()
  for (const ev of events || []) {
    if (ev.type !== ENTRY_VOIDED_TYPE) continue
    const p = parsePayload(ev.payload)
    const seq = Number(p?.seq || 0)
    if (seq > 0) out.add(seq)
  }
  return out
}

/**
 * 从链上事件解析记账行（仅 EntryAdded，排除已作废）；可选按 tableId / rowOrder 过滤排序
 */
export async function buildEntryRows(events, schema, groupKey = '', tableId = null, rowOrder = null) {
  const fields = schema?.fields || []
  const voided = collectVoidedSeqs(events)
  const rows = []
  for (const ev of events || []) {
    if (ev.type !== ENTRY_EVENT_TYPE) continue
    if (voided.has(ev.seq)) continue
    if (tableId && payloadTableId(ev) !== tableId) continue
    const data = await resolveEntryData(ev.payload, groupKey)
    if (!data) continue
    if (data.__locked) {
      rows.push({
        seq: ev.seq,
        signerId: ev.signerId,
        createdAt: ev.createdAt,
        locked: true,
        cells: {},
      })
      continue
    }
    const cells = {}
    for (const f of fields) {
      cells[f.key] = data[f.key] ?? ''
    }
    rows.push({
      seq: ev.seq,
      tableId: payloadTableId(ev),
      signerId: ev.signerId,
      createdAt: ev.createdAt,
      locked: false,
      cells,
    })
  }

  if (Array.isArray(rowOrder) && rowOrder.length) {
    const bySeq = new Map(rows.map((r) => [r.seq, r]))
    const ordered = []
    for (const seq of rowOrder) {
      const r = bySeq.get(seq)
      if (r) {
        ordered.push(r)
        bySeq.delete(seq)
      }
    }
    const rest = [...bySeq.values()].sort((a, b) => b.seq - a.seq)
    return [...ordered, ...rest]
  }
  return rows.reverse()
}

export function displayCell(row, field, members = []) {
  if (row.locked) return '🔒 已加密'
  const v = cellValue(row, field.key)
  if (!v) return '—'
  if (field.type === 'user') {
    const m = members.find((x) => x.id === v)
    return m?.nickname || m?.username || v
  }
  return v
}

export function contentColumns(schema) {
  const fields = schema?.fields || []
  // UI chain sequence — label「#」避免与业务字段「序号」重复显示
  const cols = [{ key: '_seq', label: '#', fixed: true }]
  for (const f of fields) {
    const label = cellLabel(schema, f.key)
    // hide legacy duplicate 序号 fields created by earlier AI imports
    if (label === '序号' && (f.key === 'seq' || f.key === '序号' || f.key === 'no' || f.key === 'line_no')) {
      continue
    }
    cols.push({ key: f.key, label, field: f })
  }
  return cols
}
