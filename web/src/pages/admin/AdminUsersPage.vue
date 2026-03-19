<template>
  <div class="admin-page">
    <!-- <div class="page-header">
      <h1>用户管理</h1>
    </div> -->

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div class="form-card">
      <!-- <h2>筛选条件</h2> -->
      <form @submit.prevent="loadUsers" class="admin-form filter-form">
        <div class="form-row filter-row">
          <div class="form-group filter-group">
            <label class="filter-label">Keyword</label>
            <input v-model.trim="filters.keyword" type="text" class="filter-control" />
          </div>
          <div class="form-group filter-group">
            <label class="filter-label">Status</label>
            <el-select v-model="filters.status" class="filter-control">
              <el-option value="" label="all" />
              <el-option value="active" label="active" />
              <el-option value="disabled" label="disabled" />
              <el-option value="pending" label="pending" />
            </el-select>
          </div>
          <div class="filter-actions">
            <button type="submit" :disabled="loading" class="submit-btn">查询</button>
          </div>
        </div>

      </form>
    </div>

    <div class="table-card">
      <!-- <h2>用户列表</h2> -->
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="expand-content">
              <div class="action-section">
                <h3>用户组管理</h3>
                <form @submit.prevent="assignUserGroup(row.id)" class="inline-form">
                  <el-select
                    v-model="selectedGroupByUser[row.id]"
                    style="width: 200px"
                    :loading="groupsLoading"
                    @visible-change="handleGroupsVisibleChange"
                  >
                    <el-option value="" label="未分组" />
                    <el-option v-for="group in groups" :key="group.id" :value="group.id" :label="group.name" />
                  </el-select>
                  <el-button type="primary" native-type="submit" :loading="assigningUserID === row.id" size="small">更新</el-button>
                </form>
              </div>

              <div class="action-section">
                <h3>配额调整</h3>
                <form @submit.prevent="adjustQuota(row.id)" class="inline-form">
                  <el-input v-model.trim="quotaDelta[row.id]" type="number" placeholder="delta" style="width: 120px" size="small" />
                  <el-input v-model.trim="quotaReason[row.id]" placeholder="reason" style="width: 200px" size="small" />
                  <el-button type="primary" native-type="submit" :loading="submittingUserID === row.id" size="small">调额</el-button>
                </form>
              </div>

              <div class="action-section">
                <h3>限额查询</h3>
                <div class="inline-form">
                  <el-select
                    v-model="usageModelByUser[row.id]"
                    style="width: 200px"
                    size="small"
                    :loading="modelsLoading"
                    @visible-change="handleModelsVisibleChange"
                  >
                    <el-option value="" label="all models" />
                    <el-option v-for="model in models" :key="model.id" :value="model.id" :label="model.display_name" />
                  </el-select>
                  <el-button @click="loadLimitUsage(row.id)" :loading="loadingUsageUserID === row.id" size="small">加载限额</el-button>
                  <el-button @click="loadAdjustments(row.id)" :loading="loadingAdjustmentUserID === row.id" size="small">调整记录</el-button>
                </div>

                <div v-if="usageReports[row.id]" class="usage-report">
                  <div class="report-header">限额用量报告 (来源: {{ usageReports[row.id]?.effective_policy.source || 'none' }})</div>
                  <div class="report-grid">
                    <div class="report-item">
                      <span class="label">Hour Requests:</span>
                      <span>{{ usageReports[row.id]?.usage.hour.requests }} / adj={{ usageReports[row.id]?.adjustments.hour.requests }} / remain={{ remainingLabel(usageReports[row.id]?.remaining.hour.requests) }}</span>
                    </div>
                    <div class="report-item">
                      <span class="label">Hour Tokens:</span>
                      <span>{{ usageReports[row.id]?.usage.hour.tokens }} / adj={{ usageReports[row.id]?.adjustments.hour.tokens }} / remain={{ remainingLabel(usageReports[row.id]?.remaining.hour.tokens) }}</span>
                    </div>
                    <div class="report-item">
                      <span class="label">Week Requests:</span>
                      <span>{{ usageReports[row.id]?.usage.week.requests }} / adj={{ usageReports[row.id]?.adjustments.week.requests }} / remain={{ remainingLabel(usageReports[row.id]?.remaining.week.requests) }}</span>
                    </div>
                    <div class="report-item">
                      <span class="label">Week Tokens:</span>
                      <span>{{ usageReports[row.id]?.usage.week.tokens }} / adj={{ usageReports[row.id]?.adjustments.week.tokens }} / remain={{ remainingLabel(usageReports[row.id]?.remaining.week.tokens) }}</span>
                    </div>
                    <div class="report-item">
                      <span class="label">Lifetime Requests:</span>
                      <span>{{ usageReports[row.id]?.usage.lifetime.requests }} / adj={{ usageReports[row.id]?.adjustments.lifetime.requests }} / remain={{ remainingLabel(usageReports[row.id]?.remaining.lifetime.requests) }}</span>
                    </div>
                    <div class="report-item">
                      <span class="label">Lifetime Tokens:</span>
                      <span>{{ usageReports[row.id]?.usage.lifetime.tokens }} / adj={{ usageReports[row.id]?.adjustments.lifetime.tokens }} / remain={{ remainingLabel(usageReports[row.id]?.remaining.lifetime.tokens) }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <div class="action-section">
                <h3>新增限额调整</h3>
                <form @submit.prevent="createAdjustment(row.id)" class="adjustment-form">
                  <el-select
                    v-model="adjustmentModelByUser[row.id]"
                    placeholder="model"
                    size="small"
                    :loading="modelsLoading"
                    @visible-change="handleModelsVisibleChange"
                  >
                    <el-option value="" label="all models" />
                    <el-option v-for="model in models" :key="model.id" :value="model.id" :label="model.display_name" />
                  </el-select>
                  <el-select v-model="adjustmentMetricByUser[row.id]" placeholder="metric" size="small">
                    <el-option value="request_count" label="request_count" />
                    <el-option value="total_tokens" label="total_tokens" />
                  </el-select>
                  <el-select v-model="adjustmentWindowByUser[row.id]" placeholder="window" size="small">
                    <el-option value="rolling_hour" label="rolling_hour" />
                    <el-option value="rolling_week" label="rolling_week" />
                    <el-option value="lifetime" label="lifetime" />
                  </el-select>
                  <el-input v-model.trim="adjustmentDeltaByUser[row.id]" type="number" placeholder="delta" size="small" style="width: 120px" />
                  <el-input v-model.trim="adjustmentReasonByUser[row.id]" placeholder="reason" size="small" style="width: 200px" />
                  <el-button type="primary" native-type="submit" :loading="submittingAdjustmentUserID === row.id" size="small">新增</el-button>
                </form>

                <ul v-if="(adjustmentsByUser[row.id] ?? []).length > 0" class="adjustment-list">
                  <li v-for="adjustment in adjustmentsByUser[row.id]" :key="adjustment.id">
                    {{ adjustment.metric_type }} / {{ adjustment.window_type }} / delta={{ adjustment.delta }} / model={{ adjustment.model_id || 'all' }} / expires_at={{ adjustment.expires_at || '-' }}
                  </li>
                </ul>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" min-width="150" />
        <el-table-column prop="email" label="邮箱" min-width="200" />
        <el-table-column prop="role" label="角色" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <span class="status-badge" :class="row.status">{{ row.status }}</span>
          </template>
        </el-table-column>
        <el-table-column label="配额" width="180">
          <template #default="{ row }">{{ row.used_quota }} / {{ row.quota }}</template>
        </el-table-column>
        <el-table-column label="用户组" min-width="120">
          <template #default="{ row }">{{ row.user_group?.name || '-' }}</template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无用户" />
        </template>
      </el-table>

      <el-pagination
        v-if="total > 0"
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="loadUsers"
        @size-change="loadUsers"
        style="margin-top: 16px; justify-content: flex-end"
      />
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
const total = ref(0)
const groups = ref<UserGroupItem[]>([])
const models = ref<ModelItem[]>([])
const groupsLoading = ref(false)
const modelsLoading = ref(false)
const pagination = reactive({
  page: 1,
  pageSize: 20
})
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
let groupsRequest: Promise<void> | null = null
let modelsRequest: Promise<void> | null = null

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
      page: String(pagination.page),
      page_size: String(pagination.pageSize)
    })
    if (filters.keyword) {
      params.set('keyword', filters.keyword)
    }
    if (filters.status) {
      params.set('status', filters.status)
    }

    const data = await listAdminUsers<AdminUser[]>(auth.accessToken, params.toString())
    items.value = data
    total.value = data.length
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

