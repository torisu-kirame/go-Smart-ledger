<template>
  <svg
    class="app-icon"
    :class="sizeClass"
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    :stroke-width="strokeWidth"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <template v-for="(shape, idx) in shapes" :key="idx">
      <path v-if="shape.t === 'path'" :d="shape.d" />
      <line
        v-else-if="shape.t === 'line'"
        :x1="shape.x1"
        :y1="shape.y1"
        :x2="shape.x2"
        :y2="shape.y2"
      />
      <polyline v-else-if="shape.t === 'polyline'" :points="shape.points" />
      <circle
        v-else-if="shape.t === 'circle'"
        :cx="shape.cx"
        :cy="shape.cy"
        :r="shape.r"
      />
    </template>
  </svg>
</template>

<script setup>
import { computed } from 'vue'
import { ICON_REGISTRY } from '../icons/registry.js'

const props = defineProps({
  name: { type: String, required: true },
  size: { type: String, default: 'md' },
  strokeWidth: { type: [Number, String], default: 2 },
})

const shapes = computed(() => ICON_REGISTRY[props.name] || ICON_REGISTRY.home)
const sizeClass = computed(() => `app-icon--${props.size}`)
</script>
