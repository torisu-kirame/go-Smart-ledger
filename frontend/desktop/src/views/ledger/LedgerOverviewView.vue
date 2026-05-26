<template>
  <div class="ledger-overview">
    <div v-if="msg" class="alert alert-success">{{ msg }}</div>
    <div v-if="error" class="alert alert-error">{{ error }}</div>

    <div class="overview-layout">
      <aside class="overview-aside">
        <LedgerInfoPanel />

        <section v-if="ledger.encryption?.enabled && !groupKeyReady" class="detail-card detail-card--warn">
          <h3 class="detail-card__title">端到端加密</h3>
          <div class="form-row">
            <label>加密口令</label>
            <input v-model="e2ePassphrase" type="password" class="field-sm" />
            <button type="button" class="btn-primary" style="margin-top: 0.5rem" @click="unlockE2E">
              解锁
            </button>
          </div>
        </section>

        <section v-if="ledger.type === 'multi'" class="detail-card">
          <h3 class="detail-card__title">邀请成员</h3>
          <div class="member-add-block">
            <label>被邀请人</label>
            <MemberAddPanel
              v-model="inviteUserId"
              :multiple="false"
              :exclude-ids="inviteExcludeIds"
            />
          </div>
          <button
            class="btn-primary"
            type="button"
            :disabled="inviteBusy || !inviteUserId"
            @click="sendInvite"
          >
            发送邀请
          </button>
          <div v-if="outgoingInvites.length" class="outgoing-list">
            <h4>待处理</h4>
            <div v-for="inv in outgoingInvites" :key="inv.inviteeId" class="invite-row compact">
              <span class="mono">{{ inv.inviteeId }}</span>
              <span class="muted">{{ formatInviteTime(inv.createdAt) }}</span>
            </div>
          </div>
        </section>

        <section v-if="ledger.externalAnchor?.txHash" class="detail-card">
          <h3 class="detail-card__title">链外锚定</h3>
          <p class="mono anchor-hash">{{ ledger.externalAnchor.txHash }}</p>
          <a
            v-if="ledger.externalAnchor.explorerUrl"
            :href="ledger.externalAnchor.explorerUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="anchor-link"
          >区块浏览器</a>
        </section>
      </aside>

      <div class="overview-main">
        <section class="detail-card">
          <h3 class="detail-card__title">链上操作</h3>
          <div class="actions-row">
            <button class="btn-ghost" :disabled="busy" @click="doVerify">校验完整性</button>
            <button class="btn-primary" :disabled="busy" @click="doAnchor">封账并锚定</button>
            <button v-if="ledger.anchorStatus === 'synced'" class="btn-ghost" @click="goBackup">
              加密备份
            </button>
          </div>
        </section>

        <section v-if="ledger.approvalPolicy?.enabled" class="detail-card">
          <h3 class="detail-card__title">待审批记账</h3>
          <div v-if="!pending.length" class="muted empty-inline">暂无待审批</div>
          <div v-for="p in pending" :key="p.id" class="pending-row">
            <span class="mono">#{{ p.id.slice(0, 8) }}…</span>
            <div class="actions-row">
              <button class="btn-primary" :disabled="busy" @click="approve(p.id)">批准</button>
              <button class="btn-ghost" :disabled="busy" @click="reject(p.id)">拒绝</button>
            </div>
          </div>
        </section>

        <section class="detail-card detail-card--fill">
          <h3 class="detail-card__title">事件流水 <span class="count">({{ events.length }})</span></h3>
          <div class="table-wrap">
            <table>
              <thead>
                <tr><th>Seq</th><th>类型</th><th>签名者</th><th>哈希</th></tr>
              </thead>
              <tbody>
                <tr v-for="e in [...events].reverse()" :key="e.seq">
                  <td class="mono">{{ e.seq }}</td>
                  <td>{{ e.type }}</td>
                  <td>{{ e.signerId }}</td>
                  <td class="mono">{{ e.hash?.slice(0, 16) }}…</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '../../api/http'
import { useAuthStore } from '../../stores/auth'
import LedgerInfoPanel from '../../components/ledger/LedgerInfoPanel.vue'
import MemberAddPanel from '../../components/MemberAddPanel.vue'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { saveLocalGroupKey, unwrapGroupKey } from '../../utils/e2eCrypto'

const router = useRouter()
const auth = useAuthStore()
const {
  ledgerId,
  ledger,
  events,
  pending,
  groupKey,
  error,
  msg,
  load,
} = useLedgerDetail()

const busy = ref(false)
const e2ePassphrase = ref('')
const inviteUserId = ref('')
const inviteBusy = ref(false)
const outgoingInvites = ref([])

const inviteExcludeIds = computed(() => {
  const ids = (ledger.value?.members || []).map((m) => m.id)
  if (auth.user?.id) ids.push(auth.user.id)
  return ids
})
const groupKeyReady = computed(() => {
  if (!ledger.value?.encryption?.enabled) return true
  return !!groupKey.value
})

