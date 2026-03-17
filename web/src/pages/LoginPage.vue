<template>
  <section>
    <h2>登录</h2>
    <form @submit.prevent="submit">
      <div>
        <label for="identifier">邮箱或用户名</label>
        <input id="identifier" v-model="identifier" autocomplete="username" />
      </div>

      <div>
        <label for="password">密码</label>
        <input id="password" v-model="password" type="password" autocomplete="current-password" />
      </div>

      <button type="submit" :disabled="submitting">登录</button>
    </form>

    <p v-if="errorMessage">{{ errorMessage }}</p>
    <RouterLink to="/signup">没有账号？去注册</RouterLink>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import { ApiError, apiRequest } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface AuthSessionResponse {
  access_token: string
  expires_in: number
  user: {
    id: string
    username: string
    email: string
    role: 'user' | 'admin' | 'root'
  }
}

const identifier = ref('')
const password = ref('')
const errorMessage = ref('')
const submitting = ref(false)
const auth = useAuthStore()
const router = useRouter()

async function submit() {
  submitting.value = true
  errorMessage.value = ''

  try {
    const { data } = await apiRequest<AuthSessionResponse>('/auth/signin', {
      method: 'POST',
      body: {
        identifier: identifier.value,
        password: password.value
      }
    })

    auth.setSession(data.access_token, data.user)
    await auth.fetchMe()
    router.push(data.user.role === 'admin' || data.user.role === 'root' ? '/admin/upstreams' : '/chat')
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '登录失败'
  } finally {
    submitting.value = false
  }
}
</script>
