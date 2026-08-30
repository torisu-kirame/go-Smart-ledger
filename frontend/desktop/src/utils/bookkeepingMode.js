/** Bookkeeping is always simple sheet/entry mode. */

export const BOOKKEEPING_SIMPLE = 'simple'

export function resolveBookkeepingMode() {
  return BOOKKEEPING_SIMPLE
}

export function isSimpleBookkeeping() {
  return true
}

export function isProfessionalBookkeeping() {
  return false
}

export function bookkeepingModeLabel() {
  return '简单流水'
}

export function ledgerBookkeepingPath(ledger) {
  const id = ledger?.id
  if (!id) return '/ledgers'
  return `/ledgers/${id}/view`
}
