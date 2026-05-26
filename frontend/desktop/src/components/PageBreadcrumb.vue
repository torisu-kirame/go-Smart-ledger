<template>
  <nav class="page-breadcrumb" aria-label="页面路径">
    <ol class="page-breadcrumb__list">
      <li
        v-for="(crumb, index) in crumbs"
        :key="`${crumb.label}-${index}`"
        class="page-breadcrumb__item"
      >
        <span v-if="index > 0" class="page-breadcrumb__sep" aria-hidden="true">
          <svg
            class="page-breadcrumb__sep-icon"
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <polyline points="9 18 15 12 9 6" />
          </svg>
        </span>
        <router-link
          v-if="crumb.to && index < crumbs.length - 1"
          :to="crumb.to"
          class="page-breadcrumb__tag"
        >
          {{ crumb.label }}
        </router-link>
        <span
          v-else
          class="page-breadcrumb__tag page-breadcrumb__tag--current"
          aria-current="page"
        >
          {{ crumb.label }}
        </span>
      </li>
    </ol>
  </nav>
</template>

<script setup>
defineProps({
  crumbs: {
    type: Array,
    default: () => [],
  },
})
</script>

<style scoped>
.page-breadcrumb__list {
  display: inline-flex;
  align-items: stretch;
  flex-wrap: wrap;
  margin: 0;
  padding: 0;
  list-style: none;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  overflow: hidden;
  background: var(--bg);
  box-shadow: var(--shadow-sm);
}

.page-breadcrumb__item {
  display: inline-flex;
  align-items: stretch;
}

.page-breadcrumb__tag {
  display: inline-flex;
  align-items: center;
  padding: 0.45rem 0.85rem;
  font-size: 1.05rem;
  font-weight: 600;
  line-height: 1.3;
  color: var(--text);
  text-decoration: none;
  background: var(--bg-card);
  transition: background 0.15s ease, color 0.15s ease;
}

a.page-breadcrumb__tag:hover {
  background: var(--accent-soft);
  color: var(--accent);
}

.page-breadcrumb__tag--current {
  color: var(--accent);
  background: var(--accent-soft);
}

.page-breadcrumb__sep {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 0.2rem;
  color: var(--text-muted);
  background: var(--bg-elevated);
  border-left: 1px solid var(--border);
  border-right: 1px solid var(--border);
}

.page-breadcrumb__sep-icon {
  width: 0.85rem;
  height: 0.85rem;
  flex-shrink: 0;
}
</style>
