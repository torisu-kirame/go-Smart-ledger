<template>
  <div>
    <h2>Excel 导入</h2>
    <p class="muted">下载模板 → 填写 → 上传预览 → 确认导入并封账 → 可选加密备份</p>
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="msg" class="alert alert-success">{{ msg }}</div>

    <div class="panel">
      <button class="btn-ghost" @click="downloadTpl">下载导入模板</button>
    </div>

    <div class="panel">
      <h3>1. 选择账本</h3>
      <select v-model="ledgerId" style="max-width:320px">
        <option value="">请选择</option>
        <option v-for="l in ledgers" :key="l.id" :value="l.id">{{ l.name }} ({{ l.id.slice(0,8) }}…)</option>
      </select>
    </div>

    <div class="panel">
      <h3>2. 上传并预览</h3>
      <input type="file" accept=".xlsx,.xls" @change="onFile" />
      <p v-if="preview" class="muted">共 {{ preview.total }} 行，有效 {{ preview.valid }}，无效 {{ preview.invalid }}</p>
    </div>

    <div v-if="preview?.rows?.length" class="panel">
      <h3>预览</h3>
      <div class="table-wrap" style="max-height:280px;overflow:auto">
        <table>
          <thead><tr><th>行</th><th>日期</th><th>类型</th><th>金额</th><th>分类</th><th>状态</th></tr></thead>
          <tbody>
            <tr v-for="r in preview.rows" :key="r.line">
              <td>{{ r.line }}</td><td>{{ r.date }}</td><td>{{ r.type }}</td><td>{{ r.amount }}</td><td>{{ r.category }}</td>
              <td :style="{ color: r.error ? 'var(--danger)' : 'var(--success)' }">{{ r.error || 'OK' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="preview?.valid" class="panel">
      <h3>3. 确认导入</h3>
      <div class="form-row"><label>记账人 ID</label><input v-model="signerId" /></div>
      <label><input type="checkbox" v-model="autoAnchor" /> 导入后自动封账锚定</label><br />
      <label><input type="checkbox" v-model="autoBackup" /> 封账后自动创建加密备份</label>
      <div v-if="autoBackup" class="form-row"><label>备份密码</label><input v-model="backupPassword" type="password" /></div>
      <button class="btn-primary" :disabled="busy" @click="commit">确认导入并上链</button>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const ledgers = ref([])
const ledgerId = ref(route.query.ledgerId || '')
const preview = ref(null)
const signerId = ref('')
const autoAnchor = ref(true)
const autoBackup = ref(true)
const backupPassword = ref('')
const error = ref('')
const msg = ref('')
const busy = ref(false)

onMounted(async () => {
  ledgers.value = await api.listLedgers()
  if (ledgerId.value) {
    const l = ledgers.value.find((x) => x.id === ledgerId.value)
    if (l?.members?.[0]) signerId.value = l.members[0].id
  }
})

async function downloadTpl() {
  const res = await api.downloadTemplate()
  const blob = await res.blob()
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = 'smart-ledger-import-template.xlsx'
  a.click()
}

async function onFile(ev) {
  const file = ev.target.files?.[0]
  if (!file || !ledgerId.value) {
    error.value = '请先选择账本'
    return
  }
  error.value = ''
  try {
    preview.value = await api.importPreview(ledgerId.value, file)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '解析失败'
  }
}

async function commit() {
  if (!preview.value?.rows?.length) return
  busy.value = true
  error.value = ''
  msg.value = ''
  try {
    const res = await api.importCommit(ledgerId.value, {
      signerId: signerId.value,
      rows: preview.value.rows,
      autoAnchor: autoAnchor.value,
      autoBackup: autoBackup.value,
      backupPassword: backupPassword.value,
    })
    const imp = res.import
    msg.value = `已导入 ${imp.imported} 条，跳过 ${imp.skipped} 条` + (imp.anchorTx ? ' · 已封账' : '') + (imp.backupRef ? ` · 备份 ${imp.backupRef}` : '')
    if (imp.backupError) msg.value += ` (备份失败: ${imp.backupError})`
    router.push(`/ledgers/${ledgerId.value}`)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '导入失败'
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.muted { color: var(--text-muted); font-size: 0.875rem; }
</style>
