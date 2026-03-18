<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>用户管理</h1>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div class="form-card">
      <form @submit.prevent="loadUsers" class="admin-form">
        <div class="form-row">
          <div class="form-group">
            <label>关键词</label>
            <input v-model.trim="filters.keyword" type="text" placeholder="用户名或邮箱" />
          </div>
          <div class="form-group">
            <label>状态</label>
            <el-select v-model="filters.status">
              <el-option value="" label="全部" />
              <el-option value="active" label="Active" />
              <el-option value="disabled" label="Disabled" />
              <el-option value="pending" label="Pending" />
            </el-select>
          </div>
        </div>
        <button type="submit" :disabled="loading" class="submit-btn">查询</button>
      </form>
    </div>

    <div class="table-card">
      <h2>用户列表</h2>
      <p v-if="loading" class="loading">加载中...</p>
      <table v-else-if="items.length > 0">
        <thead>
          <tr>
            <th>用户名</th>
            <th>邮箱</th>
            <th>角色</th>
            <th>额度</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td>{{ item.username }}</td>
            <td>{{ item.email }}</td>
            <td>{{ item.role }}</td>
            <td>{{ item.quota }}</td>
            <td>
              <form @submit.prevent="adjustQuota(item.id)" class="inline-form">
                <input v-model.trim="quotaDelta[item.id]" type="number" placeholder="变动" class="small-input" />
                <input v-model.trim="quotaReason[item.id]" type="text" placeholder="原因" class="small-input" />
                <button type="submit" :disabled="submittingUserID === item.id" class="small-btn">调额</button>
              </form>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">暂无用户</p>
    </div>
  </div>
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

<style scoped>
@import '@/styles/admin.css';

.inline-form {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.small-input {
  padding: 0.5rem;
  background: var(--bg-primary);
  border: 1px solid var(--input-border);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 0.85rem;
  width: 80px;
}

.small-input:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.small-btn {
  padding: 0.5rem 0.75rem;
  background: var(--accent-primary);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.small-btn:hover:not(:disabled) {
  background: var(--accent-secondary);
}

.small-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
