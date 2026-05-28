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

    <section v-if="isSimpleLedger" class="detail-card">
      <h3 class="detail-card__title">多表（F49）</h3>
      <p class="field-hint">
        开启后可在账本内创建多张表（类似 Excel 多 sheet），流水、导入与附件均按表隔离。
      </p>
      <template v-if="isCreator">
        <div class="toggle-row">
          <span class="toggle-row__text">启用多表</span>
          <ToggleSwitch
            v-model="multiTableOn"
            :disabled="tableSaving"
            aria-label="启用多表"
            @update:model-value="onMultiTableToggle"
          />
        </div>
        <template v-if="ledger.multiTableEnabled">
          <ul class="table-list">
            <li v-for="t in ledger.tables || []" :key="t.id">
              <strong>{{ t.name }}</strong>
              <span class="mono muted">{{ t.id }}</span>
              <span class="muted">{{ (t.entrySchema?.fields || []).length }} 列</span>
            </li>
          </ul>
          <div v-if="ledger.tables?.length > 1" class="form-row" style="margin-top: 0.75rem">
            <label>新表名称</label>
            <input v-model="newTableName" class="field-sm" placeholder="例如：差旅、采购" />
            <button
              type="button"
              class="btn-primary"
              style="margin-top: 0.5rem"
              :disabled="tableSaving || !newTableName.trim()"
              @click="createTable"
            >
              添加表
            </button>
          </div>
          <p v-else class="field-hint">已启用多表。添加第二张表后即可在「流水」页切换表签。</p>
        </template>
      </template>
      <p v-else class="field-hint">仅创建者可修改多表设置。</p>
    </section>

    <section class="detail-card">
      <h3 class="detail-card__title">安全与协作</h3>

      <div v-if="isMulti && isSimpleLedger" class="policy-block">
        <h4 class="policy-block__heading">审批策略</h4>
        <p class="field-hint">
          启用后，记账需经全体成员批准后方可上链（批准人数等于当前成员总数）。
        </p>
        <template v-if="isCreator">
          <div class="toggle-row">
            <span class="toggle-row__text">启用审批流</span>
            <ToggleSwitch
              v-model="approvalEnabled"
              :disabled="policySaving"
              aria-label="启用审批流"
              @update:model-value="onApprovalToggle"
            />
          </div>
          <p v-if="approvalEnabled" class="field-hint" style="margin-top: 0.5rem">
            当前需 {{ memberCount }} 名成员全部批准（与成员总数一致）。
          </p>
        </template>
        <template v-else>
          <p class="status-line">
            {{
              ledger.approvalPolicy?.enabled
                ? `已启用（需 ${ledger.approvalPolicy.threshold} 人批准）`
                : '未启用'
            }}
          </p>
          <p class="field-hint">仅创建者可修改。</p>
        </template>
      </div>

      <div class="policy-block">
        <h4 class="policy-block__heading">端到端加密</h4>
        <p class="field-hint">
          启用后记账明细在客户端加密后再提交，服务端不保存明文。口令仅保存在本机，遗失将无法解密历史数据。
        </p>
        <template v-if="ledger.encryption?.enabled">
          <div class="status-line badge-wrap">
            <span class="badge badge-ok">已启用</span>
            <span class="muted">{{ ledger.encryption.algo || 'aes-gcm-v1' }}</span>
            <button
              v-if="canOpenPassphraseView"
              type="button"
              class="btn-icon"
              title="查看加密口令"
              aria-label="查看加密口令"
              @click="openPassphraseModal"
            >
              <AppIcon name="eye" size="sm" />
            </button>
          </div>
          <div v-if="isCreator" class="toggle-row" style="margin-top: 0.65rem">
            <span class="toggle-row__text">允许账本成员查看加密口令</span>
            <ToggleSwitch
              v-model="passphraseViewEnabled"
              :disabled="policySaving"
              aria-label="允许账本成员查看加密口令"
              @update:model-value="onPassphraseViewToggle"
            />
          </div>
          <p v-else-if="ledger.encryption.passphraseViewEnabled" class="field-hint">
            创建者已允许成员在验证登录密码后查看加密口令。
          </p>
        </template>
        <template v-else-if="isCreator">
          <div class="toggle-row">
            <span class="toggle-row__text">启用组级端到端加密</span>
            <ToggleSwitch
              v-model="enableE2E"
              :disabled="policySaving"
              aria-label="启用组级端到端加密"
            />
          </div>
          <div v-if="enableE2E" class="form-row" style="margin-top: 0.65rem">
            <label>加密口令（账本密码）</label>
            <input
              v-model="e2ePassphrase"
              type="password"
              class="field-sm"
              placeholder="为每位成员包装组密钥，请妥善保管"
              autocomplete="new-password"
            />
          </div>
          <button
            type="button"
            class="btn-primary"
            style="margin-top: 0.65rem"
            :disabled="policySaving || !canEnableE2E"
            @click="saveEncryption"
          >
            {{ policySaving ? '启用中…' : '启用加密' }}
          </button>
        </template>
        <template v-else>
          <p class="status-line">未启用</p>
          <p class="field-hint">仅创建者可启用。</p>
        </template>
      </div>
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

    <div v-if="showPassphraseModal" class="modal" @click.self="closePassphraseModal">
      <div class="modal-card">
        <h3 class="modal-title">查看加密口令</h3>
        <p v-if="viewError" class="alert alert-error">{{ viewError }}</p>

        <template v-if="viewStep === 'password'">
          <p class="field-hint">请输入您当前账号的登录密码以验证身份。</p>
          <div class="form-row">
            <label>登录密码</label>
            <input
              v-model="viewLoginPassword"
              type="password"
              class="field-sm"
              autocomplete="current-password"
              @keyup.enter="confirmViewPassphrase"
            />
          </div>
          <div class="modal-actions">
            <button type="button" class="btn-ghost" @click="closePassphraseModal">取消</button>
            <button
              type="button"
              class="btn-primary"
              :disabled="viewLoading || !viewLoginPassword"
              @click="confirmViewPassphrase"
            >
              {{ viewLoading ? '验证中…' : '确认' }}
            </button>
          </div>
        </template>

        <template v-else-if="viewStep === 'register'">
          <p class="field-hint">
            尚未登记口令副本。请输入账本加密口令，系统将以您的登录密码加密后保存，便于日后查看。
          </p>
          <div class="form-row">
            <label>账本加密口令</label>
            <input v-model="registerLedgerPassphrase" type="password" class="field-sm" autocomplete="off" />
          </div>
          <div class="modal-actions">
            <button type="button" class="btn-ghost" @click="closePassphraseModal">取消</button>
            <button
              type="button"
              class="btn-primary"
              :disabled="viewLoading || registerLedgerPassphrase.length < 6"
              @click="registerAndRevealPassphrase"
            >
              {{ viewLoading ? '保存中…' : '登记并查看' }}
            </button>
          </div>
        </template>

        <template v-else-if="viewStep === 'revealed'">
          <div class="form-row">
            <label>加密口令</label>
            <input :value="revealedPassphrase" type="text" class="field-sm" readonly />
          </div>
          <p class="field-hint">请妥善保管，遗失将无法解密历史记账数据。</p>
          <div class="modal-actions">
            <button type="button" class="btn-primary" @click="closePassphraseModal">关闭</button>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '../../api/http'
