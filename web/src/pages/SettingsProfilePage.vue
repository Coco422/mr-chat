<template>
  <div class="settings-page">
    <h1>个人资料</h1>

    <form @submit.prevent="save" class="settings-form">
      <div class="form-group">
        <label for="display-name">显示名称</label>
        <input id="display-name" v-model="displayName" />
      </div>

      <div class="form-group">
        <label for="avatar-url">头像 URL</label>
        <input id="avatar-url" v-model="avatarUrl" />
      </div>

      <div class="form-group">
        <label for="timezone">时区</label>
        <input id="timezone" v-model="timezone" />
      </div>

      <div class="form-group">
        <label for="locale">语言</label>
        <input id="locale" v-model="locale" />
      </div>

      <button type="submit" :disabled="saving" class="save-btn">
        {{ saving ? '保存中...' : '保存' }}
      </button>
    </form>

    <p v-if="message" class="message" :class="{ error: message.includes('失败') }">{{ message }}</p>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { ApiError } from '@/lib/api'
import { getCurrentUser, updateCurrentUser } from '@/api/user'
import { useAuthStore, type CurrentUser } from '@/stores/auth'

const auth = useAuthStore()
const displayName = ref('')
const avatarUrl = ref('')
const timezone = ref('Asia/Shanghai')
const locale = ref('zh-CN')
const saving = ref(false)
const message = ref('')

onMounted(() => {
  void loadProfile()
})

async function loadProfile() {
  if (!auth.accessToken) {
    return
  }

  try {
    const data = await getCurrentUser<CurrentUser>(auth.accessToken)
    auth.setSession(auth.accessToken, data)
    displayName.value = data.display_name
    avatarUrl.value = data.avatar_url ?? ''
    timezone.value = data.settings?.timezone ?? 'Asia/Shanghai'
    locale.value = data.settings?.locale ?? 'zh-CN'
  } catch (error) {
    message.value = error instanceof ApiError ? error.message : '加载 profile 失败'
  }
}

async function save() {
  if (!auth.accessToken) {
    return
  }

  saving.value = true
  message.value = ''

  try {
    const data = await updateCurrentUser<CurrentUser>(auth.accessToken, {
      display_name: displayName.value,
      avatar_url: avatarUrl.value,
      settings: {
        timezone: timezone.value,
        locale: locale.value
      }
    })

    auth.setSession(auth.accessToken, data)
    message.value = '保存成功'
  } catch (error) {
    message.value = error instanceof ApiError ? error.message : '保存失败'
  } finally {
    saving.value = false
  }
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

.settings-form {
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  border-radius: 12px;
  padding: 2rem;
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
