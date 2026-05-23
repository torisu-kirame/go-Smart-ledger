<template>
  <div class="page chain-page">
    <header class="head">
      <div>
        <h2>{{ t('chain.title') }}</h2>
      </div>
      <div class="tabs">
        <button
          type="button"
          :class="{ active: tab === 'dashboard' }"
          @click="tab = 'dashboard'"
        >
          {{ t('chain.tabOfficial') }}
        </button>
        <button
          type="button"
          :class="{ active: tab === 'overview' }"
          @click="tab = 'overview'; loadOverview()"
        >
          {{ t('chain.tabOverview') }}
        </button>
      </div>
    </header>

    <div v-if="tab === 'dashboard'" class="frame-wrap">
      <iframe
        :key="locale"
        class="explorer-frame"
        :src="dashboardSrc"
        :title="t('chain.iframeTitle')"
      />
    </div>

    <template v-else>
      <div v-if="error" class="alert alert-error">{{ error }}</div>
      <div class="grid-3">
        <div class="card">
          <h4>{{ t('chain.cardNode') }}</h4>
          <div class="val">{{ status?.online ? t('chain.nodeOnline') : t('chain.nodeOffline') }}</div>
        </div>
        <div class="card">
          <h4>{{ t('chain.cardRole') }}</h4>
          <div class="val">{{ status?.role || '—' }}</div>
        </div>
        <div class="card">
          <h4>{{ t('chain.cardQueue') }}</h4>
          <div class="val">{{ status?.queuePending ?? 0 }} {{ t('chain.pendingRetry') }}</div>
        </div>
      </div>

      <div class="panel" v-if="consensus">
        <h3>{{ t('chain.consensus') }}</h3>
        <pre class="json-block">{{ pretty(consensus) }}</pre>
      </div>

      <div class="panel">
        <div class="panel-head">
          <h3>{{ t('chain.recentBlocks') }}</h3>
          <button class="btn-ghost" type="button" @click="loadOverview">{{ t('chain.refresh') }}</button>
        </div>
        <table v-if="blocks.length" class="data-table">
          <thead>
            <tr>
              <th>{{ t('chain.colHeight') }}</th>
              <th>{{ t('chain.colHash') }}</th>
              <th>{{ t('chain.colTxs') }}</th>
              <th>{{ t('chain.colTime') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="b in blocks" :key="b.height">
              <td>{{ b.height }}</td>
              <td class="mono">{{ short(b.hash) }}</td>
              <td>{{ b.txCount ?? b.transactions?.length ?? '—' }}</td>
              <td>{{ b.timestamp || b.time || '—' }}</td>
            </tr>
          </tbody>
        </table>
        <p v-else class="muted">{{ t('chain.noBlocks') }}</p>
      </div>

      <div class="panel">
        <div class="panel-head">
          <h3>{{ t('chain.queueTitle') }}</h3>
          <button class="btn-ghost" type="button" @click="loadQueue">{{ t('chain.refresh') }}</button>
        </div>
        <p v-if="!queue.length" class="muted">{{ t('chain.noQueue') }}</p>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th>{{ t('chain.colLabel') }}</th>
              <th>{{ t('chain.colLedger') }}</th>
              <th>{{ t('chain.colStatus') }}</th>
              <th>{{ t('chain.colAttempts') }}</th>
              <th>{{ t('chain.colError') }}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in queue" :key="item.id">
              <td>{{ item.label }}</td>
              <td class="mono">{{ item.ledgerId || '—' }}</td>
              <td>{{ queueStatusLabel(item.status) }}</td>
              <td>{{ item.attempts }}</td>
              <td class="err-cell">{{ item.lastError || '—' }}</td>
              <td>
                <button
                  v-if="item.status !== 'done'"
                  class="btn-ghost"
                  type="button"
                  @click="retry(item.id)"
                >
                  {{ t('chain.retry') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { api, ApiError } from '../api/http'
import { useI18n } from '../composables/useI18n'
import { dashboardPath } from '../utils/locale'

const { locale, t } = useI18n()

const tab = ref('dashboard')
const status = ref(null)
const consensus = ref(null)
const blocks = ref([])
const queue = ref([])
const error = ref('')

const dashboardSrc = computed(() => dashboardPath(locale.value))

function queueStatusLabel(s) {
  const map = {
    pending: 'chain.statusPending',
    retrying: 'chain.statusRetrying',
    failed: 'chain.statusFailed',
    done: 'chain.statusDone',
  }
  const key = map[s]
  return key ? t(key) : s
}

function pretty(v) {
  return JSON.stringify(v, null, 2)
}

function short(h) {
  if (!h || typeof h !== 'string') return '—'
  return h.length > 16 ? `${h.slice(0, 10)}…${h.slice(-6)}` : h
}

async function loadOverview() {
  error.value = ''
  try {
    status.value = await api.chainStatus()
    try {
      consensus.value = await api.chainConsensus()
    } catch {
      consensus.value = null
    }
    const res = await api.chainBlocks(1, 10)
    blocks.value = res?.blocks || res?.items || (Array.isArray(res) ? res : [])
    await loadQueue()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t('chain.loadFail')
  }
}

async function loadQueue() {
  const res = await api.chainQueue()
  queue.value = res?.items || []
}

async function retry(id) {
  try {
    await api.retryChainQueue(id)
    await loadQueue()
    status.value = await api.chainStatus()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t('chain.retryFail')
  }
}

watch(locale, () => {
  if (tab.value === 'overview') loadOverview()
})

onMounted(() => {
  if (tab.value === 'overview') loadOverview()
})
</script>

<style scoped>
.chain-page { display: flex; flex-direction: column; gap: 1rem; min-height: 0; }
.head { display: flex; flex-wrap: wrap; align-items: flex-start; justify-content: space-between; gap: 1rem; }
.head h2 { margin: 0; }
.muted { color: var(--text-muted); font-size: 0.875rem; }
.tabs { display: flex; gap: 0.5rem; }
.tabs button {
  padding: 0.4rem 0.85rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  cursor: pointer;
}
.tabs button.active {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
}
.frame-wrap {
  flex: 1 1 auto;
  height: min(720px, calc(100vh - 11rem));
  min-height: 480px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  background: #0d1117;
}
.explorer-frame {
  display: block;
  width: 100%;
  height: 100%;
  min-height: 480px;
  border: 0;
}
.panel-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.75rem; }
.panel-head h3 { margin: 0; }
.json-block {
  font-size: 0.75rem;
  overflow: auto;
  max-height: 200px;
  margin: 0;
  padding: 0.75rem;
  background: var(--surface-elevated);
  border-radius: var(--radius-sm);
}
.data-table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
.data-table th,
.data-table td { padding: 0.5rem 0.65rem; border-bottom: 1px solid var(--border); text-align: left; }
.mono { font-family: ui-monospace, monospace; font-size: 0.8rem; }
.err-cell { max-width: 240px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