import { useAuthStore } from '../../stores/auth'
import AppSelect from '../../components/AppSelect.vue'
import AppIcon from '../../components/AppIcon.vue'
import ToggleSwitch from '../../components/ToggleSwitch.vue'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { clearInfoUnlockSession } from '../../composables/useLedgerDetail'
import {
  normalizeStorageLocation,
  STORAGE_LOCATIONS,
} from '../../utils/ledgerStorage'
import {
  buildEncryptionForCreate,
  saveLocalGroupKey,
  saveLocalPassphrase,
  loadLocalPassphrase,
  wrapPassphraseForLoginView,
  unwrapPassphraseForLoginView,
} from '../../utils/e2eCrypto'
import { DEFAULT_ENTRY_SCHEMA } from '../../utils/entrySchema'

const router = useRouter()
const auth = useAuthStore()
const { ledgerId, ledger, error, msg, isSimpleLedger, applyLedgerUpdate } = useLedgerDetail()

const nameDraft = ref('')
const saving = ref(false)
const storageSaving = ref(false)
const policySaving = ref(false)
const approvalEnabled = ref(false)
const passphraseViewEnabled = ref(false)
const enableE2E = ref(false)
const e2ePassphrase = ref('')
const showArchiveConfirm = ref(false)
const confirmPassword = ref('')
const archiving = ref(false)
const tableSaving = ref(false)
const multiTableOn = ref(false)
const newTableName = ref('')

