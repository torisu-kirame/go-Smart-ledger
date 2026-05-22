<template>
  <div v-if="ledger">
    <h2><router-link to="/ledgers" style="color:var(--text-muted)">←</router-link> {{ ledger.name }}</h2>
    <div v-if="msg" class="alert alert-success">{{ msg }}</div>
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div class="grid-3">
      <div class="card"><h4>类型</h4><div class="val">{{ ledger.type === 'multi' ? '多人' : '私人' }}</div></div>
      <div class="card"><h4>序号</h4><div class="val mono">{{ ledger.latestSeq }}</div></div>
      <div class="card"><h4>锚定</h4><div class="val"><span :class="['badge', ledger.anchorStatus==='synced'?'badge-ok':'badge-pending']">{{ ledger.anchorStatus }}</span></div></div>
    </div>
    <div class="panel">
      <h3>链上操作</h3>
      <button class="btn-ghost" :disabled="busy" @click="doVerify">校验完整性</button>
      <button class="btn-primary" :disabled="busy" style="margin-left:0.5rem" @click="doAnchor">封账并锚定</button>
      <button v-if="ledger.anchorStatus==='synced'" class="btn-ghost" style="margin-left:0.5rem" @click="goBackup">加密备份</button>
      <router-link v-if="ledger.anchorStatus==='synced'" to="/import" style="margin-left:0.5rem">继续导入</router-link>
    </div>
    <div class="panel">
      <h3>记一笔</h3>
      <form @submit.prevent="addEntry" class="form-grid">
        <div class="form-row"><label>记账人</label>
          <select v-model="entry.signerId"><option v-for="m in ledger.members" :key="m.id" :value="m.id">{{ m.id }}</option></select>
        </div>
        <div class="form-row"><label>日期</label><input v-model="entry.date" type="date" required /></div>
        <div class="form-row"><label>类型</label>
          <select v-model="entry.type"><option value="expense">支出</option><option value="income">收入</option></select>
        </div>
        <div class="form-row"><label>金额</label><input v-model="entry.amount" required /></div>
        <div class="form-row"><label>分类</label><input v-model="entry.category" /></div>
        <div class="form-row"><label>备注</label><input v-model="entry.note" /></div>
      </form>
      <button class="btn-primary" :disabled="busy" @click="addEntry">提交到链</button>
    </div>
    <div class="panel">
      <h3>事件流水 ({{ events.length }})</h3>
      <div class="table-wrap">
        <table>
          <thead><tr><th>Seq</th><th>类型</th><th>签名者</th><th>哈希</th></tr></thead>
          <tbody>
            <tr v-for="e in [...events].reverse()" :key="e.seq">
              <td class="mono">{{ e.seq }}</td><td>{{ e.type }}</td><td>{{ e.signerId }}</td>
              <td class="mono">{{ e.hash?.slice(0,16) }}…</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, ApiError } from '../api/http'

const route = useRoute()
const router = useRouter()
const id = route.params.id
const ledger = ref(null)
const events = ref([])
const error = ref('')
const msg = ref('')
const busy = ref(false)
const entry = reactive({ signerId: '', date: new Date().toISOString().slice(0, 10), type: 'expense', amount: '', category: '', note: '' })

async function load() {
  ledger.value = await api.getLedger(id)
  events.value = await api.listEvents(id)
  if (!entry.signerId && ledger.value.members[0]) entry.signerId = ledger.value.members[0].id
}

onMounted(load)

async function addEntry() {
  busy.value = true
  error.value = ''
  msg.value = ''
  try {
    await api.appendEntry(id, { ...entry, signerId: entry.signerId })
    msg.value = '记账成功'
    entry.amount = ''
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '失败'
  } finally {
    busy.value = false
  }
}

async function doAnchor() {
  busy.value = true
  try {
    const r = await api.anchor(id)
    msg.value = `封账成功 · ${r.status}`
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '锚定失败'
  } finally {
    busy.value = false
  }
}

async function doVerify() {
  const r = await api.verify(id)
  msg.value = r.valid ? 'Merkle 校验通过' : '校验未通过'
}

function goBackup() {
  router.push({ path: '/backup', query: { ledgerId: id } })
}
</script>

<style scoped>
.form-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 0.75rem; }
</style>