async function unlockE2E() {
  if (!ledger.value?.encryption?.enabled) return
  const uid = auth.user?.id
  const wrapped = ledger.value.encryption.wrappedKeys?.[uid]
  if (!wrapped) {
    error.value = '未找到您的密钥包装'
    return
  }
  try {
    groupKey.value = await unwrapGroupKey(wrapped, e2ePassphrase.value, ledgerId, uid)
    saveLocalGroupKey(ledgerId, groupKey.value)
    msg.value = '加密账本已解锁'
  } catch {
    error.value = '口令错误或密钥损坏'
  }
}

async function approve(pendingId) {
  busy.value = true
  try {
    const res = await api.approvePending(ledgerId, pendingId)
    msg.value = res.status === 'committed' ? '已批准并上链' : '已记录批准'
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '批准失败'
  } finally {
    busy.value = false
  }
}

async function reject(pendingId) {
  busy.value = true
  try {
    await api.rejectPending(ledgerId, pendingId)
    msg.value = '已拒绝'
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '操作失败'
  } finally {
    busy.value = false
  }
}

async function doAnchor() {
  busy.value = true
  try {
    const r = await api.anchor(ledgerId)
    msg.value = r.externalAnchor?.txHash
      ? `封账成功 · 链外锚定 ${r.externalAnchor.txHash.slice(0, 14)}…`
      : `封账成功 · ${r.status}`
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '锚定失败'
  } finally {
    busy.value = false
  }
}

async function doVerify() {
  const r = await api.verify(ledgerId)
  msg.value = r.valid ? 'Merkle 校验通过' : '校验未通过'
}

function goBackup() {
  router.push({ path: '/backup', query: { ledgerId } })
}

function formatInviteTime(t) {
  if (!t) return ''
  return new Date(t).toLocaleString()
}

async function loadOutgoingInvites() {
  if (ledger.value?.type !== 'multi') {
    outgoingInvites.value = []
    return
  }
  try {
    const res = await api.listLedgerInvites(ledgerId)
    outgoingInvites.value = res.invites || []
  } catch {
    outgoingInvites.value = []
  }
}

function inviteErrorMessage(e, fallback) {
  if (!(e instanceof ApiError)) return fallback
  const m = e.message || ''
  if (m.includes('already a member')) return '对方已是账本成员'
  if (m.includes('invite already pending')) return '已向该用户发送过邀请'
  if (m.includes('cannot invite yourself')) return '不能邀请自己'
  return m || fallback
}

async function sendInvite() {
  const targetId = typeof inviteUserId.value === 'string' ? inviteUserId.value.trim() : ''
  if (!targetId) {
    error.value = '请填写被邀请人'
    return
  }
  inviteBusy.value = true
  error.value = ''
  try {
    await api.inviteMember(ledgerId, targetId)
    msg.value = '邀请已发送'
    inviteUserId.value = ''
    await loadOutgoingInvites()
  } catch (e) {
    error.value = inviteErrorMessage(e, '邀请失败')
  } finally {
    inviteBusy.value = false
  }
}

watch(
  () => [ledger.value?.id, ledger.value?.type],
  ([, type]) => {
    if (type === 'multi') loadOutgoingInvites()
    else outgoingInvites.value = []
  },
  { immediate: true }
)
</script>

<style scoped>
.overview-layout {
  display: grid;
  grid-template-columns: minmax(16rem, 22rem) 1fr;
  gap: 1rem;
  align-items: start;
}
.detail-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1rem 1.1rem;
  margin-bottom: 1rem;
  box-shadow: var(--shadow-sm);
}
.detail-card--warn {
  border-color: color-mix(in srgb, var(--accent) 35%, var(--border));
  background: color-mix(in srgb, var(--accent-soft) 40%, var(--bg-card));
}
.detail-card--fill {
  margin-bottom: 0;
}
.detail-card__title {
  margin: 0 0 0.85rem;
  font-size: 0.8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.detail-card__title .count {
  font-weight: 600;
  text-transform: none;
  letter-spacing: 0;
}
.pending-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
  padding: 0.55rem 0;
  border-bottom: 1px solid var(--border);
  flex-wrap: wrap;
}
.pending-row:last-child {
  border-bottom: none;
}
.mono {
  font-family: ui-monospace, monospace;
}
.anchor-hash {
  font-size: 0.8rem;
  word-break: break-all;
  margin: 0 0 0.5rem;
}
.anchor-link {
  font-size: 0.875rem;
  color: var(--accent);
}
.member-add-block {
  display: grid;
  gap: 0.45rem;
  margin-bottom: 0.75rem;
}
.member-add-block label {
  font-size: 0.8125rem;
  color: var(--text-muted);
}
.outgoing-list {
  margin-top: 0.85rem;
  padding-top: 0.75rem;
  border-top: 1px dashed var(--border);
}
.outgoing-list h4 {
  margin: 0 0 0.45rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
}
.invite-row.compact {
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.35rem 0;
  font-size: 0.8125rem;
}
.empty-inline {
  font-size: 0.875rem;
  padding: 0.25rem 0;
}
@media (max-width: 900px) {
  .overview-layout {
    grid-template-columns: 1fr;
  }
  .overview-aside {
    display: contents;
  }
  .overview-main .detail-card--fill {
    order: 10;
  }
}
</style>
