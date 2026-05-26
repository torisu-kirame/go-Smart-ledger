<template>
  <div class="ledger-settings">
    <div v-if="msg" class="alert alert-success">{{ msg }}</div>
    <div v-if="error" class="alert alert-error">{{ error }}</div>

    <section class="detail-card">
      <h3 class="detail-card__title">基础设置</h3>
      <div class="form-row">
        <label>账本名称</label>
        <input
          v-model="nameDraft"
          class="field-sm"
          :disabled="!isCreator || saving"
          maxlength="64"
        />
        <p v-if="!isCreator" class="field-hint">仅创建者可修改账本名称。</p>
      </div>
      <div class="form-row">
        <label>账本位置</label>
        <AppSelect
          :model-value="storageLoc"
          :options="storageOptions"
          :disabled="storageSaving"
          @update:model-value="onStorageChange"
        />
      </div>
      <button
        v-if="isCreator"
        type="button"
        class="btn-primary"
        :disabled="saving || !nameChanged"
        @click="saveName"
      >
        {{ saving ? '保存中…' : '保存名称' }}
      </button>
    </section>

    <section class="detail-card detail-card--danger">
      <h3 class="detail-card__title">注销账本</h3>
      <p class="danger-hint">
        注销后账本将从您的列表中移除，链上历史数据仍保留。此操作不可撤销，请确认已备份重要数据。
      </p>
      <template v-if="!isCreator">
        <p class="muted">仅账本创建者可以注销账本。</p>
      </template>
      <template v-else>
        <div v-if="!showArchiveConfirm" class="actions-row">
          <button type="button" class="btn-danger" @click="showArchiveConfirm = true">
            注销账本
          </button>
        </div>
        <div v-else class="archive-confirm">
          <p class="field-hint">请输入登录密码以确认注销：</p>
          <input v-model="confirmPassword" type="password" class="field-sm" autocomplete="current-password" />
          <div class="actions-row" style="margin-top: 0.75rem">
            <button type="button" class="btn-ghost" :disabled="archiving" @click="cancelArchive">
              取消
            </button>
            <button
              type="button"
              class="btn-danger"
              :disabled="archiving || !confirmPassword"
              @click="doArchive"
            >
              {{ archiving ? '注销中…' : '确认注销' }}
            </button>
          </div>
        </div>
      </template>
    </section>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '../../api/http'
import { useAuthStore } from '../../stores/auth'
import AppSelect from '../../components/AppSelect.vue'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { clearInfoUnlockSession } from '../../composables/useLedgerDetail'
import {
  normalizeStorageLocation,
  STORAGE_LOCATIONS,
} from '../../utils/ledgerStorage'

const router = useRouter()
const auth = useAuthStore()
const { ledgerId, ledger, error, msg } = useLedgerDetail()

const nameDraft = ref('')
const saving = ref(false)
const storageSaving = ref(false)
const showArchiveConfirm = ref(false)
const confirmPassword = ref('')
const archiving = ref(false)

const storageOptions = STORAGE_LOCATIONS.map((o) => ({ value: o.id, label: o.label }))
const storageLoc = computed(() => normalizeStorageLocation(ledger.value?.storageLocation))
const isCreator = computed(() => ledger.value?.creatorId === auth.user?.id)
const nameChanged = computed(
  () => nameDraft.value.trim() && nameDraft.value.trim() !== ledger.value?.name
)

watch(
  () => ledger.value?.name,
  (n) => {
    nameDraft.value = n || ''
  },
  { immediate: true }
)

async function saveName() {
  const name = nameDraft.value.trim()
  if (!name) return
  saving.value = true
  error.value = ''
  try {
    const updated = await api.updateLedger(ledgerId, { name })
    ledger.value = { ...ledger.value, ...updated }
    msg.value = '账本名称已更新'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '保存失败'
  } finally {
    saving.value = false
  }
}

async function onStorageChange(value) {
  const loc = normalizeStorageLocation(value)
  storageSaving.value = true
  error.value = ''
  try {
    const updated = await api.setLedgerStorageLocation(ledgerId, loc)
    ledger.value = { ...ledger.value, ...updated }
    msg.value = '账本位置已更新'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '更新失败'
  } finally {
    storageSaving.value = false
  }
}

function cancelArchive() {
  showArchiveConfirm.value = false
  confirmPassword.value = ''
}

async function doArchive() {
  archiving.value = true
  error.value = ''
  try {
    await api.verifyPassword(confirmPassword.value)
    await api.archiveLedger(ledgerId)
    clearInfoUnlockSession(ledgerId)
    await router.push('/ledgers')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '注销失败'
  } finally {
    archiving.value = false
  }
}
</script>

<style scoped>
.ledger-settings {
  max-width: 36rem;
  display: grid;
  gap: 1rem;
}
.detail-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1rem 1.1rem;
  box-shadow: var(--shadow-sm);
}
.detail-card--danger {
  border-color: color-mix(in srgb, var(--danger) 40%, var(--border));
}
.detail-card__title {
  margin: 0 0 0.85rem;
  font-size: 0.8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.danger-hint {
  margin: 0 0 0.85rem;
  font-size: 0.875rem;
  line-height: 1.45;
  color: var(--text-muted);
}
.field-hint {
  margin: 0.35rem 0 0;
  font-size: 0.8125rem;
  color: var(--text-muted);
}
.btn-danger {
  background: var(--danger);
  color: #fff;
  border: none;
  border-radius: var(--radius-sm);
  padding: 0.45rem 0.9rem;
  font-weight: 600;
  cursor: pointer;
}
.btn-danger:hover:not(:disabled) {
  filter: brightness(1.08);
}
.btn-danger:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
</style>
