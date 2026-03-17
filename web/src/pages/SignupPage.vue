<template>
  <section>
    <h2>注册</h2>
    <form @submit.prevent="submit">
      <div>
        <label for="username">用户名</label>
        <input id="username" v-model="username" autocomplete="username" />
      </div>

      <div>
        <label for="email">邮箱</label>
        <input id="email" v-model="email" autocomplete="email" />
      </div>

      <div>
        <label for="password">密码</label>
        <input id="password" v-model="password" type="password" autocomplete="new-password" />
      </div>

      <button type="submit" :disabled="submitting">注册</button>
    </form>

    <p v-if="errorMessage">{{ errorMessage }}</p>
    <RouterLink to="/login">已有账号？去登录</RouterLink>
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

const username = ref('')
const email = ref('')
const password = ref('')
const errorMessage = ref('')
const submitting = ref(false)
const auth = useAuthStore()
const router = useRouter()

async function submit() {
  submitting.value = true
  errorMessage.value = ''

  try {
    const { data } = await apiRequest<AuthSessionResponse>('/auth/signup', {
      method: 'POST',
      body: {
        username: username.value,
        email: email.value,
        password: password.value
      }
    })

    auth.setSession(data.access_token, data.user)
    await auth.fetchMe()
    router.push('/chat')
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '注册失败'
  } finally {
    submitting.value = false
  }
}
</script>
