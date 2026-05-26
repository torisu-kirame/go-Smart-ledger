<template>
  <div class="form-grid" :class="{ 'form-grid--column': column }">
    <div v-for="f in schema.fields" :key="f.key" class="form-row">
      <label>{{ f.label }}<span v-if="f.required" class="req">*</span></label>
      <AppSelect
        v-if="f.type === 'user'"
        v-model="model[f.key]"
        :options="memberOptions"
        placeholder="请选择"
      />
      <input
        v-else-if="f.type === 'date'"
        v-model="model[f.key]"
        type="date"
        :required="f.required"
      />
      <input
        v-else-if="f.type === 'number'"
        v-model="model[f.key]"
        type="text"
        inputmode="decimal"
        :placeholder="f.label"
        :required="f.required"
      />
      <input
        v-else
        v-model="model[f.key]"
        type="text"
        :placeholder="f.label"
        :required="f.required"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import AppSelect from './AppSelect.vue'

const props = defineProps({
  schema: { type: Object, required: true },
  model: { type: Object, required: true },
  members: { type: Array, default: () => [] },
  /** 纵向单列：标签在上、输入在下 */
  column: { type: Boolean, default: false },
})

const memberOptions = computed(() =>
  props.members.map((m) => ({
    value: m.id,
    label: m.nickname || m.username || m.id,
  }))
)
</script>

<style scoped>
.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(11rem, 1fr));
  gap: 0.75rem;
  max-width: 48rem;
}
.form-grid .form-row {
  max-width: none;
  margin-bottom: 0;
}
.form-grid input,
.form-grid :deep(.app-select) {
  max-width: 100%;
}
.form-grid--column {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: none;
}
.form-grid--column .form-row {
  display: grid;
  gap: 0.35rem;
}
.form-grid--column label {
  font-size: 0.8125rem;
  color: var(--text-muted);
}
.req { color: var(--danger); margin-left: 0.15rem; }
</style>
