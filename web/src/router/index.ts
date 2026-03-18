import { createRouter, createWebHistory } from 'vue-router'

import AdminLayout from '@/layouts/AdminLayout.vue'
import AppLayout from '@/layouts/AppLayout.vue'
import AuthLayout from '@/layouts/AuthLayout.vue'
import AdminAuditLogsPage from '@/pages/admin/AdminAuditLogsPage.vue'
import AdminChannelsPage from '@/pages/admin/AdminChannelsPage.vue'
import AdminModelsPage from '@/pages/admin/AdminModelsPage.vue'
import AdminRedeemCodesPage from '@/pages/admin/AdminRedeemCodesPage.vue'
import AdminUserGroupsPage from '@/pages/admin/AdminUserGroupsPage.vue'
import AdminUpstreamsPage from '@/pages/admin/AdminUpstreamsPage.vue'
import AdminUsersPage from '@/pages/admin/AdminUsersPage.vue'
import ChatPage from '@/pages/ChatPage.vue'
import ForbiddenPage from '@/pages/ForbiddenPage.vue'
import LoginPage from '@/pages/LoginPage.vue'
import NotFoundPage from '@/pages/NotFoundPage.vue'
import SettingsProfilePage from '@/pages/SettingsProfilePage.vue'
import SettingsSecurityPage from '@/pages/SettingsSecurityPage.vue'
import SignupPage from '@/pages/SignupPage.vue'
import UsagePage from '@/pages/UsagePage.vue'
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
      component: AuthLayout,
      meta: { guestOnly: true },
      children: [
        {
          path: 'login',
          name: 'login',
          component: LoginPage
        },
        {
          path: 'signup',
          name: 'signup',
          component: SignupPage
        }
      ]
    },
    {
      path: '/',
      component: AppLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: 'chat/:conversationId?',
          name: 'chat',
          component: ChatPage
        },
        {
          path: 'usage',
          name: 'usage',
          component: UsagePage
        },
        {
          path: 'settings/profile',
          name: 'settings-profile',
          component: SettingsProfilePage
        },
        {
          path: 'settings/security',
          name: 'settings-security',
          component: SettingsSecurityPage
        }
      ]
    },
    {
      path: '/admin',
      component: AdminLayout,
      meta: { requiresAuth: true, adminOnly: true },
      children: [
        {
          path: '',
          redirect: '/admin/upstreams'
        },
        {
          path: 'upstreams',
          name: 'admin-upstreams',
          component: AdminUpstreamsPage
        },
        {
          path: 'models',
          name: 'admin-models',
          component: AdminModelsPage
        },
        {
          path: 'channels',
          name: 'admin-channels',
          component: AdminChannelsPage
        },
        {
          path: 'user-groups',
          name: 'admin-user-groups',
          component: AdminUserGroupsPage
        },
        {
          path: 'users',
          name: 'admin-users',
          component: AdminUsersPage
        },
        {
          path: 'redeem-codes',
          name: 'admin-redeem-codes',
          component: AdminRedeemCodesPage
        },
        {
          path: 'audit-logs',
          name: 'admin-audit-logs',
          component: AdminAuditLogsPage
        }
      ]
    },
    {
      path: '/403',
      name: 'forbidden',
      component: ForbiddenPage
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: NotFoundPage
    }
  ]
})

let refreshPromise: Promise<boolean> | null = null

// async function ensureSession(auth: ReturnType<typeof useAuthStore>) {
//   if (auth.isAuthenticated) {
//     return true
//   }

//   if (!refreshPromise) {
//     refreshPromise = auth.refreshSession().finally(() => {
//       refreshPromise = null
//     })
//   }

//   return refreshPromise
// }

// router.beforeEach(async (to) => {
//   const auth = useAuthStore()

//   if (to.meta.guestOnly) {
//     const recovered = await ensureSession(auth)
//     if (recovered) {
//       return { name: 'chat' }
//     }

//     return true
//   }

//   if (to.meta.requiresAuth) {
//     const recovered = await ensureSession(auth)
//     if (!recovered) {
//       return { name: 'login' }
//     }

//     if (!auth.user) {
//       await auth.fetchMe().catch(() => null)
//     }
//   }

//   if (to.meta.adminOnly && !auth.isAdmin) {
//     return { name: 'forbidden' }
//   }

//   return true
// })

export default router
