<template>
  <nav class="settings-nav" :aria-label="ariaLabel">
    <button
      v-for="item in items"
      :key="item.id"
      type="button"
      class="nav-item"
      :class="{ active: active === item.id }"
      @click="$emit('select', item.id)"
    >
      <span class="nav-label">{{ item.label }}</span>
      <span v-if="item.desc" class="nav-desc">{{ item.desc }}</span>
    </button>
  </nav>
</template>

<script setup>
defineProps({
  items: { type: Array, required: true },
  active: { type: String, required: true },
  ariaLabel: { type: String, default: '' },
})

defineEmits(['select'])
</script>

<style scoped>
.settings-nav {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.15rem;
  width: 100%;
  padding: 0.55rem 0.75rem;
  border: 1px solid transparent;
  border-radius: 10px;
  background: transparent;
  color: var(--text-muted);
  font: inherit;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease, border-color 0.15s ease;
}

.nav-item:hover {
  background: var(--hover);
  color: var(--text);
}

.nav-item.active {
  border-color: var(--accent);
  background: var(--accent-soft);
  color: var(--accent);
}

.nav-label {
  font-size: 0.9rem;
  font-weight: 700;
  line-height: 1.2;
}

.nav-desc {
  font-size: 0.72rem;
  font-weight: 500;
  opacity: 0.85;
  line-height: 1.3;
}

.nav-item.active .nav-desc {
  color: inherit;
}

@media (max-width: 720px) {
  .settings-nav {
    flex-direction: row;
    flex-wrap: wrap;
    gap: 0.4rem;
  }

  .nav-item {
    width: auto;
    flex: 1 1 auto;
    min-width: 7rem;
  }

  .nav-desc {
    display: none;
  }
}
</style>
