import { computed, inject, provide, ref, unref, watch } from 'vue'
import { api } from '../api/http'
import { loadLocalGroupKey } from '../utils/e2eCrypto'
import { refreshLedgerFromServer } from '../utils/ledgerRefresh'
import { resolveSchema } from '../utils/entrySchema'
import {
  isProfessionalBookkeeping,
  isSimpleBookkeeping,
  resolveBookkeepingMode,
} from '../utils/bookkeepingMode'

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

/**
 * @param {import('vue').Ref<string> | string} ledgerIdSource route param id
 */
export function provideLedgerDetail(ledgerIdSource) {
  const ledgerId = computed(() => unref(ledgerIdSource))

  const ledger = ref(null)
  const events = ref([])
  const pending = ref([])
  const groupKey = ref('')
  const loading = ref(false)
  const contentLoading = ref(false)
  const error = ref('')
  const msg = ref('')

  const schema = computed(() => resolveSchema(ledger.value))
  const bookkeepingMode = computed(() => resolveBookkeepingMode(ledger.value))
  const isSimpleLedger = computed(() => isSimpleBookkeeping(ledger.value))
  const isProfessionalLedger = computed(() => isProfessionalBookkeeping(ledger.value))
  const memberOptions = computed(() =>
    (ledger.value?.members || []).map((m) => ({ id: m.id, username: m.id }))
  )

  function patchLedger(partial) {
    if (!partial) return
    ledger.value = ledger.value ? { ...ledger.value, ...partial } : partial
  }

  async function load() {
    const id = ledgerId.value
    if (!id) return
    contentLoading.value = true
    loading.value = true
    error.value = ''
    try {
      const refreshed = await refreshLedgerFromServer(api, id)
      ledger.value = refreshed.ledger
      events.value = refreshed.events
    } catch (e) {
      error.value = e?.message || '加载账本失败'
    } finally {
      contentLoading.value = false
      loading.value = false
    }
    if (isSimpleBookkeeping(ledger.value) && ledger.value?.approvalPolicy?.enabled) {
      try {
        const res = await api.listPending(id)
        pending.value = res.pending || []
      } catch {
        pending.value = []
      }
    } else {
      pending.value = []
    }
    const saved = loadLocalGroupKey(id)
    if (saved) groupKey.value = saved
  }

  /** Merge API ledger payload then reload events / meta from server. */
  async function applyLedgerUpdate(updated) {
    patchLedger(updated)
    await load()
  }

  watch(
    ledgerId,
    (id, prev) => {
      if (!id || id === prev) return
      ledger.value = null
      events.value = []
      pending.value = []
      error.value = ''
      msg.value = ''
      load()
    },
    { immediate: true }
  )

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
    bookkeepingMode,
    isSimpleLedger,
    isProfessionalLedger,
    memberOptions,
    load,
    patchLedger,
    applyLedgerUpdate,
  }
  provide(LEDGER_DETAIL_KEY, ctx)
  return ctx
}

export function useLedgerDetail() {
  const ctx = inject(LEDGER_DETAIL_KEY)
  if (!ctx) throw new Error('useLedgerDetail must be used inside LedgerDetailLayout')
  return ctx
}
