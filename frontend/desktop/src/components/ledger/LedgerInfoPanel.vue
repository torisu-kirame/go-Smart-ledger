<template>
  <section class="detail-card ledger-info">
    <h3 class="detail-card__title">账本信息</h3>

    <div v-if="!unlocked" class="ledger-info__lock">
      <p class="muted ledger-info__hint">
        {{ lockHint }}
      </p>
      <div class="form-row">
        <label>{{ passwordLabel }}</label>
        <input
          v-model="passphrase"
          type="password"
          class="field-sm"
          autocomplete="off"
          :placeholder="passwordPlaceholder"
          @keyup.enter="tryUnlock"
        />
        <button
          type="button"
          class="btn-primary"
          style="margin-top: 0.5rem"
          :disabled="unlocking || !passphrase"
          @click="tryUnlock"
        >
          {{ unlocking ? '验证中…' : '查看信息' }}
        </button>
      </div>
      <p v-if="unlockError" class="alert alert-error" style="margin-top: 0.65rem">{{ unlockError }}</p>
    </div>

    <dl v-else class="info-grid">
      <div class="info-row">
        <dt>账本 ID</dt>
        <dd class="mono">{{ ledger.id }}</dd>
      </div>
      <div class="info-row">
        <dt>名称</dt>
        <dd>{{ ledger.name }}</dd>
      </div>
      <div class="info-row">
        <dt>类型</dt>
        <dd>{{ ledger.type === 'multi' ? '多人账本' : '私人账本' }}</dd>
      </div>
      <div class="info-row">
        <dt>记账方式</dt>
        <dd>{{ bookkeepingModeLabel(bookkeepingMode) }}</dd>
      </div>
      <div class="info-row">
        <dt>创建者</dt>
        <dd class="mono">{{ ledger.creatorId }}</dd>
      </div>
      <div v-if="ledger.ledgerAddress" class="info-row">
        <dt>账本地址</dt>
        <dd class="mono break-all">{{ ledger.ledgerAddress }}</dd>
      </div>
      <div class="info-row">
        <dt>存储位置</dt>
        <dd>{{ storageLabel }}</dd>
      </div>
      <div class="info-row">
        <dt>端到端加密</dt>
        <dd>{{ ledger.encryption?.enabled ? '已启用' : '未启用' }}</dd>
      </div>
      <div v-if="isSimpleLedger" class="info-row">
        <dt>审批策略</dt>
        <dd>
          {{
            ledger.approvalPolicy?.enabled
              ? `已启用（阈值 ${ledger.approvalPolicy.threshold}）`
              : '未启用'
          }}
        </dd>
      </div>
      <div class="info-row">
        <dt>最新序号 / Merkle 根</dt>
        <dd>
          <span class="mono">{{ ledger.latestSeq }}</span>
          <span class="muted"> · </span>
          <span class="mono break-all">{{ ledger.latestRoot || '—' }}</span>
        </dd>
      </div>
      <div class="info-row">
        <dt>锚定状态</dt>
        <dd>{{ ledger.anchorStatus }}</dd>
      </div>
      <div v-if="ledger.createdAt" class="info-row">
        <dt>创建时间</dt>
        <dd>{{ formatTime(ledger.createdAt) }}</dd>
      </div>
      <div v-if="ledger.updatedAt" class="info-row">
        <dt>更新时间</dt>
        <dd>{{ formatTime(ledger.updatedAt) }}</dd>
      </div>
      <div v-if="ledger.lastBackupRef || ledger.lastBackupCid" class="info-row">
        <dt>最近备份</dt>
        <dd class="mono break-all">
          <span v-if="ledger.lastBackupRef">{{ ledger.lastBackupRef }}</span>
          <span v-if="ledger.lastBackupCid" class="muted"> CID {{ ledger.lastBackupCid }}</span>
        </dd>
      </div>
      <div v-if="ledger.members?.length" class="info-row info-row--block">
        <dt>成员（{{ ledger.members.length }}）</dt>
        <dd>
          <ul class="member-list">
            <li v-for="m in ledger.members" :key="m.id">
              <span class="mono">{{ m.id }}</span>
              <span v-if="m.role" class="muted"> · {{ m.role }}</span>
              <span v-if="m.address" class="mono member-addr" :title="m.address">{{ shortAddr(m.address) }}</span>
            </li>
          </ul>
        </dd>
      </div>
      <div v-if="schemaLabel && isSimpleLedger" class="info-row">
        <dt>流水模板</dt>
        <dd>{{ schemaLabel }}</dd>
      </div>
    </dl>
  </section>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { api, ApiError } from '../../api/http'
