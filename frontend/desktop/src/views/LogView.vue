<template>
  <div class="page log-page">
    <PageHeader :crumbs="crumbs" :subtitle="t('logs.subtitle')" />

    <div class="panel log-panel">
      <div class="log-toolbar">
        <span class="muted">{{ t('logs.count', { n: store.logs.length }) }}</span>
        <button type="button" class="btn-ghost" :disabled="!store.logs.length" @click="store.clearLogs()">
          {{ t('logs.clear') }}
        </button>
      </div>
      <div v-if="!store.logs.length" class="muted log-empty">{{ t('logs.empty') }}</div>
      <ul v-else class="log-list">
        <li v-for="entry in store.logs" :key="entry.id" class="log-item" :class="`log-item--${entry.level}`">
          <time class="log-item__time">{{ formatTime(entry.at) }}</time>
          <span class="log-item__text">{{ entry.text }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import { useI18n } from '../composables/useI18n'
import { useNotifyStore } from '../stores/notify'

const { t } = useI18n()
const store = useNotifyStore()

const crumbs = computed(() => [
  { label: t('layout.nav.chain'), to: '/chain' },
  { label: t('layout.nav.logs') },
])

function formatTime(iso) {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}
</script>

<style scoped>
.log-panel {
  padding: 1rem 1.15rem;
}
.log-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}
.log-empty {
  padding: 1.5rem 0;
}
.log-list {
  margin: 0;
  padding: 0;
  list-style: none;
  max-height: calc(100vh - 12rem);
  overflow: auto;
}
.log-item {
  display: grid;
  grid-template-columns: 11rem 1fr;
  gap: 0.75rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border);
  font-size: 0.875rem;
}
.log-item--success .log-item__text {
  color: #86efac;
}
.log-item--error .log-item__text {
  color: #fca5a5;
}
.log-item__time {
  font-size: 0.75rem;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
}
</style>
