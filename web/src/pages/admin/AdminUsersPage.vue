<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>用户管理</h1>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

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
        <div>
          {{ item.username }} / {{ item.email }} / {{ item.role }} / quota={{ item.quota }} / group={{ item.user_group?.name || item.user_group_id || '-' }}
        </div>
        <form @submit.prevent="assignUserGroup(item.id)">
          <label>
            user group
            <select v-model="selectedGroupByUser[item.id]">
              <option value="">未分组</option>
              <option v-for="group in groups" :key="group.id" :value="group.id">
                {{ group.name }}
              </option>
            </select>
          </label>
          <button type="submit" :disabled="assigningUserID === item.id">更新分组</button>
        </form>
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
        <div>
          <label>
            usage model
            <select v-model="usageModelByUser[item.id]">
              <option value="">all models</option>
              <option v-for="model in models" :key="model.id" :value="model.id">
                {{ model.display_name }}
              </option>
            </select>
          </label>
          <button type="button" @click="loadLimitUsage(item.id)" :disabled="loadingUsageUserID === item.id">加载限额用量</button>
          <button type="button" @click="loadAdjustments(item.id)" :disabled="loadingAdjustmentUserID === item.id">加载调整记录</button>
        </div>
        <div v-if="usageReports[item.id]">
          <div>policy source={{ usageReports[item.id]?.effective_policy.source || 'none' }}</div>
          <div>hour requests: used={{ usageReports[item.id]?.usage.hour.requests }} / adj={{ usageReports[item.id]?.adjustments.hour.requests }} / remaining={{ remainingLabel(usageReports[item.id]?.remaining.hour.requests) }}</div>
          <div>hour tokens: used={{ usageReports[item.id]?.usage.hour.tokens }} / adj={{ usageReports[item.id]?.adjustments.hour.tokens }} / remaining={{ remainingLabel(usageReports[item.id]?.remaining.hour.tokens) }}</div>
          <div>week requests: used={{ usageReports[item.id]?.usage.week.requests }} / adj={{ usageReports[item.id]?.adjustments.week.requests }} / remaining={{ remainingLabel(usageReports[item.id]?.remaining.week.requests) }}</div>
          <div>week tokens: used={{ usageReports[item.id]?.usage.week.tokens }} / adj={{ usageReports[item.id]?.adjustments.week.tokens }} / remaining={{ remainingLabel(usageReports[item.id]?.remaining.week.tokens) }}</div>
          <div>lifetime requests: used={{ usageReports[item.id]?.usage.lifetime.requests }} / adj={{ usageReports[item.id]?.adjustments.lifetime.requests }} / remaining={{ remainingLabel(usageReports[item.id]?.remaining.lifetime.requests) }}</div>
          <div>lifetime tokens: used={{ usageReports[item.id]?.usage.lifetime.tokens }} / adj={{ usageReports[item.id]?.adjustments.lifetime.tokens }} / remaining={{ remainingLabel(usageReports[item.id]?.remaining.lifetime.tokens) }}</div>
        </div>
        <form @submit.prevent="createAdjustment(item.id)">
          <label>
            adjust model
            <select v-model="adjustmentModelByUser[item.id]">
              <option value="">all models</option>
              <option v-for="model in models" :key="model.id" :value="model.id">
                {{ model.display_name }}
              </option>
            </select>
          </label>
          <label>
            metric
            <select v-model="adjustmentMetricByUser[item.id]">
              <option value="request_count">request_count</option>
              <option value="total_tokens">total_tokens</option>
            </select>
          </label>
          <label>
            window
            <select v-model="adjustmentWindowByUser[item.id]">
              <option value="rolling_hour">rolling_hour</option>
              <option value="rolling_week">rolling_week</option>
              <option value="lifetime">lifetime</option>
            </select>
          </label>
          <label>
            delta
            <input v-model.trim="adjustmentDeltaByUser[item.id]" type="number" />
          </label>
          <label>
            reason
            <input v-model.trim="adjustmentReasonByUser[item.id]" type="text" />
          </label>
          <button type="submit" :disabled="submittingAdjustmentUserID === item.id">新增限额调整</button>
        </form>
        <ul v-if="(adjustmentsByUser[item.id] ?? []).length > 0">
          <li v-for="adjustment in adjustmentsByUser[item.id]" :key="adjustment.id">
            {{ adjustment.metric_type }} / {{ adjustment.window_type }} / delta={{ adjustment.delta }} / model={{ adjustment.model_id || 'all' }} / expires_at={{ adjustment.expires_at || '-' }}
          </li>
        </ul>
      </li>
    </ul>
    <p v-else>暂无用户</p>
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
  display_name: string
  role: string
  status: string
  quota: number
  used_quota: number
  user_group_id: string | null
  user_group: UserGroupItem | null
}

interface UserGroupItem {
  id: string
  name: string
  status: string
}

interface ModelItem {
  id: string
  model_key: string
  display_name: string
}

interface UsageCounter {
  requests: number
  tokens: number
}

interface UsageCounters {
  hour: UsageCounter
  week: UsageCounter
  lifetime: UsageCounter
}

interface UsageReport {
  user_id: string
  user_group_id: string | null
  model_id: string | null
  effective_policy: {
    source: string
  }
  usage: UsageCounters
  adjustments: UsageCounters
  remaining: UsageCounters
}

