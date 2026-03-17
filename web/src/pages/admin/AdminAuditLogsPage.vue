<template>
  <section>
    <h1>Audit Logs</h1>
    <p v-if="errorMessage">{{ errorMessage }}</p>

    <form @submit.prevent="loadLogs">
      <div>
        <label>
          Action
          <input v-model.trim="filters.action" type="text" />
        </label>
      </div>
      <div>
        <label>
          Resource Type
          <input v-model.trim="filters.resourceType" type="text" />
        </label>
      </div>
      <div>
        <label>
          Result
          <select v-model="filters.result">
            <option value="">all</option>
            <option value="success">success</option>
            <option value="failure">failure</option>
          </select>
        </label>
      </div>
      <button type="submit" :disabled="loading">查询</button>
    </form>

    <hr />

    <p v-if="loading">加载中...</p>
    <ul v-else-if="items.length > 0">
      <li v-for="item in items" :key="item.id">
        {{ item.action }} / {{ item.resource_type }} / {{ item.resource_id || '-' }} /
        {{ item.result }} / {{ item.actor_user_id || '-' }} / {{ item.created_at }}
      </li>
    </ul>
    <p v-else>暂无审计日志</p>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { ApiError, apiRequest } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface AuditLogItem {
  id: string
  action: string
  resource_type: string
  resource_id: string | null
  actor_user_id: string | null
  result: string
  created_at: string
}

const auth = useAuthStore()
const loading = ref(false)
const errorMessage = ref('')
const items = ref<AuditLogItem[]>([])
const filters = reactive({
  action: '',
  resourceType: '',
  result: ''
})

onMounted(async () => {
  await loadLogs()
})

async function loadLogs() {
  loading.value = true
  errorMessage.value = ''

  try {
    const params = new URLSearchParams({
      page: '1',
      page_size: '50'
    })
    if (filters.action) {
      params.set('action', filters.action)
    }
    if (filters.resourceType) {
      params.set('resource_type', filters.resourceType)
    }
    if (filters.result) {
      params.set('result', filters.result)
    }

    const { data } = await apiRequest<AuditLogItem[]>(`/admin/audit-logs?${params.toString()}`, {
      accessToken: auth.accessToken
    })
    items.value = data
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    loading.value = false
  }
}

function toErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    return `${error.code}: ${error.message}`
  }
  return '请求失败'
}
</script>
