import { computed, inject, provide, ref } from 'vue'
import { api } from '../api/http'
import { loadLocalGroupKey } from '../utils/e2eCrypto'
import { refreshLedgerFromServer } from '../utils/ledgerRefresh'
import { resolveSchema } from '../utils/entrySchema'

const LEDGER_DETAIL_KEY = Symbol('ledgerDetail')
const INFO_UNLOCK_PREFIX = 'sl-info-unlock:'

export function getInfoUnlockSession(ledgerId) {
  if (typeof sessionStorage === 'undefined') return false
  return sessionStorage.getItem(INFO_UNLOCK_PREFIX + ledgerId) === '1'
}

export function setInfoUnlockSession(ledgerId) {
  if (typeof sessionStorage === 'undefined') return
  sessionStorage.setItem(INFO_UNLOCK_PREFIX + ledgerId, '1')
}

export function clearInfoUnlockSession(ledgerId) {
  if (typeof sessionStorage === 'undefined') return
  sessionStorage.removeItem(INFO_UNLOCK_PREFIX + ledgerId)
}

export function provideLedgerDetail(ledgerId) {
  const ledger = ref(null)
  const events = ref([])
  const pending = ref([])
  const groupKey = ref('')
  const loading = ref(false)
  const contentLoading = ref(false)
  const error = ref('')
  const msg = ref('')

  const schema = computed(() => resolveSchema(ledger.value))
  const memberOptions = computed(() =>
    (ledger.value?.members || []).map((m) => ({ id: m.id, username: m.id }))
  )

  async function load() {
    contentLoading.value = true
    loading.value = true
    try {
      const refreshed = await refreshLedgerFromServer(api, ledgerId)
      ledger.value = refreshed.ledger
      events.value = refreshed.events
    } finally {
      contentLoading.value = false
      loading.value = false
    }
    if (ledger.value?.approvalPolicy?.enabled) {
      const res = await api.listPending(ledgerId)
      pending.value = res.pending || []
    } else {
      pending.value = []
    }
    const saved = loadLocalGroupKey(ledgerId)
    if (saved) groupKey.value = saved
  }

  const ctx = {
    ledgerId,
    ledger,
    events,
    pending,
    groupKey,
    loading,
    contentLoading,
    error,
    msg,
    schema,
    memberOptions,
    load,
  }
  provide(LEDGER_DETAIL_KEY, ctx)
  return ctx
}

export function useLedgerDetail() {
  const ctx = inject(LEDGER_DETAIL_KEY)
  if (!ctx) throw new Error('useLedgerDetail must be used inside LedgerDetailLayout')
  return ctx
}
