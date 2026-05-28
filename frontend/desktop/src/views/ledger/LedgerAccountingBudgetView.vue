<template>
  <div class="acct-budget">
    <section class="detail-card">
      <h3 class="detail-card__title">预算编制（F43）</h3>
      <div class="form-row inline">
        <label>会计期间</label>
        <input v-model="period" type="month" class="field-sm" />
        <button type="button" class="btn-ghost" :disabled="busy" @click="loadBudget">加载</button>
        <button type="button" class="btn-primary" :disabled="busy" @click="saveBudget">保存预算</button>
        <button type="button" class="btn-ghost" :disabled="busy" @click="loadAnalysis">执行分析</button>
      </div>
      <p class="field-hint">
        按<strong>科目</strong>或<strong>项目</strong>设定期间预算；执行分析对比凭证实际发生额，超支行将高亮预警。
      </p>

      <div v-if="alerts.length" class="alert-box">
        <strong>超支预警（{{ alerts.length }}）</strong>
        <ul>
          <li v-for="a in alerts" :key="a.scope + a.scopeKey">
            {{ a.scopeLabel }}：预算 {{ a.budget }}，实际 {{ a.actual }}（{{ a.utilizationPct }}）
          </li>
        </ul>
      </div>

      <h4 class="sub-title">预算行</h4>
      <div v-for="(ln, i) in lines" :key="ln.id || i" class="budget-row">
        <select v-model="ln.scope" class="field-sm">
          <option value="account">科目</option>
          <option value="project">项目</option>
        </select>
        <AppSelect
          v-if="ln.scope === 'account'"
          v-model="ln.scopeKey"
          :options="accountOptions"
          sm
          placeholder="科目"
        />
        <input
          v-else
          v-model="ln.scopeKey"
          type="text"
          class="field-sm"
          placeholder="项目名称"
        />
        <input v-model="ln.amount" type="text" class="field-sm amount" placeholder="预算金额" />
        <DeleteButton icon-only sm title="删除" @click="lines.splice(i, 1)" />
      </div>
      <button type="button" class="btn-ghost" @click="addLine">+ 预算行</button>
    </section>

    <section v-if="analysis" class="detail-card">
      <h3 class="detail-card__title">执行分析 · {{ analysis.period }}</h3>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>范围</th>
              <th>预算</th>
              <th>实际</th>
              <th>差异</th>
              <th>执行率</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in analysis.lines"
              :key="row.scope + row.scopeKey"
              :class="{ 'row-warn': row.overBudget }"
            >
              <td>{{ row.scopeLabel || row.scopeKey }}</td>
              <td class="mono">{{ row.budget }}</td>
              <td class="mono">{{ row.actual }}</td>
              <td class="mono">{{ row.variance }}</td>
              <td>{{ row.utilizationPct || '—' }}</td>
              <td>
                <span v-if="row.overBudget" class="badge badge-pending">超支</span>
                <span v-else class="badge badge-ok">正常</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { api, ApiError } from '../../api/http'
import AppSelect from '../../components/AppSelect.vue'
import DeleteButton from '../../components/DeleteButton.vue'
import { useLedgerAccounting } from '../../composables/provideAccounting'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { useNotify } from '../../composables/useNotify'

const { chart, busy } = useLedgerAccounting()
const { ledgerId } = useLedgerDetail()
const notify = useNotify()

const period = ref(new Date().toISOString().slice(0, 7))
const lines = ref([])
const analysis = ref(null)

const accountOptions = computed(() =>
  (chart.value.accounts || [])
    .filter((a) => a.active && (a.category === 'expense' || a.category === 'revenue'))
    .map((a) => ({ value: a.code, label: `${a.code} ${a.name}` }))
)

const alerts = computed(() => analysis.value?.alerts || [])

function addLine() {
  lines.value.push({
    id: '',
    scope: 'account',
    scopeKey: accountOptions.value[0]?.value || '6602',
    amount: '',
  })
}

async function loadBudget() {
  busy.value = true
  try {
    const b = await api.getAccountingBudget(ledgerId.value, period.value)
    lines.value = (b.lines || []).map((ln) => ({ ...ln }))
    if (!lines.value.length) addLine()
    analysis.value = null
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '加载失败')
  } finally {
    busy.value = false
  }
}

async function saveBudget() {
  busy.value = true
  try {
    const p = period.value.replace('/', '-')
    await api.putAccountingBudget(ledgerId.value, {
      period: p,
      lines: lines.value.filter((ln) => ln.scopeKey && ln.amount),
    })
    notify.success('预算已保存')
    await loadBudget()
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '保存失败')
  } finally {
    busy.value = false
  }
}

async function loadAnalysis() {
  busy.value = true
  try {
    const p = period.value.replace('/', '-')
    analysis.value = await api.getAccountingBudgetAnalysis(ledgerId.value, p)
    notify.success('执行分析已生成')
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '分析失败')
  } finally {
    busy.value = false
  }
}

loadBudget()
</script>

<style scoped>
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
}
.form-row.inline {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}
.sub-title {
  margin: 1rem 0 0.5rem;
  font-size: 0.875rem;
}
.budget-row {
  display: grid;
  grid-template-columns: 6rem 1fr 6rem auto;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  align-items: center;
}
.alert-box {
  background: color-mix(in srgb, var(--accent-soft) 50%, var(--bg-card));
  border: 1px solid color-mix(in srgb, var(--accent) 35%, var(--border));
  border-radius: var(--radius-sm);
  padding: 0.65rem 0.85rem;
  margin-bottom: 0.75rem;
  font-size: 0.8125rem;
}
.alert-box ul {
  margin: 0.35rem 0 0;
  padding-left: 1.1rem;
}
.row-warn {
  background: color-mix(in srgb, var(--accent-soft) 25%, transparent);
}
table {
  width: 100%;
  font-size: 0.8125rem;
  border-collapse: collapse;
}
th,
td {
  padding: 0.4rem 0.5rem;
  text-align: left;
  border-bottom: 1px solid var(--border);
}
</style>
