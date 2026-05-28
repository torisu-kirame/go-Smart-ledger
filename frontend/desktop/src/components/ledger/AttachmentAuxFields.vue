<template>
  <div class="aux-fields">
    <label class="aux-fields__label">辅助核算</label>
    <div class="aux-fields__grid">
      <input
        :value="model.department"
        type="text"
        placeholder="部门"
        class="field-sm"
        :disabled="disabled"
        @input="emitField('department', $event.target.value)"
      />
      <input
        :value="model.project"
        type="text"
        placeholder="项目"
        class="field-sm"
        :disabled="disabled"
        @input="emitField('project', $event.target.value)"
      />
      <input
        :value="model.counterparty"
        type="text"
        placeholder="往来"
        class="field-sm"
        :disabled="disabled"
        @input="emitField('counterparty', $event.target.value)"
      />
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  model: {
    type: Object,
    default: () => ({ department: '', project: '', counterparty: '' }),
  },
  disabled: { type: Boolean, default: false },
})

const emit = defineEmits(['update:model'])

function emitField(key, value) {
  emit('update:model', { ...props.model, [key]: value })
}
</script>

<style scoped>
.aux-fields__label {
  display: block;
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-bottom: 0.35rem;
}
.aux-fields__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.35rem;
}
@media (max-width: 640px) {
  .aux-fields__grid {
    grid-template-columns: 1fr;
  }
}
</style>
