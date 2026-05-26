<template>
  <div class="ledger-view-page">
    <div v-if="msg" class="alert alert-success">{{ msg }}</div>
    <div v-if="error" class="alert alert-error">{{ error }}</div>

    <div v-if="ledger.encryption?.enabled && !groupKeyReady" class="detail-card unlock-panel">
      <h3 class="detail-card__title">端到端加密</h3>
      <div class="form-row">
        <label>加密口令</label>
        <input v-model="e2ePassphrase" type="password" class="field-sm" />
        <button type="button" class="btn-primary" style="margin-top: 0.5rem" @click="unlockE2E">解锁</button>
      </div>
    </div>

    <LedgerContentPanel
      expanded
      :events="events"
      :schema="schema"
      :members="memberOptions"
      :group-key="groupKey"
      :loading="contentLoading"
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
import { computed, reactive, ref } from 'vue'
import { api, ApiError } from '../../api/http'
import AppIcon from '../../components/AppIcon.vue'
import EntryFormFields from '../../components/EntryFormFields.vue'
import LedgerContentPanel from '../../components/LedgerContentPanel.vue'
import { useAuthStore } from '../../stores/auth'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { emptyEntryData } from '../../utils/entrySchema'
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
  memberOptions,
  error,
  msg,
  load,
} = useLedgerDetail()

const busy = ref(false)
const e2ePassphrase = ref('')
const entryData = reactive({})
const showEntryModal = ref(false)

const groupKeyReady = computed(() => {
  if (!ledger.value?.encryption?.enabled) return true
  return !!groupKey.value
})
const canAddEntry = computed(() => groupKeyReady.value)
const entryButtonTitle = computed(() =>
  canAddEntry.value ? '记一笔' : '请先解锁加密账本后再记账'
)
const entryModalHint = computed(() => {
  if (ledger.value?.encryption?.enabled && !groupKeyReady.value) {
    return '该账本已启用端到端加密，请先解锁。'
  }
  if (ledger.value?.approvalPolicy?.enabled) {
    return '提交后将进入审批流程。'
  }
  return ''
})

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
.entry-modal {
  width: min(100%, 26rem);
}
.entry-modal h3 {
  margin: 0 0 1rem;
}
.entry-modal-hint {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: -0.5rem 0 1rem;
}
.entry-modal .modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 1.25rem;
}
</style>
