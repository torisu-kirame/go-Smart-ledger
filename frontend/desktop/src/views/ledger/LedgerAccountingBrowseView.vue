<template>
  <div class="acct-browse">
    <section class="detail-card">
      <h3 class="detail-card__title">已过账凭证</h3>
      <p class="field-hint">
        复式记账要求：至少 2 行分录，每行只填借方或贷方，且借方合计 = 贷方合计。
      </p>
      <div v-if="!journals.length" class="muted">暂无凭证</div>
      <div v-for="j in journals" :key="j.id" class="journal-card">
        <div class="journal-card__head">
          <span class="mono">#{{ j.eventSeq || '—' }}</span>
          <span>{{ j.date }}</span>
          <span class="muted">{{ j.description || '—' }}</span>
        </div>
        <div class="table-wrap">
          <table class="journal-table">
            <thead>
              <tr>
                <th>科目</th>
                <th>借方</th>
                <th>贷方</th>
                <th>往来方</th>
                <th>项目</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(ln, idx) in j.lines" :key="idx">
                <td class="mono">{{ accountName(ln.accountCode) }}</td>
                <td>{{ ln.debit || '—' }}</td>
                <td>{{ ln.credit || '—' }}</td>
                <td>{{ ln.counterparty || '—' }}</td>
                <td>{{ ln.project || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>

    <button
      type="button"
      class="fab-entry btn-primary"
      :disabled="busy || !chart.accounts?.length"
      title="录入复式凭证"
      @click="showJournalModal = true"
    >
      <AppIcon name="plus" size="sm" />
      <span>记一笔</span>
    </button>

    <div v-if="showJournalModal" class="modal" @click.self="closeModal">
      <form class="modal-card entry-modal wide" @submit.prevent="submitJournal">
        <h3>记一笔（复式凭证）</h3>
        <div class="form-row">
          <label>凭证日期</label>
          <input v-model="journalForm.date" type="date" required class="field-sm" />
        </div>
        <div class="form-row">
          <label>摘要</label>
          <input v-model="journalForm.description" class="field-sm" placeholder="可选" />
        </div>
        <div v-for="(ln, i) in journalForm.lines" :key="i" class="journal-line">
          <AppSelect v-model="ln.accountCode" :options="accountOptions" sm placeholder="科目" />
          <input v-model="ln.debit" placeholder="借方" class="field-sm amount" />
          <input v-model="ln.credit" placeholder="贷方" class="field-sm amount" />
          <input v-model="ln.counterparty" placeholder="往来方" class="field-sm" />
          <input v-model="ln.project" placeholder="项目" class="field-sm" />
          <DeleteButton icon-only sm title="删除行" @click="journalForm.lines.splice(i, 1)" />
        </div>
        <button type="button" class="btn-ghost" @click="addJournalLine">+ 分录行</button>
        <div class="modal-actions">
          <button type="button" class="btn-ghost" @click="closeModal">取消</button>
          <button type="submit" class="btn-primary" :disabled="busy">
            {{ busy ? '过账中…' : '过账' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { api, ApiError } from '../../api/http'
import AppIcon from '../../components/AppIcon.vue'
import AppSelect from '../../components/AppSelect.vue'
import DeleteButton from '../../components/DeleteButton.vue'
import { useLedgerAccounting } from '../../composables/provideAccounting'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { useNotify } from '../../composables/useNotify'
import { validateJournalPayload } from '../../utils/journalEntry'

const { chart, journals, busy, loadAll } = useLedgerAccounting()
const { ledgerId, load: reloadLedger } = useLedgerDetail()
const notify = useNotify()

const showJournalModal = ref(false)
const journalForm = reactive({
  date: new Date().toISOString().slice(0, 10),
  description: '',
  lines: [
    { accountCode: '1002', debit: '', credit: '', counterparty: '', project: '' },
    { accountCode: '6001', debit: '', credit: '', counterparty: '', project: '' },
  ],
})

const accountOptions = computed(() =>
  (chart.value.accounts || [])
    .filter((a) => a.active)
    .map((a) => ({ value: a.code, label: `${a.code} ${a.name}` }))
)

const accountIndex = computed(() => {
  const m = {}
  for (const a of chart.value.accounts || []) {
    m[a.code] = a.name
  }
  return m
})

function accountName(code) {
  const n = accountIndex.value[code]
  return n ? `${code} ${n}` : code
}

function addJournalLine() {
  journalForm.lines.push({
    accountCode: '',
    debit: '',
    credit: '',
    counterparty: '',
    project: '',
  })
}

function closeModal() {
  if (busy.value) return
  showJournalModal.value = false
}

async function submitJournal() {
  const check = validateJournalPayload({
    date: journalForm.date,
    lines: journalForm.lines,
  })
  if (!check.ok) {
    notify.error(check.message)
    return
  }
  busy.value = true
  try {
    await api.postAccountingJournal(ledgerId.value, {
      date: check.date,
      description: journalForm.description,
      lines: check.lines,
    })
    notify.success('凭证已过账')
    journalForm.description = ''
    showJournalModal.value = false
    await loadAll(ledgerId, true)
    await reloadLedger()
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '过账失败')
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.acct-browse {
  position: relative;
  padding-bottom: 4.5rem;
}
.detail-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1rem 1.1rem;
  margin-bottom: 1rem;
  box-shadow: var(--shadow-sm);
}
.detail-card__title {
  margin: 0 0 0.85rem;
  font-size: 0.8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.field-hint {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: 0 0 0.75rem;
  line-height: 1.45;
}
.journal-card {
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 0.65rem;
  margin-top: 0.65rem;
}
.journal-card__head {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  font-size: 0.8125rem;
  margin-bottom: 0.45rem;
}
.journal-table {
  width: 100%;
  font-size: 0.8125rem;
  border-collapse: collapse;
}
.journal-table th,
.journal-table td {
  padding: 0.35rem 0.5rem;
  text-align: left;
  border-bottom: 1px solid var(--border);
}
.journal-table thead th {
  font-weight: 600;
  color: var(--text-muted);
  background: var(--bg-elevated);
}
.fab-entry {
  position: fixed;
  right: 2rem;
  bottom: 2rem;
  z-index: 40;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.65rem 1.15rem;
  border-radius: 999px;
  border: none;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.35);
  font-size: 0.9375rem;
  font-weight: 600;
}
.fab-entry:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.entry-modal.wide {
  width: min(100%, 36rem);
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 1rem;
}
.journal-line {
  display: grid;
  grid-template-columns: 1fr 4.5rem 4.5rem 5rem 4rem auto;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  align-items: center;
}
@media (max-width: 720px) {
  .journal-line {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
