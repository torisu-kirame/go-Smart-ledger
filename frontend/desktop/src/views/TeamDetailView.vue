<template>
  <div v-if="team" class="page team-detail">
    <PageHeader
      :crumbs="crumbs"
      subtitle="团队 Chat 与关联账本；账本数据仍须各自接受邀请后查看。"
    />

    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="success" class="alert alert-success">{{ success }}</div>

    <div class="layout">
      <section class="panel chat-panel">
        <h3>团队聊天</h3>
        <div ref="chatBox" class="chat-box">
          <div v-if="!messages.length" class="muted">暂无消息，发一条打个招呼吧</div>
          <div
            v-for="m in messages"
            :key="m.id"
            class="msg"
            :class="{ mine: m.senderId === auth.user?.id }"
          >
            <div class="msg-meta">
              <span class="sender">{{ m.senderNickname || m.senderUsername || m.senderId }}</span>
              <span class="time">{{ formatTime(m.createdAt) }}</span>
            </div>
            <div v-if="m.type === 'text'" class="msg-body">{{ m.body }}</div>
            <div v-else class="msg-file">
              <a v-if="fileLinks[m.id]" :href="fileLinks[m.id]" target="_blank" rel="noopener noreferrer" class="file-link">
                <AppIcon name="paperclip" size="sm" />
                <span>{{ m.fileName }} ({{ formatSize(m.fileSize) }})</span>
              </a>
              <span v-else class="muted">加载附件…</span>
            </div>
          </div>
        </div>
        <FileUploadZone
          compact
          block
          class="chat-upload"
          :disabled="sending"
          title="点击或拖拽发送文件"
          @file="onFilePick"
        />
        <form class="chat-compose" @submit.prevent="sendText">
          <input v-model="draft" placeholder="输入消息…" maxlength="4000" />
          <button class="btn-primary chat-send" type="submit" :disabled="sending || !draft.trim()">
            <AppIcon name="send" size="sm" />
            <span>发送</span>
          </button>
        </form>
      </section>

      <section class="panel ledgers-panel">
        <h3>关联账本</h3>
        <ul v-if="ledgerRows.length" class="ledger-list">
          <li v-for="row in ledgerRows" :key="row.id">
            <div>
              <strong>{{ row.name }}</strong>
              <span class="mono">{{ shortId(row.id) }}</span>
            </div>
            <div class="ledger-actions">
              <router-link :to="`/ledgers/${row.id}`">打开</router-link>
              <router-link :to="{ path: '/ledgers', query: { invite: row.id } }">邀请成员</router-link>
              <DeleteButton
                v-if="isCreator"
                icon-only
                sm
                title="移除此关联"
                @click="unlinkLedger(row.id)"
              />
            </div>
          </li>
        </ul>
        <p v-else class="muted">尚未关联账本</p>

        <div v-if="isCreator" class="add-ledger">
          <label>添加多人账本</label>
          <AppSelect v-model="addLedgerId" :options="availableLedgerOptions" placeholder="选择账本" />
          <button class="btn-ghost" type="button" :disabled="!addLedgerId || busy" @click="linkLedger">关联</button>
        </div>

        <h4 class="members-title">成员 ({{ team.members?.length || 0 }})</h4>
        <ul class="member-list">
          <li v-for="m in team.members" :key="m.userId">
            <img class="mini-avatar" :src="api.userAvatarUrl(m.userId)" alt="" />
            <span>{{ m.nickname || m.username }}</span>
            <span v-if="m.userId === team.creatorId" class="tag">创建者</span>
          </li>
        </ul>
      </section>
    </div>
  </div>
  <div v-else-if="loading" class="page muted">加载中…</div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import AppIcon from '../components/AppIcon.vue'
