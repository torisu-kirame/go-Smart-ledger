<template>
  <div class="page">
    <header class="page-header">
      <h2>账本管理</h2>
      <div class="header-actions">
        <button class="btn-ghost" type="button" :disabled="syncingAll || !list.length" @click="syncAllLocal">
          {{ syncingAll ? '同步中…' : '全部同步到本机' }}
        </button>
        <button class="btn-primary" type="button" @click="openCreate">创建账本</button>
      </div>
    </header>
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="success" class="alert alert-success">{{ success }}</div>

    <section v-if="incomingInvites.length" class="panel panel-highlight">
      <h3>收到的账本邀请</h3>
      <p class="section-hint">仅被邀请并接受后，您才有权查看对应账本数据。</p>
      <div v-for="inv in incomingInvites" :key="inv.ledgerId + inv.inviterId" class="invite-row">
        <div>
          <strong>{{ inviteLedgerName(inv.ledgerId) }}</strong>
          <span class="mono">ID {{ inv.ledgerId }}</span>
          <span v-if="inv.inviterId" class="inviter">邀请人 {{ inv.inviterId }}</span>
        </div>
        <button class="btn-primary" :disabled="inviteBusy" @click="acceptInvite(inv.ledgerId)">接受加入</button>
      </div>
    </section>

    <section ref="inviteSectionRef" class="panel">
      <h3>邀请成员加入账本</h3>
      <p class="section-hint">向多人账本发送邀请；对方在「收到的账本邀请」中同意后成为成员。团队本身不自动授予账本权限。</p>
      <div v-if="!multiLedgers.length" class="muted">暂无多人账本，请先创建或接受邀请加入。</div>
      <template v-else>
        <div class="form-row">
          <label>选择账本</label>
          <AppSelect v-model="inviteLedgerId" :options="multiLedgerOptions" />
        </div>
        <div class="form-row member-add-block">
          <label>被邀请人</label>
          <MemberAddPanel
            v-model="inviteUserId"
            :multiple="false"
            :exclude-ids="inviteExcludeIds"
          />
        </div>
        <button class="btn-primary" :disabled="inviteBusy || !inviteUserId || !inviteLedgerId" @click="sendInvite">
          发送邀请
        </button>
        <div v-if="outgoingInvites.length" class="outgoing-list">
          <h4>待对方处理的邀请</h4>
          <div v-for="inv in outgoingInvites" :key="inv.inviteeId" class="invite-row compact">
            <span>用户 <span class="mono">{{ inv.inviteeId }}</span></span>
            <span class="muted">{{ formatInviteTime(inv.createdAt) }}</span>
          </div>
        </div>
      </template>
    </section>

    <div class="panel">
      <div v-if="!list.length" class="muted empty">暂无账本，点击右上角创建</div>
      <div v-else class="table-wrap">
        <table>
          <thead><tr><th>名称</th><th>类型</th><th>账本地址</th><th>序号</th><th>锚定</th><th></th></tr></thead>
          <tbody>
            <tr v-for="l in list" :key="l.id">
              <td>{{ l.name }}</td>
              <td><span :class="['badge', l.type === 'multi' ? 'badge-multi' : 'badge-private']">{{ l.type === 'multi' ? '多人' : '私人' }}</span></td>
              <td class="mono" :title="l.ledgerAddress">{{ shortAddr(l.ledgerAddress) }}</td>
              <td class="mono">{{ l.latestSeq }}</td>
              <td><span :class="['badge', l.anchorStatus === 'synced' ? 'badge-ok' : 'badge-pending']">{{ l.anchorStatus }}</span></td>
              <td class="row-actions">
                <router-link :to="`/ledgers/${l.id}`">详情</router-link>
                <button
                  v-if="l.type === 'multi'"
                  type="button"
                  class="btn-link"
                  @click="focusInvite(l.id)"
                >
                  邀请
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div v-if="show" class="modal">
      <form class="modal-card" @submit.prevent="create">
        <h3>创建账本</h3>
        <div class="form-row readonly-row">
          <label>创建者（本人）</label>
          <span class="mono uid">{{ auth.user?.id || '—' }}</span>
        </div>
        <div class="form-row">
          <label>类型</label>
          <AppSelect v-model="form.type" :options="ledgerTypeOptions" />
        </div>
        <div class="form-row"><label>名称</label><input v-model="form.name" required /></div>
        <div class="form-row">
          <label>记账模板</label>
          <AppSelect v-model="form.templateId" :options="templateOptions" />
        </div>
        <div v-if="form.templateId === 'custom'" class="custom-schema">
          <div v-for="(f, i) in form.customFields" :key="i" class="form-row member">
            <input v-model="f.key" placeholder="字段 key（英文）" />
            <input v-model="f.label" placeholder="显示名" />
            <AppSelect v-model="f.type" sm class="member-select" :options="FIELD_TYPE_OPTIONS" />
            <label class="check"><input type="checkbox" v-model="f.required" /> 必填</label>
            <DeleteButton icon-only sm title="删除字段" @click="form.customFields.splice(i, 1)" />
          </div>
          <button type="button" class="btn-ghost" @click="addCustomField">+ 字段</button>
        </div>
        <div v-if="form.type === 'multi'" class="form-row">
          <label class="inline-check">
            <input v-model="form.enableE2E" type="checkbox" />
            启用组级端到端加密
          </label>
        </div>
        <div v-if="form.type === 'multi' && form.enableE2E" class="form-row">
          <label>加密口令（仅本机解密，勿丢失）</label>
          <input v-model="form.e2ePassphrase" type="password" placeholder="创建后用于加解密记账数据" />
        </div>
        <div v-if="form.type === 'multi'" class="form-row member-add-block">
          <label>邀请首批成员（可选）</label>
          <p class="field-hint">创建后向对方发送加入申请，对方同意后才成为成员。</p>
          <MemberAddPanel
            v-model="form.otherMemberIds"
            :multiple="true"
            :exclude-ids="auth.user?.id ? [auth.user.id] : []"
          />
        </div>
        <div style="display:flex;gap:0.5rem;justify-content:flex-end;margin-top:1rem">
          <button type="button" class="btn-ghost" @click="show = false">取消</button>
          <button class="btn-primary" :disabled="saving">{{ saving ? '创建中…' : '创建' }}</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'
