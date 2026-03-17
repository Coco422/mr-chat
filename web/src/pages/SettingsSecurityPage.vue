<template>
  <section>
    <h1>Security Settings</h1>

    <div>
      <h2>Security Info</h2>
      <pre>{{ securityInfo }}</pre>
    </div>

    <form @submit.prevent="updatePassword">
      <div>
        <label for="current-password">Current Password</label>
        <input id="current-password" v-model="currentPassword" type="password" />
      </div>

      <div>
        <label for="new-password">New Password</label>
        <input id="new-password" v-model="newPassword" type="password" />
      </div>

      <button type="submit" :disabled="submitting">Update Password</button>
    </form>

    <p v-if="message">{{ message }}</p>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { ApiError, apiRequest } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface SecurityInfoResponse {
  last_login_at: string | null
  password_updated_at: string | null
  has_password: boolean
}

const auth = useAuthStore()
const securityInfo = ref<SecurityInfoResponse | null>(null)
const currentPassword = ref('')
const newPassword = ref('')
const submitting = ref(false)
const message = ref('')

onMounted(() => {
  void loadSecurityInfo()
})

async function loadSecurityInfo() {
  if (!auth.accessToken) {
    return
  }

  try {
    const { data } = await apiRequest<SecurityInfoResponse>('/users/me/security', {
      accessToken: auth.accessToken
    })
    securityInfo.value = data
  } catch (error) {
    message.value = error instanceof ApiError ? error.message : '加载安全信息失败'
  }
}

async function updatePassword() {
  if (!auth.accessToken) {
    return
  }

  submitting.value = true
  message.value = ''

  try {
    await apiRequest('/users/me/password', {
      method: 'PUT',
      accessToken: auth.accessToken,
      body: {
        current_password: currentPassword.value,
        new_password: newPassword.value
      }
    })

    currentPassword.value = ''
    newPassword.value = ''
    message.value = '密码已更新'
    await loadSecurityInfo()
  } catch (error) {
    message.value = error instanceof ApiError ? error.message : '更新密码失败'
  } finally {
    submitting.value = false
  }
}
</script>
