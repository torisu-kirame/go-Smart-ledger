<template>
  <div class="page" @click="closeMenu">
    <header class="page-header">
      <div>
        <h2>团队</h2>
        <p class="page-subtitle">
          团队是协作入口；点击团队进入聊天。关联账本仅为快捷入口，查看账本须各自接受邀请。
        </p>
      </div>
      <div class="header-actions">
        <button
          v-if="totalUnread > 0"
          class="btn-ghost"
          type="button"
          :disabled="markingAll"
          @click="markAllRead"
        >
          全部已读
        </button>
        <button class="btn-primary team-create" type="button" @click="openCreate">
          <AppIcon name="plus" size="sm" />
          <span>创建团队</span>
        </button>
      </div>
    </header>

    <div v-if="error" class="alert alert-error">{{ error }}</div>

    <div class="panel team-list-panel">
      <div v-if="!teams.length" class="muted empty">暂无团队，点击右上角创建</div>
      <router-link
        v-for="t in teams"
        :key="t.id"
        :to="`/teams/${t.id}`"
        class="team-row"
        :class="{ unread: t.unreadCount > 0 }"
        @contextmenu.prevent="openMenu($event, t)"
      >
        <TeamAvatar :team-id="t.id" :size="48" />
        <div class="team-body">
          <div class="team-title-row">
            <span class="team-name">{{ t.name }}</span>
            <span v-if="t.lastMessage?.createdAt" class="team-time">{{ formatListTime(t.lastMessage.createdAt) }}</span>
          </div>
          <p class="team-preview">
            <template v-if="t.lastMessage">
              <span v-if="t.lastMessage.senderNickname" class="sender">{{ t.lastMessage.senderNickname }}: </span>
              {{ t.lastMessage.preview }}
            </template>
            <span v-else class="muted-inline">暂无消息</span>
          </p>
        </div>
        <span v-if="t.unreadCount > 0" class="unread-badge" :title="`${t.unreadCount} 条未读`">
          {{ t.unreadCount > 99 ? '99+' : t.unreadCount }}
        </span>
      </router-link>
    </div>

    <div
      v-if="menu.show"
      class="context-menu"
      :style="{ left: menu.x + 'px', top: menu.y + 'px' }"
      @click.stop
    >
      <button type="button" @click="markOneRead(menu.team)">标记已读</button>
      <button type="button" @click="goTeam(menu.team)">进入团队</button>
    </div>

    <div v-if="show" class="modal">
      <form class="modal-card wide" @submit.prevent="create">
        <h3>创建团队</h3>
        <div class="form-row">
          <label>团队名称</label>
          <input v-model="form.name" required placeholder="例如：家庭账本协作组" />
        </div>
        <div class="form-row">
          <label>关联多人账本（至少 1 个，可多选）</label>
          <div v-if="!multiLedgers.length" class="muted">暂无多人账本，请先在账本管理创建</div>
          <div v-else class="ledger-picks">
            <label v-for="l in multiLedgers" :key="l.id" class="check-row">
              <input type="checkbox" :value="l.id" v-model="form.ledgerIds" />
              <span>{{ l.name }}（{{ shortId(l.id) }}）</span>
            </label>
          </div>
        </div>
        <div class="form-row">
          <label>邀请好友（至少 1 人）</label>
          <MemberAddPanel
            v-model="form.memberUserIds"
            :multiple="true"
            :exclude-ids="auth.user?.id ? [auth.user.id] : []"
          />
        </div>
        <div class="modal-actions">
          <button type="button" class="btn-ghost" @click="show = false">取消</button>
          <button class="btn-primary" :disabled="saving">创建</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'
import AppIcon from '../components/AppIcon.vue'
import MemberAddPanel from '../components/MemberAddPanel.vue'
import TeamAvatar from '../components/TeamAvatar.vue'

const router = useRouter()
const auth = useAuthStore()
const teams = ref([])
const totalUnread = ref(0)
const ledgers = ref([])
const error = ref('')
const show = ref(false)
const saving = ref(false)
const markingAll = ref(false)
const form = reactive({ name: '', ledgerIds: [], memberUserIds: [] })
const menu = reactive({ show: false, x: 0, y: 0, team: null })
let pollTimer = null

const multiLedgers = computed(() => ledgers.value.filter((l) => l.type === 'multi'))

function shortId(id) {
  return id?.length > 12 ? `${id.slice(0, 8)}…` : id
}

function formatListTime(t) {
  if (!t) return ''
  const d = new Date(t)
  const now = new Date()
  const sameDay =
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate()
  if (sameDay) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return d.toLocaleDateString([], { month: 'numeric', day: 'numeric' })
}