import AppSelect from '../components/AppSelect.vue'
import DeleteButton from '../components/DeleteButton.vue'
import MemberAddPanel from '../components/MemberAddPanel.vue'
import { DEFAULT_ENTRY_SCHEMA, FIELD_TYPE_OPTIONS } from '../utils/entrySchema'
import { buildEncryptionForCreate, saveLocalGroupKey } from '../utils/e2eCrypto'
import { syncLedgerToLocal } from '../localdb/db'

const router = useRouter()
const auth = useAuthStore()
const list = ref([])
const incomingInvites = ref([])
const outgoingInvites = ref([])
const inviteLedgerId = ref('')
const inviteUserId = ref('')
const inviteBusy = ref(false)
const inviteSectionRef = ref(null)
const syncingAll = ref(false)
const error = ref('')
const success = ref('')
const show = ref(false)
const saving = ref(false)
const templates = ref([DEFAULT_ENTRY_SCHEMA])
const form = reactive({
  type: 'private',
  name: '',
  templateId: 'default',
  customFields: [{ key: '', label: '', type: 'text', required: true }],
  otherMemberIds: [],
  enableE2E: false,
  e2ePassphrase: '',
})

const ledgerTypeOptions = [
  { value: 'private', label: '私人（1人）' },
  { value: 'multi', label: '多人（邀请加入）' },
]

const templateOptions = computed(() => [
  ...templates.value.map((t) => ({
    value: t.templateId,
    label: `${templateLabel(t)}${t.builtin ? '（内置）' : ''}`,
  })),
  { value: 'custom', label: '临时自定义（不保存）' },
])

const multiLedgers = computed(() => list.value.filter((l) => l.type === 'multi'))
const multiLedgerOptions = computed(() =>
  multiLedgers.value.map((l) => ({
    value: l.id,
    label: `${l.name}（${shortAddr(l.ledgerAddress)}）`,
  }))
)
const inviteExcludeIds = computed(() => {
  const ledger = list.value.find((l) => l.id === inviteLedgerId.value)
  const ids = (ledger?.members || []).map((m) => m.id)
  if (auth.user?.id) ids.push(auth.user.id)
  return ids
})

const selectedTemplateFields = computed(() => {
  if (form.templateId === 'custom') return ''
  const t = templates.value.find((x) => x.templateId === form.templateId)
  return (t?.fields || []).map((f) => f.label).join('、')
})

function templateLabel(t) {
  return t.name || t.templateId
}

function addCustomField() {
  form.customFields.push({ key: '', label: '', type: 'text', required: false })
}

function shortAddr(a) {
  if (!a) return '—'
  return a.length > 14 ? `${a.slice(0, 8)}…${a.slice(-6)}` : a
}

function resetForm() {
  form.type = 'private'
  form.name = ''
  form.templateId = 'default'
  form.customFields = [{ key: '', label: '', type: 'text', required: true }]
  form.otherMemberIds = []
  form.enableE2E = false
  form.e2ePassphrase = ''
}

