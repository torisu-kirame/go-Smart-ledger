import { decryptEntryData } from './e2eCrypto'
import { cellLabel, cellValue } from './entrySchema'

export const ENTRY_EVENT_TYPE = 'EntryAdded'

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

/** 从链上事件解析记账行（仅 EntryAdded）；可选按 tableId 过滤 */
export async function buildEntryRows(events, schema, groupKey = '', tableId = null) {
  const fields = schema?.fields || []
  const rows = []
  for (const ev of events || []) {
    if (ev.type !== ENTRY_EVENT_TYPE) continue
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
  return [
    { key: '_seq', label: '序号', fixed: true },
    ...fields.map((f) => ({ key: f.key, label: cellLabel(schema, f.key), field: f })),
  ]
}
