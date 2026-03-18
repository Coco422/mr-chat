<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>用户管理</h1>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div class="form-card">
      <h2>筛选条件</h2>
      <form @submit.prevent="loadUsers" class="admin-form">
        <div class="form-row">
          <div class="form-group">
            <label>Keyword</label>
            <input v-model.trim="filters.keyword" type="text" />
          </div>
          <div class="form-group">
            <label>Status</label>
            <el-select v-model="filters.status" style="width: 100%">
              <el-option value="" label="all" />
              <el-option value="active" label="active" />
              <el-option value="disabled" label="disabled" />
              <el-option value="pending" label="pending" />
            </el-select>
          </div>
        </div>
        <button type="submit" :disabled="loading" class="submit-btn">查询</button>
      </form>
    </div>

    <div class="table-card">
      <h2>用户列表</h2>
      <p v-if="loading" class="loading">加载中...</p>
      <div v-else-if="items.length > 0" class="user-list">
        <article v-for="item in items" :key="item.id" class="user-card">
          <div class="user-info">
            <div class="user-header">
              <span class="username">{{ item.username }}</span>
              <span class="status-badge" :class="item.status">{{ item.status }}</span>
            </div>
            <div class="user-meta">{{ item.email }}</div>
            <div class="user-meta">
              role={{ item.role }} | quota={{ item.quota }} | used={{ item.used_quota }} | group={{ item.user_group?.name || item.user_group_id || '-' }}
            </div>
          </div>

          <div class="user-actions">
            <form @submit.prevent="assignUserGroup(item.id)" class="inline-form">
              <label>用户组</label>
              <el-select v-model="selectedGroupByUser[item.id]" class="small-select">
                <el-option value="" label="未分组" />
                <el-option v-for="group in groups" :key="group.id" :value="group.id" :label="group.name" />
              </el-select>
              <button type="submit" :disabled="assigningUserID === item.id" class="small-btn">更新</button>
            </form>

            <form @submit.prevent="adjustQuota(item.id)" class="inline-form">
              <label>调额</label>
              <input v-model.trim="quotaDelta[item.id]" type="number" placeholder="delta" class="small-input" />
              <input v-model.trim="quotaReason[item.id]" type="text" placeholder="reason" class="reason-input" />
              <button type="submit" :disabled="submittingUserID === item.id" class="small-btn">调额</button>
            </form>

            <div class="inline-form">
              <label>限额模型</label>
              <el-select v-model="usageModelByUser[item.id]" class="small-select">
                <el-option value="" label="all models" />
                <el-option v-for="model in models" :key="model.id" :value="model.id" :label="model.display_name" />
              </el-select>
              <button type="button" @click="loadLimitUsage(item.id)" :disabled="loadingUsageUserID === item.id" class="small-btn">加载限额</button>
              <button type="button" @click="loadAdjustments(item.id)" :disabled="loadingAdjustmentUserID === item.id" class="small-btn">调整记录</button>
            </div>
          </div>

          <div v-if="usageReports[item.id]" class="usage-report">
            <div class="report-header">限额用量报告 (来源: {{ usageReports[item.id]?.effective_policy.source || 'none' }})</div>
            <div class="report-grid">
              <div class="report-item">
                <span class="label">Hour Requests:</span>
                <span>{{ usageReports[item.id]?.usage.hour.requests }} / adj={{ usageReports[item.id]?.adjustments.hour.requests }} / remain={{ remainingLabel(usageReports[item.id]?.remaining.hour.requests) }}</span>
              </div>
              <div class="report-item">
                <span class="label">Hour Tokens:</span>
                <span>{{ usageReports[item.id]?.usage.hour.tokens }} / adj={{ usageReports[item.id]?.adjustments.hour.tokens }} / remain={{ remainingLabel(usageReports[item.id]?.remaining.hour.tokens) }}</span>
              </div>
              <div class="report-item">
                <span class="label">Week Requests:</span>
                <span>{{ usageReports[item.id]?.usage.week.requests }} / adj={{ usageReports[item.id]?.adjustments.week.requests }} / remain={{ remainingLabel(usageReports[item.id]?.remaining.week.requests) }}</span>
              </div>
              <div class="report-item">
                <span class="label">Week Tokens:</span>
                <span>{{ usageReports[item.id]?.usage.week.tokens }} / adj={{ usageReports[item.id]?.adjustments.week.tokens }} / remain={{ remainingLabel(usageReports[item.id]?.remaining.week.tokens) }}</span>
              </div>
              <div class="report-item">
                <span class="label">Lifetime Requests:</span>
                <span>{{ usageReports[item.id]?.usage.lifetime.requests }} / adj={{ usageReports[item.id]?.adjustments.lifetime.requests }} / remain={{ remainingLabel(usageReports[item.id]?.remaining.lifetime.requests) }}</span>
              </div>
              <div class="report-item">
                <span class="label">Lifetime Tokens:</span>
                <span>{{ usageReports[item.id]?.usage.lifetime.tokens }} / adj={{ usageReports[item.id]?.adjustments.lifetime.tokens }} / remain={{ remainingLabel(usageReports[item.id]?.remaining.lifetime.tokens) }}</span>
              </div>
            </div>
          </div>

          <form @submit.prevent="createAdjustment(item.id)" class="adjustment-form">
            <label>
              adjust model
              <el-select v-model="adjustmentModelByUser[item.id]">
                <el-option value="" label="all models" />
                <el-option v-for="model in models" :key="model.id" :value="model.id" :label="model.display_name" />
              </el-select>
            </label>
            <label>
              metric
              <el-select v-model="adjustmentMetricByUser[item.id]">
                <el-option value="request_count" label="request_count" />
                <el-option value="total_tokens" label="total_tokens" />
              </el-select>
            </label>
            <label>
              window
              <el-select v-model="adjustmentWindowByUser[item.id]">
                <el-option value="rolling_hour" label="rolling_hour" />
                <el-option value="rolling_week" label="rolling_week" />
                <el-option value="lifetime" label="lifetime" />
              </el-select>
            </label>
            <label>
              delta
              <input v-model.trim="adjustmentDeltaByUser[item.id]" type="number" />
            </label>
            <label>
              reason
              <input v-model.trim="adjustmentReasonByUser[item.id]" type="text" />
            </label>
            <button type="submit" :disabled="submittingAdjustmentUserID === item.id" class="small-btn">新增限额调整</button>
          </form>

          <ul v-if="(adjustmentsByUser[item.id] ?? []).length > 0" class="adjustment-list">
            <li v-for="adjustment in adjustmentsByUser[item.id]" :key="adjustment.id">
              {{ adjustment.metric_type }} / {{ adjustment.window_type }} / delta={{ adjustment.delta }} / model={{ adjustment.model_id || 'all' }} / expires_at={{ adjustment.expires_at || '-' }}
            </li>
          </ul>
        </article>
      </div>
      <p v-else class="empty">暂无用户</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { ApiError } from '@/lib/api'
