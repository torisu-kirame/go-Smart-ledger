/** Mirrors backend accounting.DefaultChart accounts (without ledger id). */
export const DEFAULT_ACCOUNTS = [
  { code: '1001', name: '库存现金', category: 'asset', active: true },
  { code: '1002', name: '银行存款', category: 'asset', active: true },
  { code: '1122', name: '应收账款', category: 'asset', active: true },
  { code: '2202', name: '应付账款', category: 'liability', active: true },
  { code: '4001', name: '实收资本', category: 'equity', active: true },
  { code: '6001', name: '主营业务收入', category: 'revenue', active: true },
  { code: '6401', name: '主营业务成本', category: 'expense', active: true },
  { code: '6601', name: '销售费用', category: 'expense', active: true },
  { code: '6602', name: '管理费用', category: 'expense', active: true },
]

export const ACCOUNT_CATEGORIES = [
  { value: 'asset', label: '资产' },
  { value: 'liability', label: '负债' },
  { value: 'equity', label: '权益' },
  { value: 'revenue', label: '收入' },
  { value: 'expense', label: '费用' },
]

export function categoryLabel(c) {
  const m = Object.fromEntries(ACCOUNT_CATEGORIES.map((x) => [x.value, x.label]))
  return m[c] || c
}

export function defaultChartPayload(ledgerId) {
  return {
    ledgerId,
    accounts: DEFAULT_ACCOUNTS.map((a) => ({ ...a })),
  }
}

export function validateChartPayload(chart) {
  const accounts = chart?.accounts || []
  if (!accounts.length) return { ok: false, message: '至少需要一个科目' }
  const seen = new Set()
  for (const a of accounts) {
    const code = String(a.code || '').trim()
    const name = String(a.name || '').trim()
    if (!code || seen.has(code)) return { ok: false, message: '科目编码不能为空且不能重复' }
    if (!name) return { ok: false, message: `科目 ${code} 缺少名称` }
    if (!ACCOUNT_CATEGORIES.some((c) => c.value === a.category)) {
      return { ok: false, message: `科目 ${code} 类别无效` }
    }
    seen.add(code)
  }
  return { ok: true }
}
