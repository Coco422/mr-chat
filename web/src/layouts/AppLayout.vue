<template>
  <div>
    <header>
      <h1>MrChat</h1>
      <p v-if="auth.user">
        当前用户：{{ auth.user.username }} / {{ auth.user.role }}
      </p>
      <button type="button" @click="handleSignOut">Sign out</button>
    </header>

    <nav>
      <RouterLink to="/chat">Chat</RouterLink>
      <span> | </span>
      <RouterLink to="/usage">Usage</RouterLink>
      <span> | </span>
      <RouterLink to="/settings/profile">Profile</RouterLink>
      <span> | </span>
      <RouterLink to="/settings/security">Security</RouterLink>
      <template v-if="auth.isAdmin">
        <span> | </span>
        <RouterLink to="/admin/upstreams">Admin</RouterLink>
      </template>
    </nav>

    <RouterView />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

onMounted(() => {
  if (auth.isAuthenticated && !auth.user) {
    void auth.fetchMe()
  }
})

async function handleSignOut() {
  await auth.signOut()
  router.push({ name: 'login' })
}
</script>
