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
      <AppIcon :name="sectionIcon(id)" size="sm" class="settings-nav-item__icon" />
      <span>{{ t(`settings.nav.${id}`) }}</span>
    </button>
  </nav>
</template>

<script setup>
import { computed } from 'vue'
import AppIcon from '../AppIcon.vue'
import { useI18n } from '../../composables/useI18n'
import { SETTINGS_ICON_BY_SECTION } from '../../icons/registry.js'
import { SETTINGS_SECTIONS } from '../../utils/settingsSections'

defineProps({
  active: { type: String, required: true },
})

defineEmits(['select'])

const { t } = useI18n()
const sections = SETTINGS_SECTIONS
const ariaLabel = computed(() => t('settings.nav.label'))

function sectionIcon(id) {
  return SETTINGS_ICON_BY_SECTION[id] || 'settings'
}
</script>

<style scoped>
.settings-nav {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 10.5rem;
}

.settings-nav-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
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

.settings-nav-item__icon {
  flex-shrink: 0;
  opacity: 0.85;
}

.settings-nav-item:hover {
  background: var(--hover);
  color: var(--text);
}

.settings-nav-item.active {
  background: var(--accent-soft);
  color: var(--accent);
}

.settings-nav-item.active .settings-nav-item__icon {
  color: var(--accent);
}
</style>
