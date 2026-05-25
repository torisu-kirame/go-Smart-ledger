<template>
  <nav class="settings-nav" :aria-label="ariaLabel">
    <button
      v-for="id in sections"
      :key="id"
      type="button"
      class="settings-nav-item"
      :class="{ active: active === id }"
      @click="$emit('select', id)"
    >
      {{ t(`settings.nav.${id}`) }}
    </button>
  </nav>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from '../../composables/useI18n'
import { SETTINGS_SECTIONS } from '../../utils/settingsSections'

defineProps({
  active: { type: String, required: true },
})

defineEmits(['select'])

const { t } = useI18n()
const sections = SETTINGS_SECTIONS
const ariaLabel = computed(() => t('settings.nav.label'))
</script>

<style scoped>
.settings-nav {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 9.5rem;
}

.settings-nav-item {
  text-align: left;
  padding: 0.55rem 0.85rem;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: var(--text-muted);
  font: inherit;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}

.settings-nav-item:hover {
  background: var(--hover);
  color: var(--text);
}

.settings-nav-item.active {
  background: var(--accent-soft);
  color: var(--accent);
}
</style>
