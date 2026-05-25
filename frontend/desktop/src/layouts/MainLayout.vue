<template>
  <div class="shell">
    <aside class="sidebar">
      <div class="brand">
        <strong>Smart Ledger</strong>
      </div>
      <nav>
        <router-link
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          custom
          v-slot="{ href, navigate, isActive, isExactActive }"
        >
          <a
            :href="href"
            class="nav-item"
            :class="{ active: navActive(item, isActive, isExactActive) }"
            @click="navigate"
          >
            {{ item.label }}
          </a>
        </router-link>
      </nav>
      <div class="foot">
        <button class="settings-btn btn-ghost" type="button" @click="goSettings()">{{ t('layout.settings') }}</button>
        <a class="user-link" href="/settings#account" @click.prevent="goSettings('#account')">
          <img v-if="auth.user?.id" class="foot-avatar" :src="footAvatar" alt="" />
          <span>{{ auth.user?.nickname || auth.user?.username }}</span>
        </a>
        <button class="btn-ghost foot-logout" type="button" @click="onLogout">{{ t('layout.logout') }}</button>
        <router-link to="/chain" class="ext-link">{{ t('layout.chainLink') }}</router-link>
      </div>
    </aside>
    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/http'
import { useAuthStore } from '../stores/auth'
import { useI18n } from '../composables/useI18n'

const auth = useAuthStore()
const router = useRouter()
const { t } = useI18n()

const navItems = computed(() => [
  { to: '/', label: t('layout.nav.home'), exact: true },
  { to: '/ledgers', label: t('layout.nav.ledgers'), exact: false },
  { to: '/entry-templates', label: t('layout.nav.templates'), exact: true },
  { to: '/import', label: t('layout.nav.import'), exact: true },
  { to: '/backup', label: t('layout.nav.backup'), exact: true },
  { to: '/friends', label: t('layout.nav.friends'), exact: true },
  { to: '/teams', label: t('layout.nav.teams'), exact: true },
  { to: '/chain', label: t('layout.nav.chain'), exact: true },
])

const footAvatar = computed(() =>
  auth.user?.id ? api.userAvatarUrl(auth.user.id) : ''
)

function navActive(item, isActive, isExactActive) {
  if (item.exact) return isExactActive
  return isActive
}

function goSettings(hash = '') {
  const h = typeof hash === 'string' && hash.startsWith('#') ? hash : ''
  router.push(h ? { path: '/settings', hash: h } : '/settings')
}

onMounted(() => {
  document.documentElement.classList.add('app-shell-active')
  if (!auth.loading && !auth.isLoggedIn) router.replace('/login')
})

onUnmounted(() => {
  document.documentElement.classList.remove('app-shell-active')
})

async function onLogout() {
  await auth.logout()
  router.replace('/login')
}
</script>

<style scoped>
.shell {
  display: flex;
  height: 100vh;
  max-height: 100vh;
  overflow: hidden;
}

.sidebar {
  width: 240px;
  flex-shrink: 0;
  height: 100vh;
  max-height: 100vh;
  background: var(--bg-elevated);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 1rem 0;
  overflow: hidden;
}

.brand {
  flex-shrink: 0;
  padding: 0 1rem 1rem;
  border-bottom: 1px solid var(--border);
}
nav {
  flex: 1;
  min-height: 0;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  overflow-y: auto;
}
.nav-item {
  display: block;
  padding: 0.62rem 0.7rem;
  border-radius: 10px;
  color: var(--text-muted);
  text-decoration: none;
  transition: background-color 0.18s ease, color 0.18s ease;
}
.nav-item:hover { background: var(--hover); color: var(--text); }
.nav-item.active {
  background: var(--accent-soft);
  color: var(--accent);
  font-weight: 600;
}
.foot {
  flex-shrink: 0;
  padding: 1rem;
  border-top: 1px solid var(--border);
  font-size: 0.75rem;
  color: var(--text-muted);
}
.user-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--text);
  text-decoration: none;
  margin-bottom: 0.25rem;
  cursor: pointer;
}
.user-link:hover { color: var(--accent); }
.foot-avatar { width: 28px; height: 28px; border-radius: 50%; object-fit: cover; border: 1px solid var(--border); }
.foot-logout { width: 100%; margin: 0.5rem 0; }
.settings-btn { width: 100%; margin-bottom: 0.75rem; }
.ext-link { display: inline-block; margin-top: 0.25rem; }
.main {
  flex: 1;
  min-width: 0;
  min-height: 0;
  height: 100vh;
  padding: 1.5rem 2rem;
  overflow-x: hidden;
  overflow-y: auto;
}

@media (max-width: 900px) {
  .shell {
    flex-direction: column;
    height: auto;
    max-height: none;
    overflow: visible;
  }
  .sidebar {
    width: 100%;
    height: auto;
    max-height: none;
    overflow: visible;
  }
  nav {
    flex: none;
    min-height: auto;
    flex-direction: row;
    flex-wrap: wrap;
    overflow: visible;
  }
  .main {
    height: auto;
    min-height: 50vh;
    overflow: visible;
    padding: 1rem;
  }
}
</style>
