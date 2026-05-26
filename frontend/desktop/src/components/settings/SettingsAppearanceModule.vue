<template>
  <div class="settings-module">
    <div class="form-row theme-picker-row">
      <label>{{ t('settings.theme.appearance') }}</label>
      <div class="theme-list">
        <ThemePreviewCard
          v-for="p in THEME_PRESETS"
          :key="p.id"
          :theme-id="p.id"
          :accent-id="accentId"
          :label="themePresetName(p)"
          :active="themeId === p.id"
          @select="setThemeId(p.id)"
        />
      </div>
    </div>

    <div class="form-row">
      <label>{{ t('settings.theme.accent') }}</label>
      <div class="accent-grid">
        <button
          v-for="p in ACCENT_PRESETS"
          :key="p.id"
          type="button"
          class="accent-chip"
          :class="{ active: accentId === p.id }"
          :title="accentName(p)"
          @click="setAccentColor(p.id)"
        >
          <span class="swatch" :style="{ background: p.swatch }" />
          <span>{{ accentName(p) }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import ThemePreviewCard from '../ThemePreviewCard.vue'
import { useI18n } from '../../composables/useI18n'
import { ACCENT_PRESETS, getAccent, getTheme, setAccent, setTheme, THEME_PRESETS } from '../../utils/theme'

const { locale, t } = useI18n()

const themeId = ref(getTheme())
const accentId = ref(getAccent())

const ACCENT_I18N = {
  blue: { zh: '天蓝', en: 'Blue' },
  teal: { zh: '青绿', en: 'Teal' },
  violet: { zh: '紫罗兰', en: 'Violet' },
  amber: { zh: '琥珀', en: 'Amber' },
  rose: { zh: '玫红', en: 'Rose' },
}

function accentName(p) {
  return ACCENT_I18N[p.id]?.[locale.value] ?? p.name
}

function themePresetName(p) {
  return t(`settings.theme.${p.i18nKey}`)
}

function setThemeId(id) {
  themeId.value = setTheme(id)
}

function setAccentColor(id) {
  accentId.value = setAccent(id)
}
</script>

<style scoped>
.theme-picker-row {
  max-width: none;
}

.theme-list {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.accent-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.accent-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-muted);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
}

.accent-chip.active {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-soft);
}

.swatch {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 1px solid rgba(128, 128, 128, 0.25);
}
</style>
