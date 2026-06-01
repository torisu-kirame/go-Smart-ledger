<template>
  <div class="page" :class="{ 'page--embedded': embedded }">
    <PageHeader v-if="!embedded" :crumbs="crumbs">
      <template #actions>
        <button class="btn-primary" @click="openCreate">新建模板</button>
      </template>
    </PageHeader>
    <div v-else class="embedded-toolbar">
      <button class="btn-primary" @click="openCreate">新建模板</button>
    </div>
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="msg" class="alert alert-success">{{ msg }}</div>

    <p class="section-hint panel-hint">
      <strong>账户通用模板</strong>：在此创建或修改后，将自动同步到所有使用相同模板的简单流水账本；任意账本内的「模板」页看到的列表一致。
    </p>

    <div class="panel">
      <div v-if="!templates.length" class="muted empty">加载中…</div>
      <div v-for="t in templates" :key="t.templateId" class="tpl-card">
        <div class="tpl-head">
          <div>
            <h3>
              {{ t.name }}
              <span v-if="t.builtin" class="badge badge-ok">内置</span>
              <span v-else class="badge badge-multi">自定义</span>
            </h3>
          </div>
          <div v-if="!t.builtin" class="actions">
            <button class="btn-ghost" @click="openEdit(t)">编辑</button>
            <DeleteButton @click="remove(t)" />
          </div>
        </div>
        <div class="fields">
          <span v-for="f in t.fields" :key="f.key" class="field-chip">
            {{ f.label }}<span v-if="f.required" class="req">*</span>
          </span>
        </div>
      </div>
    </div>

    <div v-if="showModal" class="modal">
      <form class="modal-card wide" @submit.prevent="save">
        <h3>{{ editing ? '编辑模板' : '新建模板' }}</h3>
        <div class="form-row">
          <label>模板名称</label>
          <input v-model="form.name" required placeholder="例如：家庭日常记账" />
        </div>
        <div v-for="(f, i) in form.fields" :key="i" class="field-row">
          <input v-model="f.key" placeholder="key（英文）" required />
          <input v-model="f.label" placeholder="显示名" required />
          <AppSelect v-model="f.type" sm :options="FIELD_TYPE_OPTIONS" />
          <label class="check"><input type="checkbox" v-model="f.required" /> 必填</label>
          <DeleteButton icon-only sm title="删除字段" @click="form.fields.splice(i, 1)" />
        </div>
        <button type="button" class="btn-ghost" @click="addField">+ 字段</button>
        <div class="foot-actions">
          <button type="button" class="btn-ghost" @click="showModal = false">取消</button>
          <button class="btn-primary" :disabled="saving">{{ saving ? '保存中…' : '保存' }}</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, ApiError } from '../api/http'
import AppSelect from '../components/AppSelect.vue'
import PageHeader from '../components/PageHeader.vue'
import { usePageCrumbs } from '../composables/usePageCrumbs'
import {
  loadEntryTemplates,
  notifyEntryTemplatesChanged,
  syncEntryTemplateToLedgers,
  useEntryTemplates,
} from '../composables/useEntryTemplates'
import DeleteButton from '../components/DeleteButton.vue'
import { DEFAULT_ENTRY_SCHEMA, FIELD_TYPE_OPTIONS } from '../utils/entrySchema'

const route = useRoute()
const embedded = computed(() => Boolean(route.params.id))
const { crumbs } = usePageCrumbs()
const { templates, refreshEntryTemplates } = useEntryTemplates()
const error = ref('')
const msg = ref('')
const showModal = ref(false)
const saving = ref(false)
const editing = ref(null)
const form = reactive({
  name: '',
  fields: [{ key: 'bookkeeper', label: '记账人', type: 'user', required: true }],
})

const typeLabels = { text: '文本', number: '数字', date: '日期', user: '用户' }
function typeLabel(t) {
  return typeLabels[t] || t
}

async function load() {
  error.value = ''
  try {
    await refreshEntryTemplates()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '加载失败'
  }
}

