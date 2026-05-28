<template>
  <div class="acct-ops">
    <section v-if="section === 'period'" class="detail-card">
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

    <section v-else-if="section === 'report'" class="detail-card">
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

    <section v-else-if="section === 'attach'" class="detail-card">
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

    <section v-else-if="section === 'bank'" class="detail-card">
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
import { computed, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, ApiError } from '../../api/http'
import FileUploadZone from '../../components/FileUploadZone.vue'
import { useLedgerAccounting } from '../../composables/provideAccounting'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { useNotify } from '../../composables/useNotify'

const route = useRoute()
const { ledgerId } = useLedgerDetail()
const notify = useNotify()
const {
  periods,
  attachments,
  bankStatements,
  reports,
  busy,
  loadAll,
} = useLedgerAccounting()

const section = computed(() => route.path.split('/').pop())
const reportPeriod = ref('')
const attachSeq = ref(1)
const matchDraft = reactive({})

function lid() {
  return ledgerId.value
}

async function closePeriod(period) {
  busy.value = true
  try {
    await api.closeAccountingPeriod(lid(), period)
    notify.success(`${period} 已结账锁定`)
    await loadAll(ledgerId, true)
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '结账失败')
  } finally {
    busy.value = false
  }
}

async function reopenPeriod(period) {
  busy.value = true
  try {
    await api.reopenAccountingPeriod(lid(), period)
    notify.success(`${period} 已重新开放`)
    await loadAll(ledgerId, true)
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '操作失败')
  } finally {
    busy.value = false
  }
}

async function loadReports() {
  busy.value = true
  try {
    const p = reportPeriod.value ? reportPeriod.value.replace('/', '-') : ''
    reports.value = await api.getAccountingReports(lid(), p)
    notify.success('报表已生成')
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '报表生成失败')
  } finally {
    busy.value = false
  }
}

async function onAttachFile(file) {
  if (!file || !attachSeq.value) return
  busy.value = true
  try {
    await api.uploadAccountingAttachment(lid(), attachSeq.value, file, '')
    notify.success('附件已上传')
    await loadAll(ledgerId, true)
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '上传失败')
  } finally {
    busy.value = false
  }
}

async function onBankImport(file) {
  if (!file) return
  busy.value = true
  try {
    await api.importBankStatement(lid(), file)
    notify.success('对账单已导入')
    await loadAll(ledgerId, true)
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '导入失败')
  } finally {
    busy.value = false
  }
}

async function matchLine(stmtId, lineId) {
  const seq = matchDraft[lineId]
  if (!seq) return
  busy.value = true
  try {
    await api.matchBankLine(lid(), stmtId, lineId, seq)
    notify.success('已匹配')
    await loadAll(ledgerId, true)
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '匹配失败')
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.detail-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1rem 1.1rem;
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
.bank-stmt h4 {
  margin: 1rem 0 0.35rem;
  font-size: 0.875rem;
}
</style>
