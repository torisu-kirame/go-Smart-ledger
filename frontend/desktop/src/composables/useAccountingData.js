import { ref } from 'vue'
import { api, ApiError } from '../api/http'
import { defaultChartPayload } from '../utils/chartOfAccounts'
import { useNotify } from './useNotify'

/** Shared accounting state for ledger finance sub-routes. */
const chart = ref({ accounts: [] })
const journals = ref([])
const periods = ref([])
const attachments = ref([])
const bankStatements = ref([])
const reports = ref(null)
const busy = ref(false)
let loadToken = 0

export function useAccountingData() {
  const notify = useNotify()

  async function loadAll(ledgerId, isProfessional) {
    const id = typeof ledgerId === 'string' ? ledgerId : ledgerId?.value
    if (!id || !isProfessional) return
    const token = ++loadToken
    busy.value = true
    try {
      chart.value = await api.getAccountingChart(id)
      const jr = await api.listAccountingJournals(id)
      journals.value = jr.journals || []
      const pr = await api.listAccountingPeriods(id)
      periods.value = pr.periods || []
      const at = await api.listAccountingAttachments(id)
      attachments.value = at.attachments || []
      const bs = await api.listBankStatements(id)
      bankStatements.value = bs.statements || []
    } catch (e) {
      if (token === loadToken) {
        notify.error(e instanceof ApiError ? e.message : '加载失败')
      }
    } finally {
      if (token === loadToken) busy.value = false
    }
  }

  async function saveChart(ledgerId) {
    const id = typeof ledgerId === 'string' ? ledgerId : ledgerId?.value
    busy.value = true
    try {
      chart.value = await api.putAccountingChart(id, chart.value)
      notify.success('科目表已保存')
    } catch (e) {
      notify.error(e instanceof ApiError ? e.message : '保存失败')
      throw e
    } finally {
      busy.value = false
    }
  }

  async function resetChart(ledgerId) {
    const id = typeof ledgerId === 'string' ? ledgerId : ledgerId?.value
    busy.value = true
    try {
      const def = defaultChartPayload(id)
      chart.value = await api.putAccountingChart(id, def)
      notify.success('已恢复默认科目表')
    } catch (e) {
      notify.error(e instanceof ApiError ? e.message : '操作失败')
    } finally {
      busy.value = false
    }
  }

  return {
    chart,
    journals,
    periods,
    attachments,
    bankStatements,
    reports,
    busy,
    loadAll,
    saveChart,
    resetChart,
  }
}
