<template>
  <div class="ledger-import">
    <section class="detail-card">
      <h3 class="detail-card__title">导入数据</h3>
      <p class="field-hint">
        支持 <strong>.xlsx</strong> / <strong>.csv</strong>。上传后可选择
        <strong>追加到已有表</strong>（自动补齐缺失字段）或 <strong>导入到新表</strong>。
        若尚未开启多表，导入时会自动开启。
      </p>
      <FileUploadZone
        block
        accept=".xlsx,.xls,.csv,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
        title="点击或拖拽上传 Excel 或 CSV"
        hint="首行作为列标题；数据从第二行起"
        :disabled="busy"
        @file="onUpload"
      />
    </section>

    <section v-if="preview" class="detail-card">
      <h3 class="detail-card__title">导入目标</h3>
      <div class="target-modes" role="radiogroup" aria-label="导入目标">
        <label class="target-mode" :class="{ active: targetMode === 'existing' }">
          <input
            v-model="targetMode"
            type="radio"
            value="existing"
            :disabled="!existingTables.length"
          />
          <span>导入到已有表</span>
        </label>
        <label class="target-mode" :class="{ active: targetMode === 'new' }">
          <input v-model="targetMode" type="radio" value="new" />
          <span>导入到新表</span>
        </label>
      </div>
      <p v-if="!existingTables.length" class="field-hint">当前账本尚无表，将创建新表。</p>
      <div v-if="targetMode === 'existing'" class="form-row">
        <label>选择表</label>
        <select v-model="selectedTableId" class="field-sm">
          <option v-for="t in existingTables" :key="t.id" :value="t.id">
            {{ t.name || t.id }}
          </option>
        </select>
      </div>
      <div v-else class="form-row">
        <label>新表名称</label>
        <input v-model="tableName" class="field-sm" placeholder="导入数据" />
      </div>
      <div class="import-meta muted-row">
        <span class="muted">文件类型：{{ preview.fileKind }}</span>
        <span v-if="preview.willEnableMultiTable" class="badge badge-pending">将自动开启多表</span>
      </div>
      <p class="muted">
        有效 {{ preview.valid }} 行 · 无效 {{ preview.invalid }} 行 · 共 {{ preview.total }} 行
      </p>
      <p class="field-hint">字段：{{ schemaLabels }}</p>
      <div class="table-wrap" style="max-height: 280px; overflow: auto">
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
              <td :style="{ color: r.error ? 'var(--danger)' : 'var(--success)' }">
                {{ r.error || 'OK' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section v-if="preview?.valid" class="detail-card">
      <h3 class="detail-card__title">确认导入</h3>
      <div v-if="needsSigner" class="form-row">
        <label>默认记账人 ID</label>
        <input v-model="signerId" placeholder="用户列未填时使用" class="field-sm" />
      </div>
      <label class="inline-check">
        <input v-model="autoAnchor" type="checkbox" />
        导入后自动封账锚定
      </label>
      <div class="actions-row" style="margin-top: 0.85rem">
        <button type="button" class="btn-primary" :disabled="busy || !canCommit" @click="commit">
          {{ busy ? '导入中…' : commitLabel }}
        </button>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '../../api/http'
import { useAuthStore } from '../../stores/auth'
import FileUploadZone from '../../components/FileUploadZone.vue'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { useNotify } from '../../composables/useNotify'
import { cellValue } from '../../utils/entrySchema'

const router = useRouter()
const auth = useAuthStore()
const notify = useNotify()
const { ledgerId, applyLedgerUpdate, activeTableId, tables } = useLedgerDetail()

const preview = ref(null)
const tableName = ref('')
const targetMode = ref('new')
const selectedTableId = ref('')
const signerId = ref(auth.user?.id || '')
const autoAnchor = ref(false)
const busy = ref(false)

const existingTables = computed(() => (tables.value || []).filter((t) => t?.id))
const previewFields = computed(() => preview.value?.entrySchema?.fields || [])
const schemaLabels = computed(() => previewFields.value.map((f) => f.label).join('、'))
const needsSigner = computed(() =>
  previewFields.value.some((f) => f.type === 'user' && f.key === 'bookkeeper')
)
const canCommit = computed(() => {
  if (!preview.value?.valid) return false
  if (targetMode.value === 'existing') return !!selectedTableId.value
  return true
})
const commitLabel = computed(() => {
  if (targetMode.value === 'existing') {
    const t = existingTables.value.find((x) => x.id === selectedTableId.value)
    return t ? `追加到「${t.name || t.id}」` : '确认导入'
  }
  const name = tableName.value.trim() || '导入数据'
  return `创建表「${name}」并导入`
})

watch(
  existingTables,
  (list) => {
    if (!list.length) {
      targetMode.value = 'new'
      selectedTableId.value = ''
      return
    }
    if (!selectedTableId.value || !list.some((t) => t.id === selectedTableId.value)) {
      selectedTableId.value =
        activeTableId.value && list.some((t) => t.id === activeTableId.value)
          ? activeTableId.value
          : list[0].id
    }
  },
  { immediate: true }
)

async function onUpload(file) {
  if (!file) return
  busy.value = true
  preview.value = null
  try {
    const res = await api.importAdaptivePreview(ledgerId.value, file)
    preview.value = res
    tableName.value = res.proposedTableName || ''
    if (existingTables.value.length) {
      targetMode.value = 'existing'
      selectedTableId.value =
        activeTableId.value && existingTables.value.some((t) => t.id === activeTableId.value)
          ? activeTableId.value
          : existingTables.value[0].id
    } else {
      targetMode.value = 'new'
    }
    notify.success('已解析文件，请确认导入目标与预览')
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '解析失败')
  } finally {
    busy.value = false
  }
}

async function commit() {
  if (!preview.value?.rows?.length || !canCommit.value) return
  busy.value = true
  try {
    const body = {
      signerId: signerId.value || auth.user?.id,
      entrySchema: preview.value.entrySchema,
      rows: (preview.value.rows || []).map((r) => ({
        line: r.line,
        cells: r.cells || {},
        error: r.error || '',
        date: r.date || '',
        type: r.type || '',
        amount: r.amount || '',
        category: r.category || '',
        note: r.note || '',
        counterparty: r.counterparty || '',
      })),
      autoAnchor: autoAnchor.value,
    }
    if (targetMode.value === 'existing') {
      body.tableId = selectedTableId.value
    } else {
      body.tableName = tableName.value.trim() || '导入数据'
    }
    const res = await api.importAdaptiveCommit(ledgerId.value, body)
    const a = res.adaptive
    const imported = Number(a?.import?.imported || 0)
    const skipped = Number(a?.import?.skipped || 0)
    if (imported <= 0) {
      notify.error(
        skipped
          ? `未写入任何行（跳过 ${skipped} 行），请检查必填字段与数据格式`
          : '未写入任何行'
      )
      return
    }
    const modeTip = a.mode === 'appended' ? '已追加到表' : '已导入新表'
    const skipTip = skipped ? `，跳过 ${skipped} 行` : ''
    notify.success(`${modeTip}「${a.tableName}」：${imported} 条${skipTip}`)
    if (a.tableId) activeTableId.value = a.tableId
    if (res.ledger) {
      await applyLedgerUpdate(res.ledger)
      if (a.tableId) activeTableId.value = a.tableId
    }
    router.push(`/ledgers/${ledgerId.value}/view`)
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '导入失败')
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.detail-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1rem 1.1rem;
  margin-bottom: 1rem;
  box-shadow: var(--shadow-sm);
}
.detail-card__title {
  margin: 0 0 0.65rem;
  font-size: 0.8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.field-hint {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: 0 0 0.75rem;
  line-height: 1.45;
}
.import-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
  margin-bottom: 0.65rem;
}
.import-meta input {
  min-width: 12rem;
}
.muted-row {
  margin-top: 0.35rem;
}
.target-modes {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 0.85rem;
}
.target-mode {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 0.875rem;
  cursor: pointer;
  background: var(--bg);
}
.target-mode.active {
  border-color: var(--accent);
  color: var(--accent);
  background: color-mix(in srgb, var(--accent) 8%, transparent);
}
.target-mode input {
  margin: 0;
}
.form-row {
  display: grid;
  gap: 0.35rem;
  margin-bottom: 0.75rem;
}
.form-row select.field-sm,
.form-row input.field-sm {
  max-width: 22rem;
}
.inline-check {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.875rem;
}
.link-accent {
  margin-left: auto;
  font-size: 0.8125rem;
  color: var(--accent);
  text-decoration: none;
}
.link-accent:hover {
  text-decoration: underline;
}
.bank-stmt h4 {
  margin: 0.35rem 0;
  font-size: 0.875rem;
  font-weight: 600;
}
</style>
