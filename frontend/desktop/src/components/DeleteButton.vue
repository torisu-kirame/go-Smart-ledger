<template>
  <button
    type="button"
    class="btn-delete"
    :class="{ 'btn-delete--icon-only': iconOnly, 'btn-delete--sm': sm }"
    :disabled="disabled"
    :title="title || (iconOnly ? label : undefined)"
    :aria-label="label"
    @click="$emit('click', $event)"
  >
    <svg
      class="btn-delete__icon"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden="true"
    >
      <polyline points="3 6 5 6 21 6" />
      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
      <line x1="10" y1="11" x2="10" y2="17" />
      <line x1="14" y1="11" x2="14" y2="17" />
    </svg>
    <span v-if="!iconOnly" class="btn-delete__label">{{ label }}</span>
  </button>
</template>

<script setup>
defineProps({
  label: { type: String, default: '删除' },
  iconOnly: { type: Boolean, default: false },
  sm: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  title: { type: String, default: '' },
})

defineEmits(['click'])
</script>