function buildEntrySchema() {
  if (form.templateId === 'custom') {
    const fields = form.customFields
      .filter((f) => f.key.trim() && f.label.trim())
      .map((f) => ({
        key: f.key.trim(),
        label: f.label.trim(),
        type: f.type,
        required: !!f.required,
      }))
    if (!fields.length) throw new ApiError('请至少定义 1 个自定义字段', 400)
    return { templateId: 'custom', fields }
  }
  const t = templates.value.find((x) => x.templateId === form.templateId)
  if (t?.fields?.length) {
    return { templateId: t.templateId, fields: t.fields }
  }
  return { templateId: form.templateId }
}

function openCreate() {
  error.value = ''
  success.value = ''
  resetForm()
  show.value = true
}

watch(() => form.type, () => {
  if (form.type === 'multi' && !form.otherMemberIds.length) {
    form.otherMemberIds = []
  }
})

function inviteLedgerName(ledgerId) {
  const l = list.value.find((x) => x.id === ledgerId)
  return l?.name || '账本'
}

function formatInviteTime(t) {
  if (!t) return ''
  return new Date(t).toLocaleString()
}

async function loadOutgoingInvites() {
  if (!inviteLedgerId.value) {
    outgoingInvites.value = []
    return
  }
  try {
    const res = await api.listLedgerInvites(inviteLedgerId.value)
    outgoingInvites.value = res.invites || []
  } catch {
    outgoingInvites.value = []
  }
}

async function loadInvites() {
  try {
    const res = await api.listMyInvites()
    incomingInvites.value = res.invites || []
  } catch {
    incomingInvites.value = []
  }
}

async function load() {
  const data = await api.listLedgers()
  list.value = Array.isArray(data) ? data : []
  if (!inviteLedgerId.value && multiLedgers.value.length) {
    inviteLedgerId.value = multiLedgers.value[0].id
  } else if (inviteLedgerId.value && !multiLedgers.value.some((l) => l.id === inviteLedgerId.value)) {
    inviteLedgerId.value = multiLedgers.value[0]?.id || ''
  }
  await Promise.all([loadInvites(), loadOutgoingInvites()])
}

watch(inviteLedgerId, () => {
  loadOutgoingInvites()
})

async function acceptInvite(ledgerId) {
  inviteBusy.value = true
  error.value = ''
  try {
    await api.acceptInvite(ledgerId)
    success.value = '已加入账本'
    await load()
    router.push(`/ledgers/${ledgerId}`)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '接受失败'
  } finally {
    inviteBusy.value = false
  }
}

