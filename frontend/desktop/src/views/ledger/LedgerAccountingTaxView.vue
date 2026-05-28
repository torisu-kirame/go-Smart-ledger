<template>
  <div class="acct-tax">
    <section class="detail-card">
      <h3 class="detail-card__title">税务模板（F47）</h3>
      <div class="form-row inline">
        <label>内置模板</label>
        <select v-model="selectedPreset" class="field-sm">
          <option value="">— 选择 —</option>
          <option v-for="p in presets" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
        <button type="button" class="btn-ghost" :disabled="busy" @click="applyPreset">应用模板</button>
        <button type="button" class="btn-primary" :disabled="busy" @click="saveTemplate">保存</button>
      </div>
      <div class="form-grid">
        <label>计税模式</label>
        <select v-model="template.mode" class="field-sm">
          <option value="general">一般纳税人</option>
          <option value="simple">简易计税</option>
          <option value="none">免税/不征税</option>
        </select>
        <label>销项税率</label>
        <input v-model="template.defaultOutputRate" class="field-sm" placeholder="0.13" />
        <label>进项税率</label>
        <input v-model="template.defaultInputRate" class="field-sm" placeholder="0.13" />
        <label>简易征收率</label>
        <input v-model="template.simpleLevyRate" class="field-sm" placeholder="0.03" />
      </div>
      <p class="field-hint">
        过账时在分录行选择税务类别（应税/免税/零税率）；收入类计入销项，费用类计入进项。
      </p>
    </section>

    <section class="detail-card">
      <h3 class="detail-card__title">税务报表</h3>
      <div class="form-row inline">
        <label>期间</label>
        <input v-model="period" type="month" class="field-sm" />
        <button type="button" class="btn-primary" :disabled="busy" @click="loadReport">生成报表</button>
      </div>
      <template v-if="report">
        <div class="tax-totals">
          <div>销项税：<strong class="mono">{{ report.outputTaxTotal }}</strong></div>
          <div>进项税：<strong class="mono">{{ report.inputTaxTotal }}</strong></div>
          <div>应纳：<strong class="mono">{{ report.netPayable }}</strong></div>
        </div>
        <table v-if="report.lines?.length">
          <thead>
            <tr>
              <th>日期</th>
              <th>科目</th>
              <th>计税基数</th>
              <th>税率</th>
              <th>税额</th>
              <th>类型</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(ln, i) in report.lines" :key="i">
              <td>{{ ln.date }}</td>
              <td class="mono">{{ ln.accountCode }} {{ ln.accountName }}</td>
              <td class="mono">{{ ln.baseAmount }}</td>
              <td>{{ ln.taxRate }}</td>
              <td class="mono">{{ ln.taxAmount }}</td>
              <td>{{ taxKindLabel(ln.kind) }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="muted">该期间无应税分录</div>
      </template>
    </section>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { api, ApiError } from '../../api/http'
import { useLedgerAccounting } from '../../composables/provideAccounting'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { useNotify } from '../../composables/useNotify'

const { busy } = useLedgerAccounting()
const { ledgerId } = useLedgerDetail()
const notify = useNotify()

const presets = ref([])
const selectedPreset = ref('')
const template = ref({
  mode: 'general',
  defaultOutputRate: '0.13',
  defaultInputRate: '0.13',
  simpleLevyRate: '0.03',
})
const period = ref(new Date().toISOString().slice(0, 7))
const report = ref(null)

function taxKindLabel(k) {
  if (k === 'output') return '销项'
  if (k === 'input') return '进项'
  if (k === 'levy') return '简易征收'
  return k
}

async function loadPresets() {
  const res = await api.getAccountingTaxPresets(ledgerId.value)
  presets.value = res.presets || []
}

async function loadTemplate() {
  try {
    template.value = await api.getAccountingTax(ledgerId.value)
  } catch {
    /* default */
  }
}

async function applyPreset() {
  if (!selectedPreset.value) return
  busy.value = true
  try {
    template.value = await api.applyAccountingTaxPreset(ledgerId.value, selectedPreset.value)
    notify.success('模板已应用')
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '应用失败')
  } finally {
    busy.value = false
  }
}

async function saveTemplate() {
  busy.value = true
  try {
    template.value = await api.putAccountingTax(ledgerId.value, template.value)
    notify.success('税务设置已保存')
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '保存失败')
  } finally {
    busy.value = false
  }
}

async function loadReport() {
  busy.value = true
  try {
    const p = period.value.replace('/', '-')
    report.value = await api.getAccountingTaxReport(ledgerId.value, p)
    notify.success('税务报表已生成')
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '生成失败')
  } finally {
    busy.value = false
  }
}

loadPresets()
loadTemplate()
</script>

<style scoped>
.detail-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1rem 1.1rem;
  margin-bottom: 1rem;
}
.detail-card__title {
  margin: 0 0 0.85rem;
  font-size: 0.8125rem;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
}
.field-hint {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: 0.5rem 0 0;
}
.form-grid {
  display: grid;
  grid-template-columns: 7rem 1fr;
  gap: 0.5rem;
  align-items: center;
  max-width: 24rem;
}
.form-row.inline {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 0.75rem;
}
.tax-totals {
  display: flex;
  flex-wrap: wrap;
  gap: 1.5rem;
  margin: 0.75rem 0;
  font-size: 0.875rem;
}
table {
  width: 100%;
  font-size: 0.8125rem;
  border-collapse: collapse;
}
th,
td {
  padding: 0.35rem 0.5rem;
  border-bottom: 1px solid var(--border);
  text-align: left;
}
</style>
