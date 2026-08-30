import { resolveSchemaWithTemplates } from './entrySchema'

export const DEFAULT_TABLE_ID = 'default'

export function ledgerTables(ledger, templates = []) {
  if (ledger?.multiTableEnabled) {
    return Array.isArray(ledger.tables) ? ledger.tables : []
  }
  if (ledger?.tables?.length) return ledger.tables
  return [
    {
      id: DEFAULT_TABLE_ID,
      name: '默认',
      entrySchema: resolveSchemaWithTemplates(ledger, templates),
      sortOrder: 0,
    },
  ]
}

export function isMultiTableLedger(ledger) {
  return !!ledger?.multiTableEnabled
}

export function tableById(ledger, tableId) {
  const id = tableId || DEFAULT_TABLE_ID
  return ledgerTables(ledger).find((t) => t.id === id) || ledgerTables(ledger)[0]
}

export function resolveSchemaForTable(ledger, tableId, templates) {
  const list = ledgerTables(ledger, templates)
  if (ledger?.multiTableEnabled && !list.length) {
    return { templateId: 'custom', fields: [] }
  }
  const t = tableById(ledger, tableId)
  const tableSchema = t?.entrySchema
  if (tableSchema?.templateId === 'custom') {
    return {
      templateId: 'custom',
      fields: Array.isArray(tableSchema.fields) ? tableSchema.fields.map((f) => ({ ...f })) : [],
    }
  }
  if (tableSchema?.templateId && tableSchema.templateId !== 'custom') {
    return resolveSchemaWithTemplates({ entrySchema: tableSchema }, templates)
  }
  if (tableSchema?.fields?.length) return tableSchema
  return resolveSchemaWithTemplates(ledger, templates)
}

export function entryTableIdFromPayload(payload) {
  if (!payload) return DEFAULT_TABLE_ID
  const p = typeof payload === 'string' ? JSON.parse(payload) : payload
  return p?.tableId || DEFAULT_TABLE_ID
}
