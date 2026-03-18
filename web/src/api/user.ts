import { apiRequest } from '@/lib/api'

export interface SecurityInfoResponse {
  last_login_at: string | null
  password_updated_at: string | null
  has_password: boolean
}

export interface QuotaResponse {
  quota: number
  used_quota: number
  remaining_quota: number
}

export interface BillingLogItem {
  id: string
  type: string
  delta_quota: number
  balance_after: number
  reason?: string
  created_at: string
}

export async function getCurrentUser<T>(accessToken: string) {
  const { data } = await apiRequest<T>('/users/me', {
    accessToken
  })
  return data
}

export async function updateCurrentUser<T>(accessToken: string, body: Record<string, unknown>) {
  const { data } = await apiRequest<T>('/users/me', {
    method: 'PUT',
    accessToken,
    body
  })
  return data
}

export async function getSecurityInfo(accessToken: string) {
  const { data } = await apiRequest<SecurityInfoResponse>('/users/me/security', {
    accessToken
  })
  return data
}

export async function updateMyPassword(accessToken: string, currentPassword: string, newPassword: string) {
  await apiRequest('/users/me/password', {
    method: 'PUT',
    accessToken,
    body: {
      current_password: currentPassword,
      new_password: newPassword
    }
  })
}

export async function getMyQuota(accessToken: string) {
  const { data } = await apiRequest<QuotaResponse>('/users/me/quota', {
    accessToken
  })
  return data
}

export async function getBillingLogs(accessToken: string, page = 1, pageSize = 20) {
  const { data, meta } = await apiRequest<BillingLogItem[]>(`/billing/logs?page=${page}&page_size=${pageSize}`, {
    accessToken
  })
  return { data, meta }
}
