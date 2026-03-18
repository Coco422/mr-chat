<template>
  <div class="usage-page">
    <div class="page-header">
      <h1>用量统计</h1>
      <button class="refresh-btn" @click="loadData" :disabled="!auth.accessToken">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <polyline points="23 4 23 10 17 10" stroke-width="2"/>
          <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" stroke-width="2"/>
        </svg>
      </button>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-label">剩余额度</div>
        <div class="stat-value">{{ formatQuota(quota?.remaining_quota || 0) }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">已使用</div>
        <div class="stat-value">{{ formatQuota(quota?.used_quota || 0) }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">总额度</div>
        <div class="stat-value">{{ formatQuota(quota?.quota || 0) }}</div>
      </div>
    </div>

    <div class="usage-section">
      <div class="section-header">
        <h2>使用记录</h2>
        <div class="range-tabs">
          <button
            v-for="range in ranges"
            :key="range.value"
            @click="usageRange = range.value; loadData()"
            class="range-tab"
            :class="{ active: usageRange === range.value }"
          >
            {{ range.label }}
          </button>
        </div>
      </div>
      <div class="logs-table">
        <table v-if="billingLogs.length > 0">
          <thead>
            <tr>
              <th>时间</th>
              <th>类型</th>
              <th>变动</th>
              <th>余额</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in billingLogs" :key="log.id">
              <td>{{ formatDate(log.created_at) }}</td>
              <td>{{ log.type }}</td>
              <td :class="log.delta_quota > 0 ? 'positive' : 'negative'">
                {{ log.delta_quota > 0 ? '+' : '' }}{{ formatQuota(log.delta_quota) }}
              </td>
              <td>{{ formatQuota(log.balance_after) }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty">暂无记录</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { ApiError, apiRequest } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface QuotaResponse {
  quota: number
  used_quota: number
  remaining_quota: number
}

interface UsageResponse {
  summary: {
    total_spent_tokens: number
    spent_today: number
    spent_in_range: number
  }
  daily: Array<{
    date: string
    spent_tokens: number
  }>
}

interface BillingSummaryResponse {
  remaining_quota: number
  consumed_total: number
  redeemed_total: number
}

interface BillingLogItem {
  id: string
  type: string
  delta_quota: number
  balance_after: number
  reason?: string
  created_at: string
}

const auth = useAuthStore()
const errorMessage = ref('')
const usageRange = ref<'7d' | '30d' | 'month'>('7d')
const quota = ref<QuotaResponse | null>(null)
const usage = ref<UsageResponse | null>(null)
const billingSummary = ref<BillingSummaryResponse | null>(null)
const billingLogs = ref<BillingLogItem[]>([])
const logsMeta = ref<Record<string, unknown> | null>(null)

const ranges: Array<{ value: '7d' | '30d' | 'month'; label: string }> = [
  { value: '7d', label: '最近7天' },
  { value: '30d', label: '最近30天' },
  { value: 'month', label: '本月' }
]

onMounted(() => {
  void loadData()
})

async function loadData() {
  if (!auth.accessToken) {
    return
  }

  errorMessage.value = ''

  try {
    const [quotaResult, logsResult] = await Promise.all([
      apiRequest<QuotaResponse>('/users/me/quota', {
        accessToken: auth.accessToken
      }),
      apiRequest<BillingLogItem[]>('/billing/logs?page=1&page_size=20', {
        accessToken: auth.accessToken
      })
    ])

    quota.value = quotaResult.data
    billingLogs.value = logsResult.data
    logsMeta.value = logsResult.meta ?? null
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '加载数据失败'
  }
}

function formatQuota(amount: number) {
  return (amount / 1000).toFixed(2) + ' K'
}

function formatDate(date: string) {
  return new Date(date).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.usage-page {
  padding: 2rem;
  max-width: 1200px;
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

.refresh-btn {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.refresh-btn:hover:not(:disabled) {
  background: var(--accent-primary);
  color: white;
}

.error {
  padding: 1rem;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid var(--error-color);
  border-radius: 8px;
  color: var(--error-color);
  margin-bottom: 1.5rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.stat-card {
  padding: 1.5rem;
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  border-radius: 12px;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 600;
  color: var(--text-primary);
}

.usage-section {
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  border-radius: 12px;
  padding: 1rem;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
}

.section-header h2 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.range-tabs {
  display: flex;
  gap: 0.5rem;
}

.range-tab {
  padding: 0.5rem 1rem;
  background: transparent;
  border: 1px solid var(--input-border);
  border-radius: 8px;
  color: var(--text-secondary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.range-tab:hover {
  background: var(--input-bg);
  color: var(--text-primary);
}

.range-tab.active {
  background: var(--accent-primary);
  color: white;
  border-color: var(--accent-primary);
}

.logs-table table {
  width: 100%;
  border-collapse: collapse;
}

.logs-table th {
  text-align: left;
  padding: 0.75rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--input-border);
}

.logs-table td {
  padding: 0.875rem 0.75rem;
  font-size: 0.9rem;
  color: var(--text-primary);
  border-bottom: 1px solid var(--input-border);
}

.logs-table tr:last-child td {
  border-bottom: none;
}

.logs-table .positive {
  color: #10b981;
}

.logs-table .negative {
  color: var(--error-color);
}

.empty {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}
</style>
