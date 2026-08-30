<template>
  <div class="shell">
    <aside class="sidebar">
      <router-link to="/" class="brand" @click="onBrandClick">
        <span class="brand-icon" aria-hidden="true">
          <AppIcon name="brand" size="lg" />
        </span>
        <span class="brand-text">
          <strong>Smart Ledger</strong>
          <small>{{ t('layout.tagline') }}</small>
        </span>
      </router-link>

      <nav class="sidebar-nav" :aria-label="t('layout.nav.label')">
        <div v-for="group in navGroups" :key="group.id" class="nav-group">
          <span class="nav-group-label">{{ group.label }}</span>
          <router-link
            v-for="item in group.items"
            :key="item.to"
            :to="item.to"
            custom
            v-slot="{ href, navigate }"
          >
            <a
              :href="href"
              class="nav-item"
              :class="{ active: navItemActive(item) }"
              @click="onNavClick(item, navigate, $event)"
            >
              <AppIcon :name="item.icon" size="sm" class="nav-item__icon" />
              <span>{{ item.label }}</span>
            </a>
          </router-link>
        </div>
      </nav>

      <div class="foot">
        <a
          class="user-card"
          href="/settings#account"
          @click.prevent="goSettings('#account')"
        >
          <img v-if="auth.user?.id" class="foot-avatar" :src="footAvatar" alt="" />
          <span v-else class="foot-avatar foot-avatar--placeholder">
            <AppIcon name="user" size="sm" />
          </span>
          <span class="user-meta">
            <span class="user-name">{{ auth.user?.nickname || auth.user?.username }}</span>
            <span class="user-hint">{{ t('layout.profileHint') }}</span>
          </span>
          <AppIcon name="chevron-right" size="sm" class="user-chevron" />
        </a>
        <button class="foot-logout icon-btn icon-btn--ghost" type="button" @click="onLogout">
          <AppIcon name="logout" size="sm" />
          <span>{{ t('layout.logout') }}</span>
        </button>
      </div>
    </aside>
    <main class="main" :class="{ 'main--split': panes.split }">
      <WorkspaceTabBar v-if="showWorkspaceTabs" />
      <div v-if="panes.split" class="split-workspace">
        <section
          class="split-pane"
          :class="{ 'split-pane--focus': panes.focus === 'left' }"
          @mousedown="panes.setFocus('left')"
        >
          <component :is="paneView(panes.leftId)" />
        </section>
        <div
          class="split-resizer"
          title="Drag to resize"
          @mousedown.prevent="startResize"
        />
        <section
          class="split-pane"
          :class="{ 'split-pane--focus': panes.focus === 'right' }"
          @mousedown="panes.setFocus('right')"
        >
          <component :is="paneView(panes.rightId)" />
        </section>
      </div>
      <router-view v-else />
    </main>
    <ToastStack />
  </div>
</template>

