import { inject, provide } from 'vue'
import { useAccountingData } from './useAccountingData'

const ACCOUNTING_KEY = Symbol('ledgerAccounting')

export function provideLedgerAccounting(ledgerId, isProfessional) {
  const ctx = useAccountingData()
  provide(ACCOUNTING_KEY, { ...ctx, ledgerId, isProfessional })
  return ctx
}

export function useLedgerAccounting() {
  const ctx = inject(ACCOUNTING_KEY)
  if (!ctx) throw new Error('useLedgerAccounting must be used inside LedgerAccountingLayout')
  return ctx
}
