<template>
  <div class="page">
    <PageHeader :crumbs="crumbs" />
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="msg" class="alert alert-success">{{ msg }}</div>

    <div class="panel">
      <button class="btn-ghost" :disabled="!ledgerId" @click="downloadTpl">下载导入模板</button>
    </div>

    <div class="panel">
      <h3>1. 选择账本</h3>
      <AppSelect
        v-model="ledgerId"
        sm
        placeholder="请选择账本"
        :options="ledgerOptions"
        @change="onLedgerChange"
      />
    </div>

    <div v-if="currentLedger && isProfessionalLedger" class="alert alert-error">
      所选账本为<strong>专业复式</strong>模式，请使用账本「财务」页录入凭证；Excel 流水导入仅适用于简单流水账本。
    </div>

    <div v-if="currentLedger && isMultiTable" class="panel">
      <h3>2. 选择导入目标表</h3>
      <p class="field-hint">多表账本将按工作表名称匹配表名；也可指定单表导入。</p>
      <AppSelect
        v-model="importTableId"
        sm
        placeholder="全部工作表（按表名匹配）"
        :options="importTableOptions"
        @change="preview = null"
      />
    </div>

    <div class="panel">
      <h3>{{ isMultiTable ? '3' : '2' }}. 上传并预览</h3>
      <FileUploadZone
        block
        accept=".xlsx,.xls,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.ms-excel"
        title="点击或拖拽上传 Excel"
        hint="支持 .xlsx / .xls"
        :disabled="!ledgerId || isProfessionalLedger"
        @file="onUploadFile"
      />
    </div>

    <template v-if="multiSheetPreview.length">
      <div v-for="sh in multiSheetPreview" :key="sh.tableId" class="panel">
        <h3>预览 · {{ sh.tableName }}</h3>
        <p class="muted">有效 {{ sh.valid }} / 无效 {{ sh.invalid }}（sheet: {{ sh.sheetName || '—' }}）</p>
        <div class="table-wrap" style="max-height:240px;overflow:auto">
          <table>
            <thead>
              <tr>
                <th>行</th>
                <th v-for="f in sh.entrySchema?.fields || []" :key="f.key">{{ f.label }}</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in sh.rows" :key="r.line">
                <td>{{ r.line }}</td>
                <td v-for="f in sh.entrySchema?.fields || []" :key="f.key">{{ cellValue(r, f.key) }}</td>
                <td :style="{ color: r.error ? 'var(--danger)' : 'var(--success)' }">
                  {{ r.error || 'OK' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <button
          v-if="sh.valid > 0"
          type="button"
          class="btn-primary"
          style="margin-top: 0.65rem"
          :disabled="busy"
          @click="commitSheet(sh)"
        >
          导入此表（{{ sh.valid }} 条）
        </button>
      </div>
    </template>

    <div v-else-if="preview?.rows?.length" class="panel">
      <h3>预览</h3>
      <div class="table-wrap" style="max-height:280px;overflow:auto">
        <table>
          <thead>
            <tr>
              <th>行</th>
              <th v-for="f in previewFields" :key="f.key">{{ f.label }}</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in preview.rows" :key="r.line">
              <td>{{ r.line }}</td>
              <td v-for="f in previewFields" :key="f.key">{{ cellValue(r, f.key) }}</td>
              <td :style="{ color: r.error ? 'var(--danger)' : 'var(--success)' }">{{ r.error || 'OK' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="canCommitSingle" class="panel">
      <h3>{{ isMultiTable ? '4' : '3' }}. 确认导入</h3>
      <div v-if="needsSigner" class="form-row">
        <label>默认记账人 ID</label>
        <input v-model="signerId" placeholder="Excel 未填记账人列时使用" />
      </div>
      <label class="inline-check"><input type="checkbox" v-model="autoAnchor" /> 导入后自动封账锚定</label>
      <label class="inline-check"><input type="checkbox" v-model="autoBackup" /> 封账后自动创建加密备份</label>
      <div v-if="autoBackup" class="form-row"><label>备份密码</label><input v-model="backupPassword" type="password" /></div>
      <button class="btn-primary" :disabled="busy" @click="commit">确认导入并上链</button>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'
import AppSelect from '../components/AppSelect.vue'
import PageHeader from '../components/PageHeader.vue'
import { usePageCrumbs } from '../composables/usePageCrumbs'
import FileUploadZone from '../components/FileUploadZone.vue'
import { cellValue, resolveSchema } from '../utils/entrySchema'
import { isProfessionalBookkeeping } from '../utils/bookkeepingMode'
import { isMultiTableLedger, ledgerTables } from '../utils/ledgerTables'
import { useNotify } from '../composables/useNotify'

const route = useRoute()
const { crumbs } = usePageCrumbs()
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
const importTableId = ref('')
const multiSheetPreview = ref([])
const notify = useNotify()

const currentLedger = computed(() => ledgers.value.find((x) => x.id === ledgerId.value))
const isProfessionalLedger = computed(() => isProfessionalBookkeeping(currentLedger.value))
const isMultiTable = computed(() => isMultiTableLedger(currentLedger.value))
const importTableOptions = computed(() => [
  { value: '', label: '全部工作表（按表名匹配）' },
  ...ledgerTables(currentLedger.value).map((t) => ({ value: t.id, label: t.name })),
])
const canCommitSingle = computed(
  () => preview.value?.valid > 0 && !multiSheetPreview.value.length
)
const currentSchema = computed(() =>
  preview.value?.entrySchema || resolveSchema(currentLedger.value)
)
const previewFields = computed(() => currentSchema.value?.fields || [])
const schemaLabels = computed(() => previewFields.value.map((f) => f.label).join('、'))
const needsSigner = computed(() =>
  previewFields.value.some((f) => f.type === 'user' && f.key === 'bookkeeper')
)

const ledgerOptions = computed(() =>
  ledgers.value.map((l) => ({
    value: l.id,
    label: `${l.name} (${l.id.slice(0, 8)}…)`,
  }))
)

function onLedgerChange() {
  preview.value = null
  multiSheetPreview.value = []
  importTableId.value = ''
  if (ledgerId.value && auth.user?.id) signerId.value = auth.user.id
}

onMounted(async () => {
  ledgers.value = await api.listLedgers()
  onLedgerChange()
})

async function downloadTpl() {
  if (!ledgerId.value) {
    error.value = '请先选择账本'
    return
  }
  const res = await api.downloadTemplate(ledgerId.value)
  const blob = await res.blob()
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = 'smart-ledger-import-template.xlsx'
  a.click()
}

async function onUploadFile(file) {
  if (!file || !ledgerId.value) {
    error.value = '请先选择账本'
    return
  }
  error.value = ''
  try {
    preview.value = null
    multiSheetPreview.value = []
    const res = await api.importPreview(
      ledgerId.value,
      file,
      importTableId.value || ''
    )
    if (res.sheets?.length) {
      multiSheetPreview.value = res.sheets
      notify.success('已解析多表工作簿')
    } else {
      preview.value = res
    }
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '解析失败'
    notify.error(error.value)
  }
}

async function commitSheet(sh) {
  if (!sh?.rows?.length || !sh.valid) return
  busy.value = true
  error.value = ''
  try {
    const res = await api.importCommit(ledgerId.value, {
      signerId: signerId.value || auth.user?.id,
      tableId: sh.tableId,
      rows: sh.rows.filter((r) => !r.error),
      autoAnchor: autoAnchor.value,
      autoBackup: autoBackup.value,
      backupPassword: backupPassword.value,
    })
    const imp = res.import
    notify.success(`表「${sh.tableName}」已导入 ${imp.imported} 条`)
    router.push(`/ledgers/${ledgerId.value}/view`)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '导入失败'
    notify.error(error.value)
  } finally {
    busy.value = false
  }
}

async function commit() {
  if (!preview.value?.rows?.length) return
  busy.value = true
  error.value = ''
  msg.value = ''
  try {
    const res = await api.importCommit(ledgerId.value, {
      signerId: signerId.value || auth.user?.id,
      tableId: preview.value.tableId || importTableId.value || '',
      rows: preview.value.rows,
      autoAnchor: autoAnchor.value,
      autoBackup: autoBackup.value,
      backupPassword: backupPassword.value,
    })
    const imp = res.import
    msg.value = `已导入 ${imp.imported} 条，跳过 ${imp.skipped} 条` + (imp.anchorTx ? ' · 已封账' : '') + (imp.backupRef ? ` · 备份 ${imp.backupRef}` : '')
    if (imp.backupError) msg.value += ` (备份失败: ${imp.backupError})`
    notify.success(msg.value)
    router.push(`/ledgers/${ledgerId.value}/view`)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '导入失败'
    notify.error(error.value)
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.muted { color: var(--text-muted); font-size: 0.875rem; }
</style>