<script setup>
import { computed, defineAsyncComponent, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppIcon from '../components/AppIcon.vue'
import ToastStack from '../components/ToastStack.vue'
import WorkspaceTabBar from '../components/WorkspaceTabBar.vue'
import { api } from '../api/http'
import { useAuthStore } from '../stores/auth'
import { useWorkspacePanes, tabFromPath } from '../stores/workspacePanes'
import { useI18n } from '../composables/useI18n'
import { NAV_ICON_BY_ROUTE } from '../icons/registry.js'

const AiAssistantView = defineAsyncComponent(() => import('../views/AiAssistantView.vue'))
const LedgersView = defineAsyncComponent(() => import('../views/LedgersView.vue'))

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const panes = useWorkspacePanes()
const { t } = useI18n()

const leftWidthPct = ref(50)

function navItem(to, labelKey, exact = true) {
  return {
    to,
    label: t(labelKey),
    exact,
    icon: NAV_ICON_BY_ROUTE[to] || 'home',
  }
}

const navGroups = computed(() => [
  {
    id: 'workspace',
    label: t('layout.navGroup.workspace'),
    items: [
      navItem('/', 'layout.nav.home', true),
      // 账本管理在上，AI 助手在下
      navItem('/ledgers', 'layout.nav.ledgers', false),
      navItem('/assistant', 'layout.nav.assistant', true),
    ],
  },
  {
    id: 'data',
    label: t('layout.navGroup.data'),
    items: [
      navItem('/backup', 'layout.nav.backup', true),
    ],
  },
  {
    id: 'collab',
    label: t('layout.navGroup.collab'),
    items: [
      navItem('/friends', 'layout.nav.friends', true),
      navItem('/teams', 'layout.nav.teams', true),
    ],
  },
  {
    id: 'system',
    label: t('layout.navGroup.system'),
    items: [
      navItem('/chain', 'layout.nav.chain', true),
      navItem('/logs', 'layout.nav.logs', true),
      navItem('/settings', 'layout.settings', true),
    ],
  },
])

const footAvatar = computed(() =>
  auth.user?.id ? api.userAvatarUrl(auth.user.id) : ''
)

const showWorkspaceTabs = computed(() => {
  if (panes.split) return true
  return !!tabFromPath(route.path) || panes.openTabs.length > 0
})

function paneView(id) {
  if (id === 'assistant') return AiAssistantView
  if (id === 'ledgers') return LedgersView
  return LedgersView
}

function navItemActive(item) {
  const navRoot = route.meta?.navRoot
  if (navRoot && item.to === navRoot) return true
  if (panes.split) {
    const focused = panes.focus === 'right' ? panes.rightTab : panes.leftTab
    if (focused && item.to === focused.path) return true
  }
  const path = route.path
  if (item.to === '/') return path === '/' || path === ''
  if (path === item.to) return true
  return path.startsWith(`${item.to}/`)
}

function onNavClick(item, navigate, ev) {
  const tab = tabFromPath(item.to)
  if (tab) {
    if (panes.split) {
      panes.placeOnSide(tab.id, panes.focus)
    } else {
      panes.openSingle(tab.id)
    }
  } else if (panes.split) {
    panes.exitSplit(panes.focus === 'right' ? panes.rightId : panes.leftId)
  }
  navigate(ev)
}

function goSettings(hash = '') {
  const h = typeof hash === 'string' && hash.startsWith('#') ? hash : ''
  if (panes.split) panes.exitSplit()
  router.push(h ? { path: '/settings', hash: h } : '/settings')
}

function onBrandClick(ev) {
  if (panes.split) panes.exitSplit()
  if (router.currentRoute.value.path === '/') {
    ev.preventDefault()
  }
}

watch(
  () => route.path,
  (path) => {
    if (path.startsWith('/ledgers/') && path !== '/ledgers') {
      if (panes.split) panes.exitSplit('ledgers')
      return
    }
    panes.syncFromRoute(path)
  },
  { immediate: true }
)

function startResize(e) {
  const root = e.currentTarget.parentElement
  if (!root) return
  const rect = root.getBoundingClientRect()
  function onMove(ev) {
    const pct = ((ev.clientX - rect.left) / rect.width) * 100
    leftWidthPct.value = Math.min(75, Math.max(25, pct))
    root.style.setProperty('--split-left', `${leftWidthPct.value}%`)
  }
  function onUp() {
    window.removeEventListener('mousemove', onMove)
    window.removeEventListener('mouseup', onUp)
  }
  window.addEventListener('mousemove', onMove)
  window.addEventListener('mouseup', onUp)
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
  width: var(--sidebar-width);
  flex-shrink: 0;
  height: 100vh;
  max-height: 100vh;
  background: var(--bg-elevated);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 0.85rem 0 0;
  overflow: hidden;
}

.brand {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.35rem 1rem 1rem;
  margin: 0 0.5rem 0.5rem;
  border-bottom: 1px solid var(--border);
  text-decoration: none;
  color: inherit;
  border-radius: var(--radius-sm);
  transition: background 0.15s ease;
}

.brand:hover {
  background: var(--hover);
}

.brand-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 12px;
  background: linear-gradient(145deg, var(--accent-soft), color-mix(in srgb, var(--accent) 22%, transparent));
  color: var(--accent);
  flex-shrink: 0;
}