import {
  createAdminUserLimitAdjustment,
  getAdminUserLimitUsage,
  listAdminModels,
  listAdminUserGroups,
  listAdminUserLimitAdjustments,
  listAdminUsers,
  updateAdminUserGroup,
  updateAdminUserQuota
} from '@/api/admin'
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
      listAdminUserGroups<UserGroupItem[]>(auth.accessToken),
      listAdminModels<ModelItem[]>(auth.accessToken)
    ])

    groups.value = groupResponse
    models.value = modelResponse
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

    const data = await listAdminUsers<AdminUser[]>(auth.accessToken, params.toString())
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
    await updateAdminUserGroup(auth.accessToken, userID, selectedGroupByUser[userID] || null)

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
    await updateAdminUserQuota(
      auth.accessToken,
      userID,
      Number(quotaDelta[userID] || 0),
      quotaReason[userID] || ''
    )

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
    usageReports[userID] = await getAdminUserLimitUsage<UsageReport>(
      auth.accessToken,
      userID,
      usageModelByUser[userID] || undefined
    )
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

    adjustmentsByUser[userID] = await listAdminUserLimitAdjustments<UserLimitAdjustment[]>(
      auth.accessToken,
      userID,
      params.toString()
    )
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
    await createAdminUserLimitAdjustment(auth.accessToken, userID, {
      model_id: adjustmentModelByUser[userID] || null,
      metric_type: adjustmentMetricByUser[userID] || 'request_count',
      window_type: adjustmentWindowByUser[userID] || 'rolling_hour',
      delta: Number(adjustmentDeltaByUser[userID] || 0),
      reason: adjustmentReasonByUser[userID] || null
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
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}

.user-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.user-card {
  border: 1px solid var(--input-border);
  border-radius: 12px;
  padding: 1rem;
  background: var(--bg-primary);
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.user-header {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.username {
  color: var(--text-primary);
  font-size: 1rem;
  font-weight: 600;
}

.user-meta {
  color: var(--text-secondary);
  font-size: 0.86rem;
  word-break: break-all;
}

.user-actions {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
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

.small-select {
  min-width: 180px;
}

.reason-input {
  padding: 0.5rem;
  background: var(--bg-primary);
  border: 1px solid var(--input-border);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 0.85rem;
  min-width: 180px;
}

.small-input:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.reason-input:focus {
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

.usage-report {
  border: 1px dashed var(--input-border);
  border-radius: 10px;
  padding: 0.75rem;
  background: var(--input-bg);
}

.report-header {
  color: var(--text-primary);
  font-size: 0.86rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.report-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 0.5rem;
}

.report-item {
  display: flex;
  gap: 0.4rem;
  align-items: baseline;
  font-size: 0.82rem;
  color: var(--text-secondary);
}

.report-item .label {
  color: var(--text-primary);
  font-weight: 500;
}

.adjustment-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.6rem;
  align-items: end;
}

.adjustment-form label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  color: var(--text-secondary);
  font-size: 0.82rem;
}

.adjustment-form :deep(.el-select) {
  width: 100%;
}

.adjustment-form input {
  width: 100%;
  padding: 0.5rem;
  background: var(--bg-primary);
  border: 1px solid var(--input-border);
  border-radius: 6px;
  color: var(--text-primary);
}

.adjustment-form input:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.adjustment-list {
  margin: 0;
  padding-left: 1.1rem;
  color: var(--text-secondary);
  font-size: 0.82rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
</style>