import AppSelect from '../components/AppSelect.vue'
import DeleteButton from '../components/DeleteButton.vue'
import FileUploadZone from '../components/FileUploadZone.vue'
import PageHeader from '../components/PageHeader.vue'
import { usePageCrumbs } from '../composables/usePageCrumbs'
import { api, ApiError, fetchTeamChatFileBlob } from '../api/http'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const auth = useAuthStore()
const teamId = computed(() => route.params.teamId)

const team = ref(null)
const { crumbs } = usePageCrumbs(computed(() => team.value?.name || '…'))
const ledgers = ref([])
const messages = ref([])
const fileLinks = ref({})
const draft = ref('')
const addLedgerId = ref('')
const error = ref('')
const success = ref('')
const loading = ref(true)
const sending = ref(false)
const busy = ref(false)
const chatBox = ref(null)
let pollTimer = null
const blobUrls = []

const isCreator = computed(() => team.value?.creatorId === auth.user?.id)

const ledgerRows = computed(() => {
  const ids = team.value?.ledgerIds || (team.value?.ledgerId ? [team.value.ledgerId] : [])
  return ids.map((id) => {
    const l = ledgers.value.find((x) => x.id === id)
    return { id, name: l?.name || `账本 ${id}` }
  })
})

const linkedSet = computed(() => new Set(team.value?.ledgerIds || []))

const availableLedgerOptions = computed(() =>
  ledgers.value
    .filter((l) => l.type === 'multi' && !linkedSet.value.has(l.id))
    .map((l) => ({ value: l.id, label: `${l.name}（${shortId(l.id)}）` }))
)

function shortId(id) {
  return id?.length > 12 ? `${id.slice(0, 8)}…` : id
}

function formatTime(t) {
  if (!t) return ''
  return new Date(t).toLocaleString()
}

