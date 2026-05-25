<template>
  <div class="page dashboard">
    <header class="page-header">
      <h2>{{ t('dashboard.title') }}</h2>
      <button class="btn-ghost" type="button" :disabled="loading" @click="load">
        {{ t('dashboard.refresh') }}
      </button>
    </header>

    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="chainPending > 0 || chainFailed > 0" class="alert alert-warn chain-alert">
      <span>
        {{ queueAlertText }}
      </span>
      <router-link to="/chain">{{ t('dashboard.openChain') }}</router-link>
    </div>

    <div class="grid-stats">
      <div class="card">
        <h4>{{ t('dashboard.cardGateway') }}</h4>
        <div class="val">{{ gatewayOk ? t('dashboard.ok') : '…' }}</div>
      </div>
      <div class="card">
        <h4>{{ t('dashboard.cardChain') }}</h4>
        <div class="val" :class="chainOnline ? 'ok' : 'bad'">
          {{ chainOnline ? t('chain.nodeOnline') : t('chain.nodeOffline') }}
        </div>
      </div>
      <div class="card">
        <h4>{{ t('dashboard.cardHeight') }}</h4>
        <div class="val mono">{{ chainHeight }}</div>
      </div>
      <div class="card">
        <h4>{{ t('dashboard.cardLedgers') }}</h4>
        <div class="val">{{ ledgers.length }}</div>
      </div>
    </div>

    <div class="panel">
      <h3>{{ t('dashboard.quickTitle') }}</h3>
      <div class="quick-grid">
        <router-link
          v-for="item in quickLinks"
          :key="item.to"
          :to="item.to"
          class="quick-card"
        >
          <span class="quick-label">{{ item.label }}</span>
        </router-link>
      </div>
    </div>

    <div class="panel guide-panel">
      <h3>{{ t('dashboard.guideTitle') }}</h3>
      <ol class="guide-steps">
        <li v-for="(step, i) in guideSteps" :key="i">{{ step }}</li>
      </ol>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { api, ApiError } from '../api/http'
import { useI18n } from '../composables/useI18n'

const { t } = useI18n()

const loading = ref(false)
const error = ref('')
const gatewayOk = ref(false)
const chainOnline = ref(false)
const chainHeight = ref('—')
const ledgers = ref([])
const chainPending = ref(0)
const chainFailed = ref(0)

const quickLinks = computed(() => [
  { to: '/ledgers', label: t('layout.nav.ledgers') },
  { to: '/entry-templates', label: t('layout.nav.templates') },
  { to: '/import', label: t('layout.nav.import') },
  { to: '/backup', label: t('layout.nav.backup') },
  { to: '/friends', label: t('layout.nav.friends') },
  { to: '/teams', label: t('layout.nav.teams') },
  { to: '/chain', label: t('layout.nav.chain') },
  { to: '/settings', label: t('layout.settings') },
])

const guideSteps = computed(() => [
  t('dashboard.guide1'),
  t('dashboard.guide2'),
  t('dashboard.guide3'),
  t('dashboard.guide4'),
  t('dashboard.guide5'),
])

const queueAlertText = computed(() => {
  const p = chainPending.value
  const f = chainFailed.value
  if (f > 0) return t('dashboard.queueAlertBoth').replace('{pending}', p).replace('{failed}', f)
  return t('dashboard.queueAlertPending').replace('{pending}', p)
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [healthRes, chainRes, ledgerList] = await Promise.all([
      api.health().catch(() => null),
      api.chainStatus().catch(() => null),
      api.listLedgers(),
    ])
    gatewayOk.value = healthRes?.status === 'ok' || healthRes?.gateway === 'ok'
    ledgers.value = Array.isArray(ledgerList) ? ledgerList : []

    if (chainRes) {
      chainOnline.value = !!chainRes.online
      chainHeight.value =
        chainRes.height != null && chainRes.height !== ''
          ? String(chainRes.height)
          : '—'
      chainPending.value = chainRes.queuePending ?? healthRes?.chainQueuePending ?? 0
      chainFailed.value = chainRes.queueFailed ?? healthRes?.chainQueueFailed ?? 0
    } else if (healthRes) {
      chainOnline.value = !!healthRes.miniLedgerOnline
      chainPending.value = healthRes.chainQueuePending ?? 0
      chainFailed.value = healthRes.chainQueueFailed ?? 0
    }
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t('dashboard.loadFail')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.dashboard .page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.25rem;
}
.dashboard .page-header h2 {
  margin: 0;
}
.grid-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 1rem;
  margin-bottom: 1.25rem;
}
.val.ok { color: var(--success); }
.val.bad { color: var(--text-muted); }
.quick-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 0.65rem;
}
.quick-card {
  display: block;
  padding: 0.85rem 1rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text);
  text-decoration: none;
  font-weight: 500;
  transition: border-color 0.15s ease, background 0.15s ease;
}
.quick-card:hover {
  border-color: var(--accent);
  background: var(--accent-soft);
  color: var(--accent);
}
.guide-steps {
  margin: 0;
  padding-left: 1.25rem;
  line-height: 1.65;
  color: var(--text);
}
.guide-steps li + li {
  margin-top: 0.5rem;
}
.chain-alert {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}
.chain-alert a {
  color: inherit;
  font-weight: 600;
  text-decoration: none;
}
.chain-alert a:hover {
  text-decoration: underline;
}
</style>
