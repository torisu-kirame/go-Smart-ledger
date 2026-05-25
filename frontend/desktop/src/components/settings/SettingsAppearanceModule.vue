<template>
  <div class="settings-module">
    <div class="form-row">
      <label>{{ t('settings.theme.appearance') }}</label>
      <div class="theme-grid">
        <button
          v-for="p in THEME_PRESETS"
          :key="p.id"
          type="button"
          class="theme-chip"
          :class="{ active: themeId === p.id }"
          @click="setThemeId(p.id)"
        >
          <span class="theme-preview" :data-theme-id="p.id" />
          <span>{{ themePresetName(p) }}</span>
        </button>
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
.theme-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.theme-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.35rem 0.7rem;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-muted);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
}

.theme-chip:hover {
  background: var(--hover);
  color: var(--text);
}

.theme-chip.active {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-soft);
}

.theme-preview {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  border: 1px solid rgba(128, 128, 128, 0.35);
  flex-shrink: 0;
}

.theme-preview[data-theme-id='classic-light'] {
  background: #f3f6fc;
  border-color: #d8e1f1;
}

.theme-preview[data-theme-id='classic-dark'] {
  background: #0b1020;
  border-color: #2b3a58;
}

.theme-preview[data-theme-id='deep-dark'] {
  background: #050508;
  border-color: #1c1c22;
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
  border: 1px solid rgba(255, 255, 255, 0.2);
}
</style>
