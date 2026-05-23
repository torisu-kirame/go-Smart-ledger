/** 内置记账模板（与后端 default 一致） */
export const DEFAULT_TEMPLATE_ID = 'default'

export const DEFAULT_ENTRY_SCHEMA = {
  templateId: DEFAULT_TEMPLATE_ID,
  fields: [
    { key: 'bookkeeper', label: '记账人', type: 'user', required: true },
    { key: 'payee', label: '收账人', type: 'text', required: false },
    { key: 'amount', label: '金额', type: 'number', required: true },
    { key: 'date', label: '日期', type: 'date', required: true },
    { key: 'note', label: '备注', type: 'text', required: false },
  ],
}

export function resolveSchema(ledger) {
  if (ledger?.entrySchema?.fields?.length) return ledger.entrySchema
  return DEFAULT_ENTRY_SCHEMA
}

export function emptyEntryData(schema, defaults = {}) {
  const data = {}
  for (const f of schema.fields || []) {
    data[f.key] = defaults[f.key] ?? ''
  }
  return data
}

export function cellLabel(schema, key) {
  return schema.fields?.find((f) => f.key === key)?.label || key
}

export function cellValue(row, key) {
  if (row.cells && row.cells[key] !== undefined) return row.cells[key]
  return row[key] ?? ''
}
