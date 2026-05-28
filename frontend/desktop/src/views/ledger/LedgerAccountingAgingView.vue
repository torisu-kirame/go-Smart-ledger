<template>
  <div class="acct-aging">
    <section class="detail-card">
      <h3 class="detail-card__title">应收 / 应付账龄（F45）</h3>
      <div class="form-row inline">
        <label>截止日期</label>
        <input v-model="asOf" type="date" class="field-sm" />
        <label>应收科目</label>
        <input v-model="recvAccounts" type="text" class="field-sm mono" placeholder="1122" />
        <label>应付科目</label>
        <input v-model="payAccounts" type="text" class="field-sm mono" placeholder="2202" />
        <button type="button" class="btn-primary" :disabled="busy" @click="loadReport">生成账龄</button>
      </div>
      <p class="field-hint">
        按凭证分录上的<strong>往来方</strong>汇总未清余额，FIFO 核销后按账龄区间（0–30 / 31–60 / 61–90 / 90+ 天）列示。
        过账时请在分录行填写「往来方」。
      </p>
    </section>

    <section v-if="report" class="detail-card">
      <h3 class="detail-card__title">汇总 · 截至 {{ report.asOf }}</h3>
      <div v-if="!report.summaries?.length" class="muted">暂无未清应收/应付</div>
      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>类型</th>
              <th>往来方</th>
              <th>科目</th>
              <th>合计</th>
              <th>0–30天</th>
              <th>31–60天</th>
              <th>61–90天</th>
              <th>90天+</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in report.summaries" :key="s.kind + s.counterparty + s.accountCode">
              <td>{{ s.kind === 'receivable' ? '应收' : '应付' }}</td>
              <td>{{ s.counterparty }}</td>
              <td class="mono">{{ s.accountCode }} {{ s.accountName }}</td>
              <td class="mono"><strong>{{ s.total }}</strong></td>
              <td class="mono">{{ s.current || '—' }}</td>
              <td class="mono">{{ s.days31_60 || '—' }}</td>
              <td class="mono">{{ s.days61_90 || '—' }}</td>
              <td class="mono">{{ s.over90 || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <h4 v-if="report.items?.length" class="sub-title">未清明细</h4>
      <div v-if="report.items?.length" class="table-wrap">
        <table class="detail-table">
          <thead>
            <tr>
              <th>往来方</th>
              <th>类型</th>
              <th>凭证日</th>
              <th>天数</th>
              <th>区间</th>
              <th>金额</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(it, idx) in report.items" :key="idx">
              <td>{{ it.counterparty }}</td>
              <td>{{ it.kind === 'receivable' ? '应收' : '应付' }}</td>
              <td>{{ it.date }}</td>
              <td>{{ it.days }}</td>
              <td>{{ it.bucket }}</td>
              <td class="mono">{{ it.amount }}</td>
            </tr>
          </tbody>
        </table>
      </div>
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

const asOf = ref(new Date().toISOString().slice(0, 10))
const recvAccounts = ref('1122')
const payAccounts = ref('2202')
const report = ref(null)

async function loadReport() {
  busy.value = true
  try {
    report.value = await api.getAccountingAging(ledgerId.value, {
      asOf: asOf.value,
      receivableAccounts: recvAccounts.value,
      payableAccounts: payAccounts.value,
    })
    notify.success('账龄表已生成')
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '生成失败')
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
  margin: 0;
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
.detail-table {
  margin-top: 0.5rem;
}
</style>
