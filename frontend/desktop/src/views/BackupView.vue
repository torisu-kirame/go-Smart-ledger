<template>
  <div class="page">
    <h2>备份 / 恢复</h2>
    <p class="page-desc">
      对已<strong>封账锚定</strong>的账本创建加密备份；本地磁盘与 IPFS 双写，备份 CID 写入链上事件。
      恢复可将快照<strong>写回</strong>账本（覆盖模式）。
    </p>
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="msg" class="alert alert-success">{{ msg }}</div>

    <div class="panel">
      <h3>创建备份</h3>
      <div class="form-row">
        <label>账本</label>
        <AppSelect
          v-model="ledgerId"
          placeholder="请选择账本"
          :options="sealedOptions"
        />
        <small v-if="!sealed.length" style="color:var(--warning)">暂无已封账账本，请先在账本详情中封账锚定</small>
      </div>
      <div class="form-row"><label>备份密码</label><input v-model="password" type="password" /></div>
      <button class="btn-primary" :disabled="!ledgerId || !password" @click="doBackup">创建加密备份（本地 + IPFS）</button>
      <div v-if="backupRef" class="result">
        <p class="mono">本地 ref: {{ backupRef }}</p>
        <p v-if="ipfsCid" class="mono">IPFS CID: {{ ipfsCid }}</p>
        <p v-if="anchoredOnChain" class="ok">已上链记录 BackupAnchored 事件</p>
        <p v-if="anchorError" class="warn">链上记录失败: {{ anchorError }}</p>
      </div>
    </div>

    <div class="panel">
      <h3>恢复</h3>
      <div class="form-row"><label>备份引用 ref</label><input v-model="restoreRef" placeholder="ledgerId/xxxx" /></div>
      <div class="form-row"><label>IPFS CID（可选，本地缺失时从 IPFS 拉取）</label><input v-model="restoreIpfsCid" class="mono" /></div>
      <div class="form-row"><label>密码</label><input v-model="restorePassword" type="password" /></div>
      <div class="form-row">
        <label>目标账本 ID</label>
        <input v-model="restoreLedgerId" placeholder="写回此账本" />
      </div>
      <label class="inline-check"><input type="checkbox" v-model="overwrite" /> 覆盖写入（目标账本已有数据时需勾选）</label>
      <div style="margin-top:0.75rem;display:flex;gap:0.5rem">
        <button class="btn-ghost" @click="doRestorePreview">解密并预览</button>
        <button class="btn-primary" :disabled="!canRestore" @click="doRestoreCommit">写回账本</button>
      </div>
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
import AppSelect from '../components/AppSelect.vue'
import { api, ApiError } from '../api/http'

const route = useRoute()
const ledgers = ref([])
const ledgerId = ref(route.query.ledgerId || '')
const password = ref('')
const backupRef = ref('')
const ipfsCid = ref('')
const anchoredOnChain = ref(false)
const anchorError = ref('')
const restoreRef = ref('')
const restoreIpfsCid = ref('')
const restorePassword = ref('')
const restoreLedgerId = ref('')
const overwrite = ref(false)
const snapshot = ref(null)
const error = ref('')
const msg = ref('')

const sealed = computed(() => ledgers.value.filter((l) => l.anchorStatus === 'synced'))
const sealedOptions = computed(() =>
  sealed.value.map((l) => ({ value: l.id, label: l.name }))
)
const canRestore = computed(
  () => restoreRef.value && restorePassword.value && restoreLedgerId.value
)

onMounted(async () => {
  ledgers.value = await api.listLedgers()
  if (ledgerId.value) restoreLedgerId.value = ledgerId.value
})

async function doBackup() {
  error.value = ''
  msg.value = ''
  anchorError.value = ''
  try {
    const r = await api.ledgerBackup(ledgerId.value, password.value)
    backupRef.value = r.ref
    ipfsCid.value = r.ipfsCid || ''
    anchoredOnChain.value = !!r.anchoredOnChain
    anchorError.value = r.anchorError || ''
    restoreRef.value = r.ref
    restoreIpfsCid.value = r.ipfsCid || ''
    restoreLedgerId.value = ledgerId.value
    msg.value = r.ipfsCid
      ? '备份成功：已双写本地与 IPFS'
      : '备份成功（仅本地；IPFS 未启用或上传失败）'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '备份失败'
  }
}

async function doRestorePreview() {
  error.value = ''
  snapshot.value = null
  try {
    snapshot.value = await api.restorePreview(restoreLedgerId.value, {
      ref: restoreRef.value,
      password: restorePassword.value,
      ipfsCid: restoreIpfsCid.value,
    })
    msg.value = '解密成功，可预览或写回'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '恢复失败'
  }
}

async function doRestoreCommit() {
  if (!confirm('确认将快照写回账本？覆盖模式下将替换链上该账本的元数据与事件。')) return
  error.value = ''
  try {
    const r = await api.restoreCommit(restoreLedgerId.value, {
      ref: restoreRef.value,
      password: restorePassword.value,
      ipfsCid: restoreIpfsCid.value,
      overwrite: overwrite.value,
    })
    msg.value = `已写回账本，当前序号 ${r.latestSeq}`
    snapshot.value = null
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '写回失败'
  }
}
</script>

<style scoped>
.muted { color: var(--text-muted); }
.snap { max-height: 320px; overflow: auto; font-size: 0.75rem; background: var(--bg); padding: 1rem; border-radius: 8px; }
.result { margin-top: 0.75rem; font-size: 0.875rem; }
.ok { color: var(--success); }
.warn { color: var(--warning); }
.check { display: flex; align-items: center; gap: 0.35rem; margin: 0.5rem 0; font-size: 0.875rem; }
</style>
