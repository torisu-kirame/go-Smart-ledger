<template>
  <div>
    <div style="display:flex;justify-content:space-between;align-items:center">
      <h2>账本管理</h2>
      <button class="btn-primary" @click="show = true">创建账本</button>
    </div>
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div class="panel">
      <div class="table-wrap">
        <table>
          <thead><tr><th>名称</th><th>类型</th><th>序号</th><th>锚定</th><th></th></tr></thead>
          <tbody>
            <tr v-for="l in list" :key="l.id">
              <td>{{ l.name }}</td>
              <td><span :class="['badge', l.type === 'multi' ? 'badge-multi' : 'badge-private']">{{ l.type === 'multi' ? '多人' : '私人' }}</span></td>
              <td class="mono">{{ l.latestSeq }}</td>
              <td><span :class="['badge', l.anchorStatus === 'synced' ? 'badge-ok' : 'badge-pending']">{{ l.anchorStatus }}</span></td>
              <td><router-link :to="`/ledgers/${l.id}`">详情</router-link></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div v-if="show" class="modal">
      <form class="modal-card" @submit.prevent="create">
        <h3>创建账本</h3>
        <div class="form-row">
          <label>类型</label>
          <select v-model="form.type">
            <option value="private">私人（1人）</option>
            <option value="multi">多人（≥2人）</option>
          </select>
        </div>
        <div class="form-row"><label>名称</label><input v-model="form.name" required /></div>
        <div class="form-row"><label>创建者 ID</label><input v-model="form.creatorId" required /></div>
        <div v-for="(m, i) in form.members" :key="i" class="form-row member">
          <input v-model="m.id" placeholder="成员 ID" required />
          <input v-model="m.address" placeholder="地址" />
          <button v-if="form.type === 'multi' && form.members.length > 2" type="button" class="btn-ghost" @click="form.members.splice(i,1)">删</button>
        </div>
        <button v-if="form.type === 'multi'" type="button" class="btn-ghost" @click="form.members.push({id:'',address:''})">+ 成员</button>
        <div style="display:flex;gap:0.5rem;justify-content:flex-end;margin-top:1rem">
          <button type="button" class="btn-ghost" @click="show=false">取消</button>
          <button class="btn-primary" :disabled="saving">创建</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, watch } from 'vue'
import { api, ApiError } from '../api/http'

const list = ref([])
const error = ref('')
const show = ref(false)
const saving = ref(false)
const form = reactive({
  type: 'private',
  name: '',
  creatorId: 'user1',
  members: [{ id: 'user1', address: '' }],
})

watch(() => form.type, (t) => {
  if (t === 'private') form.members = [form.members[0] || { id: '', address: '' }]
  else if (form.members.length < 2) form.members.push({ id: '', address: '' })
})

async function load() {
  list.value = await api.listLedgers()
}

onMounted(load)

async function create() {
  saving.value = true
  error.value = ''
  try {
    await api.createLedger({ ...form, members: form.members.filter((m) => m.id) })
    show.value = false
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '创建失败'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.modal { position: fixed; inset: 0; background: rgba(0,0,0,.65); display: flex; align-items: center; justify-content: center; z-index: 50; }
.modal-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: 12px; padding: 1.5rem; width: 480px; max-height: 90vh; overflow: auto; }
.member { display: grid; grid-template-columns: 1fr 1fr auto; gap: 0.5rem; }
</style>