.expand-content {
  padding: 16px 48px;
  background: var(--bg-secondary);
}

.filter-form {
  gap: 0;
}

.filter-row {
  grid-template-columns: minmax(220px, 320px) minmax(180px, 240px) auto;
  align-items: center;
  gap: 0.75rem;
}

.filter-group {
  flex-direction: row;
  align-items: center;
  gap: 0.5rem;
}

.filter-label {
  width: 52px;
  flex-shrink: 0;
  margin: 0;
}

.filter-control {
  flex: 1;
  min-width: 0;
}

.filter-actions {
  display: flex;
  justify-content: flex-start;
  align-items: center;
}

.filter-actions .submit-btn {
  margin-top: 0;
  padding: 0.35rem 1.1rem;
  white-space: nowrap;
}

.action-section {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--input-border);
}

.action-section:last-child {
  border-bottom: none;
  margin-bottom: 0;
  padding-bottom: 0;
}

.action-section h3 {
  margin: 0 0 12px 0;
  font-size: 0.9rem;
  color: var(--text-primary);
  font-weight: 600;
}

.inline-form {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.usage-report {
  margin-top: 12px;
  border: 1px dashed var(--input-border);
  border-radius: 8px;
  padding: 12px;
  background: var(--bg-primary);
}

.report-header {
  color: var(--text-primary);
  font-size: 0.86rem;
  font-weight: 600;
  margin-bottom: 8px;
}

.report-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 8px;
}

.report-item {
  display: flex;
  gap: 6px;
  align-items: baseline;
  font-size: 0.82rem;
  color: var(--text-secondary);
}

.report-item .label {
  color: var(--text-primary);
  font-weight: 500;
}

.adjustment-form {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.adjustment-list {
  margin: 12px 0 0 0;
  padding-left: 20px;
  color: var(--text-secondary);
  font-size: 0.82rem;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

@media (max-width: 768px) {
  .filter-row {
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }

  .filter-group {
    flex-direction: column;
    align-items: stretch;
    gap: 0.5rem;
  }

  .filter-label {
    width: auto;
  }

  .filter-actions {
    justify-content: stretch;
  }

  .filter-actions .submit-btn {
    width: 100%;
  }
}
</style>
