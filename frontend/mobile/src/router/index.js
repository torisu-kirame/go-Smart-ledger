import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const MobileLayout = () => import('../layouts/MobileLayout.vue')

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { public: true, hideTabbar: true },
    },
    {
      path: '/',
      component: MobileLayout,
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('../views/HomeView.vue'),
          meta: { tab: 'home', title: '首页' },
        },
        {
          path: 'ledgers',
          name: 'ledgers',
          component: () => import('../views/LedgersView.vue'),
          meta: { tab: 'ledgers', title: '账本' },
        },
        {
          path: 'ledgers/:id',
          name: 'ledger-detail',
          component: () => import('../views/LedgerDetailView.vue'),
          meta: { hideTabbar: true, title: '账本详情' },
        },
        {
          path: 'collab',
          name: 'collab',
          component: () => import('../views/CollabView.vue'),
          meta: { tab: 'collab', title: '协作' },
        },
        {
          path: 'profile',
          name: 'profile',
          component: () => import('../views/ProfileView.vue'),
          meta: { tab: 'profile', title: '我的' },
        },
      ],
    },
  ],
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (auth.loading) await auth.bootstrap()
  if (to.meta.public) {
    if (auth.isLoggedIn) return { path: '/' }
    return true
  }
  if (!auth.isLoggedIn) return { path: '/login', query: { redirect: to.fullPath } }
  return true
})

export default router
