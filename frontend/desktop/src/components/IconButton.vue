<template>
  <component
    :is="tag"
    :type="tag === 'button' ? 'button' : undefined"
    :href="tag === 'a' ? href : undefined"
    class="icon-btn"
    :class="[`icon-btn--${variant}`, { 'icon-btn--active': active }]"
    :disabled="tag === 'button' ? disabled : undefined"
    :title="title"
    :aria-label="ariaLabel || title"
  >
    <AppIcon :name="icon" :size="iconSize" />
    <span v-if="label" class="icon-btn__label">{{ label }}</span>
  </component>
</template>

<script setup>
import { computed } from 'vue'
import AppIcon from './AppIcon.vue'

const props = defineProps({
  icon: { type: String, required: true },
  label: { type: String, default: '' },
  title: { type: String, default: '' },
  ariaLabel: { type: String, default: '' },
  variant: { type: String, default: 'ghost' },
  size: { type: String, default: 'md' },
  active: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  tag: { type: String, default: 'button' },
  href: { type: String, default: '' },
})

const iconSize = computed(() => (props.size === 'sm' ? 'sm' : 'md'))
</script>