interface UserLimitAdjustment {
  id: string
  model_id: string | null
  metric_type: string
  window_type: string
  delta: number
  expires_at: string | null
}

const auth = useAuthStore()
const loading = ref(false)
const submittingUserID = ref('')
const assigningUserID = ref('')
const loadingUsageUserID = ref('')
const loadingAdjustmentUserID = ref('')
const submittingAdjustmentUserID = ref('')
const errorMessage = ref('')
const items = ref<AdminUser[]>([])
const groups = ref<UserGroupItem[]>([])
const models = ref<ModelItem[]>([])
const filters = reactive({
  keyword: '',
  status: ''
})
const quotaDelta = reactive<Record<string, string>>({})
const quotaReason = reactive<Record<string, string>>({})
const selectedGroupByUser = reactive<Record<string, string>>({})
const usageModelByUser = reactive<Record<string, string>>({})
const usageReports = reactive<Record<string, UsageReport | undefined>>({})
const adjustmentsByUser = reactive<Record<string, UserLimitAdjustment[]>>({})
const adjustmentModelByUser = reactive<Record<string, string>>({})
const adjustmentMetricByUser = reactive<Record<string, string>>({})
const adjustmentWindowByUser = reactive<Record<string, string>>({})
const adjustmentDeltaByUser = reactive<Record<string, string>>({})
const adjustmentReasonByUser = reactive<Record<string, string>>({})

onMounted(async () => {
  await reloadAll()
})

async function reloadAll() {
  await Promise.all([loadUsers(), loadReferenceData()])
}

async function loadReferenceData() {
  try {
    const [groupResponse, modelResponse] = await Promise.all([
      apiRequest<UserGroupItem[]>('/admin/user-groups', {
        accessToken: auth.accessToken
      }),
      apiRequest<ModelItem[]>('/admin/models', {
        accessToken: auth.accessToken
      })
    ])

    groups.value = groupResponse.data
    models.value = modelResponse.data
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  }
}

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
    for (const item of data) {
      selectedGroupByUser[item.id] = item.user_group_id ?? ''
      usageModelByUser[item.id] = usageModelByUser[item.id] ?? ''
      adjustmentModelByUser[item.id] = adjustmentModelByUser[item.id] ?? ''
      adjustmentMetricByUser[item.id] = adjustmentMetricByUser[item.id] ?? 'request_count'
      adjustmentWindowByUser[item.id] = adjustmentWindowByUser[item.id] ?? 'rolling_hour'
      adjustmentDeltaByUser[item.id] = adjustmentDeltaByUser[item.id] ?? ''
      adjustmentReasonByUser[item.id] = adjustmentReasonByUser[item.id] ?? ''
    }
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function assignUserGroup(userID: string) {
  assigningUserID.value = userID
  errorMessage.value = ''

  try {
    await apiRequest(`/admin/users/${userID}/group`, {
      method: 'PUT',
      accessToken: auth.accessToken,
      body: {
        user_group_id: selectedGroupByUser[userID] || null
      }
    })

    await loadUsers()
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    assigningUserID.value = ''
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

async function loadLimitUsage(userID: string) {
  loadingUsageUserID.value = userID
  errorMessage.value = ''

  try {
    const params = new URLSearchParams()
    if (usageModelByUser[userID]) {
      params.set('model_id', usageModelByUser[userID])
    }

    const suffix = params.size > 0 ? `?${params.toString()}` : ''
    const { data } = await apiRequest<UsageReport>(`/admin/users/${userID}/limit-usage${suffix}`, {
      accessToken: auth.accessToken
    })
    usageReports[userID] = data
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    loadingUsageUserID.value = ''
  }
}

async function loadAdjustments(userID: string) {
  loadingAdjustmentUserID.value = userID
  errorMessage.value = ''

  try {
    const params = new URLSearchParams({
      page: '1',
      page_size: '20'
    })
    if (usageModelByUser[userID]) {
      params.set('model_id', usageModelByUser[userID])
    }

    const { data } = await apiRequest<UserLimitAdjustment[]>(`/admin/users/${userID}/limit-adjustments?${params.toString()}`, {
      accessToken: auth.accessToken
    })
    adjustmentsByUser[userID] = data
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    loadingAdjustmentUserID.value = ''
  }
}

async function createAdjustment(userID: string) {
  submittingAdjustmentUserID.value = userID
  errorMessage.value = ''

  try {
    await apiRequest(`/admin/users/${userID}/limit-adjustments`, {
      method: 'POST',
      accessToken: auth.accessToken,
      body: {
        model_id: adjustmentModelByUser[userID] || null,
        metric_type: adjustmentMetricByUser[userID] || 'request_count',
        window_type: adjustmentWindowByUser[userID] || 'rolling_hour',
        delta: Number(adjustmentDeltaByUser[userID] || 0),
        reason: adjustmentReasonByUser[userID] || null
      }
    })

    adjustmentDeltaByUser[userID] = ''
    adjustmentReasonByUser[userID] = ''
    await Promise.all([loadLimitUsage(userID), loadAdjustments(userID)])
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    submittingAdjustmentUserID.value = ''
  }
}

function remainingLabel(value: number | undefined) {
  if (value == null) {
    return '-'
  }
  if (value < 0) {
    return 'unlimited'
  }
  return String(value)
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
