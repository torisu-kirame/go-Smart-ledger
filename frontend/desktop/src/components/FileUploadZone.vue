<template>
  <label
    class="upload-zone"
    :class="{
      'is-dragover': dragover,
      'is-disabled': disabled,
      'is-compact': compact,
      'is-block': block,
    }"
    @dragover.prevent="onDragover"
    @dragleave.prevent="onDragleave"
    @drop.prevent="onDrop"
  >
    <input
      ref="inputRef"
      type="file"
      class="upload-input"
      :accept="accept"
      :disabled="disabled"
      :multiple="multiple"
      @change="onInputChange"
    />
    <div class="upload-body">
      <div class="upload-icon" aria-hidden="true">
        <AppIcon name="upload" size="md" />
      </div>
      <p class="upload-title">{{ title }}</p>
      <p v-if="hint" class="upload-hint">{{ hint }}</p>
      <p v-if="fileName" class="upload-file mono">{{ fileName }}</p>
    </div>
  </label>
</template>

<script setup>
import { ref } from 'vue'
import AppIcon from './AppIcon.vue'

const props = defineProps({
  accept: { type: String, default: '' },
  title: { type: String, default: '点击或拖拽上传' },
  hint: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  compact: { type: Boolean, default: false },
  block: { type: Boolean, default: false },
  multiple: { type: Boolean, default: false },
})

const emit = defineEmits(['file'])

const inputRef = ref(null)
const dragover = ref(false)
const fileName = ref('')

function emitFile(file) {
  if (!file) return
  fileName.value = file.name
  emit('file', file)
}

function onInputChange(ev) {
  const file = ev.target.files?.[0]
  emitFile(file)
  ev.target.value = ''
}

function onDragover() {
  if (props.disabled) return
  dragover.value = true
}

function onDragleave() {
  dragover.value = false
}

function onDrop(ev) {
  dragover.value = false
  if (props.disabled) return
  const file = ev.dataTransfer?.files?.[0]
  emitFile(file)
}

function clear() {
  fileName.value = ''
  if (inputRef.value) inputRef.value.value = ''
}

defineExpose({ clear })
</script>

<style scoped>
.upload-zone {
  position: relative;
  display: block;
  width: 100%;
  max-width: 28rem;
  min-height: 132px;
  padding: 1.35rem 1rem;
  border: 2px dashed color-mix(in srgb, var(--accent) 55%, var(--border));
  border-radius: var(--radius);
  background: color-mix(in srgb, var(--accent) 8%, var(--bg-card));
  color: var(--text-muted);
  text-align: center;
  cursor: pointer;
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease,
    box-shadow 0.18s ease;
}

.upload-zone.is-block {
  max-width: none;
}

.upload-zone.is-compact {
  min-height: 96px;
  max-width: 100%;
  padding: 0.85rem 0.75rem;
}

.upload-zone:hover:not(.is-disabled) {
  border-color: var(--accent);
  background: var(--accent-soft);
}

.upload-zone.is-dragover {
  border-color: var(--accent);
  background: var(--accent-soft);
  box-shadow: 0 0 0 3px var(--accent-soft);
}

.upload-zone.is-disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.upload-input {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.upload-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.3rem;
  pointer-events: none;
}

.upload-icon {
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: color-mix(in srgb, var(--accent) 22%, transparent);
  color: var(--accent);
  margin-bottom: 0.15rem;
}

.upload-zone.is-compact .upload-icon {
  width: 2.25rem;
  height: 2.25rem;
}

.upload-title {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text);
  font-weight: 600;
  line-height: 1.35;
}

.upload-zone.is-compact .upload-title {
  font-size: 0.8125rem;
  font-weight: 500;
}

.upload-hint {
  margin: 0;
  font-size: 0.78rem;
  color: var(--text-muted);
  line-height: 1.35;
}

.upload-file {
  margin: 0.2rem 0 0;
  font-size: 0.78rem;
  color: var(--accent);
  word-break: break-all;
}
</style>
