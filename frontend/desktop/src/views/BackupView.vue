<template>
  <div>
    <h2>备份 / 恢复</h2>
    <p class="muted">仅支持对已<strong>封账锚定</strong>（anchorStatus=synced）的账本创建加密备份；恢复仅预览快照内容。</p>
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="msg" class="alert alert-success">{{ msg }}</div>

    <div class="panel">
      <h3>创建备份</h3>
      <div class="form-row">
        <label>账本</label>
        <select v-model="ledgerId">
          <option value="">请选择</option>
          <option v-for="l in sealed" :key="l.id" :value="l.id">{{ l.name }}</option>
        </select>
        <small v-if="!sealed.length" style="color:var(--warning)">暂无已封账账本，请先在账本详情中封账锚定</small>
      </div>
      <div class="form-row"><label>备份密码</label><input v-model="password" type="password" /></div>
      <button class="btn-primary" :disabled="!ledgerId || !password" @click="doBackup">创建加密备份</button>
      <p v-if="backupRef" class="mono">备份引用: {{ backupRef }}</p>
    </div>

    <div class="panel">
      <h3>恢复预览</h3>
      <div class="form-row"><label>备份引用 ref</label><input v-model="restoreRef" placeholder="ledgerId/xxxx" /></div>
      <div class="form-row"><label>密码</label><input v-model="restorePassword" type="password" /></div>
      <div class="form-row">
        <label>关联账本 ID（用于 API）</label>
        <input v-model="restoreLedgerId" placeholder="与备份时账本一致" />
      </div>
      <button class="btn-ghost" @click="doRestorePreview">解密并预览</button>
    </div>

    <div v-if="snapshot" class="panel">
      <h3>快照预览</h3>
      <p class="mono">账本 {{ snapshot.ledgerId }} · 导出 {{ snapshot.exportedAt }}</p>
      <p>事件数: {{ snapshot.events?.length || 0 }}</p>
      <pre class="snap">{{ JSON.stringify(snapshot, null, 2) }}</pre>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, ApiError } from '../api/http'

const route = useRoute()
const ledgers = ref([])
const ledgerId = ref(route.query.ledgerId || '')
const password = ref('')
const backupRef = ref('')
const restoreRef = ref('')
const restorePassword = ref('')
const restoreLedgerId = ref('')
const snapshot = ref(null)
const error = ref('')
const msg = ref('')

const sealed = computed(() => ledgers.value.filter((l) => l.anchorStatus === 'synced'))

onMounted(async () => {
  ledgers.value = await api.listLedgers()
  if (ledgerId.value) restoreLedgerId.value = ledgerId.value
})

async function doBackup() {
  error.value = ''
  msg.value = ''
  try {
    const r = await api.ledgerBackup(ledgerId.value, password.value)
    backupRef.value = r.ref
    restoreRef.value = r.ref
    restoreLedgerId.value = ledgerId.value
    msg.value = '备份成功，请妥善保存 ref 与密码'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '备份失败'
  }
}

async function doRestorePreview() {
  error.value = ''
  snapshot.value = null
  try {
    snapshot.value = await api.restorePreview(restoreLedgerId.value, restoreRef.value, restorePassword.value)
    msg.value = '解密成功'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '恢复失败'
  }
}
</script>

<style scoped>
.muted { color: var(--text-muted); }
.snap { max-height: 320px; overflow: auto; font-size: 0.75rem; background: var(--bg); padding: 1rem; border-radius: 8px; }
</style>
