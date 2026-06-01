/** 内置记账模板（与后端 default 一致） */
export const DEFAULT_TEMPLATE_ID = 'default'

export const FIELD_TYPE_OPTIONS = [
  { value: 'text', label: '文本' },
  { value: 'number', label: '数字' },
  { value: 'date', label: '日期' },
  { value: 'user', label: '用户' },
]

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

export function resolveSchema(ledger, templates) {
  return resolveSchemaWithTemplates(ledger, templates)
}

/** 按 templateId 从账户通用模板库解析最新字段（跨账本同步） */
export function resolveSchemaWithTemplates(ledger, templates = []) {
  const ledgerSchema = ledger?.entrySchema
  const tid = ledgerSchema?.templateId
  if (tid && tid !== 'custom') {
    const global = templates.find((t) => t.templateId === tid)
    if (global?.fields?.length) {
      return {
        templateId: tid,
        fields: global.fields.map((f) => ({ ...f })),
      }
    }
  }
  if (ledgerSchema?.fields?.length) {
    return ledgerSchema
  }
  const builtin = templates.find((t) => t.templateId === DEFAULT_TEMPLATE_ID)
  if (builtin?.fields?.length) {
    return { templateId: builtin.templateId, fields: builtin.fields.map((f) => ({ ...f })) }
  }
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
