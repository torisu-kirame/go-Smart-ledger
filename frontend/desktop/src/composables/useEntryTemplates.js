import { ref } from 'vue'
import { api } from '../api/http'

export const ENTRY_TEMPLATES_CHANGED = 'smart-ledger-entry-templates-changed'

const templates = ref([])
const loading = ref(false)
const loaded = ref(false)

function notifyChanged() {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new Event(ENTRY_TEMPLATES_CHANGED))
  }
}

/** 账户级通用模板（所有简单流水账本共享） */
export async function loadEntryTemplates(force = false) {
  if (loading.value) return templates.value
  if (loaded.value && !force) return templates.value
  loading.value = true
  try {
    const res = await api.listEntryTemplates()
    templates.value = res.templates || []
    loaded.value = true
    return templates.value
  } finally {
    loading.value = false
  }
}

export function notifyEntryTemplatesChanged() {
  notifyChanged()
}

export function useEntryTemplates() {
  return {
    templates,
    loading,
    loaded,
    loadEntryTemplates,
    refreshEntryTemplates: () => loadEntryTemplates(true),
    notifyEntryTemplatesChanged,
  }
}

/**
 * 模板变更后同步到所有引用该模板的简单流水账本（服务端持久化）。
 */
export async function syncEntryTemplateToLedgers(templateId, fields) {
  if (!templateId || templateId === 'custom' || !fields?.length) return null
  return api.syncEntryTemplate({ templateId, fields })
}

export function findEntryTemplate(templateId, list = templates.value) {
  return (list || []).find((t) => t.templateId === templateId) || null
}
