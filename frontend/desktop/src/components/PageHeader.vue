<template>
  <header class="page-header">
    <div class="page-header__main">
      <PageBreadcrumb :crumbs="crumbs" />
      <p v-if="subtitle" class="page-subtitle">{{ subtitle }}</p>
    </div>
    <div v-if="$slots.actions" class="page-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<script setup>
import PageBreadcrumb from './PageBreadcrumb.vue'

defineProps({
  crumbs: {
    type: Array,
    default: () => [],
  },
  subtitle: {
    type: String,
    default: '',
  },
})
</script>
