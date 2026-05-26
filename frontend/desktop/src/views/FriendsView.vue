<template>
  <div class="page">
    <PageHeader
      :crumbs="crumbs"
      subtitle="搜索用户、处理好友申请与协作邀请。"
    />

    <section v-if="incoming.length" class="card highlight">
      <h3>收到的好友申请</h3>
      <table class="table">
        <thead>
          <tr><th>用户</th><th>申请时间</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="req in incoming" :key="req.fromUserId">
            <td>
              <div class="user-cell">
                <img class="mini-avatar" :src="api.userAvatarUrl(req.fromUserId)" alt="" />
                <span>{{ req.nickname || req.username }}（ID: {{ req.fromUserId }}）</span>
              </div>
            </td>
            <td>{{ formatTime(req.createdAt) }}</td>
            <td class="actions">
              <button class="btn-primary" :disabled="acting" @click="onAccept(req.fromUserId)">同意</button>
              <button class="btn-ghost" :disabled="acting" @click="onReject(req.fromUserId)">拒绝</button>
            </td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="card">
      <h3>搜索用户</h3>
      <div class="search-row">
        <input v-model="searchId" placeholder="输入对方用户 ID" />
        <button class="btn-primary" :disabled="loadingSearch" @click="onSearch">搜索</button>
      </div>
      <div v-if="searchResult" class="search-result">
        <div class="user-cell">
          <img class="mini-avatar" :src="api.userAvatarUrl(searchResult.id)" alt="" />
          <span>{{ searchResult.nickname || searchResult.username }}（ID: {{ searchResult.id }}）</span>
        </div>
        <button
          class="btn-primary"
          :disabled="adding || searchResult.id === auth.user?.id"
          @click="onSendRequest(searchResult.id)"
        >
          {{ searchResult.id === auth.user?.id ? '不能添加自己' : '发送好友申请' }}
        </button>
      </div>
      <p v-if="searchError" class="err">{{ searchError }}</p>
    </section>

    <section v-if="outgoing.length" class="card">
      <h3>已发送的申请</h3>
      <table class="table">
        <thead>
          <tr><th>用户 ID</th><th>用户名</th><th>发送时间</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="req in outgoing" :key="req.toUserId">
            <td>{{ req.toUserId }}</td>
            <td>
              <div class="user-cell">
                <img class="mini-avatar" :src="api.userAvatarUrl(req.toUserId)" alt="" />
                <span>{{ req.nickname || req.username }}</span>
              </div>
            </td>
            <td>{{ formatTime(req.createdAt) }}</td>
            <td>
              <button class="btn-ghost" :disabled="acting" @click="onCancel(req.toUserId)">撤回</button>
            </td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="card">
      <h3>好友列表</h3>
      <button class="btn-ghost" :disabled="loadingList" @click="reloadAll">刷新</button>
      <p v-if="listError" class="err">{{ listError }}</p>
      <table v-if="friends.length" class="table">
        <thead>
          <tr><th>用户 ID</th><th>用户名</th><th>添加时间</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="f in friends" :key="f.id">
            <td>{{ f.id }}</td>
            <td>
              <div class="user-cell">
                <img class="mini-avatar" :src="api.userAvatarUrl(f.id)" alt="" />
                <span>{{ f.nickname || f.username }}</span>
              </div>
            </td>
            <td>{{ formatTime(f.createdAt) }}</td>
            <td><DeleteButton sm @click="onRemove(f.id)" /></td>
          </tr>
        </tbody>
      </table>
      <p v-else-if="!loadingList" class="muted">暂无好友</p>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import DeleteButton from '../components/DeleteButton.vue'
import { usePageCrumbs } from '../composables/usePageCrumbs'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const { crumbs } = usePageCrumbs()
const searchId = ref('')
const searchResult = ref(null)
const searchError = ref('')
const loadingSearch = ref(false)
const friends = ref([])
const incoming = ref([])
const outgoing = ref([])
const listError = ref('')
const loadingList = ref(false)
const adding = ref(false)
const acting = ref(false)

function formatTime(t) {
  if (!t) return '-'
  return new Date(t).toLocaleString()
}