.brand-text {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 0;
}

.brand-text strong {
  font-size: 0.95rem;
  letter-spacing: 0.01em;
}

.brand-text small {
  font-size: 0.68rem;
  color: var(--text-muted);
  font-weight: 500;
}

.sidebar-nav {
  flex: 1;
  min-height: 0;
  padding: 0.5rem 0.65rem;
  overflow-y: auto;
}

.nav-group + .nav-group {
  margin-top: 0.85rem;
}

.nav-group-label {
  display: block;
  padding: 0 0.55rem;
  margin-bottom: 0.35rem;
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-muted);
  opacity: 0.85;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.55rem 0.65rem;
  border-radius: 10px;
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 600;
  transition: background-color 0.18s ease, color 0.18s ease;
}

.nav-item__icon {
  opacity: 0.9;
}

.nav-item:hover {
  background: var(--hover);
  color: var(--text);
}

.nav-item.active {
  background: var(--accent-soft);
  color: var(--accent);
}

.nav-item.active .nav-item__icon {
  color: var(--accent);
}

.foot {
  flex-shrink: 0;
  padding: 0.85rem;
  border-top: 1px solid var(--border);
}

.user-card {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.55rem 0.6rem;
  margin-bottom: 0.5rem;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text);
  text-decoration: none;
  cursor: pointer;
  transition: border-color 0.15s ease, background 0.15s ease;
}

.user-card:hover {
  border-color: color-mix(in srgb, var(--accent) 45%, var(--border));
  background: var(--hover);
}

.foot-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--border);
  flex-shrink: 0;
}

.foot-avatar--placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--accent-soft);
  color: var(--accent);
}

.user-meta {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}

.user-name {
  font-size: 0.85rem;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-hint {
  font-size: 0.68rem;
  color: var(--text-muted);
}

.user-chevron {
  color: var(--text-muted);
  flex-shrink: 0;
}

.foot-logout {
  width: 100%;
  justify-content: center;
}

.main {
  flex: 1;
  min-width: 0;
  min-height: 0;
  height: 100vh;
  padding: 1.5rem 2rem 2rem;
  overflow-x: hidden;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.main--split {
  overflow: hidden;
}

.split-workspace {
  --split-left: 50%;
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: var(--split-left) 6px minmax(0, 1fr);
  gap: 0;
  overflow: hidden;
}

.split-pane {
  min-width: 0;
  min-height: 0;
  overflow: auto;
  padding: 0 0.35rem;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
}

.split-pane--focus {
  border-color: color-mix(in srgb, var(--accent) 35%, var(--border));
  background: color-mix(in srgb, var(--accent) 4%, transparent);
}

.split-resizer {
  cursor: col-resize;
  background: var(--border);
  border-radius: 2px;
  margin: 0.25rem 0;
}

.split-resizer:hover {
  background: var(--accent);
}

@media (max-width: 960px) {
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
    padding-bottom: 0.5rem;
  }

  .sidebar-nav {
    flex: none;
    min-height: auto;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(9.5rem, 1fr));
    gap: 0.25rem;
  }

  .nav-group {
    margin-top: 0;
  }

  .nav-group-label {
    grid-column: 1 / -1;
  }

  .main {
    height: auto;
    min-height: 50vh;
    overflow: visible;
    padding: 1rem;
  }

  .main--split {
    overflow: visible;
  }

  .split-workspace {
    grid-template-columns: 1fr;
    grid-template-rows: auto auto auto;
  }

  .split-resizer {
    display: none;
  }
}
</style>