async function load() {
  error.value = ''
  try {
    const [t, l] = await Promise.all([api.listTeams(), api.listLedgers()])
    teams.value = t.teams || []
    totalUnread.value = t.totalUnread ?? teams.value.reduce((s, x) => s + (x.unreadCount || 0), 0)
    ledgers.value = Array.isArray(l) ? l : []
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '加载失败'
  }
}

function openCreate() {
  form.name = ''
  form.ledgerIds = []
  form.memberUserIds = []
  show.value = true
}

async function create() {
  if (form.ledgerIds.length < 1) {
    error.value = '请至少选择 1 个多人账本'
    return
  }
  const members = (form.memberUserIds || []).map((id) => String(id).trim()).filter(Boolean)
  if (members.length < 1) {
    error.value = '请至少邀请 1 位好友'
    return
  }
  saving.value = true
  error.value = ''
  try {
    await api.createTeam({
      name: form.name.trim(),
      ledgerIds: form.ledgerIds,
      ledgerType: 'multi',
      memberUserIds: members,
    })
    show.value = false
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '创建失败'
  } finally {
    saving.value = false
  }
}

function openMenu(ev, team) {
  menu.show = true
  menu.x = ev.clientX
  menu.y = ev.clientY
  menu.team = team
}

function closeMenu() {
  menu.show = false
  menu.team = null
}

async function markOneRead(team) {
  if (!team?.id) return
  closeMenu()
  try {
    await api.markTeamRead(team.id)
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '操作失败'
  }
}

async function markAllRead() {
  markingAll.value = true
  error.value = ''
  try {
    await api.markAllTeamsRead()
    await load()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '操作失败'
  } finally {
    markingAll.value = false
  }
}

function goTeam(team) {
  closeMenu()
  if (team?.id) router.push(`/teams/${team.id}`)
}

onMounted(() => {
  load()
  pollTimer = setInterval(() => load().catch(() => {}), 12000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.section-hint {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: 0.35rem 0 0;
  max-width: 42rem;
}
.header-actions { display: flex; gap: 0.5rem; flex-wrap: wrap; align-items: center; }
.team-create {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}
.page-header { display: flex; justify-content: space-between; align-items: flex-start; gap: 1rem; flex-wrap: wrap; }
.team-list-panel { padding: 0.35rem 0; }
.team-row {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 0.75rem 1rem;
  text-decoration: none;
  color: inherit;
  border-bottom: 1px solid var(--border);
  position: relative;
  transition: background 0.12s ease;
}
.team-row:last-child { border-bottom: none; }
.team-row:hover { background: var(--hover); }
.team-row.unread { background: color-mix(in srgb, var(--accent) 6%, transparent); }
.team-body { flex: 1; min-width: 0; }
.team-title-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 0.5rem;
  margin-bottom: 0.2rem;
}
.team-name {
  font-weight: 600;
  font-size: 0.95rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.team-time {
  font-size: 0.75rem;
  color: var(--text-muted);
  flex-shrink: 0;
}
.team-preview {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.team-row.unread .team-preview { color: var(--text); }
.team-preview .sender { color: var(--text-muted); }
.muted-inline { color: var(--text-muted); }
.unread-badge {
  flex-shrink: 0;
  min-width: 1.25rem;
  height: 1.25rem;
  padding: 0 0.35rem;
  border-radius: 999px;
  background: var(--danger);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
  box-shadow: 0 0 0 2px var(--bg-card);
}
.empty { padding: 2rem; text-align: center; }
.context-menu {
  position: fixed;
  z-index: 60;
  min-width: 8.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  box-shadow: var(--shadow-lg);
  padding: 0.35rem 0;
}
.context-menu button {
  display: block;
  width: 100%;
  text-align: left;
  padding: 0.5rem 0.85rem;
  border: none;
  background: none;
  color: var(--text);
  font-size: 0.875rem;
  cursor: pointer;
}
.context-menu button:hover { background: var(--hover); }
.modal { position: fixed; inset: 0; background: rgba(0,0,0,.65); display: flex; align-items: center; justify-content: center; z-index: 50; }
.modal-card.wide { width: min(100%, 520px); background: var(--bg-card); border: 1px solid var(--border); border-radius: 12px; padding: 1.5rem; max-height: 90vh; overflow: auto; }
.ledger-picks { max-height: 160px; overflow-y: auto; }
.check-row { display: flex; align-items: center; gap: 0.5rem; margin: 0.35rem 0; cursor: pointer; }
.modal-actions { display: flex; gap: 0.5rem; justify-content: flex-end; margin-top: 1rem; }
.form-row { margin-bottom: 0.9rem; }
.form-row label { display: block; font-size: 0.8125rem; color: var(--text-muted); margin-bottom: 0.35rem; }
</style>
