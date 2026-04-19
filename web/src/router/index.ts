import { createRouter, createWebHistory } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/chat'
    },
    {
      path: '/',
      component: () => import('@/layouts/AuthLayout.vue'),
      meta: { guestOnly: true },
      children: [
        {
          path: 'login',
          name: 'login',
          component: () => import('@/pages/LoginPage.vue')
        },
        {
          path: 'signup',
          name: 'signup',
          component: () => import('@/pages/SignupPage.vue')
        }
      ]
    },
    {
      path: '/',
      component: () => import('@/layouts/AppLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: 'chat/:conversationId?',
          name: 'chat',
          component: () => import('@/pages/ChatPage.vue')
        },
        {
          path: 'usage',
          name: 'usage',
          component: () => import('@/pages/UsagePage.vue')
        },
        {
          path: 'settings/profile',
          name: 'settings-profile',
          component: () => import('@/pages/SettingsProfilePage.vue')
        },
        {
          path: 'settings/security',
          name: 'settings-security',
          component: () => import('@/pages/SettingsSecurityPage.vue')
        }
      ]
    },
    {
      path: '/admin',
      component: () => import('@/layouts/AdminLayout.vue'),
      meta: { requiresAuth: true, adminOnly: true },
      children: [
        {
          path: '',
          redirect: '/admin/upstreams'
        },
        {
          path: 'upstreams',
          name: 'admin-upstreams',
          component: () => import('@/pages/admin/AdminUpstreamsPage.vue')
        },
        {
          path: 'models',
          name: 'admin-models',
          component: () => import('@/pages/admin/AdminModelsPage.vue')
        },
        {
          path: 'channels',
          name: 'admin-channels',
          component: () => import('@/pages/admin/AdminChannelsPage.vue')
        },
        {
          path: 'user-groups',
          name: 'admin-user-groups',
          component: () => import('@/pages/admin/AdminUserGroupsPage.vue')
        },
        {
          path: 'users',
          name: 'admin-users',
          component: () => import('@/pages/admin/AdminUsersPage.vue')
        },
        {
          path: 'redeem-codes',
          name: 'admin-redeem-codes',
          component: () => import('@/pages/admin/AdminRedeemCodesPage.vue')
        },
        {
          path: 'audit-logs',
          name: 'admin-audit-logs',
          component: () => import('@/pages/admin/AdminAuditLogsPage.vue')
        }
      ]
    },
    {
      path: '/403',
      name: 'forbidden',
      component: () => import('@/pages/ForbiddenPage.vue')
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/pages/NotFoundPage.vue')
    }
  ]
})

let refreshPromise: Promise<boolean> | null = null

async function ensureSession(auth: ReturnType<typeof useAuthStore>) {
  if (auth.isAuthenticated) {
    return true
  }

  if (!refreshPromise) {
    refreshPromise = auth.refreshSession().finally(() => {
      refreshPromise = null
    })
  }

  return refreshPromise
}

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (to.meta.guestOnly) {
    const recovered = await ensureSession(auth)
    if (recovered) {
      return { name: 'chat' }
    }

    return true
  }

  if (to.meta.requiresAuth) {
    const recovered = await ensureSession(auth)
    if (!recovered) {
      return { name: 'login' }
    }

    if (!auth.user) {
      const currentUser = await auth.fetchMe().catch(() => null)
      if (!currentUser) {
        return { name: 'login' }
      }
    }
  }

  if (to.meta.adminOnly && !auth.isAdmin) {
    return { name: 'forbidden' }
  }

  return true
})

export default router
