<template>
  <div
    class="ws-tabs"
    @dragover.prevent="onBarDragOver"
    @drop.prevent="onBarDrop"
  >
    <div class="ws-tabs__list" role="tablist">
      <button
        v-for="tab in panes.openTabs"
        :key="tab.id"
        type="button"
        class="ws-tab"
        role="tab"
        draggable="true"
        :class="{
          'ws-tab--active': isActive(tab),
          'ws-tab--dragging': dragId === tab.id,
        }"
        :title="t('layout.split.dragHint')"
        @click="activate(tab)"
        @dragstart="onDragStart($event, tab)"
        @dragend="onDragEnd"
      >
        <AppIcon :name="tab.icon" size="sm" />
        <span>{{ t(tab.labelKey) }}</span>
        <span v-if="panes.split && panes.leftId === tab.id" class="ws-tab__side">L</span>
        <span v-if="panes.split && panes.rightId === tab.id" class="ws-tab__side">R</span>
      </button>
    </div>
    <button
      v-if="panes.split"
      type="button"
      class="ws-tabs__exit icon-btn btn-ghost btn-sm"
      :title="t('layout.split.exit')"
      @click="exitSplit"
    >
      <AppIcon name="x" size="sm" />
      <span>{{ t('layout.split.exit') }}</span>
    </button>

    <div v-if="dragId" class="ws-drop-layer" aria-hidden="true">
      <div
        class="ws-drop-zone ws-drop-zone--left"
        :class="{ 'ws-drop-zone--hot': dropSide === 'left' }"
        @dragover.prevent="dropSide = 'left'"
        @dragleave="onZoneLeave('left')"
        @drop.prevent="onDropSide('left')"
      >
        <span>{{ t('layout.split.openLeft') }}</span>
      </div>
      <div
        class="ws-drop-zone ws-drop-zone--right"
        :class="{ 'ws-drop-zone--hot': dropSide === 'right' }"
        @dragover.prevent="dropSide = 'right'"
        @dragleave="onZoneLeave('right')"
        @drop.prevent="onDropSide('right')"
      >
        <span>{{ t('layout.split.openRight') }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppIcon from './AppIcon.vue'
import { useI18n } from '../composables/useI18n'
import { useWorkspacePanes } from '../stores/workspacePanes'

const panes = useWorkspacePanes()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const dragId = ref(null)
const dropSide = ref(null)

const show = computed(() => panes.openTabs.length > 0)

defineExpose({ show })

function isActive(tab) {
  if (panes.split) {
    const focused = panes.focus === 'right' ? panes.rightId : panes.leftId
    return focused === tab.id
  }
  return tabFromRoute() === tab.id
}

function tabFromRoute() {
  if (route.path.startsWith('/assistant')) return 'assistant'
  if (route.path === '/ledgers' || route.path.startsWith('/ledgers/')) return 'ledgers'
  return null
}

function activate(tab) {
  if (panes.split) {
    if (panes.leftId === tab.id) panes.setFocus('left')
    else if (panes.rightId === tab.id) panes.setFocus('right')
    else panes.placeOnSide(tab.id, panes.focus)
  } else {
    panes.openSingle(tab.id)
  }
  if (route.path !== tab.path) router.push(tab.path)
}

function exitSplit() {
  const keep = panes.focus === 'right' ? panes.rightId : panes.leftId
  panes.exitSplit(keep)
  const tab = panes.leftTab
  if (tab && route.path !== tab.path) router.push(tab.path)
}

function onDragStart(e, tab) {
  dragId.value = tab.id
  dropSide.value = null
  e.dataTransfer.effectAllowed = 'move'
  e.dataTransfer.setData('text/plain', tab.id)
  try {
    e.dataTransfer.setData('application/x-smart-ledger-tab', tab.id)
  } catch {
    /* ignore */
  }
}

function onDragEnd() {
  dragId.value = null
  dropSide.value = null
}

function onBarDragOver() {
  /* allow drop on zones only */
}

function onBarDrop() {
  /* zones handle drops */
}

function onZoneLeave(side) {
  if (dropSide.value === side) dropSide.value = null
}

function onDropSide(side) {
  const id = dragId.value
  if (!id) return
  panes.placeOnSide(id, side)
  const tab = panes.openTabs.find((t) => t.id === id)
  if (tab && route.path !== tab.path) router.push(tab.path)
  dragId.value = null
  dropSide.value = null
}
</script>

<style scoped>
.ws-tabs {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: -0.35rem 0 0.75rem;
  padding-bottom: 0.55rem;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.ws-tabs__list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  flex: 1;
  min-width: 0;
}

.ws-tab {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.7rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  color: var(--text-muted);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: grab;
}

.ws-tab:active {
  cursor: grabbing;
}

.ws-tab--active {
  color: var(--accent);
  border-color: color-mix(in srgb, var(--accent) 45%, var(--border));
  background: color-mix(in srgb, var(--accent) 12%, transparent);
}

.ws-tab--dragging {
  opacity: 0.55;
}

.ws-tab__side {
  font-size: 0.65rem;
  font-weight: 700;
  padding: 0.05rem 0.28rem;
  border-radius: 4px;
  background: color-mix(in srgb, var(--accent) 18%, transparent);
  color: var(--accent);
}

.ws-tabs__exit {
  flex-shrink: 0;
  gap: 0.25rem;
  font-size: 0.75rem;
}

.ws-drop-layer {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: grid;
  grid-template-columns: 1fr 1fr;
  pointer-events: none;
  background: color-mix(in srgb, var(--bg) 35%, transparent);
}

.ws-drop-zone {
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 4.5rem 0.75rem 0.75rem;
  border: 2px dashed color-mix(in srgb, var(--accent) 45%, var(--border));
  border-radius: var(--radius);
  background: color-mix(in srgb, var(--accent) 8%, transparent);
  color: var(--accent);
  font-weight: 700;
  font-size: 0.95rem;
  transition: background 0.15s, border-color 0.15s;
}

.ws-drop-zone--hot {
  background: color-mix(in srgb, var(--accent) 22%, transparent);
  border-color: var(--accent);
}
</style>