const showPassphraseModal = ref(false)
const viewStep = ref('password')
const viewLoginPassword = ref('')
const viewError = ref('')
const viewLoading = ref(false)
const revealedPassphrase = ref('')
const registerLedgerPassphrase = ref('')

const storageOptions = STORAGE_LOCATIONS.map((o) => ({ value: o.id, label: o.label }))
const storageLoc = computed(() => normalizeStorageLocation(ledger.value?.storageLocation))
const isCreator = computed(() => ledger.value?.creatorId === auth.user?.id)
const isMulti = computed(() => ledger.value?.type === 'multi')
const memberCount = computed(() => ledger.value?.members?.length || 1)
const nameChanged = computed(
  () => nameDraft.value.trim() && nameDraft.value.trim() !== ledger.value?.name
)

const canOpenPassphraseView = computed(() => {
  if (!ledger.value?.encryption?.enabled) return false
  if (isCreator.value) return true
  return !!ledger.value.encryption.passphraseViewEnabled
})

const canEnableE2E = computed(
  () => enableE2E.value && e2ePassphrase.value.length >= 6 && !ledger.value?.encryption?.enabled
)

watch(
  () => ledger.value?.name,
  (n) => {
    nameDraft.value = n || ''
  },
  { immediate: true }
)

watch(
  () => ledger.value?.approvalPolicy,
  (ap) => {
    approvalEnabled.value = !!ap?.enabled
  },
  { immediate: true, deep: true }
)

watch(
  () => ledger.value?.encryption?.passphraseViewEnabled,
  (v) => {
    passphraseViewEnabled.value = !!v
  },
  { immediate: true }
)

watch(
  () => ledger.value?.multiTableEnabled,
  (v) => {
    multiTableOn.value = !!v
  },
  { immediate: true }
)

async function onMultiTableToggle(enabled) {
  tableSaving.value = true
  error.value = ''
  try {
    const updated = await api.setLedgerMultiTable(ledgerId.value, enabled)
    await applyLedgerUpdate(updated)
    msg.value = enabled ? '已启用多表' : '已关闭多表'
  } catch (e) {
    multiTableOn.value = !enabled
    error.value = e instanceof ApiError ? e.message : '操作失败'
  } finally {
    tableSaving.value = false
  }
}

async function createTable() {
  const name = newTableName.value.trim()
  if (!name) return
  tableSaving.value = true
  error.value = ''
  try {
    const updated = await api.createLedgerTable(ledgerId.value, {
      name,
      entrySchema: DEFAULT_ENTRY_SCHEMA,
    })
    await applyLedgerUpdate(updated)
    newTableName.value = ''
    msg.value = `已添加表「${name}」`
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '添加失败'
  } finally {
    tableSaving.value = false
  }
}

function resetPassphraseModal() {
  viewStep.value = 'password'
  viewLoginPassword.value = ''
  viewError.value = ''
  revealedPassphrase.value = ''
  registerLedgerPassphrase.value = ''
}

function openPassphraseModal() {
  resetPassphraseModal()
  showPassphraseModal.value = true
}

function closePassphraseModal() {
  showPassphraseModal.value = false
  resetPassphraseModal()
}

async function onApprovalToggle(enabled) {
  approvalEnabled.value = enabled
  policySaving.value = true
  error.value = ''
  try {
    const updated = await api.setLedgerApprovalPolicy(ledgerId.value, {
      enabled,
      threshold: enabled ? memberCount.value : 1,
    })
    await applyLedgerUpdate(updated)
    msg.value = enabled ? '审批策略已启用' : '审批策略已关闭'
  } catch (e) {
    approvalEnabled.value = !enabled
    error.value = e instanceof ApiError ? e.message : '保存失败'
  } finally {
    policySaving.value = false
  }
}

async function onPassphraseViewToggle(enabled) {
  policySaving.value = true
  error.value = ''
  const prev = !enabled
  try {
    const updated = await api.setLedgerPassphraseViewPolicy(ledgerId.value, { enabled })
    await applyLedgerUpdate(updated)
    msg.value = enabled ? '已允许成员查看加密口令' : '已关闭成员查看加密口令'
    if (enabled && loadLocalPassphrase(ledgerId.value)) {
      openPassphraseModal()
      viewStep.value = 'password'
    }
  } catch (e) {
    passphraseViewEnabled.value = prev
    error.value = e instanceof ApiError ? e.message : '保存失败'
  } finally {
    policySaving.value = false
  }
}

