<template>
  <div class="page">
    <h2>账本邀请</h2>
    <p class="page-desc">接受邀请后将加入对应多人账本（F18）</p>
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="msg" class="alert alert-success">{{ msg }}</div>
    <div class="panel">
      <div v-if="!invites.length" class="muted">暂无待处理邀请</div>
      <div v-for="inv in invites" :key="inv.ledgerId" class="invite-row">
        <div>
          <strong>账本 ID</strong>
          <span class="mono">{{ inv.ledgerId }}</span>
          <p class="muted">邀请人 {{ inv.inviterId }}</p>
        </div>
        <button class="btn-primary" :disabled="busy" @click="accept(inv.ledgerId)">接受加入</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '../api/http'

const router = useRouter()
const invites = ref([])
const error = ref('')
const msg = ref('')
const busy = ref(false)

async function load() {
  error.value = ''
  try {
    const res = await api.listMyInvites()
    invites.value = res.invites || []
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '加载失败'
  }
}

async function accept(ledgerId) {
  busy.value = true
  error.value = ''
  try {
    await api.acceptInvite(ledgerId)
    msg.value = '已加入账本'
    await load()
    router.push(`/ledgers/${ledgerId}`)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '接受失败'
  } finally {
    busy.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.invite-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--border);
}
.mono { font-family: ui-monospace, monospace; color: var(--accent); margin-left: 0.35rem; }
.muted { color: var(--text-muted); font-size: 0.85rem; margin: 0.25rem 0 0; }
</style>
