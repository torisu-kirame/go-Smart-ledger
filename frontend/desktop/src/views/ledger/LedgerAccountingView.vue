<template>
  <div class="ledger-accounting">
    <div v-if="msg" class="alert alert-success">{{ msg }}</div>
    <div v-if="error" class="alert alert-error">{{ error }}</div>

    <nav class="acct-tabs" aria-label="财务功能">
      <button
        v-for="t in tabs"
        :key="t.id"
        type="button"
        class="acct-tab"
        :class="{ active: tab === t.id }"
        @click="tab = t.id"
      >
        {{ t.label }}
      </button>
    </nav>

    <!-- F38 科目与凭证 -->
    <section v-show="tab === 'coa'" class="detail-card">
      <h3 class="detail-card__title">会计科目表</h3>
      <div class="table-wrap">
        <table>
          <thead>
            <tr><th>编码</th><th>名称</th><th>类别</th><th>状态</th></tr>
          </thead>
          <tbody>
            <tr v-for="a in chart.accounts" :key="a.code">
              <td class="mono">{{ a.code }}</td>
              <td>{{ a.name }}</td>
              <td>{{ categoryLabel(a.category) }}</td>
              <td>{{ a.active ? '启用' : '停用' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <button type="button" class="btn-ghost" style="margin-top: 0.75rem" :disabled="busy" @click="resetChart">
        恢复默认科目表
      </button>
    </section>

    <section v-show="tab === 'journal'" class="detail-card">
      <h3 class="detail-card__title">记账凭证（复式）</h3>
      <form class="journal-form" @submit.prevent="postJournal">
        <div class="form-row">
          <label>凭证日期</label>
          <input v-model="journalForm.date" type="date" required class="field-sm" />
        </div>
        <div class="form-row">
          <label>摘要</label>
          <input v-model="journalForm.description" class="field-sm" placeholder="可选" />
        </div>
        <div v-for="(ln, i) in journalForm.lines" :key="i" class="journal-line">
          <AppSelect
            v-model="ln.accountCode"
            :options="accountOptions"
            sm
            placeholder="科目"
          />
          <input v-model="ln.debit" placeholder="借方" class="field-sm amount" />
          <input v-model="ln.credit" placeholder="贷方" class="field-sm amount" />
          <DeleteButton icon-only sm title="删除行" @click="journalForm.lines.splice(i, 1)" />
        </div>
        <button type="button" class="btn-ghost" @click="addJournalLine">+ 分录行</button>
        <button type="submit" class="btn-primary" :disabled="busy">过账</button>
      </form>
      <h4 class="sub-title">已过账凭证</h4>
      <div v-if="!journals.length" class="muted">暂无凭证</div>
      <div v-for="j in journals" :key="j.id" class="journal-card">
        <div class="journal-card__head">
          <span class="mono">#{{ j.eventSeq || '—' }}</span>
          <span>{{ j.date }}</span>
          <span class="muted">{{ j.description }}</span>
        </div>
        <table class="mini-table">
          <tr v-for="(ln, idx) in j.lines" :key="idx">
            <td class="mono">{{ ln.accountCode }}</td>
            <td>{{ ln.debit || '—' }}</td>
            <td>{{ ln.credit || '—' }}</td>
          </tr>
        </table>
      </div>
    </section>

    <!-- F39 期间 -->
    <section v-show="tab === 'period'" class="detail-card">
      <h3 class="detail-card__title">会计期间</h3>
      <div v-if="!periods.length" class="muted">过账后将自动出现对应月份</div>
      <div v-for="p in periods" :key="p.period" class="period-row">
        <span class="mono">{{ p.period }}</span>
        <span :class="['badge', p.status === 'closed' ? 'badge-pending' : 'badge-ok']">
          {{ p.status === 'closed' ? '已结账' : '开放' }}
        </span>
        <div class="actions-row">
          <button
            v-if="p.status !== 'closed'"
            type="button"
            class="btn-primary"
            :disabled="busy"
            @click="closePeriod(p.period)"
          >
            月结锁定
          </button>
          <button
            v-else
            type="button"
            class="btn-ghost"
            :disabled="busy"
            @click="reopenPeriod(p.period)"
          >
            反结账
          </button>
        </div>
      </div>
    </section>

    <!-- F40 报表 -->
    <section v-show="tab === 'report'" class="detail-card">
      <h3 class="detail-card__title">财务报表</h3>
      <div class="form-row inline">
        <label>期间</label>
        <input v-model="reportPeriod" type="month" class="field-sm" />
        <button type="button" class="btn-primary" :disabled="busy" @click="loadReports">生成</button>
      </div>
      <template v-if="reports">
        <h4 class="sub-title">试算平衡</h4>
        <p class="muted">
          {{ reports.trialBalance?.balanced ? '借贷平衡 ✓' : '借贷不平衡' }}
        </p>
        <h4 class="sub-title">资产负债表（汇总）</h4>
        <ul class="report-list">
          <li v-for="ln in reports.balanceSheet?.lines || []" :key="ln.code">
            {{ ln.code }} {{ ln.name }}：<strong>{{ ln.amount }}</strong>
          </li>
        </ul>
        <h4 class="sub-title">利润表（汇总）</h4>
        <p>本期合计：<strong>{{ reports.incomeStatement?.total }}</strong></p>
        <h4 class="sub-title">现金流量（简化）</h4>
        <p>净变动：<strong>{{ reports.cashFlow?.netChange }}</strong></p>
      </template>
    </section>

    <!-- F44 附件 -->
    <section v-show="tab === 'attach'" class="detail-card">
      <h3 class="detail-card__title">凭证附件</h3>
      <div class="form-row">
        <label>事件 Seq</label>
        <input v-model.number="attachSeq" type="number" min="1" class="field-sm" />
      </div>
      <FileUploadZone
        block
        :disabled="busy || !attachSeq"
        title="点击或拖拽上传凭证附件"
        :hint="attachSeq ? '' : '请先填写上方事件 Seq'"
        @file="onAttachFile"
      />
      <div v-if="!attachments.length" class="muted">暂无附件</div>
      <ul class="attach-list">
        <li v-for="a in attachments" :key="a.id">
          <span>#{{ a.entrySeq }}</span>
          <span>{{ a.filename }}</span>
          <span v-if="a.cid" class="mono muted">CID {{ a.cid.slice(0, 12) }}…</span>
        </li>
      </ul>
    </section>

    <!-- F42 对账 -->
    <section v-show="tab === 'bank'" class="detail-card">
      <h3 class="detail-card__title">银行对账</h3>
      <FileUploadZone
        block
        accept=".csv,text/csv"
        :disabled="busy"
        title="点击或拖拽导入银行 CSV"
        hint="CSV 需含日期、金额列；可选摘要列"
        @file="onBankImport"
      />
      <div v-for="stmt in bankStatements" :key="stmt.id" class="bank-stmt">
        <h4>{{ stmt.filename || stmt.id }} · {{ stmt.lines?.length || 0 }} 笔</h4>
        <div v-for="ln in stmt.lines" :key="ln.id" class="bank-line">
          <span>{{ ln.date }}</span>
          <span>{{ ln.description }}</span>
          <span class="mono">{{ ln.amount }}</span>
          <template v-if="ln.matchedSeq">
            <span class="badge badge-ok">已匹配 #{{ ln.matchedSeq }}</span>
          </template>
          <template v-else>
            <input
              v-model.number="matchDraft[ln.id]"
              type="number"
              min="1"
              placeholder="Seq"
              class="field-sm match-input"
            />
            <button type="button" class="btn-ghost" @click="matchLine(stmt.id, ln.id)">匹配</button>
          </template>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { api, ApiError } from '../../api/http'
import AppSelect from '../../components/AppSelect.vue'
import DeleteButton from '../../components/DeleteButton.vue'
import FileUploadZone from '../../components/FileUploadZone.vue'
import { useLedgerDetail } from '../../composables/useLedgerDetail'

const { ledgerId, error, msg } = useLedgerDetail()

const tabs = [
  { id: 'coa', label: '科目' },
  { id: 'journal', label: '凭证' },
  { id: 'period', label: '期间' },
  { id: 'report', label: '报表' },
  { id: 'attach', label: '附件' },
  { id: 'bank', label: '对账' },
]

const tab = ref('coa')
const busy = ref(false)
const chart = ref({ accounts: [] })
const journals = ref([])
const periods = ref([])
const reports = ref(null)
const reportPeriod = ref('')
const attachments = ref([])
const attachSeq = ref(1)
const bankStatements = ref([])
const matchDraft = reactive({})

const journalForm = reactive({
  date: new Date().toISOString().slice(0, 10),
  description: '',
  lines: [
    { accountCode: '1002', debit: '', credit: '' },
    { accountCode: '6001', debit: '', credit: '' },
  ],
})

const accountOptions = computed(() =>
  (chart.value.accounts || [])
    .filter((a) => a.active)
    .map((a) => ({ value: a.code, label: `${a.code} ${a.name}` }))
)

function categoryLabel(c) {
  const m = {
    asset: '资产',
    liability: '负债',
    equity: '权益',
    revenue: '收入',
    expense: '费用',
  }
  return m[c] || c
}

function addJournalLine() {
  journalForm.lines.push({ accountCode: '', debit: '', credit: '' })
}

async function loadAll() {
  busy.value = true
  error.value = ''
  try {
    chart.value = await api.getAccountingChart(ledgerId)
    const jr = await api.listAccountingJournals(ledgerId)
    journals.value = jr.journals || []
    const pr = await api.listAccountingPeriods(ledgerId)
    periods.value = pr.periods || []
    const at = await api.listAccountingAttachments(ledgerId)
    attachments.value = at.attachments || []
    const bs = await api.listBankStatements(ledgerId)
    bankStatements.value = bs.statements || []
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '加载失败'
  } finally {
    busy.value = false
  }
}

async function resetChart() {
  busy.value = true
  try {
    const def = await api.getAccountingChart(ledgerId)
    await api.putAccountingChart(ledgerId, def)
    msg.value = '已恢复默认科目表'
    await loadAll()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '操作失败'
  } finally {
    busy.value = false
  }
}

async function postJournal() {
  busy.value = true
  try {
    await api.postAccountingJournal(ledgerId, {
      date: journalForm.date,
      description: journalForm.description,
      lines: journalForm.lines.filter((l) => l.accountCode),
    })
    msg.value = '凭证已过账'
    await loadAll()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '过账失败'
  } finally {
    busy.value = false
  }
}

async function closePeriod(period) {
  busy.value = true
  try {
    await api.closeAccountingPeriod(ledgerId, period)
    msg.value = `${period} 已结账锁定`
    await loadAll()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '结账失败'
  } finally {
    busy.value = false
  }
}

async function reopenPeriod(period) {
  busy.value = true
  try {
    await api.reopenAccountingPeriod(ledgerId, period)
    msg.value = `${period} 已重新开放`
    await loadAll()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '操作失败'
  } finally {
    busy.value = false
  }
}

async function loadReports() {
  busy.value = true
  try {
    const p = reportPeriod.value ? reportPeriod.value.replace('/', '-') : ''
    reports.value = await api.getAccountingReports(ledgerId, p)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '报表生成失败'
  } finally {
    busy.value = false
  }
}

async function onAttachFile(file) {
  if (!file || !attachSeq.value) return
  busy.value = true
  try {
    await api.uploadAccountingAttachment(ledgerId, attachSeq.value, file)
    msg.value = '附件已上传'
    await loadAll()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '上传失败'
  } finally {
    busy.value = false
  }
}

async function onBankImport(file) {
  if (!file) return
  busy.value = true
  try {
    await api.importBankStatement(ledgerId, file)
    msg.value = '对账单已导入'
    await loadAll()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '导入失败'
  } finally {
    busy.value = false
  }
}

async function matchLine(stmtId, lineId) {
  const seq = matchDraft[lineId]
  if (!seq) return
  busy.value = true
  try {
    await api.matchBankLine(ledgerId, stmtId, lineId, seq)
    msg.value = '已匹配'
    await loadAll()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '匹配失败'
  } finally {
    busy.value = false
  }
}

onMounted(loadAll)
</script>

<style scoped>
.acct-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-bottom: 1rem;
}
.acct-tab {
  padding: 0.4rem 0.85rem;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  color: var(--text-muted);
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
}
.acct-tab.active {
  color: var(--accent);
  border-color: color-mix(in srgb, var(--accent) 50%, var(--border));
  background: var(--accent-soft);
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
.sub-title {
  margin: 1rem 0 0.5rem;
  font-size: 0.875rem;
}
.journal-line {
  display: grid;
  grid-template-columns: 1fr 5rem 5rem auto;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  align-items: center;
}
.journal-card {
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 0.65rem;
  margin-top: 0.5rem;
}
.journal-card__head {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  font-size: 0.8125rem;
  margin-bottom: 0.35rem;
}
.mini-table {
  width: 100%;
  font-size: 0.8125rem;
}
.period-row,
.bank-line {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border);
}
.form-row.inline {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.report-list {
  margin: 0;
  padding-left: 1.1rem;
  font-size: 0.875rem;
}
.attach-list {
  margin: 0.5rem 0 0;
  padding: 0;
  list-style: none;
  font-size: 0.875rem;
}
.attach-list li {
  display: flex;
  gap: 0.75rem;
  padding: 0.35rem 0;
}
.match-input {
  width: 5rem;
}
.field-hint {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: 0.5rem 0;
}
</style>
