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
    <div class="panel schema-panel">
      <h3>记账字段</h3>
      <p class="muted">
        模板：{{ ledger.entrySchema?.templateId || 'default' }} ·
        {{ schemaFieldsText }}
      </p>
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
      <EntryFormFields :schema="schema" :model="entryData" :members="memberOptions" />
      <button class="btn-primary" :disabled="busy" style="margin-top:0.75rem" @click="addEntry">提交到链</button>
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
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'
import EntryFormFields from '../components/EntryFormFields.vue'
import { emptyEntryData, resolveSchema } from '../utils/entrySchema'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const id = route.params.id
const ledger = ref(null)
const events = ref([])
const error = ref('')
const msg = ref('')
const busy = ref(false)
const entryData = reactive({})

const schema = computed(() => resolveSchema(ledger.value))
const schemaFieldsText = computed(() =>
  (schema.value.fields || []).map((f) => f.label).join('、')
)
const memberOptions = computed(() =>
  (ledger.value?.members || []).map((m) => ({ id: m.id, username: m.id }))
)

function initEntryDefaults() {
  const uid = auth.user?.id || ''
  const defaults = { bookkeeper: uid, date: new Date().toISOString().slice(0, 10) }
  const data = emptyEntryData(schema.value, defaults)
  Object.keys(entryData).forEach((k) => delete entryData[k])
  Object.assign(entryData, data)
}

async function load() {
  ledger.value = await api.getLedger(id)
  events.value = await api.listEvents(id)
  initEntryDefaults()
}

onMounted(load)

async function addEntry() {
  busy.value = true
  error.value = ''
  msg.value = ''
  try {
    const signerId = entryData.bookkeeper || auth.user?.id || ''
    await api.appendEntry(id, {
      signerId,
      schemaId: schema.value.templateId,
      data: { ...entryData },
    })
    msg.value = '记账成功'
    entryData.amount = ''
    entryData.note = ''
    entryData.payee = ''
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
.schema-panel .muted { font-size: 0.875rem; color: var(--text-muted); }
</style>
