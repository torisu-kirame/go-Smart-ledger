<template>
  <div class="shell">
    <aside class="sidebar">
      <div class="brand">
        <strong>Smart Ledger</strong>
        <small>桌面控制台</small>
      </div>
      <nav>
        <router-link to="/">概览</router-link>
        <router-link to="/ledgers">账本管理</router-link>
        <router-link to="/import">Excel 导入</router-link>
        <router-link to="/backup">备份 / 恢复</router-link>
        <router-link to="/friends">好友</router-link>
      </nav>
      <div class="foot">
        <div>{{ auth.user?.username }}</div>
        <div v-if="auth.user?.id" class="uid">ID: {{ auth.user.id }}</div>
        <button class="btn-ghost" style="width:100%;margin:0.5rem 0" @click="onLogout">退出</button>
        <a href="http://localhost:24441/dashboard" target="_blank" rel="noreferrer">MiniLedger 浏览器 →</a>
      </div>
    </aside>
    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

onMounted(() => {
  if (!auth.loading && !auth.isLoggedIn) router.replace('/login')
})

async function onLogout() {
  await auth.logout()
  router.replace('/login')
}
</script>

<style scoped>
.shell { display: flex; min-height: 100vh; }
.sidebar {
  width: 220px; background: var(--bg-elevated); border-right: 1px solid var(--border);
  display: flex; flex-direction: column; padding: 1rem 0;
}
.brand { padding: 0 1rem 1rem; border-bottom: 1px solid var(--border); }
.brand small { display: block; color: var(--text-muted); margin-top: 0.25rem; }
nav { flex: 1; padding: 0.75rem; display: flex; flex-direction: column; gap: 0.25rem; }
nav a { padding: 0.55rem 0.65rem; border-radius: 8px; color: var(--text-muted); }
nav a.router-link-active { background: rgba(61,139,253,.15); color: var(--accent); }
.foot { padding: 1rem; border-top: 1px solid var(--border); font-size: 0.75rem; color: var(--text-muted); }
.uid { margin-top: 0.25rem; color: var(--accent); font-family: monospace; }
.main { flex: 1; padding: 1.5rem; overflow: auto; }
</style>
