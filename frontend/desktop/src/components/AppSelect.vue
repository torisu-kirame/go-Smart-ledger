<template>
  <div
    ref="rootRef"
    class="app-select"
    :class="{ open: isOpen, disabled, 'field-sm': sm }"
  >
    <button
      type="button"
      class="app-select-trigger"
      :disabled="disabled"
      :aria-expanded="isOpen"
      aria-haspopup="listbox"
      @click="toggle"
    >
      <span class="app-select-value" :class="{ placeholder: !hasValue }">{{ displayText }}</span>
      <span class="app-select-chevron" aria-hidden="true" />
    </button>
    <ul v-show="isOpen" class="app-select-menu" role="listbox">
      <li
        v-for="opt in options"
        :key="String(opt.value)"
        role="option"
        class="app-select-option"
        :class="{
          selected: isSelected(opt.value),
          disabled: opt.disabled,
        }"
        :aria-selected="isSelected(opt.value)"
        @click="choose(opt)"
      >
        {{ opt.label }}
      </li>
    </ul>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  options: { type: Array, default: () => [] },
  placeholder: { type: String, default: '请选择' },
  disabled: { type: Boolean, default: false },
  sm: { type: Boolean, default: false },
})

const emit = defineEmits(['update:modelValue', 'change'])

const rootRef = ref(null)
const isOpen = ref(false)

const hasValue = computed(() => {
  const v = props.modelValue
  return v !== '' && v !== null && v !== undefined
})

const displayText = computed(() => {
  if (!hasValue.value) return props.placeholder
  const hit = props.options.find((o) => String(o.value) === String(props.modelValue))
  return hit?.label ?? props.placeholder
})

function isSelected(value) {
  return String(value) === String(props.modelValue)
}

function toggle() {
  if (props.disabled) return
  isOpen.value = !isOpen.value
}

function close() {
  isOpen.value = false
}

function choose(opt) {
  if (opt.disabled) return
  const next = opt.value
  if (String(next) !== String(props.modelValue)) {
    emit('update:modelValue', next)
    emit('change', next)
  }
  close()
}

function onDocClick(ev) {
  if (!rootRef.value?.contains(ev.target)) close()
}

function onKeydown(ev) {
  if (ev.key === 'Escape') close()
}

onMounted(() => {
  document.addEventListener('click', onDocClick)
  document.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  document.removeEventListener('click', onDocClick)
  document.removeEventListener('keydown', onKeydown)
})
</script>

<style scoped>
.app-select {
  position: relative;
  width: 100%;
  max-width: var(--field-max, 22rem);
}

.app-select.field-sm {
  max-width: var(--field-sm, 16rem);
}

.app-select-trigger {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.55rem 0.8rem;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text);
  font: inherit;
  text-align: left;
  cursor: pointer;
  transition: border-color 0.16s ease, box-shadow 0.16s ease;
}

.app-select-trigger:hover:not(:disabled) {
  border-color: color-mix(in srgb, var(--accent) 50%, var(--border));
}

.app-select.open .app-select-trigger,
.app-select-trigger:focus-visible {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-soft);
}

.app-select-value.placeholder {
  color: var(--text-muted);
}

.app-select-chevron {
  width: 0.55rem;
  height: 0.55rem;
  border-right: 2px solid var(--text-muted);
  border-bottom: 2px solid var(--text-muted);
  transform: rotate(45deg) translateY(-2px);
  flex-shrink: 0;
  transition: transform 0.15s ease;
}

.app-select.open .app-select-chevron {
  transform: rotate(-135deg) translateY(2px);
}

.app-select-menu {
  position: absolute;
  z-index: 40;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  margin: 0;
  padding: 0.35rem;
  list-style: none;
  max-height: 240px;
  overflow-y: auto;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--bg-card);
  box-shadow: var(--shadow-lg);
}

.app-select-option {
  padding: 0.5rem 0.65rem;
  border-radius: 8px;
  font-size: 0.875rem;
  color: var(--text);
  cursor: pointer;
  transition: background-color 0.12s ease, color 0.12s ease;
}

.app-select-option:hover:not(.disabled) {
  background: var(--hover);
}

.app-select-option.selected {
  background: var(--accent-soft);
  color: var(--accent);
  font-weight: 600;
}

.app-select-option.disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.app-select.disabled .app-select-trigger {
  opacity: 0.55;
  cursor: not-allowed;
}
</style>