async function afterTemplateMutation(templateId, fields) {
  notifyEntryTemplatesChanged()
  try {
    const res = await syncEntryTemplateToLedgers(templateId, fields)
    if (res?.updatedLedgers > 0) {
      msg.value = `已同步到 ${res.updatedLedgers} 个账本`
    }
  } catch {
    /* 模板已保存，账本同步失败不阻断 */
  }
}

function resetFormFromDefault() {
  form.name = ''
  form.fields = DEFAULT_ENTRY_SCHEMA.fields.map((f) => ({ ...f }))
}

function openCreate() {
  editing.value = null
  resetFormFromDefault()
  form.name = '我的记账模板'
  showModal.value = true
}

function openEdit(t) {
  editing.value = t
  form.name = t.name
  form.fields = t.fields.map((f) => ({ ...f }))
  showModal.value = true
}

function addField() {
  form.fields.push({ key: '', label: '', type: 'text', required: false })
}

async function save() {
  saving.value = true
  error.value = ''
  msg.value = ''
  const body = {
    name: form.name.trim(),
    fields: form.fields.filter((f) => f.key.trim() && f.label.trim()),
  }
  try {
    if (editing.value) {
      await api.updateEntryTemplate(editing.value.templateId, body)
      msg.value = '模板已更新'
      await afterTemplateMutation(editing.value.templateId, body.fields)
    } else {
      const created = await api.createEntryTemplate(body)
      msg.value = '模板已创建'
      await afterTemplateMutation(created.templateId, body.fields)
    }
    showModal.value = false
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '保存失败'
  } finally {
    saving.value = false
  }
}

async function remove(t) {
  if (!confirm(`确定删除模板「${t.name}」？`)) return
  error.value = ''
  try {
    await api.deleteEntryTemplate(t.templateId)
    msg.value = '已删除'
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '删除失败'
  }
}

onMounted(load)
watch(() => route.params.id, load)
</script>

<style scoped>
.muted { color: var(--text-muted); font-size: 0.875rem; }
.empty { padding: 2rem; text-align: center; }
.tpl-card {
  background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius);
  padding: 1rem; margin-bottom: 0.75rem;
}
.panel-hint {
  margin: 0 0 1rem;
  font-size: 0.875rem;
  color: var(--text-muted);
  line-height: 1.45;
}
.tpl-head { display: flex; justify-content: space-between; gap: 1rem; }
.tpl-head h3 { margin: 0 0 0.35rem; font-size: 1rem; }
.actions { display: flex; gap: 0.35rem; flex-shrink: 0; }
.fields { display: flex; flex-wrap: wrap; gap: 0.35rem; margin-top: 0.75rem; }
.field-chip {
  font-size: 0.75rem; background: rgba(61,139,253,.1); border: 1px solid var(--border);
  padding: 0.2rem 0.5rem; border-radius: 6px;
}
.req { color: var(--danger); }
.modal { position: fixed; inset: 0; background: rgba(0,0,0,.65); display: flex; align-items: center; justify-content: center; z-index: 50; }
.modal-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: 12px; padding: 1.5rem; width: 480px; max-height: 90vh; overflow: auto; }
.modal-card.wide { width: 640px; }
.field-row { display: grid; grid-template-columns: 1fr 1fr 100px auto auto; gap: 0.5rem; margin: 0.35rem 0; align-items: center; }
.check { font-size: 0.75rem; display: flex; align-items: center; gap: 0.25rem; white-space: nowrap; }
.foot-actions { display: flex; gap: 0.5rem; justify-content: flex-end; margin-top: 1rem; }
.alert-success { background: rgba(34, 197, 94, 0.12); border: 1px solid rgba(34, 197, 94, 0.35); color: #4ade80; padding: 0.65rem 0.85rem; border-radius: 8px; margin-bottom: 0.75rem; }
.page--embedded { max-width: none; }
.embedded-toolbar { display: flex; justify-content: flex-end; margin-bottom: 0.75rem; }
</style>
