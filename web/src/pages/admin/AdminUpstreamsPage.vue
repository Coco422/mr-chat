<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>上游管理</h1>
      <button class="primary-btn" @click="showForm = !showForm">
        {{ showForm ? '取消' : '+ 新建上游' }}
      </button>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div v-if="showForm" class="form-card">
      <h2>创建上游</h2>
      <form @submit.prevent="createUpstream" class="admin-form">
        <div class="form-row">
          <div class="form-group">
            <label>名称</label>
            <input v-model.trim="form.name" type="text" required />
          </div>
          <div class="form-group">
            <label>Provider Type</label>
            <input v-model.trim="form.providerType" type="text" />
          </div>
        </div>

        <div class="form-group">
          <label>Base URL</label>
          <input v-model.trim="form.baseURL" type="text" required />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Auth Type</label>
            <input v-model.trim="form.authType" type="text" />
          </div>
          <div class="form-group">
            <label>API Key</label>
            <input v-model.trim="form.apiKey" type="text" />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Status</label>
            <el-select v-model="form.status">
              <el-option value="active" label="Active" />
              <el-option value="disabled" label="Disabled" />
              <el-option value="maintenance" label="Maintenance" />
            </el-select>
          </div>
          <div class="form-group">
            <label>Timeout Seconds</label>
            <input v-model.number="form.timeoutSeconds" type="number" min="1" />
          </div>
          <div class="form-group">
            <label>Cooldown Seconds</label>
            <input v-model.number="form.cooldownSeconds" type="number" min="1" />
          </div>
          <div class="form-group">
            <label>Failure Threshold</label>
            <input v-model.number="form.failureThreshold" type="number" min="1" />
          </div>
        </div>

        <button type="submit" :disabled="submitting" class="submit-btn">创建上游</button>
      </form>
    </div>

    <div class="table-card">
      <div class="table-header">
        <h2>上游列表</h2>
        <button class="refresh-btn" @click="loadUpstreams" :disabled="loading">刷新</button>
      </div>

      <p v-if="loading" class="loading">加载中...</p>
      <table v-else-if="items.length > 0">
        <thead>
          <tr>
            <th>名称</th>
            <th>Provider</th>
            <th>状态</th>
            <th>Base URL</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td>{{ item.name }}</td>
            <td>{{ item.provider_type }}</td>
            <td><span class="status-badge" :class="item.status">{{ item.status }}</span></td>
            <td class="url">{{ item.base_url }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">暂无上游配置</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { ApiError, apiRequest } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface UpstreamItem {
  id: string
  name: string
  provider_type: string
  base_url: string
  status: string
}

const auth = useAuthStore()
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const showForm = ref(false)
const items = ref<UpstreamItem[]>([])
const form = reactive({
  name: '',
  providerType: 'openai_compatible',
  baseURL: '',
  authType: 'bearer',
  apiKey: '',
  status: 'active',
  timeoutSeconds: 60,
  cooldownSeconds: 60,
  failureThreshold: 3
})

onMounted(async () => {
  await loadUpstreams()
})

async function loadUpstreams() {
  loading.value = true
  errorMessage.value = ''

  try {
    const { data } = await apiRequest<UpstreamItem[]>('/admin/upstreams', {
      accessToken: auth.accessToken
    })
    items.value = data
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function createUpstream() {
  submitting.value = true
  errorMessage.value = ''

  try {
    await apiRequest('/admin/upstreams', {
      method: 'POST',
      accessToken: auth.accessToken,
      body: {
        name: form.name,
        provider_type: form.providerType,
        base_url: form.baseURL,
        auth_type: form.authType,
        auth_config: form.apiKey ? { api_key: form.apiKey } : {},
        status: form.status,
        timeout_seconds: form.timeoutSeconds,
        cooldown_seconds: form.cooldownSeconds,
        failure_threshold: form.failureThreshold,
        metadata: {}
      }
    })

    form.name = ''
    form.baseURL = ''
    form.apiKey = ''
    showForm.value = false
    await loadUpstreams()
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    submitting.value = false
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
.admin-page {
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 1.75rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.primary-btn {
  padding: 0.75rem 1.5rem;
  background: var(--accent-primary);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.primary-btn:hover {
  background: var(--accent-secondary);
}

.error {
  padding: 1rem;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid var(--error-color);
  border-radius: 8px;
  color: var(--error-color);
  margin-bottom: 1.5rem;
}

.form-card, .table-card {
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  border-radius: 12px;
  padding: 1.5rem;
  margin-bottom: 2rem;
}

.form-card h2, .table-header h2 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 1.5rem;
}

.admin-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
}

.form-group input, .form-group select {
  padding: 0.75rem;
  background: var(--bg-primary);
  border: 1px solid var(--input-border);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 0.9rem;
}

.form-group input:focus, .form-group select:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.submit-btn {
  padding: 0.875rem;
  background: var(--accent-primary);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  margin-top: 0.5rem;
}

.submit-btn:hover:not(:disabled) {
  background: var(--accent-secondary);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.table-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
}

.refresh-btn {
  padding: 0.5rem 1rem;
  background: transparent;
  border: 1px solid var(--input-border);
  border-radius: 8px;
  color: var(--text-secondary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.refresh-btn:hover:not(:disabled) {
  background: var(--accent-primary);
  color: white;
  border-color: var(--accent-primary);
}

table {
  width: 100%;
  border-collapse: collapse;
}

th {
  text-align: left;
  padding: 0.75rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--input-border);
}

td {
  padding: 0.875rem 0.75rem;
  font-size: 0.9rem;
  color: var(--text-primary);
  border-bottom: 1px solid var(--input-border);
}

tr:last-child td {
  border-bottom: none;
}

.url {
  font-family: monospace;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.status-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-badge.active {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
}

.status-badge.disabled {
  background: rgba(239, 68, 68, 0.2);
  color: var(--error-color);
}

.status-badge.maintenance {
  background: rgba(245, 158, 11, 0.2);
  color: #f59e0b;
}

.loading, .empty {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}
</style>
