import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/login', component: () => import('../views/LoginView.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    children: [
      { path: '', component: () => import('../views/DashboardView.vue') },
      { path: 'entry-templates', component: () => import('../views/EntryTemplatesView.vue') },
      { path: 'ledgers', component: () => import('../views/LedgersView.vue') },
      { path: 'ledgers/:id', component: () => import('../views/LedgerDetailView.vue') },
      { path: 'import', component: () => import('../views/ImportView.vue') },
      { path: 'backup', component: () => import('../views/BackupView.vue') },
      { path: 'friends', component: () => import('../views/FriendsView.vue') },
      { path: 'settings', component: () => import('../views/SettingsView.vue') },
      { path: 'profile', redirect: '/settings' },
      { path: 'teams', component: () => import('../views/TeamsView.vue') },
      { path: 'teams/:teamId', component: () => import('../views/TeamDetailView.vue') },
      { path: 'invites', redirect: '/ledgers' },
      { path: 'chain', component: () => import('../views/ChainExplorerView.vue') },
    ],
  },
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.loading && !auth.isLoggedIn && !to.meta.public) {
    return '/login'
  }
  if (auth.isLoggedIn && to.path === '/login') {
    return '/'
  }
})

export default router
