<template>
  <div class="page">
    <header class="page-header">
      <h2>账本管理</h2>
      <button class="btn-primary" @click="openCreate">创建账本</button>
    </header>
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="success" class="alert alert-success">{{ success }}</div>
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
              <td><router-link :to="`/ledgers/${l.id}`">详情</router-link></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div v-if="show" class="modal">
      <form class="modal-card" @submit.prevent="create">
        <h3>创建账本</h3>
        <p class="hint">账本 ID 与链上地址由系统自动生成（雪花 ID + HD 钱包 BIP44）</p>
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
          <p class="hint">添加自定义列（至少 1 个必填字段）</p>
          <div v-for="(f, i) in form.customFields" :key="i" class="form-row member">
            <input v-model="f.key" placeholder="字段 key（英文）" />
            <input v-model="f.label" placeholder="显示名" />
            <AppSelect v-model="f.type" sm class="member-select" :options="FIELD_TYPE_OPTIONS" />
            <label class="check"><input type="checkbox" v-model="f.required" /> 必填</label>
            <button type="button" class="btn-ghost" @click="form.customFields.splice(i, 1)">删</button>
          </div>
          <button type="button" class="btn-ghost" @click="addCustomField">+ 字段</button>
        </div>
        <p v-else class="hint">字段：{{ selectedTemplateFields }}</p>
        <div v-if="form.type === 'multi'" class="form-row">
          <label class="inline-check">
            <input v-model="form.enableE2E" type="checkbox" />
            启用组级端到端加密（F19）
          </label>
        </div>
        <div v-if="form.type === 'multi' && form.enableE2E" class="form-row">
          <label>加密口令（仅本机解密，勿丢失）</label>
          <input v-model="form.e2ePassphrase" type="password" placeholder="创建后用于加解密记账数据" />
        </div>
        <template v-if="form.type === 'multi'">
          <div class="form-row">
            <label>其他成员</label>
            <p class="hint inline">填写好友的用户 ID（创建者已自动加入）</p>
          </div>
          <div v-for="(m, i) in form.otherMembers" :key="i" class="form-row member">
            <input v-model="m.id" placeholder="成员用户 ID" required />
            <button v-if="form.otherMembers.length > 1" type="button" class="btn-ghost" @click="form.otherMembers.splice(i, 1)">删</button>
          </div>
          <button type="button" class="btn-ghost" @click="form.otherMembers.push({ id: '' })">+ 成员</button>
        </template>
        <div style="display:flex;gap:0.5rem;justify-content:flex-end;margin-top:1rem">
          <button type="button" class="btn-ghost" @click="show = false">取消</button>
          <button class="btn-primary" :disabled="saving">{{ saving ? '创建中…' : '创建' }}</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'
import AppSelect from '../components/AppSelect.vue'
import { DEFAULT_ENTRY_SCHEMA, FIELD_TYPE_OPTIONS } from '../utils/entrySchema'
import { buildEncryptionForCreate, saveLocalGroupKey } from '../utils/e2eCrypto'

const auth = useAuthStore()
const list = ref([])
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
  otherMembers: [{ id: '' }],
  enableE2E: false,
  e2ePassphrase: '',
})

const ledgerTypeOptions = [
  { value: 'private', label: '私人（1人）' },
  { value: 'multi', label: '多人（≥2人）' },
]

const templateOptions = computed(() => [
  ...templates.value.map((t) => ({
    value: t.templateId,
    label: `${templateLabel(t)}${t.builtin ? '（内置）' : ''}`,
  })),
  { value: 'custom', label: '临时自定义（不保存）' },
])

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
  form.otherMembers = [{ id: '' }]
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
  if (form.type === 'multi' && !form.otherMembers.length) {
    form.otherMembers = [{ id: '' }]
  }
})

async function load() {
  const data = await api.listLedgers()
  list.value = Array.isArray(data) ? data : []
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
  const others = form.otherMembers.map((m) => m.id.trim()).filter(Boolean)
  if (others.length < 1) {
    throw new ApiError('多人账本至少需要 1 名其他成员', 400)
  }
  return [{ id: uid, address: '' }, ...others.map((id) => ({ id, address: '' }))]
}

async function create() {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    const members = buildMembers()
    let encryption = { enabled: false }
    let groupKey = ''
    if (form.type === 'multi' && form.enableE2E && form.e2ePassphrase) {
      const enc = await buildEncryptionForCreate(members, auth.user.id, form.e2ePassphrase, 'new')
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
    show.value = false
    await load()
    success.value = `已创建账本「${created.name}」（ID: ${created.id}）`
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '创建失败'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
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
.hint { font-size: 0.75rem; color: var(--text-muted); margin: 0 0 0.5rem; }
.hint.inline { margin: 0; }
.readonly-row { display: grid; gap: 0.35rem; }
.uid { color: var(--accent); font-size: 0.875rem; }
.empty { padding: 1.5rem; text-align: center; }
.alert-success { background: rgba(34, 197, 94, 0.12); border: 1px solid rgba(34, 197, 94, 0.35); color: #4ade80; padding: 0.65rem 0.85rem; border-radius: 8px; margin-bottom: 0.75rem; }
.muted { color: var(--text-muted); }
</style>
