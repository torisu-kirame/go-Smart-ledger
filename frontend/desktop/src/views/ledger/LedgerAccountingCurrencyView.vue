<template>
  <div class="acct-currency">
    <section class="detail-card">
      <h3 class="detail-card__title">多币种设置（F46）</h3>
      <div class="form-grid">
        <label>本位币</label>
        <input v-model="settings.baseCurrency" class="field-sm mono" />
        <label>汇兑损益科目</label>
        <input v-model="settings.gainLossAccount" class="field-sm mono" />
        <label>启用币种（逗号分隔）</label>
        <input v-model="currenciesText" class="field-sm" placeholder="CNY,USD,EUR" />
      </div>
      <button type="button" class="btn-primary" :disabled="busy" @click="saveSettings">保存设置</button>
    </section>

    <section class="detail-card">
      <h3 class="detail-card__title">期末汇率</h3>
      <div class="form-row inline">
        <label>期间</label>
        <input v-model="period" type="month" class="field-sm" />
        <button type="button" class="btn-ghost" :disabled="busy" @click="loadRates">加载</button>
        <button type="button" class="btn-primary" :disabled="busy" @click="saveRates">保存汇率</button>
        <button type="button" class="btn-ghost" :disabled="busy" @click="runReval">期末重估</button>
      </div>
      <p class="field-hint">汇率含义：1 单位外币 = ? 本位币（如 USD 7.25 表示 1 美元 = 7.25 元）。</p>
      <div v-for="(r, i) in fxRates" :key="i" class="rate-row">
        <input v-model="r.currency" class="field-sm mono" placeholder="USD" />
        <input v-model="r.rate" class="field-sm" placeholder="7.25" />
        <DeleteButton icon-only sm title="删除" @click="fxRates.splice(i, 1)" />
      </div>
      <button type="button" class="btn-ghost" @click="fxRates.push({ currency: '', rate: '' })">
        + 汇率
      </button>
    </section>

    <section class="detail-card">
      <h3 class="detail-card__title">外币余额</h3>
      <button type="button" class="btn-ghost" :disabled="busy" @click="loadBalances">刷新余额</button>
      <div v-if="!balances.length" class="muted">暂无外币分录（过账时填写币种与原币金额）</div>
      <table v-else>
        <thead>
          <tr>
            <th>科目</th>
            <th>币种</th>
            <th>原币余额</th>
            <th>账面本位币</th>
            <th>隐含汇率</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in balances" :key="b.accountCode + b.currency">
            <td class="mono">{{ b.accountCode }} {{ b.accountName }}</td>
            <td>{{ b.currency }}</td>
            <td class="mono">{{ b.foreignBalance }}</td>
            <td class="mono">{{ b.bookBalance }}</td>
            <td class="mono">{{ b.impliedRate || '—' }}</td>
          </tr>
        </tbody>
      </table>
    </section>

    <section v-if="reval" class="detail-card">
      <h3 class="detail-card__title">重估结果 · {{ reval.period }}</h3>
      <p>
        合计汇兑损益：<strong class="mono">{{ reval.totalGainLoss }}</strong>
        （计入 {{ reval.gainLossAccount }}）
      </p>
      <table v-if="reval.lines?.length">
        <thead>
          <tr>
            <th>科目</th>
            <th>币种</th>
            <th>原币</th>
            <th>账面</th>
            <th>期末汇率</th>
            <th>重估后</th>
            <th>损益</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="ln in reval.lines" :key="ln.accountCode + ln.currency">
            <td>{{ ln.accountCode }} {{ ln.accountName }}</td>
            <td>{{ ln.currency }}</td>
            <td class="mono">{{ ln.foreignBalance }}</td>
            <td class="mono">{{ ln.bookBalance }}</td>
            <td class="mono">{{ ln.closingRate }}</td>
            <td class="mono">{{ ln.revaluedBalance }}</td>
            <td class="mono">{{ ln.gainLoss }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { api, ApiError } from '../../api/http'
import DeleteButton from '../../components/DeleteButton.vue'
import { useLedgerAccounting } from '../../composables/provideAccounting'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { useNotify } from '../../composables/useNotify'

const { busy } = useLedgerAccounting()
const { ledgerId } = useLedgerDetail()
const notify = useNotify()

const settings = ref({
  baseCurrency: 'CNY',
  gainLossAccount: '6603',
  currencies: ['CNY', 'USD', 'EUR', 'HKD'],
})
const currenciesText = computed({
  get: () => (settings.value.currencies || []).join(','),
  set: (v) => {
    settings.value.currencies = String(v)
      .split(',')
      .map((s) => s.trim().toUpperCase())
      .filter(Boolean)
  },
})
const period = ref(new Date().toISOString().slice(0, 7))
const fxRates = ref([])
const balances = ref([])
const reval = ref(null)

async function loadSettings() {
  try {
    settings.value = await api.getAccountingCurrency(ledgerId.value)
  } catch {
    /* defaults */
  }
}

async function saveSettings() {
  busy.value = true
  try {
    settings.value = await api.putAccountingCurrency(ledgerId.value, settings.value)
    notify.success('币种设置已保存')
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '保存失败')
  } finally {
    busy.value = false
  }
}

async function loadRates() {
  busy.value = true
  try {
    const p = period.value.replace('/', '-')
    const res = await api.getAccountingFxRates(ledgerId.value, p)
    fxRates.value = res.rates?.length ? [...res.rates] : [{ currency: 'USD', rate: '' }]
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '加载失败')
  } finally {
    busy.value = false
  }
}

async function saveRates() {
  busy.value = true
  try {
    const p = period.value.replace('/', '-')
    await api.putAccountingFxRates(ledgerId.value, {
      period: p,
      rates: fxRates.value.filter((r) => r.currency && r.rate),
    })
    notify.success('期末汇率已保存')
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '保存失败')
  } finally {
    busy.value = false
  }
}

async function loadBalances() {
  busy.value = true
  try {
    const res = await api.getAccountingFCBalances(ledgerId.value)
    balances.value = res.balances || []
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '加载失败')
  } finally {
    busy.value = false
  }
}

async function runReval() {
  busy.value = true
  try {
    const p = period.value.replace('/', '-')
    reval.value = await api.getAccountingRevaluation(ledgerId.value, p)
    notify.success('重估完成')
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '重估失败')
  } finally {
    busy.value = false
  }
}

loadSettings()
loadBalances()
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
  margin: 0 0 0.75rem;
}
.form-grid {
  display: grid;
  grid-template-columns: 8rem 1fr;
  gap: 0.5rem 0.75rem;
  align-items: center;
  margin-bottom: 0.75rem;
}
.form-row.inline {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 0.5rem;
}
.rate-row {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  align-items: center;
}
table {
  width: 100%;
  font-size: 0.8125rem;
  border-collapse: collapse;
  margin-top: 0.5rem;
}
th,
td {
  padding: 0.35rem 0.5rem;
  border-bottom: 1px solid var(--border);
  text-align: left;
}
</style>
