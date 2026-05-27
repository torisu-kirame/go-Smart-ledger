/** Prepare and validate journal lines before posting. */

function parseAmount(s) {
  const t = String(s ?? '')
    .trim()
    .replace(/,/g, '')
  if (!t) return 0
  const n = Number(t)
  if (!Number.isFinite(n) || n < 0) return NaN
  return Math.round(n * 100) / 100
}

/**
 * @returns {{ ok: true, lines: Array<{accountCode:string,debit:string,credit:string}> } | { ok: false, message: string }}
 */
export function validateJournalPayload({ date, lines }) {
  if (!date || String(date).trim().length < 10) {
    return { ok: false, message: '请选择凭证日期' }
  }

  const prepared = []
  for (const ln of lines || []) {
    const accountCode = String(ln.accountCode ?? '').trim()
    if (!accountCode) continue

    const debitRaw = String(ln.debit ?? '').trim()
    const creditRaw = String(ln.credit ?? '').trim()
    const debit = parseAmount(debitRaw)
    const credit = parseAmount(creditRaw)

    if (Number.isNaN(debit) || Number.isNaN(credit)) {
      return { ok: false, message: '金额格式无效，请填写非负数字' }
    }
    if (debit > 0 && credit > 0) {
      return {
        ok: false,
        message: '每一行只能填借方或贷方之一，不能同时填写',
      }
    }
    if (debit <= 0 && credit <= 0) {
      continue
    }
    prepared.push({
      accountCode,
      debit: debit > 0 ? String(debitRaw) : '',
      credit: credit > 0 ? String(creditRaw) : '',
    })
  }

  if (prepared.length < 2) {
    return {
      ok: false,
      message: '至少需要 2 行有效分录（科目 + 借方或贷方金额）',
    }
  }

  let debitSum = 0
  let creditSum = 0
  for (const ln of prepared) {
    debitSum += parseAmount(ln.debit)
    creditSum += parseAmount(ln.credit)
  }
  if (debitSum <= 0 || creditSum <= 0) {
    return {
      ok: false,
      message: '请同时填写借方合计与贷方合计，且均大于 0',
    }
  }
  if (Math.abs(debitSum - creditSum) > 0.001) {
    return {
      ok: false,
      message: `借贷不平衡：借方 ${debitSum.toFixed(2)}，贷方 ${creditSum.toFixed(2)}，两者必须相等`,
    }
  }

  return { ok: true, lines: prepared, date: String(date).trim() }
}
