<template>
  <div>
    <h2>系统概览</h2>
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div class="grid-3">
      <div class="card"><h4>网关</h4><div class="val">{{ health?.status || '…' }}</div></div>
      <div class="card"><h4>MiniLedger</h4><div class="val">{{ health?.miniLedgerOnline ? '在线' : '离线' }}</div></div>
      <div class="card"><h4>账本数</h4><div class="val">{{ ledgers.length }}</div></div>
    </div>
    <div class="panel">
      <h3>快捷入口</h3>
      <router-link to="/import"><button class="btn-primary">Excel 导入</button></router-link>
      <router-link to="/backup" style="margin-left:0.5rem"><button class="btn-ghost">备份恢复</button></router-link>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { api, ApiError } from '../api/http'

const health = ref(null)
const ledgers = ref([])
const error = ref('')

onMounted(async () => {
  try {
    health.value = await api.health()
    ledgers.value = await api.listLedgers()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '加载失败'
  }
})
</script>
