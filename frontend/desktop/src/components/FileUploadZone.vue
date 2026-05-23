<template>
  <label
    class="upload-zone"
    :class="{ 'is-dragover': dragover, 'is-disabled': disabled, 'is-compact': compact }"
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
      @change="onInputChange"
    />
    <div class="upload-body">
      <div class="upload-icon" aria-hidden="true">↑</div>
      <p class="upload-title">{{ title }}</p>
      <p v-if="fileName" class="upload-file mono">{{ fileName }}</p>
    </div>
  </label>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  accept: { type: String, default: '' },
  title: { type: String, default: '拖拽文件到此处，或点击选择' },
  disabled: { type: Boolean, default: false },
  compact: { type: Boolean, default: false },
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
  min-height: 140px;
  max-width: 28rem;
  padding: 1.25rem 1rem;
  border: 2px dashed var(--border);
  border-radius: var(--radius);
  background: color-mix(in srgb, var(--accent) 6%, var(--bg));
  color: var(--text-muted);
  text-align: center;
  cursor: pointer;
  transition: border-color 0.18s ease, background-color 0.18s ease, box-shadow 0.18s ease;
}

.upload-zone.is-compact {
  min-height: 100px;
  max-width: 16rem;
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
  opacity: 0.65;
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
  gap: 0.35rem;
  pointer-events: none;
}

.upload-icon {
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--accent-soft);
  color: var(--accent);
  font-size: 1.1rem;
  font-weight: 700;
}

.upload-title {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text);
  font-weight: 600;
}

.upload-file {
  margin: 0.25rem 0 0;
  font-size: 0.78rem;
  color: var(--accent);
  word-break: break-all;
}
</style>
