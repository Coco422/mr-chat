<template>
  <section>
    <h1>Usage</h1>
    <button type="button" @click="loadData">刷新</button>
    <p v-if="errorMessage">{{ errorMessage }}</p>

    <div>
      <h2>Quota</h2>
      <pre>{{ quota }}</pre>
    </div>

    <div>
      <h2>Usage Range</h2>
      <select v-model="usageRange" @change="loadData">
        <option value="7d">7d</option>
        <option value="30d">30d</option>
        <option value="month">month</option>
      </select>
      <pre>{{ usage }}</pre>
    </div>

    <div>
      <h2>Billing Summary</h2>
      <pre>{{ billingSummary }}</pre>
    </div>

    <div>
      <h2>Billing Logs</h2>
      <pre>{{ billingLogs }}</pre>
      <p v-if="logsMeta">
        page={{ logsMeta.page }} page_size={{ logsMeta.page_size }} total={{ logsMeta.total }}
      </p>
    </div>
  </section>
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

onMounted(() => {
  void loadData()
})

async function loadData() {
  if (!auth.accessToken) {
    return
  }

  errorMessage.value = ''

  try {
    const [quotaResult, usageResult, summaryResult, logsResult] = await Promise.all([
      apiRequest<QuotaResponse>('/users/me/quota', {
        accessToken: auth.accessToken
      }),
      apiRequest<UsageResponse>(`/users/me/usage?range=${usageRange.value}`, {
        accessToken: auth.accessToken
      }),
      apiRequest<BillingSummaryResponse>('/billing/summary', {
        accessToken: auth.accessToken
      }),
      apiRequest<BillingLogItem[]>('/billing/logs?page=1&page_size=20', {
        accessToken: auth.accessToken
      })
    ])

    quota.value = quotaResult.data
    usage.value = usageResult.data
    billingSummary.value = summaryResult.data
    billingLogs.value = logsResult.data
    logsMeta.value = logsResult.meta ?? null
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '加载 usage 数据失败'
  }
}
</script>
