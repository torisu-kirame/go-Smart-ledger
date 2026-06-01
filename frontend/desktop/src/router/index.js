import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/login', component: () => import('../views/LoginView.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    children: [
      {
        path: '',
        component: () => import('../views/DashboardView.vue'),
        meta: { titleKey: 'layout.nav.home', navRoot: '/' },
      },
      {
        path: 'entry-templates',
        redirect: '/ledgers',
      },
      {
        path: 'assistant',
        component: () => import('../views/AiAssistantView.vue'),
        meta: { titleKey: 'layout.nav.assistant', navRoot: '/assistant' },
      },
      {
        path: 'ledgers',
        component: () => import('../views/LedgersView.vue'),
        meta: { titleKey: 'layout.nav.ledgers', navRoot: '/ledgers' },
      },
      {
        path: 'ledgers/:id',
        component: () => import('../layouts/LedgerDetailLayout.vue'),
        meta: {
          titleKey: 'layout.nav.ledgers',
          breadcrumbParent: '/ledgers',
          navRoot: '/ledgers',
        },
        children: [
          {
            path: '',
            component: () => import('../views/ledger/LedgerOverviewView.vue'),
          },
          {
            path: 'view',
            component: () => import('../views/ledger/LedgerContentView.vue'),
          },
          {
            path: 'import',
            component: () => import('../views/ledger/LedgerImportView.vue'),
          },
          {
            path: 'accounting',
            component: () => import('../views/ledger/LedgerAccountingLayout.vue'),
            children: [
              { path: '', redirect: (to) => `${to.path}/view` },
              {
                path: 'view',
                component: () => import('../views/ledger/LedgerAccountingBrowseView.vue'),
              },
              {
                path: 'coa',
                component: () => import('../views/ledger/LedgerAccountingCoaView.vue'),
              },
              {
                path: 'period',
                component: () => import('../views/ledger/LedgerAccountingOpsView.vue'),
              },
              {
                path: 'report',
                component: () => import('../views/ledger/LedgerAccountingOpsView.vue'),
              },
              {
                path: 'attach',
                component: () => import('../views/ledger/LedgerAccountingOpsView.vue'),
              },
              {
                path: 'bank',
                component: () => import('../views/ledger/LedgerAccountingOpsView.vue'),
              },
              {
                path: 'budget',
                component: () => import('../views/ledger/LedgerAccountingBudgetView.vue'),
              },
              {
                path: 'aging',
                component: () => import('../views/ledger/LedgerAccountingAgingView.vue'),
              },
              {
                path: 'currency',
                component: () => import('../views/ledger/LedgerAccountingCurrencyView.vue'),
              },
              {
                path: 'tax',
                component: () => import('../views/ledger/LedgerAccountingTaxView.vue'),
              },
            ],
          },
          {
            path: 'templates',
            component: () => import('../views/EntryTemplatesView.vue'),
          },
          {
            path: 'settings',
            component: () => import('../views/ledger/LedgerSettingsView.vue'),
          },
        ],
      },
      {
        path: 'backup',
        component: () => import('../views/BackupView.vue'),
        meta: { titleKey: 'layout.nav.backup', navRoot: '/backup' },
      },
      {
        path: 'friends',
        component: () => import('../views/FriendsView.vue'),
        meta: { titleKey: 'layout.nav.friends', navRoot: '/friends' },
      },
      {
        path: 'settings',
        component: () => import('../views/SettingsView.vue'),
        meta: { titleKey: 'layout.settings', navRoot: '/settings' },
      },
      { path: 'profile', redirect: '/settings' },
      {
        path: 'teams',
        component: () => import('../views/TeamsView.vue'),
        meta: { titleKey: 'layout.nav.teams', navRoot: '/teams' },
      },
      {
        path: 'teams/:teamId',
        component: () => import('../views/TeamDetailView.vue'),
        meta: {
          titleKey: 'layout.nav.teams',
          breadcrumbParent: '/teams',
          navRoot: '/teams',
        },
      },
      { path: 'invites', redirect: '/ledgers' },
      {
        path: 'chain',
        component: () => import('../views/ChainExplorerView.vue'),
        meta: { titleKey: 'layout.nav.chain', navRoot: '/chain' },
      },
      {
        path: 'logs',
        component: () => import('../views/LogView.vue'),
        meta: { titleKey: 'layout.nav.logs', navRoot: '/logs' },
      },
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
