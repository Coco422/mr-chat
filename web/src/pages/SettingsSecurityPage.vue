<template>
  <div class="settings-page">
    <h1>安全设置</h1>

    <div class="info-card" v-if="securityInfo">
      <div class="info-item">
        <span class="info-label">最后登录</span>
        <span class="info-value">{{ securityInfo.last_login_at ? formatDate(securityInfo.last_login_at) : '未知' }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">密码更新</span>
        <span class="info-value">{{ securityInfo.password_updated_at ? formatDate(securityInfo.password_updated_at) : '未设置' }}</span>
      </div>
    </div>

    <form @submit.prevent="updatePassword" class="settings-form">
      <h2>修改密码</h2>

      <div class="form-group">
        <label for="current-password">当前密码</label>
        <input id="current-password" v-model="currentPassword" type="password" autocomplete="current-password" />
      </div>

      <div class="form-group">
        <label for="new-password">新密码</label>
        <input id="new-password" v-model="newPassword" type="password" autocomplete="new-password" />
      </div>

      <button type="submit" :disabled="submitting" class="save-btn">
        {{ submitting ? '更新中...' : '更新密码' }}
      </button>
    </form>

    <p v-if="message" class="message" :class="{ error: message.includes('失败') }">{{ message }}</p>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { ApiError } from '@/lib/api'
import { getSecurityInfo, type SecurityInfoResponse, updateMyPassword } from '@/api/user'
import { useAuthStore } from '@/stores/auth'

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
    securityInfo.value = await getSecurityInfo(auth.accessToken)
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
    await updateMyPassword(auth.accessToken, currentPassword.value, newPassword.value)

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

function formatDate(date: string) {
  return new Date(date).toLocaleString('zh-CN')
}
</script>

<style scoped>
.settings-page {
  padding: 2rem;
  max-width: 600px;
  margin: 0 auto;
}

.settings-page h1 {
  font-size: 1.75rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 2rem;
}

.info-card {
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  border-radius: 12px;
  padding: 1.5rem;
  margin-bottom: 2rem;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--input-border);
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.info-value {
  font-size: 0.9rem;
  color: var(--text-primary);
  font-weight: 500;
}

.settings-form {
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  border-radius: 12px;
  padding: 2rem;
}

.settings-form h2 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 1.5rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group:last-of-type {
  margin-bottom: 2rem;
}

.form-group label {
  display: block;
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
}

.form-group input {
  width: 100%;
  padding: 0.75rem 1rem;
  background: var(--bg-primary);
  border: 1px solid var(--input-border);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.form-group input:focus {
  outline: none;
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px var(--accent-glow);
}

.save-btn {
  width: 100%;
  padding: 0.875rem;
  background: var(--accent-primary);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.save-btn:hover:not(:disabled) {
  background: var(--accent-secondary);
}

.save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.message {
  margin-top: 1rem;
  padding: 1rem;
  border-radius: 8px;
  font-size: 0.9rem;
  background: rgba(94, 184, 229, 0.1);
  color: var(--accent-primary);
  border: 1px solid var(--accent-primary);
}

.message.error {
  background: rgba(239, 68, 68, 0.1);
  color: var(--error-color);
  border-color: var(--error-color);
}
</style>
