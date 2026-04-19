import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { apiRequest, registerAuthSessionStore } from '@/lib/api'
import { refreshToken, signOut as signOutRequest } from '@/api/auth'

const accessTokenKey = 'mrchat.access_token'
const userKey = 'mrchat.user'

export interface AuthUser {
  id: string
  username: string
  email: string
  role: 'user' | 'admin' | 'root'
  settings?: {
    timezone?: string
    locale?: string
  }
}

export interface CurrentUser extends AuthUser {
  display_name: string
  avatar_url: string | null
  status: string
  quota: number
  used_quota: number
  created_at: string
  updated_at: string
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref(localStorage.getItem(accessTokenKey) ?? '')
  const user = ref<CurrentUser | AuthUser | null>(readStoredUser())

  const isAuthenticated = computed(() => accessToken.value.length > 0)
  const role = computed(() => user.value?.role ?? 'guest')
  const isAdmin = computed(() => role.value === 'admin' || role.value === 'root')

  function setSession(token: string, nextUser: CurrentUser | AuthUser) {
    accessToken.value = token
    user.value = nextUser
    localStorage.setItem(accessTokenKey, token)
    localStorage.setItem(userKey, JSON.stringify(nextUser))
  }

  function clearSession() {
    accessToken.value = ''
    user.value = null
    localStorage.removeItem(accessTokenKey)
    localStorage.removeItem(userKey)
  }

  async function refreshSession() {
    try {
      const data = await refreshToken()
      setSession(data.access_token, data.user)
      return true
    } catch {
      clearSession()
      return false
    }
  }

  async function fetchMe() {
    if (!accessToken.value) {
      return null
    }

    try {
      const { data } = await apiRequest<CurrentUser>('/users/me', {
        accessToken: accessToken.value
      })
      user.value = data
      localStorage.setItem(userKey, JSON.stringify(data))
      return data
    } catch {
      clearSession()
      return null
    }
  }

  async function signOut() {
    try {
      await signOutRequest()
    } finally {
      clearSession()
    }
  }

  return {
    accessToken,
    user,
    role,
    isAuthenticated,
    isAdmin,
    setSession,
    clearSession,
    refreshSession,
    fetchMe,
    signOut
  }
})

registerAuthSessionStore(() => useAuthStore())

function readStoredUser(): CurrentUser | AuthUser | null {
  const raw = localStorage.getItem(userKey)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as CurrentUser | AuthUser
  } catch {
    localStorage.removeItem(userKey)
    return null
  }
}
