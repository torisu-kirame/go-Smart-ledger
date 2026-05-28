<template>
  <div class="ledger-accounting">
    <section class="acct-summary detail-card">
      <div class="acct-summary__stats">
        <div>
          <span class="acct-summary__label">科目</span>
          <strong>{{ chart.accounts?.length || 0 }}</strong>
        </div>
        <div>
          <span class="acct-summary__label">已过账凭证</span>
          <strong>{{ journals.length }}</strong>
        </div>
        <div>
          <span class="acct-summary__label">开放期间</span>
          <strong>{{ openPeriodCount }}</strong>
        </div>
      </div>
      <p v-if="!chart.accounts?.length" class="field-hint">
        科目表尚未初始化，请在「科目」页恢复默认或添加科目后再记账。
      </p>
    </section>

    <nav class="acct-tabs" aria-label="财务功能">
      <router-link
        v-for="t in tabs"
        :key="t.path"
        :to="`${basePath}/${t.path}`"
        class="acct-tab"
        :class="{ active: isTabActive(t.path) }"
      >
        {{ t.label }}
      </router-link>
    </nav>

    <router-view />
  </div>
</template>

<script setup>
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { provideLedgerAccounting } from '../../composables/provideAccounting'

const route = useRoute()
const { ledgerId, isProfessionalLedger } = useLedgerDetail()

const basePath = computed(() => `/ledgers/${ledgerId.value}/accounting`)

const tabs = [
  { path: 'view', label: '查看' },
  { path: 'coa', label: '科目' },
  { path: 'period', label: '期间' },
  { path: 'report', label: '报表' },
  { path: 'budget', label: '预算' },
  { path: 'aging', label: '账龄' },
  { path: 'currency', label: '外币' },
  { path: 'tax', label: '税务' },
  { path: 'attach', label: '附件' },
  { path: 'bank', label: '对账' },
]

const { chart, journals, periods, loadAll } = provideLedgerAccounting(ledgerId, isProfessionalLedger)

const openPeriodCount = computed(
  () => periods.value.filter((p) => p.status !== 'closed').length
)

function isTabActive(path) {
  const seg = route.path.split('/').pop()
  if (path === 'view') return seg === 'view' || seg === 'accounting'
  return seg === path
}

watch(
  () => [ledgerId.value, isProfessionalLedger.value],
  () => {
    if (isProfessionalLedger.value) loadAll(ledgerId, true)
  },
  { immediate: true }
)
</script>

<style scoped>
.ledger-accounting {
  position: relative;
  padding-bottom: 1rem;
}
.acct-summary__stats {
  display: flex;
  flex-wrap: wrap;
  gap: 1.5rem 2rem;
}
.acct-summary__label {
  display: block;
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-bottom: 0.2rem;
}
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
  text-decoration: none;
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
.field-hint {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: 0.5rem 0 0;
}
</style>
