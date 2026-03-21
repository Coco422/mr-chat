import { apiRequest } from '@/lib/api'

export async function getAdminReferences<T>(accessToken: string) {
  const { data } = await apiRequest<T>('/admin/references', {
    accessToken
  })
  return data
}

export async function listAdminChannels<T>(accessToken: string) {
  const { data } = await apiRequest<T>('/admin/channels', {
    accessToken
  })
  return data
}

export async function getAdminChannel<T>(accessToken: string, channelID: string) {
  const { data } = await apiRequest<T>(`/admin/channels/${channelID}`, {
    accessToken
  })
  return data
}

export async function createAdminChannel(accessToken: string, body: Record<string, unknown>) {
  await apiRequest('/admin/channels', {
    method: 'POST',
    accessToken,
    body
  })
}

export async function listAdminModels<T>(accessToken: string) {
  const { data } = await apiRequest<T>('/admin/models', {
    accessToken
  })
  return data
}

export async function getAdminModel<T>(accessToken: string, modelID: string) {
  const { data } = await apiRequest<T>(`/admin/models/${modelID}`, {
    accessToken
  })
  return data
}

export async function importAdminModels<T>(accessToken: string, body: Record<string, unknown>) {
  const { data } = await apiRequest<T>('/admin/models/import', {
    method: 'POST',
    accessToken,
    body
  })
  return data
}

export async function createAdminModel(accessToken: string, body: Record<string, unknown>) {
  await apiRequest('/admin/models', {
    method: 'POST',
    accessToken,
    body
  })
}

export async function listAdminUpstreams<T>(accessToken: string) {
  const { data } = await apiRequest<T>('/admin/upstreams', {
    accessToken
  })
  return data
}

export async function getAdminUpstream<T>(accessToken: string, upstreamID: string) {
  const { data } = await apiRequest<T>(`/admin/upstreams/${upstreamID}`, {
    accessToken
  })
  return data
}

export async function discoverAdminUpstreamModels<T>(accessToken: string, upstreamID: string) {
  const { data } = await apiRequest<T>(`/admin/upstreams/${upstreamID}/discovered-models`, {
    accessToken
  })
  return data
}

export async function createAdminUpstream(accessToken: string, body: Record<string, unknown>) {
  await apiRequest('/admin/upstreams', {
    method: 'POST',
    accessToken,
    body
  })
}

export async function listAdminUserGroups<T>(accessToken: string) {
  const { data } = await apiRequest<T>('/admin/user-groups', {
    accessToken
  })
  return data
}

export async function getAdminUserGroup<T>(accessToken: string, groupID: string) {
  const { data } = await apiRequest<T>(`/admin/user-groups/${groupID}`, {
    accessToken
  })
  return data
}

export async function createAdminUserGroup(accessToken: string, body: Record<string, unknown>) {
  await apiRequest('/admin/user-groups', {
    method: 'POST',
    accessToken,
    body
  })
}

export async function getAdminUserGroupLimits<T>(accessToken: string, groupID: string) {
  const { data } = await apiRequest<T>(`/admin/user-groups/${groupID}/limits`, {
    accessToken
  })
  return data
}

export async function updateAdminUserGroupLimits(accessToken: string, groupID: string, body: Record<string, unknown>) {
  await apiRequest(`/admin/user-groups/${groupID}/limits`, {
    method: 'PUT',
    accessToken,
    body
  })
}

export async function listAdminAuditLogs<T>(accessToken: string, query: string) {
  const { data } = await apiRequest<T>(`/admin/audit-logs?${query}`, {
    accessToken
  })
  return data
}

export async function listAdminUsers<T>(accessToken: string, query: string) {
  const { data } = await apiRequest<T>(`/admin/users?${query}`, {
    accessToken
  })
  return data
}

export async function updateAdminUserGroup(accessToken: string, userID: string, userGroupID: string | null) {
  await apiRequest(`/admin/users/${userID}/group`, {
    method: 'PUT',
    accessToken,
    body: {
      user_group_id: userGroupID 
    }
  })
}

export async function updateAdminUserQuota(accessToken: string, userID: string, delta: number, reason: string) {
  await apiRequest(`/admin/users/${userID}/quota`, {
    method: 'PUT',
    accessToken,
    body: {
      delta,
      reason
    }
  })
}

export async function getAdminUserLimitUsage<T>(accessToken: string, userID: string, modelID?: string) {
  const suffix = modelID ? `?model_id=${encodeURIComponent(modelID)}` : ''
  const { data } = await apiRequest<T>(`/admin/users/${userID}/limit-usage${suffix}`, {
    accessToken
  })
  return data
}

export async function listAdminUserLimitAdjustments<T>(
  accessToken: string,
  userID: string,
  query: string
) {
  const { data } = await apiRequest<T>(`/admin/users/${userID}/limit-adjustments?${query}`, {
    accessToken
  })
  return data
}

export async function createAdminUserLimitAdjustment(
  accessToken: string,
  userID: string,
  body: Record<string, unknown>
) {
  await apiRequest(`/admin/users/${userID}/limit-adjustments`, {
    method: 'POST',
    accessToken,
    body
  })
}
