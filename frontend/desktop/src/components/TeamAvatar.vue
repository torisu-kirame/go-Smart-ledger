<template>
  <canvas
    ref="canvasRef"
    class="team-avatar"
    :width="size"
    :height="size"
    :style="{ width: size + 'px', height: size + 'px' }"
    :title="teamId"
  />
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { drawTeamIdenticon } from '../utils/teamIdenticon'

const props = defineProps({
  teamId: { type: String, required: true },
  size: { type: Number, default: 48 },
})

const canvasRef = ref(null)

async function render() {
  if (!canvasRef.value || !props.teamId) return
  await drawTeamIdenticon(canvasRef.value, props.teamId)
}

onMounted(render)
watch(() => props.teamId, render)
</script>

<style scoped>
.team-avatar {
  border-radius: 10px;
  flex-shrink: 0;
  display: block;
  border: 1px solid var(--border);
}
</style>
