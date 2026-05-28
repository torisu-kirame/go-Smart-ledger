import { DEFAULT_ENTRY_SCHEMA, resolveSchema } from './entrySchema'

export const DEFAULT_TABLE_ID = 'default'

export function ledgerTables(ledger) {
  if (ledger?.tables?.length) return ledger.tables
  return [
    {
      id: DEFAULT_TABLE_ID,
      name: '默认',
      entrySchema: resolveSchema(ledger),
      sortOrder: 0,
    },
  ]
}

export function isMultiTableLedger(ledger) {
  return !!ledger?.multiTableEnabled && (ledger.tables?.length || 0) > 0
}

export function tableById(ledger, tableId) {
  const id = tableId || DEFAULT_TABLE_ID
  return ledgerTables(ledger).find((t) => t.id === id) || ledgerTables(ledger)[0]
}

export function resolveSchemaForTable(ledger, tableId) {
  const t = tableById(ledger, tableId)
  return t?.entrySchema?.fields?.length ? t.entrySchema : resolveSchema(ledger)
}

export function entryTableIdFromPayload(payload) {
  if (!payload) return DEFAULT_TABLE_ID
  const p = typeof payload === 'string' ? JSON.parse(payload) : payload
  return p?.tableId || DEFAULT_TABLE_ID
}
