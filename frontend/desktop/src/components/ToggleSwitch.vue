<template>
  <button
    type="button"
    role="switch"
    class="toggle-switch"
    :class="{ 'toggle-switch--on': modelValue, 'toggle-switch--disabled': disabled }"
    :aria-checked="modelValue"
    :aria-label="ariaLabel || label"
    :disabled="disabled"
    @click="onClick"
  >
    <span class="toggle-switch__track" aria-hidden="true">
      <span class="toggle-switch__thumb" />
    </span>
    <span v-if="label" class="toggle-switch__label">{{ label }}</span>
  </button>
</template>

<script setup>
const props = defineProps({
  modelValue: { type: Boolean, default: false },
  label: { type: String, default: '' },
  ariaLabel: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
})

const emit = defineEmits(['update:modelValue'])

function onClick() {
  if (props.disabled) return
  emit('update:modelValue', !props.modelValue)
}
</script>

<style scoped>
.toggle-switch {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
  font: inherit;
  color: var(--text);
  text-align: left;
}

.toggle-switch--disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.toggle-switch__track {
  position: relative;
  flex-shrink: 0;
  width: 2.75rem;
  height: 1.5rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--text-muted) 35%, var(--border));
  transition: background 0.22s ease;
}

.toggle-switch--on .toggle-switch__track {
  background: var(--accent);
}

.toggle-switch__thumb {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 1.25rem;
  height: 1.25rem;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.22);
  transition: transform 0.22s cubic-bezier(0.4, 0, 0.2, 1);
}

.toggle-switch--on .toggle-switch__thumb {
  transform: translateX(1.25rem);
}

.toggle-switch__label {
  font-size: 0.875rem;
  font-weight: 500;
  line-height: 1.35;
}

.toggle-switch:focus-visible .toggle-switch__track {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}
</style>
