<template>
  <div v-if="ledger || loading" class="ledger-shell">
    <div v-if="loading && !ledger" class="ledger-shell__loading muted">加载中…</div>
    <template v-else-if="ledger">
      <PageHeader :crumbs="crumbs" :subtitle="headerSubtitle" />

      <div class="ledger-hero panel">
        <div class="ledger-hero__top">
          <div class="ledger-hero__identity">
            <h1 class="ledger-hero__name">{{ ledger.name }}</h1>
            <div class="ledger-hero__meta">
              <span :class="['badge', ledger.type === 'multi' ? 'badge-multi' : 'badge-private']">
                {{ ledger.type === 'multi' ? '多人' : '私人' }}
              </span>
              <span class="ledger-hero__meta-item mono">序号 {{ ledger.latestSeq }}</span>
              <span
                :class="['badge', ledger.anchorStatus === 'synced' ? 'badge-ok' : 'badge-pending']"
              >
                {{ ledger.anchorStatus }}
              </span>
            </div>
          </div>
        </div>

        <nav class="ledger-tabs" aria-label="账本页面">
          <router-link
            :to="basePath"
            class="ledger-tab"
            :class="{ active: activeTab === 'overview' }"
          >
            详情
          </router-link>
          <router-link
            :to="`${basePath}/view`"
            class="ledger-tab"
            :class="{ active: activeTab === 'view' }"
          >
            查看
          </router-link>
          <router-link
            :to="`${basePath}/settings`"
            class="ledger-tab"
            :class="{ active: activeTab === 'settings' }"
          >
            设置
          </router-link>
        </nav>
      </div>

      <div class="ledger-body">
        <router-view />
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import { provideLedgerDetail } from '../composables/useLedgerDetail'
import { useI18n } from '../composables/useI18n'

const route = useRoute()
const { t } = useI18n()
const ledgerId = route.params.id
const basePath = computed(() => `/ledgers/${ledgerId}`)

const { ledger, loading, load } = provideLedgerDetail(ledgerId)

const activeTab = computed(() => {
  const p = route.path
  if (p.endsWith('/settings')) return 'settings'
  if (p.endsWith('/view')) return 'view'
  return 'overview'
})

const crumbs = computed(() => {
  const name = ledger.value?.name || '…'
  const suffixMap = { overview: '详情', view: '查看', settings: '设置' }
  const suffix = suffixMap[activeTab.value] || '详情'
  return [
    { label: t('layout.nav.ledgers'), to: '/ledgers' },
    { label: `${name} ${suffix}` },
  ]
})

const headerSubtitle = computed(() => {
  if (!ledger.value?.ledgerAddress) return ''
  const a = ledger.value.ledgerAddress
  return a.length > 20 ? `${a.slice(0, 10)}…${a.slice(-8)}` : a
})

onMounted(load)
</script>

<style scoped>
.ledger-shell {
  max-width: 72rem;
}
.ledger-shell__loading {
  padding: 1rem;
}
.ledger-body {
  margin-top: 1rem;
}
.ledger-hero {
  padding: 0;
  overflow: hidden;
}
.ledger-hero__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1.25rem;
  padding: 1.15rem 1.25rem 0.85rem;
  flex-wrap: wrap;
}
.ledger-hero__name {
  margin: 0 0 0.5rem;
  font-size: 1.35rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1.25;
}
.ledger-hero__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 0.75rem;
}
.ledger-hero__meta-item {
  font-size: 0.8125rem;
  color: var(--text-muted);
}
.ledger-tabs {
  display: flex;
  gap: 0;
  padding: 0 1.25rem;
  border-top: 1px solid var(--border);
  background: var(--bg-elevated);
}
.ledger-tab {
  position: relative;
  padding: 0.7rem 1.1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-muted);
  text-decoration: none;
  transition: color 0.15s ease, background 0.15s ease;
}
.ledger-tab:hover {
  color: var(--text);
  background: var(--hover);
}
.ledger-tab.active {
  color: var(--accent);
}
.ledger-tab.active::after {
  content: '';
  position: absolute;
  left: 0.65rem;
  right: 0.65rem;
  bottom: 0;
  height: 2px;
  border-radius: 2px 2px 0 0;
  background: var(--accent);
}
</style>