function friendErrorMessage(e, fallback) {
  if (!(e instanceof ApiError)) return fallback
  const m = e.message || ''
  if (m.includes('incoming friend request pending')) {
    return '对方已向你发送申请，请在上方「收到的好友申请」中同意'
  }
  if (m.includes('friend request already sent')) return '已发送过申请，请等待对方处理'
  if (m.includes('already friends')) return '你们已是好友'
  return m || fallback
}

async function loadRequests() {
  try {
    const [inc, out] = await Promise.all([
      api.listIncomingFriendRequests(),
      api.listOutgoingFriendRequests(),
    ])
    incoming.value = inc.requests || []
    outgoing.value = out.requests || []
  } catch {
    incoming.value = []
    outgoing.value = []
  }
}

async function loadFriends() {
  loadingList.value = true
  listError.value = ''
  try {
    const res = await api.listFriends()
    friends.value = res.friends || []
  } catch (e) {
    listError.value = e instanceof ApiError ? e.message : '加载失败'
  } finally {
    loadingList.value = false
  }
}

async function reloadAll() {
  await Promise.all([loadFriends(), loadRequests()])
}

async function onSearch() {
  searchError.value = ''
  searchResult.value = null
  if (!searchId.value.trim()) {
    searchError.value = '请输入用户 ID'
    return
  }
  loadingSearch.value = true
  try {
    searchResult.value = await api.searchUser(searchId.value.trim())
  } catch (e) {
    searchError.value = e instanceof ApiError ? e.message : '搜索失败'
  } finally {
    loadingSearch.value = false
  }
}

async function onSendRequest(id) {
  adding.value = true
  searchError.value = ''
  try {
    await api.addFriend(id)
    searchResult.value = null
    searchId.value = ''
    await loadRequests()
  } catch (e) {
    searchError.value = friendErrorMessage(e, '发送申请失败')
  } finally {
    adding.value = false
  }
}

async function onAccept(fromUserId) {
  acting.value = true
  try {
    await api.acceptFriendRequest(fromUserId)
    await reloadAll()
  } catch (e) {
    listError.value = e instanceof ApiError ? e.message : '操作失败'
  } finally {
    acting.value = false
  }
}

async function onReject(fromUserId) {
  acting.value = true
  try {
    await api.rejectFriendRequest(fromUserId)
    await loadRequests()
  } catch (e) {
    listError.value = e instanceof ApiError ? e.message : '操作失败'
  } finally {
    acting.value = false
  }
}

async function onCancel(toUserId) {
  acting.value = true
  try {
    await api.cancelFriendRequest(toUserId)
    await loadRequests()
  } catch (e) {
    listError.value = e instanceof ApiError ? e.message : '操作失败'
  } finally {
    acting.value = false
  }
}

async function onRemove(id) {
  if (!confirm('确定删除该好友？')) return
  try {
    await api.removeFriend(id)
    await reloadAll()
  } catch (e) {
    listError.value = e instanceof ApiError ? e.message : '删除失败'
  }
}

onMounted(reloadAll)
</script>

<style scoped>
.muted { color: var(--text-muted); font-size: 0.875rem; }
.card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); padding: 1.25rem; margin-bottom: 1rem; }
.card.highlight { border-color: var(--accent, #3b82f6); }
.card h3 { margin: 0 0 0.75rem; font-size: 1rem; }
.search-row {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
  max-width: 26rem;
  flex-wrap: wrap;
}
.search-row input { flex: 1; min-width: 10rem; max-width: 18rem; }
.search-result { display: flex; align-items: center; justify-content: space-between; gap: 1rem; padding: 0.75rem; background: var(--bg); border-radius: 8px; flex-wrap: wrap; }
.err { color: var(--danger); font-size: 0.875rem; margin-top: 0.5rem; }
.table { width: 100%; border-collapse: collapse; margin-top: 0.75rem; font-size: 0.875rem; }
.table th, .table td { text-align: left; padding: 0.5rem; border-bottom: 1px solid var(--border); }
.actions { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.user-cell { display: flex; align-items: center; gap: 0.5rem; }
.mini-avatar { width: 32px; height: 32px; border-radius: 50%; object-fit: cover; border: 1px solid var(--border); }
</style>
