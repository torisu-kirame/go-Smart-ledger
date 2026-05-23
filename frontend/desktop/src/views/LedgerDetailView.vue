<template>
  <div v-if="ledger" class="page">
    <h2><router-link to="/ledgers" style="color:var(--text-muted)">←</router-link> {{ ledger.name }}</h2>
    <div v-if="msg" class="alert alert-success">{{ msg }}</div>
    <div v-if="error" class="alert alert-error">{{ error }}</div>

    <div class="grid-3">
      <div class="card"><h4>类型</h4><div class="val">{{ ledger.type === 'multi' ? '多人' : '私人' }}</div></div>
      <div class="card"><h4>序号</h4><div class="val mono">{{ ledger.latestSeq }}</div></div>
      <div class="card"><h4>锚定</h4><div class="val"><span :class="['badge', ledger.anchorStatus==='synced'?'badge-ok':'badge-pending']">{{ ledger.anchorStatus }}</span></div></div>
    </div>

    <div v-if="ledger.encryption?.enabled" class="panel">
      <div v-if="!groupKeyReady" class="form-row">
        <label>输入加密口令以解密/记账</label>
        <input v-model="e2ePassphrase" type="password" class="field-sm" />
        <button type="button" class="btn-ghost" style="margin-top:0.35rem" @click="unlockE2E">解锁</button>
      </div>
    </div>

    <div class="panel">
      <h3>链上操作</h3>
      <div class="actions-row">
        <button class="btn-ghost" :disabled="busy" @click="doVerify">校验完整性</button>
        <button class="btn-primary" :disabled="busy" @click="doAnchor">封账并锚定</button>
        <button class="btn-ghost" :disabled="busy" @click="doSync">同步事件</button>
        <button v-if="ledger.anchorStatus==='synced'" class="btn-ghost" @click="goBackup">加密备份</button>
      </div>
    </div>

    <div v-if="ledger.type === 'multi'" class="panel">
      <h3>邀请成员</h3>
      <div class="form-row">
        <label>好友用户 ID</label>
        <input v-model="inviteUserId" placeholder="被邀请人雪花 ID" />
      </div>
      <button class="btn-ghost" :disabled="busy || !inviteUserId" @click="sendInvite">发送邀请</button>
    </div>

    <div v-if="ledger.approvalPolicy?.enabled" class="panel">
      <h3>待审批记账</h3>
      <div v-if="!pending.length" class="muted">暂无待审批</div>
      <div v-for="p in pending" :key="p.id" class="pending-row">
        <div>
          <span class="mono">#{{ p.id.slice(0, 8) }}…</span>
        </div>
        <div class="actions-row">
          <button class="btn-primary" :disabled="busy" @click="approve(p.id)">批准</button>
          <button class="btn-ghost" :disabled="busy" @click="reject(p.id)">拒绝</button>
        </div>
      </div>
    </div>

    <div class="panel">
      <h3>记一笔</h3>
      <EntryFormFields :schema="schema" :model="entryData" :members="memberOptions" />
      <button class="btn-primary" :disabled="busy" style="margin-top:0.75rem" @click="addEntry">
        {{ ledger.approvalPolicy?.enabled ? '提交审批' : '提交到链' }}
      </button>
    </div>

    <div class="panel">
      <h3>事件流水 ({{ events.length }})</h3>
      <div class="table-wrap">
        <table>
          <thead><tr><th>Seq</th><th>类型</th><th>签名者</th><th>哈希</th></tr></thead>
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
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'
import EntryFormFields from '../components/EntryFormFields.vue'
import { emptyEntryData, resolveSchema } from '../utils/entrySchema'
import {
  encryptEntryData,
  loadLocalGroupKey,
  saveLocalGroupKey,
  unwrapGroupKey,
} from '../utils/e2eCrypto'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const id = route.params.id
const ledger = ref(null)
const events = ref([])
const pending = ref([])
const error = ref('')
const msg = ref('')
const busy = ref(false)
const inviteUserId = ref('')
const e2ePassphrase = ref('')
const groupKey = ref('')
const entryData = reactive({})

const schema = computed(() => resolveSchema(ledger.value))
const memberOptions = computed(() =>
  (ledger.value?.members || []).map((m) => ({ id: m.id, username: m.id }))
)
const groupKeyReady = computed(() => {
  if (!ledger.value?.encryption?.enabled) return true
  return !!groupKey.value
})

function initEntryDefaults() {
  const uid = auth.user?.id || ''
  const defaults = { bookkeeper: uid, date: new Date().toISOString().slice(0, 10) }
  const data = emptyEntryData(schema.value, defaults)
  Object.keys(entryData).forEach((k) => delete entryData[k])
  Object.assign(entryData, data)
}

async function load() {
  ledger.value = await api.getLedger(id)
  events.value = await api.listEvents(id)
  if (ledger.value?.approvalPolicy?.enabled) {
    const res = await api.listPending(id)
    pending.value = res.pending || []
  } else {
    pending.value = []
  }
  const saved = loadLocalGroupKey(id)
  if (saved) groupKey.value = saved
  initEntryDefaults()
}

onMounted(load)

async function unlockE2E() {
  if (!ledger.value?.encryption?.enabled) return
  const uid = auth.user?.id
  const wrapped = ledger.value.encryption.wrappedKeys?.[uid]
  if (!wrapped) {
    error.value = '未找到您的密钥包装，请联系账本创建者'
    return
  }
  try {
    groupKey.value = await unwrapGroupKey(wrapped, e2ePassphrase.value, id, uid)
    saveLocalGroupKey(id, groupKey.value)
    msg.value = '加密账本已解锁'
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
      const res = await api.proposeEntry(id, entry)
      msg.value = res.status === 'committed' ? '记账已上链' : '已提交审批，等待其他成员批准'
    } else {
      await api.appendEntry(id, entry)
      msg.value = '记账成功'
    }
    entryData.amount = ''
    entryData.note = ''
    entryData.payee = ''
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '失败'
  } finally {
    busy.value = false
  }
}

async function approve(pendingId) {
  busy.value = true
  try {
    const res = await api.approvePending(id, pendingId)
    msg.value = res.status === 'committed' ? '已批准并上链' : '已记录批准，等待更多成员'
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
    await api.rejectPending(id, pendingId)
    msg.value = '已拒绝'
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '操作失败'
  } finally {
    busy.value = false
  }
}

async function sendInvite() {
  busy.value = true
  try {
    await api.inviteMember(id, inviteUserId.value.trim())
    msg.value = '邀请已发送'
    inviteUserId.value = ''
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '邀请失败'
  } finally {
    busy.value = false
  }
}

async function doSync() {
  busy.value = true
  try {
    const since = ledger.value?.latestSeq ? Math.max(0, ledger.value.latestSeq - 20) : 0
    const res = await api.syncLedger(id, since)
    events.value = res.events || []
    if (res.ledger) ledger.value = res.ledger
    msg.value = `已同步 ${events.value.length} 条新事件`
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '同步失败'
  } finally {
    busy.value = false
  }
}

async function doAnchor() {
  busy.value = true
  try {
    const r = await api.anchor(id)
    msg.value = `封账成功 · ${r.status}`
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '锚定失败'
  } finally {
    busy.value = false
  }
}

async function doVerify() {
  const r = await api.verify(id)
  msg.value = r.valid ? 'Merkle 校验通过' : '校验未通过'
}

function goBackup() {
  router.push({ path: '/backup', query: { ledgerId: id } })
}
</script>

<style scoped>
.pending-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border);
  flex-wrap: wrap;
}
.mono { font-family: ui-monospace, monospace; }
</style>
