/** @typedef {'simple' | 'professional'} BookkeepingMode */

export const BOOKKEEPING_SIMPLE = 'simple'
export const BOOKKEEPING_PROFESSIONAL = 'professional'

export function resolveBookkeepingMode(ledger) {
  const m = ledger?.bookkeepingMode
  if (m === BOOKKEEPING_PROFESSIONAL) return BOOKKEEPING_PROFESSIONAL
  if (ledger?.entrySchema?.templateId === 'professional') return BOOKKEEPING_PROFESSIONAL
  return BOOKKEEPING_SIMPLE
}

export function isSimpleBookkeeping(ledger) {
  return resolveBookkeepingMode(ledger) === BOOKKEEPING_SIMPLE
}

export function isProfessionalBookkeeping(ledger) {
  return resolveBookkeepingMode(ledger) === BOOKKEEPING_PROFESSIONAL
}

export function bookkeepingModeLabel(mode) {
  return mode === BOOKKEEPING_PROFESSIONAL ? '专业复式' : '简单流水'
}

export function ledgerBookkeepingPath(ledger) {
  const id = ledger?.id
  if (!id) return '/ledgers'
  return isProfessionalBookkeeping(ledger)
    ? `/ledgers/${id}/accounting/view`
    : `/ledgers/${id}/view`
}
