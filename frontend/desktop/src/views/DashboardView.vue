<template>
  <div class="page">
    <h2>系统概览</h2>
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div class="grid-3">
      <div class="card"><h4>网关</h4><div class="val">{{ health?.status || '…' }}</div></div>
      <div class="card"><h4>MiniLedger</h4><div class="val">{{ health?.miniLedgerOnline ? '在线' : '离线' }}</div></div>
      <div class="card"><h4>账本数</h4><div class="val">{{ ledgers.length }}</div></div>
    </div>
    <div v-if="chainPending > 0 || chainFailed > 0" class="alert alert-warn chain-alert">
      <span>
        上链队列：{{ chainPending }} 条待重试
        <template v-if="chainFailed">，{{ chainFailed }} 条失败</template>
      </span>
      <router-link to="/chain">查看链浏览器 →</router-link>
    </div>
    <div class="panel">
      <h3>快捷入口</h3>
      <div class="quick-actions">
        <router-link to="/import"><button class="btn-primary">Excel 导入</button></router-link>
        <router-link to="/backup"><button class="btn-ghost">备份恢复</button></router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { api, ApiError } from '../api/http'

const health = ref(null)
const ledgers = ref([])
const error = ref('')
const chainPending = ref(0)
const chainFailed = ref(0)

onMounted(async () => {
  try {
    health.value = await api.health()
    chainPending.value = health.value?.chainQueuePending ?? 0
    chainFailed.value = health.value?.chainQueueFailed ?? 0
    ledgers.value = await api.listLedgers()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '加载失败'
  }
})
</script>

<style scoped>
.quick-actions { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.chain-alert {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}
.chain-alert a { color: inherit; font-weight: 600; }
</style>
