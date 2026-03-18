<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>审计日志</h1>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div class="form-card">
      <form @submit.prevent="loadLogs" class="admin-form">
        <div class="form-row">
          <div class="form-group">
            <label>Action</label>
            <input v-model.trim="filters.action" type="text" />
          </div>
          <div class="form-group">
            <label>Resource Type</label>
            <input v-model.trim="filters.resourceType" type="text" />
          </div>
          <div class="form-group">
            <label>Result</label>
            <el-select v-model="filters.result">
              <el-option value="" label="全部" />
              <el-option value="success" label="Success" />
              <el-option value="failure" label="Failure" />
            </el-select>
          </div>
        </div>
        <button type="submit" :disabled="loading" class="submit-btn">查询</button>
      </form>
    </div>

    <div class="table-card">
      <h2>日志列表</h2>
      <p v-if="loading" class="loading">加载中...</p>
      <table v-else-if="items.length > 0">
        <thead>
          <tr>
            <th>Action</th>
            <th>Resource Type</th>
            <th>Resource ID</th>
            <th>Result</th>
            <th>Actor</th>
            <th>时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td>{{ item.action }}</td>
            <td>{{ item.resource_type }}</td>
            <td>{{ item.resource_id || '-' }}</td>
            <td><span class="status-badge" :class="item.result">{{ item.result }}</span></td>
            <td>{{ item.actor_user_id || '-' }}</td>
            <td>{{ formatDate(item.created_at) }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">暂无审计日志</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { ApiError } from '@/lib/api'
import { listAdminAuditLogs } from '@/api/admin'
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

    items.value = await listAdminAuditLogs<AuditLogItem[]>(auth.accessToken, params.toString())
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

function formatDate(date: string) {
  return new Date(date).toLocaleString('zh-CN')
}
</script>

<style scoped>
@import '@/styles/admin.css';
</style>
