<template>
  <div class="page chain-page">
    <header class="head">
      <div>
        <h2>MiniLedger 链浏览器</h2>
        <p class="muted">官方 Dashboard 内嵌 + 链状态与待上链队列</p>
      </div>
      <div class="tabs">
        <button
          type="button"
          :class="{ active: tab === 'dashboard' }"
          @click="tab = 'dashboard'"
        >
          官方浏览器
        </button>
        <button
          type="button"
          :class="{ active: tab === 'overview' }"
          @click="tab = 'overview'; loadOverview()"
        >
          概览 / 队列
        </button>
      </div>
    </header>

    <div v-if="tab === 'dashboard'" class="frame-wrap">
      <iframe
        class="explorer-frame"
        src="/miniledger/dashboard"
        title="MiniLedger Dashboard"
      />
    </div>

    <template v-else>
      <div v-if="error" class="alert alert-error">{{ error }}</div>
      <div class="grid-3">
        <div class="card">
          <h4>链节点</h4>
          <div class="val">{{ status?.online ? '在线' : '离线' }}</div>
          <small v-if="status?.height != null">高度 {{ status.height }}</small>
        </div>
        <div class="card">
          <h4>共识角色</h4>
          <div class="val">{{ status?.role || '—' }}</div>
          <small>{{ status?.uptime || '' }}</small>
        </div>
        <div class="card">
          <h4>待上链队列</h4>
          <div class="val">{{ status?.queuePending ?? 0 }} 待重试</div>
          <small v-if="status?.queueFailed">{{ status.queueFailed }} 已失败</small>
        </div>
      </div>

      <div class="panel" v-if="consensus">
        <h3>共识状态</h3>
        <pre class="json-block">{{ pretty(consensus) }}</pre>
      </div>

      <div class="panel">
        <div class="panel-head">
          <h3>最近区块</h3>
          <button class="btn-ghost" type="button" @click="loadOverview">刷新</button>
        </div>
        <table v-if="blocks.length" class="data-table">
          <thead>
            <tr>
              <th>高度</th>
              <th>哈希</th>
              <th>交易数</th>
              <th>时间</th>
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
        <p v-else class="muted">暂无区块数据</p>
      </div>

      <div class="panel">
        <div class="panel-head">
          <h3>上链重试队列（F23）</h3>
          <button class="btn-ghost" type="button" @click="loadQueue">刷新</button>
        </div>
        <p v-if="!queue.length" class="muted">当前没有待重试的上链任务</p>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th>标签</th>
              <th>账本</th>
              <th>状态</th>
              <th>尝试</th>
              <th>错误</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in queue" :key="item.id">
              <td>{{ item.label }}</td>
              <td class="mono">{{ item.ledgerId || '—' }}</td>
              <td>{{ item.status }}</td>
              <td>{{ item.attempts }}</td>
              <td class="err-cell">{{ item.lastError || '—' }}</td>
              <td>
                <button
                  v-if="item.status !== 'done'"
                  class="btn-ghost"
                  type="button"
                  @click="retry(item.id)"
                >
                  立即重试
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
import { onMounted, ref } from 'vue'
import { api, ApiError } from '../api/http'

const tab = ref('dashboard')
const status = ref(null)
const consensus = ref(null)
const blocks = ref([])
const queue = ref([])
const error = ref('')

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
    error.value = e instanceof ApiError ? e.message : '加载链状态失败'
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
    error.value = e instanceof ApiError ? e.message : '重试失败'
  }
}

onMounted(() => {
  if (tab.value === 'overview') loadOverview()
})
</script>

<style scoped>
.chain-page { display: flex; flex-direction: column; gap: 1rem; min-height: 0; }
.head { display: flex; flex-wrap: wrap; align-items: flex-start; justify-content: space-between; gap: 1rem; }
.head h2 { margin: 0; }
.muted { color: var(--text-muted); margin: 0.25rem 0 0; font-size: 0.875rem; }
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
  flex: 1;
  min-height: calc(100vh - 12rem);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  background: #0d1117;
}
.explorer-frame {
  width: 100%;
  height: 100%;
  min-height: 640px;
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
.card small { display: block; margin-top: 0.35rem; color: var(--text-muted); font-size: 0.75rem; }
</style>
