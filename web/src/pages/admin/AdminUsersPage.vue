<template>
  <section>
    <h1>Users</h1>
    <p v-if="errorMessage">{{ errorMessage }}</p>

    <form @submit.prevent="loadUsers">
      <div>
        <label>
          Keyword
          <input v-model.trim="filters.keyword" type="text" />
        </label>
      </div>
      <div>
        <label>
          状态
          <select v-model="filters.status">
            <option value="">all</option>
            <option value="active">active</option>
            <option value="disabled">disabled</option>
            <option value="pending">pending</option>
          </select>
        </label>
      </div>
      <button type="submit" :disabled="loading">查询</button>
    </form>

    <hr />

    <p v-if="loading">加载中...</p>
    <ul v-else-if="items.length > 0">
      <li v-for="item in items" :key="item.id">
        <div>{{ item.username }} / {{ item.email }} / {{ item.role }} / quota={{ item.quota }}</div>
        <form @submit.prevent="adjustQuota(item.id)">
          <label>
            delta
            <input v-model.trim="quotaDelta[item.id]" type="number" />
          </label>
          <label>
            reason
            <input v-model.trim="quotaReason[item.id]" type="text" />
          </label>
          <button type="submit" :disabled="submittingUserID === item.id">调额</button>
        </form>
      </li>
    </ul>
    <p v-else>暂无用户</p>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { ApiError, apiRequest } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface AdminUser {
  id: string
  username: string
  email: string
  role: string
  quota: number
}

const auth = useAuthStore()
const loading = ref(false)
const submittingUserID = ref('')
const errorMessage = ref('')
const items = ref<AdminUser[]>([])
const filters = reactive({
  keyword: '',
  status: ''
})
const quotaDelta = reactive<Record<string, string>>({})
const quotaReason = reactive<Record<string, string>>({})

onMounted(async () => {
  await loadUsers()
})

async function loadUsers() {
  loading.value = true
  errorMessage.value = ''

  try {
    const params = new URLSearchParams({
      page: '1',
      page_size: '50'
    })
    if (filters.keyword) {
      params.set('keyword', filters.keyword)
    }
    if (filters.status) {
      params.set('status', filters.status)
    }

    const { data } = await apiRequest<AdminUser[]>(`/admin/users?${params.toString()}`, {
      accessToken: auth.accessToken
    })
    items.value = data
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function adjustQuota(userID: string) {
  submittingUserID.value = userID
  errorMessage.value = ''

  try {
    await apiRequest(`/admin/users/${userID}/quota`, {
      method: 'PUT',
      accessToken: auth.accessToken,
      body: {
        delta: Number(quotaDelta[userID] || 0),
        reason: quotaReason[userID] || ''
      }
    })

    quotaDelta[userID] = ''
    quotaReason[userID] = ''
    await loadUsers()
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    submittingUserID.value = ''
  }
}

function toErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    return `${error.code}: ${error.message}`
  }
  return '请求失败'
}
</script>
