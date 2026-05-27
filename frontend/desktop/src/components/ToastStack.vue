<template>
  <div class="toast-stack" aria-live="polite">
    <TransitionGroup name="toast">
      <div
        v-for="item in store.toasts"
        :key="item.id"
        class="toast-item"
        :class="`toast-item--${item.type}`"
        role="status"
      >
        <button
          type="button"
          class="toast-item__close"
          aria-label="关闭"
          @click="store.dismissToast(item.id)"
        >
          ×
        </button>
        <span class="toast-item__text">{{ item.text }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup>
import { useNotifyStore } from '../stores/notify'

const store = useNotifyStore()
</script>

<style scoped>
.toast-stack {
  position: fixed;
  top: 0.75rem;
  left: 0;
  right: 0;
  z-index: 10000;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.45rem;
  pointer-events: none;
  padding: 0 1rem;
  box-sizing: border-box;
}

.toast-item {
  position: relative;
  pointer-events: auto;
  max-width: min(92vw, 32rem);
  width: max-content;
  padding: 0.55rem 2rem 0.55rem 0.85rem;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  box-shadow: var(--shadow-lg);
  font-size: 0.875rem;
  line-height: 1.4;
  font-weight: 500;
}

.toast-item--success {
  background: color-mix(in srgb, #22c55e 18%, var(--bg-card));
  border-color: color-mix(in srgb, #22c55e 45%, var(--border));
  color: #bbf7d0;
}

.toast-item--error {
  background: color-mix(in srgb, var(--danger) 20%, var(--bg-card));
  border-color: color-mix(in srgb, var(--danger) 50%, var(--border));
  color: #fecaca;
}

.toast-item--info {
  background: var(--bg-card);
  border-color: var(--border);
  color: var(--text);
}

.toast-item__text {
  display: block;
  word-break: break-word;
}

.toast-item__close {
  position: absolute;
  top: 0.3rem;
  right: 0.3rem;
  width: 1.35rem;
  height: 1.35rem;
  padding: 0;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: inherit;
  opacity: 0.75;
  font-size: 1.1rem;
  line-height: 1;
  cursor: pointer;
}

.toast-item__close:hover {
  opacity: 1;
  background: rgba(255, 255, 255, 0.08);
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.22s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateY(-12px);
}

.toast-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.toast-move {
  transition: transform 0.22s ease;
}
</style>
