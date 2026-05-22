import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/login', component: () => import('../views/LoginView.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    children: [
      { path: '', component: () => import('../views/DashboardView.vue') },
      { path: 'ledgers', component: () => import('../views/LedgersView.vue') },
      { path: 'ledgers/:id', component: () => import('../views/LedgerDetailView.vue') },
      { path: 'import', component: () => import('../views/ImportView.vue') },
      { path: 'backup', component: () => import('../views/BackupView.vue') },
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
