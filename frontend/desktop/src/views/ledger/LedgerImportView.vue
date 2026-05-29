<template>
  <div class="ledger-import">
    <!-- 简单流水：Excel / CSV 自适应导入 -->
    <template v-if="isSimpleLedger">
      <section class="detail-card">
        <h3 class="detail-card__title">导入数据</h3>
        <p class="field-hint">
          支持 <strong>.xlsx</strong> / <strong>.csv</strong>。将按文件首行自动生成字段结构，并
          <strong>新建一张表</strong>写入数据；若尚未开启多表，导入时会自动开启。
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
        <h3 class="detail-card__title">预览</h3>
        <div class="import-meta">
          <span>新表名称</span>
          <input v-model="tableName" class="field-sm" />
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
          <button type="button" class="btn-primary" :disabled="busy" @click="commit">
            {{ busy ? '导入中…' : '导入为新表' }}
          </button>
        </div>
      </section>
    </template>

    <!-- 专业复式：银行对账单等 -->
    <template v-else>
      <section class="detail-card">
        <h3 class="detail-card__title">导入数据</h3>
        <p class="field-hint">
          专业复式账本使用科目与凭证记账，<strong>不支持</strong>简单流水的 Excel 模板批量入账。
          可在此导入<strong>银行对账单 CSV</strong>，随后在「财务 → 对账」与链上凭证 Seq 匹配。
          凭证请直接在「财务 → 查看」录入。
        </p>
        <div class="form-row">
          <label>银行科目代码</label>
          <input v-model="bankAccountCode" class="field-sm mono" placeholder="1002" />
          <span class="muted">默认 1002（银行存款）</span>
        </div>
        <FileUploadZone
          block
          accept=".csv,text/csv"
          title="点击或拖拽上传银行 CSV"
          hint="需含日期、金额列；可选摘要列"
          :disabled="busy"
          @file="onBankUpload"
        />
      </section>

      <section v-if="bankStatements.length" class="detail-card">
        <div class="import-meta">
          <h3 class="detail-card__title">已导入对账单</h3>
          <router-link :to="`${basePath}/accounting/bank`" class="link-accent">
            前往对账匹配 →
          </router-link>
        </div>
        <div v-for="stmt in bankStatements" :key="stmt.id" class="bank-stmt">
          <h4>{{ stmt.filename || stmt.id }} · {{ stmt.lines?.length || 0 }} 笔</h4>
        </div>
      </section>
    </template>
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
const {
  ledgerId,
  applyLedgerUpdate,
  activeTableId,
  isSimpleLedger,
  isProfessionalLedger,
} = useLedgerDetail()

const basePath = computed(() => `/ledgers/${ledgerId.value}`)

const preview = ref(null)
const tableName = ref('')
const signerId = ref(auth.user?.id || '')
const autoAnchor = ref(false)
const busy = ref(false)
const bankAccountCode = ref('1002')
const bankStatements = ref([])

const previewFields = computed(() => preview.value?.entrySchema?.fields || [])
const schemaLabels = computed(() => previewFields.value.map((f) => f.label).join('、'))
const needsSigner = computed(() =>
  previewFields.value.some((f) => f.type === 'user' && f.key === 'bookkeeper')
)

async function loadBankStatements() {
  if (!isProfessionalLedger.value || !ledgerId.value) return
  try {
    const res = await api.listBankStatements(ledgerId.value)
    bankStatements.value = res.statements || res || []
  } catch {
    bankStatements.value = []
  }
}

watch(
  () => [ledgerId.value, isProfessionalLedger.value],
  () => loadBankStatements(),
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
    notify.success('已解析文件，请确认预览')
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '解析失败')
  } finally {
    busy.value = false
  }
}

async function commit() {
  if (!preview.value?.rows?.length) return
  busy.value = true
  try {
    const res = await api.importAdaptiveCommit(ledgerId.value, {
      signerId: signerId.value || auth.user?.id,
      tableName: tableName.value.trim(),
      entrySchema: preview.value.entrySchema,
      rows: preview.value.rows,
      autoAnchor: autoAnchor.value,
    })
    const a = res.adaptive
    notify.success(`已导入新表「${a.tableName}」：${a.import.imported} 条`)
    if (res.ledger) {
      await applyLedgerUpdate(res.ledger)
      activeTableId.value = a.tableId
    }
    router.push(`/ledgers/${ledgerId.value}/view`)
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '导入失败')
  } finally {
    busy.value = false
  }
}

async function onBankUpload(file) {
  if (!file) return
  busy.value = true
  try {
    await api.importBankStatement(ledgerId.value, file, bankAccountCode.value.trim() || '1002')
    notify.success('对账单已导入')
    await loadBankStatements()
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
.form-row {
  display: grid;
  gap: 0.35rem;
  margin-bottom: 0.75rem;
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
