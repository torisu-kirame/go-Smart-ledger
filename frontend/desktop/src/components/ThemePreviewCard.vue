<template>
  <button
    type="button"
    class="theme-card"
    :class="{ active }"
    :aria-pressed="active"
    @click="$emit('select')"
  >
    <div class="theme-card__preview" :style="previewStyle">
      <div class="theme-card__layout">
        <span class="theme-card__rail" aria-hidden="true" />
        <div class="theme-card__content">
          <span class="theme-card__bar theme-card__bar--lg theme-card__bar--accent" />
          <span class="theme-card__bar theme-card__bar--md theme-card__bar--text" />
          <span class="theme-card__bar theme-card__bar--sm theme-card__bar--muted" />
          <span class="theme-card__bar theme-card__bar--md theme-card__bar--accent-dim" />
        </div>
      </div>
    </div>
    <span class="theme-card__name">{{ label }}</span>
  </button>
</template>

<script setup>
import { computed } from 'vue'
import { getPreviewAccentColors, getThemePalette } from '../utils/theme'

const props = defineProps({
  themeId: { type: String, required: true },
  accentId: { type: String, default: 'blue' },
  label: { type: String, required: true },
  active: { type: Boolean, default: false },
})

defineEmits(['select'])

const previewStyle = computed(() => {
  const p = getThemePalette(props.themeId)
  const accent = getPreviewAccentColors(props.themeId, props.accentId)
  return {
    '--tp-bg': p.bg,
    '--tp-bg-elevated': p.bgElevated,
    '--tp-bg-card': p.bgCard,
    '--tp-border': p.border,
    '--tp-text': p.text,
    '--tp-text-muted': p.textMuted,
    '--tp-accent': accent.accent,
    '--tp-accent-dim': accent.accentDim,
  }
})
</script>

<style scoped>
.theme-card {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  width: 100%;
  max-width: 20rem;
  padding: 0.5rem 0.65rem 0.5rem 0.5rem;
  border: 2px solid transparent;
  border-radius: 12px;
  background: transparent;
  cursor: pointer;
  text-align: left;
  transition: border-color 0.18s ease, background 0.18s ease, box-shadow 0.18s ease;
}

.theme-card:hover {
  background: var(--hover);
}

.theme-card.active {
  border-color: var(--accent);
  background: var(--accent-soft);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent) 25%, transparent);
}

.theme-card__preview {
  flex-shrink: 0;
  width: 4.75rem;
  height: 3.35rem;
  padding: 0.28rem;
  border-radius: 8px;
  background: var(--tp-bg);
  border: 1px solid var(--tp-border);
}

.theme-card__layout {
  display: flex;
  height: 100%;
  min-height: 2.75rem;
  border-radius: 5px;
  overflow: hidden;
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--tp-border) 55%, transparent);
}

/* 侧栏区：bg-elevated */
.theme-card__rail {
  flex: 0 0 30%;
  background: var(--tp-bg-elevated);
  border-right: 1px solid var(--tp-border);
}

/* 主内容区：bg-card */
.theme-card__content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.26rem;
  padding: 0.32rem 0.38rem;
  background: var(--tp-bg-card);
}

.theme-card__bar {
  display: block;
  height: 0.3rem;
  border-radius: 2px;
}

.theme-card__bar--lg {
  width: 90%;
}

.theme-card__bar--md {
  width: 68%;
}

.theme-card__bar--sm {
  width: 46%;
}

.theme-card__bar--accent {
  background: var(--tp-accent);
}

.theme-card__bar--accent-dim {
  background: var(--tp-accent-dim);
}

.theme-card__bar--text {
  background: var(--tp-text);
  opacity: 0.88;
}

.theme-card__bar--muted {
  background: var(--tp-text-muted);
}

.theme-card__name {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text);
  line-height: 1.3;
}

.theme-card.active .theme-card__name {
  color: var(--accent);
}
</style>
