<template>
  <router-link :to="to" class="page-back-link">
    <AppIcon name="arrow-left" size="sm" />
    <span v-if="label" class="page-back-link__label">{{ label }}</span>
  </router-link>
</template>

<script setup>
import AppIcon from './AppIcon.vue'

defineProps({
  to: { type: String, required: true },
  label: { type: String, default: '' },
})
</script>
