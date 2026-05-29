<template>
  <section class="detail-card attach-panel">
    <h3 class="detail-card__title">流水附件</h3>
    <p class="field-hint">为当前表的流水记录（EntryAdded 事件 Seq）上传佐证文件，可标注部门 / 项目 / 往来辅助信息。</p>
    <div class="form-row inline">
      <label>流水 Seq</label>
      <input v-model.number="attachSeq" type="number" min="1" class="field-sm" />
    </div>
    <AttachmentAuxFields v-model:model="uploadAux" :disabled="busy" />
    <FileUploadZone
      block
      :disabled="busy || !attachSeq"
      title="点击或拖拽上传附件"
      :hint="attachSeq ? '' : '请先填写流水 Seq'"
      @file="onFile"
    />
    <div v-if="!list.length" class="muted">暂无附件</div>
    <ul v-else class="attach-list">
      <li v-for="a in list" :key="a.id" class="attach-item">
        <div class="attach-item__head">
          <span>#{{ a.entrySeq }}</span>
          <span>{{ a.filename }}</span>
          <span v-if="a.cid" class="mono muted">CID {{ a.cid.slice(0, 12) }}…</span>
        </div>
        <AttachmentAuxFields
          :model="draftAux[a.id]"
          :disabled="savingId === a.id"
          @update:model="(v) => (draftAux[a.id] = v)"
        />
        <button
          type="button"
          class="btn-ghost btn-sm"
          :disabled="savingId === a.id"
          @click="saveAux(a.id)"
        >
          保存辅助核算
        </button>
      </li>
    </ul>
  </section>
</template>

<script setup>
import { reactive, ref, watch } from 'vue'
import { api, ApiError } from '../../api/http'
import FileUploadZone from '../FileUploadZone.vue'
import AttachmentAuxFields from './AttachmentAuxFields.vue'
import { useNotify } from '../../composables/useNotify'

const props = defineProps({
  ledgerId: { type: String, required: true },
  tableId: { type: String, default: 'default' },
})

const notify = useNotify()
const attachSeq = ref(1)
const list = ref([])
const busy = ref(false)
const savingId = ref('')
const uploadAux = ref({ department: '', project: '', counterparty: '' })
const draftAux = reactive({})

function auxFrom(att) {
  return {
    department: att.auxiliary?.department || '',
    project: att.auxiliary?.project || '',
    counterparty: att.auxiliary?.counterparty || '',
  }
}

async function load() {
  if (!props.ledgerId) return
  try {
    const res = await api.listAccountingAttachments(props.ledgerId, {
      tableId: props.tableId,
    })
    list.value = res.attachments || []
    for (const a of list.value) {
      draftAux[a.id] = auxFrom(a)
    }
  } catch {
    list.value = []
  }
}

async function onFile(file) {
  if (!file || !attachSeq.value) return
  busy.value = true
  try {
    await api.uploadAccountingAttachment(props.ledgerId, attachSeq.value, file, {
      tableId: props.tableId,
      auxiliary: uploadAux.value,
    })
    notify.success('附件已上传')
    await load()
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '上传失败')
  } finally {
    busy.value = false
  }
}

async function saveAux(attachId) {
  savingId.value = attachId
  try {
    await api.patchAttachmentAuxiliary(props.ledgerId, attachId, draftAux[attachId] || {})
    notify.success('辅助核算已保存')
    await load()
  } catch (e) {
    notify.error(e instanceof ApiError ? e.message : '保存失败')
  } finally {
    savingId.value = ''
  }
}

watch(
  () => [props.ledgerId, props.tableId],
  () => {
    load()
  },
  { immediate: true }
)
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
  margin: 0 0 0.5rem;
  font-size: 0.8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.field-hint {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: 0 0 0.75rem;
}
.form-row.inline {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}
.attach-list {
  margin: 0.5rem 0 0;
  padding: 0;
  list-style: none;
  font-size: 0.875rem;
}
.attach-item {
  padding: 0.65rem 0;
  border-bottom: 1px solid var(--border);
}
.attach-item__head {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}
.btn-sm {
  font-size: 0.8125rem;
  margin-top: 0.35rem;
}
</style>
