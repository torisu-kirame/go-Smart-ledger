<template>
  <div class="ledger-view-page">
    <div v-if="ledger.encryption?.enabled && !groupKeyReady" class="detail-card unlock-panel">
      <h3 class="detail-card__title">端到端加密</h3>
      <div class="form-row">
        <label>加密口令</label>
        <input v-model="e2ePassphrase" type="password" class="field-sm" />
        <button type="button" class="btn-primary" style="margin-top: 0.5rem" @click="unlockE2E">解锁</button>
      </div>
    </div>

    <nav v-if="hasSheets" class="table-tabs" aria-label="账本表">
      <button
        v-for="t in tables"
        :key="t.id"
        type="button"
        class="table-tab"
        :class="{ active: activeTableId === t.id }"
        @click="activeTableId = t.id"
      >
        {{ t.name }}
      </button>
      <button
        v-if="isCreator"
        type="button"
        class="table-tab table-tab--add"
        title="新建 Sheet"
        @click="openSheetModal()"
      >
        <AppIcon name="plus" size="sm" />
        新建 Sheet
      </button>
      <button
        v-if="isCreator && activeTable"
        type="button"
        class="table-tab table-tab--ghost"
        title="编辑当前 Sheet 字段"
        @click="openSheetModal(activeTable)"
      >
        编辑字段
      </button>
    </nav>

    <div v-if="!hasSheets" class="empty-sheets panel">
      <AppIcon name="grid" size="md" class="empty-sheets__icon" />
      <h3>还没有 Sheet</h3>
      <p class="muted">新建账本默认不含工作表。请先创建 Sheet，并自定义列字段后再记账。</p>
      <button
        v-if="isCreator"
        type="button"
        class="btn-primary"
        @click="openSheetModal()"
      >
        <AppIcon name="plus" size="sm" />
        创建第一个 Sheet
      </button>
      <p v-else class="muted">请联系创建者添加 Sheet。</p>
    </div>

    <LedgerContentPanel
      v-else
      expanded
      :events="events"
      :schema="schema"
      :table-id="activeTableId"
      :members="memberOptions"
      :group-key="groupKey"
      :loading="contentLoading"
    />

    <LedgerTableAttachments
      v-if="isSimpleLedger && hasSheets"
      :ledger-id="ledgerId"
      :table-id="activeTableId"
    />

    <button
      type="button"
      class="fab-entry btn-primary"
      :disabled="!canAddEntry || busy"
      :title="entryButtonTitle"
      @click="openEntryModal"
    >
      <AppIcon name="plus" size="sm" />
      <span>记一笔</span>
    </button>

    <div v-if="showSheetModal" class="modal" @click.self="closeSheetModal">
      <form class="modal-card sheet-modal" @submit.prevent="saveSheet">
        <h3>{{ editingSheetId ? '编辑 Sheet' : '新建 Sheet' }}</h3>
        <div class="form-row">
          <label>名称</label>
          <input v-model="sheetForm.name" required maxlength="64" placeholder="例如：日常开支、差旅" />
        </div>
        <div class="form-row">
          <label>字段（列）</label>
          <SchemaFieldsEditor v-model="sheetForm.fields" :disabled="sheetSaving" />
        </div>
        <div class="modal-actions">
          <button type="button" class="btn-ghost" :disabled="sheetSaving" @click="closeSheetModal">取消</button>
          <button type="submit" class="btn-primary" :disabled="sheetSaving">
            {{ sheetSaving ? '保存中…' : editingSheetId ? '保存' : '创建' }}
          </button>
        </div>
      </form>
    </div>

    <div v-if="showEntryModal" class="modal" @click.self="closeEntryModal">
      <form class="modal-card entry-modal" @submit.prevent="addEntry">
        <h3>记一笔</h3>
        <p v-if="entryModalHint" class="entry-modal-hint">{{ entryModalHint }}</p>
        <EntryFormFields column :schema="schema" :model="entryData" :members="memberOptions" />
        <div class="modal-actions">
          <button type="button" class="btn-ghost" @click="closeEntryModal">取消</button>
          <button type="submit" class="btn-primary" :disabled="busy || !canAddEntry">
            {{ busy ? '提交中…' : '记入' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { api, ApiError } from '../../api/http'
import AppIcon from '../../components/AppIcon.vue'
import EntryFormFields from '../../components/EntryFormFields.vue'
import LedgerContentPanel from '../../components/LedgerContentPanel.vue'
import SchemaFieldsEditor from '../../components/SchemaFieldsEditor.vue'
import LedgerTableAttachments from '../../components/ledger/LedgerTableAttachments.vue'
import { useAuthStore } from '../../stores/auth'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { blankFieldRows, emptyEntryData, normalizeSchemaFields } from '../../utils/entrySchema'
import {
  encryptEntryData,
  saveLocalGroupKey,
  unwrapGroupKey,
} from '../../utils/e2eCrypto'

const auth = useAuthStore()
const {
  ledgerId,
  ledger,
  events,
  groupKey,
  contentLoading,
  schema,
  activeTableId,
  tables,
  isSimpleLedger,
  memberOptions,
  error,
  msg,
  load,
  applyLedgerUpdate,
} = useLedgerDetail()

const busy = ref(false)
const e2ePassphrase = ref('')
const entryData = reactive({})
const showEntryModal = ref(false)
const showSheetModal = ref(false)
const sheetSaving = ref(false)
const editingSheetId = ref('')
const sheetForm = reactive({
  name: '',
  fields: blankFieldRows(3),
})

const groupKeyReady = computed(() => {
  if (!ledger.value?.encryption?.enabled) return true
  return !!groupKey.value
})
const isCreator = computed(() => ledger.value?.creatorId === auth.user?.id)
const hasSheets = computed(() => (tables.value?.length || 0) > 0)
const activeTable = computed(() => tables.value.find((t) => t.id === activeTableId.value) || null)
const schemaReady = computed(() => (schema.value?.fields || []).length > 0)
const canAddEntry = computed(() => groupKeyReady.value && hasSheets.value && schemaReady.value)
const entryButtonTitle = computed(() => {
  if (!hasSheets.value) return '请先创建 Sheet'
  if (!schemaReady.value) return '请先为 Sheet 添加字段'
  if (!groupKeyReady.value) return '请先解锁加密账本后再记账'
  return '记一笔'
})
const entryModalHint = computed(() => {
  if (ledger.value?.encryption?.enabled && !groupKeyReady.value) {
    return '该账本已启用端到端加密，请先解锁。'
  }
  if (ledger.value?.approvalPolicy?.enabled) {
    return '提交后将进入审批流程。'
  }
  return ''
})

watch(
  tables,
  (list) => {
    if (!list?.length) return
    if (!list.some((t) => t.id === activeTableId.value)) {
      activeTableId.value = list[0].id
    }
  },
  { immediate: true }
)

function openSheetModal(table = null) {
  editingSheetId.value = table?.id || ''
  sheetForm.name = table?.name || ''
  const fields = table?.entrySchema?.fields
  sheetForm.fields = fields?.length
    ? fields.map((f) => ({ ...f }))
    : blankFieldRows(3)
  showSheetModal.value = true
}

function closeSheetModal() {
  if (sheetSaving.value) return
  showSheetModal.value = false
  editingSheetId.value = ''
}

async function saveSheet() {
  const name = sheetForm.name.trim()
  const fields = normalizeSchemaFields(sheetForm.fields)
  if (!name) {
    error.value = '请填写 Sheet 名称'
    return
  }
  if (!fields.length) {
    error.value = '请至少定义 1 个字段'
    return
  }
  sheetSaving.value = true
  error.value = ''
  try {
    const entrySchema = { templateId: 'custom', fields }
    let updated
    if (editingSheetId.value) {
      updated = await api.updateLedgerTable(ledgerId.value, editingSheetId.value, {
        name,
        entrySchema,
      })
      msg.value = `已更新「${name}」`
    } else {
      if (!ledger.value?.multiTableEnabled) {
        await api.setLedgerMultiTable(ledgerId.value, true)
      }
      updated = await api.createLedgerTable(ledgerId.value, { name, entrySchema })
      msg.value = `已创建 Sheet「${name}」`
    }
    await applyLedgerUpdate(updated)
    const created = (updated?.tables || []).find((t) => t.name === name)
    if (created?.id) activeTableId.value = created.id
    showSheetModal.value = false
    editingSheetId.value = ''
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '保存失败'
  } finally {
    sheetSaving.value = false
  }
}

function openEntryModal() {
  if (!canAddEntry.value) return
  initEntryDefaults()
  showEntryModal.value = true
}

function closeEntryModal() {
  if (busy.value) return
  showEntryModal.value = false
}

function initEntryDefaults() {
  const uid = auth.user?.id || ''
  const defaults = { bookkeeper: uid, date: new Date().toISOString().slice(0, 10) }
  const data = emptyEntryData(schema.value, defaults)
  Object.keys(entryData).forEach((k) => delete entryData[k])
  Object.assign(entryData, data)
}

async function unlockE2E() {
  if (!ledger.value?.encryption?.enabled) return
  const uid = auth.user?.id
  const wrapped = ledger.value.encryption.wrappedKeys?.[uid]
  if (!wrapped) {
    error.value = '未找到您的密钥包装'
    return
  }
  try {
    groupKey.value = await unwrapGroupKey(wrapped, e2ePassphrase.value, ledgerId.value, uid)
    saveLocalGroupKey(ledgerId.value, groupKey.value)
    msg.value = '已解锁'
  } catch {
    error.value = '口令错误或密钥损坏'
  }
}

async function buildEntryPayload() {
  let data = { ...entryData }
  if (ledger.value?.encryption?.enabled && groupKey.value) {
    data = await encryptEntryData(groupKey.value, data)
  }
  return {
    signerId: entryData.bookkeeper || auth.user?.id || '',
    tableId: activeTableId.value,
    schemaId: schema.value.templateId,
    data,
  }
}

async function addEntry() {
  busy.value = true
  error.value = ''
  msg.value = ''
  try {
    const entry = await buildEntryPayload()
    if (ledger.value?.approvalPolicy?.enabled) {
      const res = await api.proposeEntry(ledgerId.value, entry)
      msg.value = res.status === 'committed' ? '记账已上链' : '已提交审批'
    } else {
      await api.appendEntry(ledgerId.value, entry)
      msg.value = '记账成功'
    }
    await load()
    showEntryModal.value = false
    if (msg.value.includes('成功') || msg.value.includes('上链')) {
      msg.value += ' · 已刷新账本内容'
    }
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '失败'
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.ledger-view-page {
  position: relative;
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - 12rem);
  padding-bottom: 4.5rem;
}
.detail-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1rem 1.1rem;
  margin-bottom: 1rem;
  box-shadow: var(--shadow-sm);
}
.detail-card__title {
  margin: 0 0 0.75rem;
  font-size: 0.8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.unlock-panel {
  flex-shrink: 0;
}
.ledger-view-page :deep(.ledger-content.panel) {
  flex: 1;
  margin-bottom: 0;
  box-shadow: var(--shadow-sm);
}
.empty-sheets {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.65rem;
  text-align: center;
  padding: 2.5rem 1.5rem;
  min-height: min(48vh, 420px);
}
.empty-sheets h3 {
  margin: 0;
  font-size: 1.05rem;
}
.empty-sheets__icon {
  color: var(--accent);
  opacity: 0.85;
}
.empty-sheets .btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  margin-top: 0.5rem;
}
.fab-entry {
  position: fixed;
  right: 2rem;
  bottom: 2rem;
  z-index: 40;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.65rem 1.15rem;
  border-radius: 999px;
  border: none;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.35), 0 0 0 1px color-mix(in srgb, var(--accent) 40%, transparent);
  font-size: 0.9375rem;
  font-weight: 600;
}
.fab-entry:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.entry-modal,
.sheet-modal {
  width: min(100%, 32rem);
}
.entry-modal h3,
.sheet-modal h3 {
  margin: 0 0 1rem;
}
.entry-modal-hint {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: -0.5rem 0 1rem;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 1.25rem;
}
.table-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-bottom: 0.75rem;
  align-items: center;
}
.table-tab {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.4rem 0.85rem;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  color: var(--text-muted);
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
}
.table-tab.active {
  color: var(--accent);
  border-color: color-mix(in srgb, var(--accent) 50%, var(--border));
  background: var(--accent-soft);
}
.table-tab--add {
  border-style: dashed;
  color: var(--accent);
}
.table-tab--ghost {
  background: transparent;
}
.form-row {
  display: grid;
  gap: 0.35rem;
  margin-bottom: 0.85rem;
}
.form-row label {
  font-size: 0.8125rem;
  color: var(--text-muted);
}
</style>
