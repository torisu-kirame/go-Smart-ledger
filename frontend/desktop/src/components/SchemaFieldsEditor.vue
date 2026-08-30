<template>
  <div class="schema-fields">
    <div v-for="(f, i) in model" :key="i" class="schema-fields__row">
      <input v-model="f.key" placeholder="字段 key（英文）" :disabled="disabled" />
      <input v-model="f.label" placeholder="显示名" :disabled="disabled" />
      <AppSelect v-model="f.type" sm class="schema-fields__type" :options="FIELD_TYPE_OPTIONS" :disabled="disabled" />
      <label class="schema-fields__check">
        <input v-model="f.required" type="checkbox" :disabled="disabled" />
        必填
      </label>
      <DeleteButton
        icon-only
        sm
        title="删除字段"
        :disabled="disabled || model.length <= 1"
        @click="remove(i)"
      />
    </div>
    <button type="button" class="btn-ghost" :disabled="disabled" @click="add">+ 添加字段</button>
  </div>
</template>

<script setup>
import AppSelect from './AppSelect.vue'
import DeleteButton from './DeleteButton.vue'
import { FIELD_TYPE_OPTIONS } from '../utils/entrySchema'

const model = defineModel({ type: Array, required: true })
defineProps({
  disabled: { type: Boolean, default: false },
})

function add() {
  model.value = [...model.value, { key: '', label: '', type: 'text', required: false }]
}

function remove(i) {
  if (model.value.length <= 1) return
  model.value = model.value.filter((_, idx) => idx !== i)
}
</script>

<style scoped>
.schema-fields {
  display: grid;
  gap: 0.5rem;
}
.schema-fields__row {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(7rem, 1fr));
  gap: 0.5rem;
  align-items: center;
}
.schema-fields__row input,
.schema-fields__row :deep(.app-select) {
  max-width: 10rem;
}
.schema-fields__check {
  font-size: 0.75rem;
  display: flex;
  align-items: center;
  gap: 0.25rem;
  white-space: nowrap;
  color: var(--text-muted);
}
</style>
