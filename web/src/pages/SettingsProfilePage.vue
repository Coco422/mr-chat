<template>
  <section>
    <h1>Profile Settings</h1>

    <form @submit.prevent="save">
      <div>
        <label for="display-name">Display Name</label>
        <input id="display-name" v-model="displayName" />
      </div>

      <div>
        <label for="avatar-url">Avatar URL</label>
        <input id="avatar-url" v-model="avatarUrl" />
      </div>

      <div>
        <label for="timezone">Timezone</label>
        <input id="timezone" v-model="timezone" />
      </div>

      <div>
        <label for="locale">Locale</label>
        <input id="locale" v-model="locale" />
      </div>

      <button type="submit" :disabled="saving">Save</button>
    </form>

    <p v-if="message">{{ message }}</p>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { ApiError, apiRequest } from '@/lib/api'
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
    const { data } = await apiRequest<CurrentUser>('/users/me', {
      accessToken: auth.accessToken
    })
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
    const { data } = await apiRequest<CurrentUser>('/users/me', {
      method: 'PUT',
      accessToken: auth.accessToken,
      body: {
        display_name: displayName.value,
        avatar_url: avatarUrl.value,
        settings: {
          timezone: timezone.value,
          locale: locale.value
        }
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
