<template>
  <div class="page">
    <h2>好友</h2>
    <p class="page-desc">通过用户 ID 搜索并添加好友</p>
    <p v-if="auth.user" class="my-id">我的用户 ID：<strong>{{ auth.user.id }}</strong> · {{ auth.user.username }}</p>

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
          @click="onAdd(searchResult.id)"
        >
          {{ searchResult.id === auth.user?.id ? '不能添加自己' : '添加好友' }}
        </button>
      </div>
      <p v-if="searchError" class="err">{{ searchError }}</p>
    </section>

    <section class="card">
      <h3>好友列表</h3>
      <button class="btn-ghost" :disabled="loadingList" @click="loadFriends">刷新</button>
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
            <td><button class="btn-ghost danger" @click="onRemove(f.id)">删除</button></td>
          </tr>
        </tbody>
      </table>
      <p v-else-if="!loadingList" class="muted">暂无好友</p>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const searchId = ref('')
const searchResult = ref(null)
const searchError = ref('')
const loadingSearch = ref(false)
const friends = ref([])
const listError = ref('')
const loadingList = ref(false)
const adding = ref(false)

function formatTime(t) {
  if (!t) return '-'
  return new Date(t).toLocaleString()
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

async function onAdd(id) {
  adding.value = true
  try {
    await api.addFriend(id)
    searchResult.value = null
    searchId.value = ''
    await loadFriends()
  } catch (e) {
    searchError.value = e instanceof ApiError ? e.message : '添加失败'
  } finally {
    adding.value = false
  }
}

async function onRemove(id) {
  if (!confirm('确定删除该好友？')) return
  try {
    await api.removeFriend(id)
    await loadFriends()
  } catch (e) {
    listError.value = e instanceof ApiError ? e.message : '删除失败'
  }
}

onMounted(loadFriends)
</script>

<style scoped>
.muted { color: var(--text-muted); font-size: 0.875rem; }
.my-id { margin: 0.5rem 0 1rem; font-size: 0.9rem; }
.card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); padding: 1.25rem; margin-bottom: 1rem; }
.card h3 { margin: 0 0 0.75rem; font-size: 1rem; }
.search-row {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
  max-width: 26rem;
  flex-wrap: wrap;
}
.search-row input { flex: 1; min-width: 10rem; max-width: 18rem; }
.search-result { display: flex; align-items: center; justify-content: space-between; gap: 1rem; padding: 0.75rem; background: var(--bg); border-radius: 8px; }
.err { color: var(--danger); font-size: 0.875rem; margin-top: 0.5rem; }
.danger { color: var(--danger); }
.table { width: 100%; border-collapse: collapse; margin-top: 0.75rem; font-size: 0.875rem; }
.table th, .table td { text-align: left; padding: 0.5rem; border-bottom: 1px solid var(--border); }
.user-cell { display: flex; align-items: center; gap: 0.5rem; }
.mini-avatar { width: 32px; height: 32px; border-radius: 50%; object-fit: cover; border: 1px solid var(--border); }
</style>