async function confirmViewPassphrase() {
  viewError.value = ''
  viewLoading.value = true
  try {
    await api.verifyPassword(viewLoginPassword.value)
    const uid = auth.user?.id
    const local = loadLocalPassphrase(ledgerId.value)
    if (local) {
      revealedPassphrase.value = local
      viewStep.value = 'revealed'
      return
    }
    const wrapped = ledger.value?.encryption?.passphraseWrappedKeys?.[uid]
    if (wrapped) {
      revealedPassphrase.value = await unwrapPassphraseForLoginView(
        wrapped,
        viewLoginPassword.value,
        uid
      )
      saveLocalPassphrase(ledgerId.value, revealedPassphrase.value)
      viewStep.value = 'revealed'
      return
    }
    if (isCreator.value || ledger.value?.encryption?.passphraseViewEnabled) {
      viewStep.value = 'register'
      return
    }
    viewError.value = '无法获取加密口令，请联系账本创建者'
  } catch (e) {
    viewError.value = e instanceof ApiError ? e.message : '验证失败'
  } finally {
    viewLoading.value = false
  }
}

async function registerAndRevealPassphrase() {
  viewError.value = ''
  viewLoading.value = true
  try {
    const uid = auth.user?.id
    const wrapped = await wrapPassphraseForLoginView(
      registerLedgerPassphrase.value,
      viewLoginPassword.value,
      uid
    )
    const updated = await api.registerLedgerPassphraseViewWrap(ledgerId.value, { wrapped })
    await applyLedgerUpdate(updated)
    saveLocalPassphrase(ledgerId.value, registerLedgerPassphrase.value)
    revealedPassphrase.value = registerLedgerPassphrase.value
    viewStep.value = 'revealed'
  } catch (e) {
    viewError.value = e instanceof ApiError ? e.message : '登记失败'
  } finally {
    viewLoading.value = false
  }
}

async function saveName() {
  const name = nameDraft.value.trim()
  if (!name) return
  saving.value = true
  error.value = ''
  try {
    const updated = await api.updateLedger(ledgerId.value, { name })
    await applyLedgerUpdate(updated)
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
    const updated = await api.setLedgerStorageLocation(ledgerId.value, loc)
    await applyLedgerUpdate(updated)
    msg.value = '账本位置已更新'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '更新失败'
  } finally {
    storageSaving.value = false
  }
}

async function saveEncryption() {
  if (!canEnableE2E.value) return
  policySaving.value = true
  error.value = ''
  try {
    const members = (ledger.value?.members || []).map((m) => ({
      id: m.id,
      address: m.address || '',
    }))
    const enc = await buildEncryptionForCreate(
      members,
      auth.user.id,
      e2ePassphrase.value,
      ledgerId.value
    )
    const updated = await api.enableLedgerEncryption(ledgerId.value, {
      enabled: true,
      algo: 'aes-gcm-v1',
      wrappedKeys: enc.wrappedKeys,
    })
    saveLocalGroupKey(ledgerId.value, enc._groupKey)
    saveLocalPassphrase(ledgerId.value, e2ePassphrase.value)
    await applyLedgerUpdate(updated)
    enableE2E.value = false
    e2ePassphrase.value = ''
    msg.value = '端到端加密已启用'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '启用失败'
  } finally {
    policySaving.value = false
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
    await api.archiveLedger(ledgerId.value)
    clearInfoUnlockSession(ledgerId.value)
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
.policy-block {
  padding-bottom: 1rem;
  margin-bottom: 1rem;
  border-bottom: 1px dashed var(--border);
}
.policy-block:last-child {
  padding-bottom: 0;
  margin-bottom: 0;
  border-bottom: none;
}
.policy-block__heading {
  margin: 0 0 0.35rem;
  font-size: 0.9375rem;
  font-weight: 600;
}
.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-top: 0.35rem;
}
.toggle-row__text {
  font-size: 0.875rem;
  font-weight: 500;
}
.status-line {
  margin: 0.35rem 0 0;
  font-size: 0.875rem;
}
.badge-wrap {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.btn-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg);
  color: var(--text-muted);
  cursor: pointer;
}
.btn-icon:hover {
  color: var(--text);
  border-color: var(--accent);
}
.modal-title {
  margin: 0 0 0.75rem;
  font-size: 1rem;
  font-weight: 600;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 1rem;
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