async function sendInvite() {
  const targetId = typeof inviteUserId.value === 'string' ? inviteUserId.value.trim() : ''
  if (!inviteLedgerId.value || !targetId) {
    error.value = '请选择账本并填写被邀请人'
    return
  }
  inviteBusy.value = true
  error.value = ''
  try {
    await api.inviteMember(inviteLedgerId.value, targetId)
    success.value = '邀请已发送，等待对方在「收到的账本邀请」中同意'
    inviteUserId.value = ''
    await loadOutgoingInvites()
  } catch (e) {
    error.value = inviteErrorMessage(e, '邀请失败')
  } finally {
    inviteBusy.value = false
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

function focusInvite(ledgerId) {
  inviteLedgerId.value = ledgerId
  loadOutgoingInvites()
  nextTick(() => {
    inviteSectionRef.value?.scrollIntoView?.({ behavior: 'smooth', block: 'start' })
  })
}

async function sendInvitesToUsers(ledgerId, userIds) {
  const ids = [...new Set(userIds.map((id) => String(id).trim()).filter(Boolean))]
  let sent = 0
  for (const uid of ids) {
    try {
      await api.inviteMember(ledgerId, uid)
      sent++
    } catch (e) {
      if (e instanceof ApiError && e.message?.includes('invite already pending')) {
        sent++
        continue
      }
      throw e
    }
  }
  return sent
}

async function syncAllLocal() {
  syncingAll.value = true
  error.value = ''
  success.value = ''
  let total = 0
  try {
    for (const l of list.value) {
      const res = await syncLedgerToLocal(api, l.id)
      total += res.newCount
    }
    success.value = `已将 ${list.value.length} 个账本同步到本机 SQLite（新增 ${total} 条事件）`
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '批量同步失败'
  } finally {
    syncingAll.value = false
  }
}

onMounted(async () => {
  try {
    const res = await api.listEntryTemplates()
    if (res.templates?.length) templates.value = res.templates
  } catch {
    templates.value = [DEFAULT_ENTRY_SCHEMA]
  }
  load()
})

function buildMembers() {
  const uid = auth.user?.id
  if (!uid) throw new ApiError('未登录或用户 ID 缺失，请重新登录', 401)
  if (form.type === 'private') {
    return [{ id: uid, address: '' }]
  }
  return [{ id: uid, address: '' }]
}

function pendingInviteUserIds() {
  return (form.otherMemberIds || []).map((id) => String(id).trim()).filter(Boolean)
}

async function create() {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    const members = buildMembers()
    const inviteTargets = form.type === 'multi' ? pendingInviteUserIds() : []
    let encryption = { enabled: false }
    let groupKey = ''
    if (form.type === 'multi' && form.enableE2E && form.e2ePassphrase) {
      const encMembers = [
        ...members,
        ...inviteTargets.map((id) => ({ id, address: '' })),
      ]
      const enc = await buildEncryptionForCreate(encMembers, auth.user.id, form.e2ePassphrase, 'new')
      encryption = { enabled: true, algo: 'aes-gcm-v1', wrappedKeys: enc.wrappedKeys }
      groupKey = enc._groupKey
    }
    const created = await api.createLedger({
      type: form.type,
      name: form.name.trim(),
      creatorId: auth.user.id,
      members,
      entrySchema: buildEntrySchema(),
      approvalPolicy: form.type === 'multi' ? { enabled: true, threshold: 2 } : { enabled: false },
      encryption,
    })
    if (groupKey && created.id) {
      saveLocalGroupKey(created.id, groupKey)
    }
    let inviteNote = ''
    if (form.type === 'multi' && inviteTargets.length) {
      const sent = await sendInvitesToUsers(created.id, inviteTargets)
      inviteNote = sent ? `，已向 ${sent} 人发送加入邀请` : ''
    }
    show.value = false
    await load()
    inviteLedgerId.value = created.id
    success.value = `已创建账本「${created.name}」（ID: ${created.id}）${inviteNote}`
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '创建失败'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.member-add-block {
  display: grid;
  gap: 0.5rem;
}
.member {
  max-width: none;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(7rem, 1fr));
  gap: 0.5rem;
  align-items: center;
}
.member input,
.member :deep(.app-select) {
  max-width: 10rem;
}
.custom-schema { margin: 0.5rem 0 1rem; padding: 0.75rem; border: 1px dashed var(--border); border-radius: 8px; }
.check { font-size: 0.75rem; display: flex; align-items: center; gap: 0.25rem; white-space: nowrap; }
.readonly-row { display: grid; gap: 0.35rem; }
.uid { color: var(--accent); font-size: 0.875rem; }
.empty { padding: 1.5rem; text-align: center; }
.alert-success { background: rgba(34, 197, 94, 0.12); border: 1px solid rgba(34, 197, 94, 0.35); color: #4ade80; padding: 0.65rem 0.85rem; border-radius: 8px; margin-bottom: 0.75rem; }
.muted { color: var(--text-muted); }
.header-actions { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.page-header { display: flex; align-items: center; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
.panel-highlight { border-color: var(--accent, #3b82f6); }
.section-hint { font-size: 0.8125rem; color: var(--text-muted); margin: 0 0 0.75rem; }
.invite-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 0.65rem 0;
  border-bottom: 1px solid var(--border);
  flex-wrap: wrap;
}
.invite-row .mono { font-size: 0.8125rem; color: var(--text-muted); margin-left: 0.5rem; }
.invite-row .inviter { display: block; font-size: 0.8125rem; color: var(--text-muted); margin-top: 0.2rem; }
.invite-row.compact { padding: 0.4rem 0; }
.outgoing-list { margin-top: 1rem; padding-top: 0.75rem; border-top: 1px dashed var(--border); }
.outgoing-list h4 { margin: 0 0 0.5rem; font-size: 0.875rem; font-weight: 600; }
.field-hint { font-size: 0.8125rem; color: var(--text-muted); margin: 0 0 0.5rem; }
.member-add-block { display: grid; gap: 0.5rem; margin-bottom: 0.75rem; }
.form-row { display: grid; gap: 0.35rem; margin-bottom: 0.75rem; max-width: 28rem; }
.form-row label { font-size: 0.8125rem; color: var(--text-muted); }
.row-actions { display: flex; gap: 0.75rem; align-items: center; flex-wrap: wrap; }
.btn-link {
  background: none;
  border: none;
  color: var(--accent);
  cursor: pointer;
  font-size: inherit;
  padding: 0;
  text-decoration: underline;
}
</style>
