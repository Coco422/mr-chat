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
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="action" label="Action" min-width="150" />
        <el-table-column prop="resource_type" label="Resource Type" min-width="150" />
        <el-table-column prop="resource_id" label="Resource ID" min-width="200">
          <template #default="{ row }">{{ row.resource_id || '-' }}</template>
        </el-table-column>
        <el-table-column prop="result" label="Result" width="120">
          <template #default="{ row }">
            <span class="status-badge" :class="row.result">{{ row.result }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="actor_user_id" label="Actor" min-width="200">
          <template #default="{ row }">{{ row.actor_user_id || '-' }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无审计日志" />
        </template>
      </el-table>

      <el-pagination
        v-if="total > 0"
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="loadLogs"
        @size-change="loadLogs"
        style="margin-top: 16px; justify-content: flex-end"
      />
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
const total = ref(0)
const pagination = reactive({
  page: 1,
  pageSize: 50
})
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
      page: String(pagination.page),
      page_size: String(pagination.pageSize)
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

    const data = await listAdminAuditLogs<AuditLogItem[]>(auth.accessToken, params.toString())
    items.value = data
    total.value = data.length
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