import { useAuthStore } from '../../stores/auth'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { unwrapGroupKey, saveLocalGroupKey } from '../../utils/e2eCrypto'
import { storageLocationLabel } from '../../utils/ledgerStorage'
import {
  getInfoUnlockSession,
  setInfoUnlockSession,
} from '../../composables/useLedgerDetail'
import { bookkeepingModeLabel } from '../../utils/bookkeepingMode'

const auth = useAuthStore()
const { ledgerId, ledger, groupKey, error, msg, schema, bookkeepingMode, isSimpleLedger } =
  useLedgerDetail()

const passphrase = ref('')
const unlocking = ref(false)
const unlockError = ref('')
const sessionUnlocked = ref(getInfoUnlockSession(ledgerId.value))

watch(ledgerId, (id) => {
  sessionUnlocked.value = getInfoUnlockSession(id)
})

const groupKeyReady = computed(() => {
  if (!ledger.value?.encryption?.enabled) return true
  return !!groupKey.value
})

const unlocked = computed(
  () => sessionUnlocked.value || (ledger.value?.encryption?.enabled && groupKeyReady.value)
)

const storageLabel = computed(() => storageLocationLabel(ledger.value?.storageLocation))

const schemaLabel = computed(() => {
  const s = schema.value
  if (!s) return ''
  if (s.templateId && s.templateId !== 'custom') return s.templateId
  if (s.fields?.length) return `自定义（${s.fields.length} 个字段）`
  return s.templateId || ''
})

const lockHint = computed(() => {
  if (ledger.value?.encryption?.enabled) {
    return '敏感账本信息已保护。请输入创建账本时设置的加密口令（账本密码）后查看。'
  }
  return '查看完整账本信息前，请输入当前登录账号的密码以确认身份。'
})

const passwordLabel = computed(() =>
  ledger.value?.encryption?.enabled ? '账本密码（加密口令）' : '登录密码'
)

const passwordPlaceholder = computed(() =>
  ledger.value?.encryption?.enabled ? '加密口令' : '当前账号密码'
)

watch(
  () => groupKeyReady.value,
  (ready) => {
    if (ready && ledger.value?.encryption?.enabled) {
      setInfoUnlockSession(ledgerId.value)
      sessionUnlocked.value = true
    }
  }
)

function formatTime(t) {
  if (!t) return '—'
  return new Date(t).toLocaleString()
}

function shortAddr(a) {
  if (!a || a.length <= 16) return a
  return `${a.slice(0, 8)}…${a.slice(-6)}`
}

async function tryUnlock() {
  if (!ledger.value) return
  unlockError.value = ''
  unlocking.value = true
  error.value = ''
  try {
    if (ledger.value.encryption?.enabled) {
      const uid = auth.user?.id
      const wrapped = ledger.value.encryption.wrappedKeys?.[uid]
      if (!wrapped) {
        unlockError.value = '未找到您的密钥包装'
        return
      }
      groupKey.value = await unwrapGroupKey(wrapped, passphrase.value, ledgerId.value, uid)
      saveLocalGroupKey(ledgerId.value, groupKey.value)
      setInfoUnlockSession(ledgerId.value)
      sessionUnlocked.value = true
      msg.value = '账本信息已解锁'
    } else {
      await api.verifyPassword(passphrase.value)
      setInfoUnlockSession(ledgerId.value)
      sessionUnlocked.value = true
      msg.value = '身份已验证'
    }
    passphrase.value = ''
  } catch (e) {
    unlockError.value =
      e instanceof ApiError
        ? e.message
        : ledger.value.encryption?.enabled
          ? '口令错误或密钥损坏'
          : '密码验证失败'
  } finally {
    unlocking.value = false
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
  margin: 0 0 0.85rem;
  font-size: 0.8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.ledger-info__hint {
  margin: 0 0 0.75rem;
  font-size: 0.875rem;
  line-height: 1.45;
}
.info-grid {
  margin: 0;
  display: grid;
  gap: 0.65rem;
}
.info-row {
  display: grid;
  grid-template-columns: 7.5rem 1fr;
  gap: 0.5rem 0.75rem;
  font-size: 0.875rem;
  align-items: start;
}
.info-row--block {
  grid-template-columns: 1fr;
}
.info-row--block dt {
  margin-bottom: 0.35rem;
}
.info-row dt {
  margin: 0;
  color: var(--text-muted);
  font-weight: 600;
}
.info-row dd {
  margin: 0;
}
.break-all {
  word-break: break-all;
}
.member-list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 0.35rem;
}
.member-addr {
  display: block;
  font-size: 0.75rem;
  margin-top: 0.15rem;
}
@media (max-width: 520px) {
  .info-row {
    grid-template-columns: 1fr;
  }
}
</style>