function formatSize(n) {
  if (!n) return '0 B'
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / (1024 * 1024)).toFixed(1)} MB`
}

async function loadTeam() {
  team.value = await api.getTeam(teamId.value)
}

async function loadLedgers() {
  const data = await api.listLedgers()
  ledgers.value = Array.isArray(data) ? data : []
}

function lastMessageId() {
  if (!messages.value.length) return 0
  const last = messages.value[messages.value.length - 1]
  return Number(last.id) || 0
}

async function loadMessages(initial = false) {
  const since = initial ? 0 : lastMessageId()
  const res = await api.listTeamMessages(teamId.value, since, 80)
  const incoming = res.messages || []
  if (initial) {
    messages.value = incoming
  } else if (incoming.length) {
    const seen = new Set(messages.value.map((m) => m.id))
    for (const m of incoming) {
      if (!seen.has(m.id)) messages.value.push(m)
    }
  }
  await hydrateFiles()
  await nextTick()
  if (chatBox.value) chatBox.value.scrollTop = chatBox.value.scrollHeight
}

async function hydrateFiles() {
  for (const m of messages.value) {
    if (m.type !== 'file' || fileLinks.value[m.id]) continue
    try {
      const url = await fetchTeamChatFileBlob(teamId.value, m.id)
      blobUrls.push(url)
      fileLinks.value = { ...fileLinks.value, [m.id]: url }
    } catch {
      /* 附件加载失败时保持占位 */
    }
  }
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    await Promise.all([loadTeam(), loadLedgers()])
    await loadMessages(true)
    try {
      await api.markTeamRead(teamId.value)
    } catch {
      /* 已读标记失败不阻断进入聊天 */
    }
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

async function sendText() {
  const text = draft.value.trim()
  if (!text) return
  sending.value = true
  error.value = ''
  try {
    const msg = await api.sendTeamMessage(teamId.value, text)
    messages.value.push(msg)
    draft.value = ''
    await nextTick()
    if (chatBox.value) chatBox.value.scrollTop = chatBox.value.scrollHeight
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '发送失败'
  } finally {
    sending.value = false
  }
}

async function onFilePick(file) {
  if (!file) return
  sending.value = true
  error.value = ''
  try {
    const msg = await api.uploadTeamChatFile(teamId.value, file)
    messages.value.push(msg)
    await hydrateFiles()
    await nextTick()
    if (chatBox.value) chatBox.value.scrollTop = chatBox.value.scrollHeight
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '上传失败'
  } finally {
    sending.value = false
  }
}

async function linkLedger() {
  if (!addLedgerId.value) return
  busy.value = true
  error.value = ''
  try {
    team.value = await api.addTeamLedger(teamId.value, addLedgerId.value)
    addLedgerId.value = ''
    success.value = '已关联账本'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '关联失败'
  } finally {
    busy.value = false
  }
}

async function unlinkLedger(ledgerId) {
  if (!confirm('确定移除此账本关联？')) return
  busy.value = true
  try {
    team.value = await api.removeTeamLedger(teamId.value, ledgerId)
    success.value = '已移除关联'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '操作失败'
  } finally {
    busy.value = false
  }
}

function startPoll() {
  pollTimer = setInterval(() => {
    loadMessages(false)
      .then(() => api.markTeamRead(teamId.value))
      .catch(() => {})
  }, 8000)
}

onMounted(() => {
  reload().then(startPoll)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  blobUrls.forEach((u) => URL.revokeObjectURL(u))
})

watch(teamId, (id, prev) => {
  if (id && id !== prev) reload()
})
</script>

<style scoped>
.section-hint { font-size: 0.8125rem; color: var(--text-muted); margin: 0.35rem 0 0; }
.file-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: var(--accent);
  font-size: 0.875rem;
  text-decoration: none;
}
.file-link:hover { text-decoration: underline; }
.chat-send {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}
.chat-upload {
  margin-top: 0.65rem;
}
.layout {
  display: grid;
  grid-template-columns: 1fr min(320px, 36%);
  gap: 1rem;
  align-items: start;
}
@media (max-width: 900px) {
  .layout { grid-template-columns: 1fr; }
}
.chat-panel { display: flex; flex-direction: column; min-height: 420px; }
.chat-box {
  flex: 1;
  min-height: 280px;
  max-height: 52vh;
  overflow-y: auto;
  padding: 0.75rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  margin-bottom: 0.75rem;
}
.msg { margin-bottom: 0.85rem; max-width: 85%; }
.msg.mine { margin-left: auto; text-align: right; }
.msg-meta { font-size: 0.75rem; color: var(--text-muted); margin-bottom: 0.2rem; }
.msg-meta .sender { font-weight: 600; color: var(--text); margin-right: 0.5rem; }
.msg-body {
  display: inline-block;
  padding: 0.5rem 0.75rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 10px;
  text-align: left;
  word-break: break-word;
}
.msg.mine .msg-body { background: var(--accent-soft); border-color: var(--accent); }
.msg-file a { color: var(--accent); font-size: 0.875rem; }
.chat-compose {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}
.chat-compose input { flex: 1; min-width: 12rem; }
.ledger-list { list-style: none; padding: 0; margin: 0 0 1rem; }
.ledger-list li {
  padding: 0.65rem 0;
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  flex-wrap: wrap;
  align-items: center;
}
.ledger-list .mono { display: block; font-size: 0.75rem; color: var(--text-muted); }
.ledger-actions { display: flex; gap: 0.65rem; align-items: center; flex-wrap: wrap; font-size: 0.875rem; }
.add-ledger { display: grid; gap: 0.5rem; margin-top: 0.75rem; padding-top: 0.75rem; border-top: 1px dashed var(--border); }
.add-ledger label { font-size: 0.8125rem; color: var(--text-muted); }
.members-title { margin: 1rem 0 0.5rem; font-size: 0.875rem; }
.member-list { list-style: none; padding: 0; margin: 0; }
.member-list li { display: flex; align-items: center; gap: 0.5rem; padding: 0.35rem 0; font-size: 0.875rem; }
.mini-avatar { width: 28px; height: 28px; border-radius: 50%; border: 1px solid var(--border); }
.tag { font-size: 0.7rem; color: var(--accent); margin-left: auto; }
</style>
