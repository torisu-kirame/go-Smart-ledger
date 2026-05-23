<template>
  <div class="page">
    <header class="page-header">
      <div>
        <h2>团队</h2>
      </div>
      <button class="btn-primary" @click="openCreate">创建团队</button>
    </header>

    <div v-if="error" class="alert alert-error">{{ error }}</div>

    <div class="panel">
      <div v-if="!teams.length" class="muted">暂无团队</div>
      <div v-for="t in teams" :key="t.id" class="team-card">
        <h3>{{ t.name }}</h3>
      </div>
    </div>

    <div v-if="show" class="modal">
      <form class="modal-card" @submit.prevent="create">
        <h3>创建团队</h3>
        <div class="form-row">
          <label>团队名称</label>
          <input v-model="form.name" required placeholder="例如：家庭账本协作组" />
        </div>
        <div class="form-row">
          <label>绑定多人账本</label>
          <AppSelect
            v-model="form.ledgerId"
            placeholder="请选择多人账本"
            :options="multiLedgerOptions"
          />
        </div>
        <div class="form-row">
          <label>邀请好友（至少 1 人）</label>
          <label v-for="f in friends" :key="f.id" class="check-row">
            <input type="checkbox" :value="f.id" v-model="form.memberUserIds" />
            <span>{{ f.nickname || f.username }}（ID: {{ f.id }}）</span>
          </label>
        </div>
        <div style="display:flex;gap:0.5rem;justify-content:flex-end;margin-top:1rem">
          <button type="button" class="btn-ghost" @click="show = false">取消</button>
          <button class="btn-primary" :disabled="saving">创建</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import AppSelect from '../components/AppSelect.vue'
import { api, ApiError } from '../api/http'

const teams = ref([])
const ledgers = ref([])
const friends = ref([])
const error = ref('')
const show = ref(false)
const saving = ref(false)
const form = reactive({ name: '', ledgerId: '', memberUserIds: [] })

const multiLedgers = computed(() => ledgers.value.filter((l) => l.type === 'multi'))
const multiLedgerOptions = computed(() =>
  multiLedgers.value.map((l) => ({
    value: l.id,
    label: `${l.name}（${l.ledgerAddress?.slice(0, 10) || l.id.slice(0, 10)}…）`,
  }))
)

async function load() {
  error.value = ''
  try {
    const [t, l, f] = await Promise.all([api.listTeams(), api.listLedgers(), api.listFriends()])
    teams.value = t.teams || []
    ledgers.value = l || []
    friends.value = f.friends || []
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '加载失败'
  }
}

function openCreate() {
  form.name = ''
  form.ledgerId = ''
  form.memberUserIds = []
  show.value = true
}

async function create() {
  if (form.memberUserIds.length < 1) {
    error.value = '请至少选择 1 位好友'
    return
  }
  saving.value = true
  error.value = ''
  try {
    await api.createTeam({
      name: form.name,
      ledgerId: form.ledgerId,
      ledgerType: 'multi',
      memberUserIds: form.memberUserIds,
    })
    show.value = false
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '创建失败'
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.muted { color: var(--text-muted); font-size: 0.875rem; }
.team-card {
  background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius);
  padding: 1rem; margin-bottom: 0.75rem;
}
.team-card h3 { margin: 0 0 0.5rem; font-size: 1rem; }
.modal { position: fixed; inset: 0; background: rgba(0,0,0,.65); display: flex; align-items: center; justify-content: center; z-index: 50; }
.modal-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: 12px; padding: 1.5rem; width: 480px; max-height: 90vh; overflow: auto; }
.check-row { display: flex; align-items: center; gap: 0.5rem; margin: 0.35rem 0; cursor: pointer; }
</style>
