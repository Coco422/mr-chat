import axios, { type Method, type InternalAxiosRequestConfig } from 'axios'

import { reportPerfMetric } from '@/lib/performance'

let isRefreshing = false
let failedQueue: Array<{
  resolve: (token: string) => void
  reject: (error: unknown) => void
}> = []

function processQueue(error: unknown, token: string | null = null) {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve(token!)
    }
  })
  failedQueue = []
}

const configuredApiBaseUrl = import.meta.env.VITE_API_BASE_URL?.trim() ?? ''
export const apiBaseUrl = configuredApiBaseUrl.replace(/\/$/, '')

interface ApiEnvelope<T> {
  success: boolean
  data: T
  meta?: Record<string, unknown>
  error?: {
    code: string
    message: string
    details?: unknown
  }
}

interface ApiRequestOptions {
  method?: Method
  accessToken?: string
  body?: unknown
  signal?: AbortSignal
}

export class ApiError extends Error {
  code: string
  status: number
  details?: unknown

  constructor(message: string, code: string, status: number, details?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
    this.details = details
  }
}

export const apiClient = axios.create({
  baseURL: `${apiBaseUrl}/api/v1`,
  timeout: 150000,
  withCredentials: true,
  headers: {
    Accept: 'application/json'
  }
})

apiClient.interceptors.response.use(
  (response) => response,
  async (error: unknown) => {
    const originalRequest = axios.isAxiosError(error) ? error.config : null

    if (axios.isAxiosError(error) && error.response?.status === 401 && originalRequest) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        })
          .then((token) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            return apiClient.request(originalRequest)
          })
          .catch((err) => Promise.reject(err))
      }

      isRefreshing = true

      try {
        const { useAuthStore } = await import('@/stores/auth')
        const auth = useAuthStore()
        const success = await auth.refreshSession()

        if (success) {
          processQueue(null, auth.accessToken)
          originalRequest.headers.Authorization = `Bearer ${auth.accessToken}`
          return apiClient.request(originalRequest)
        } else {
          processQueue(new Error('Token refresh failed'), null)
          if (typeof window !== 'undefined') {
            window.location.href = '/login'
          }
          return Promise.reject(normalizeAxiosError(error))
        }
      } catch (refreshError) {
        processQueue(refreshError, null)
        if (typeof window !== 'undefined') {
          window.location.href = '/login'
        }
        return Promise.reject(normalizeAxiosError(error))
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(normalizeAxiosError(error))
  }
)

export async function apiRequest<T = unknown>(path: string, options: ApiRequestOptions = {}) {
  const method = String(options.method ?? 'GET').toUpperCase()
  const startTime = getNow()

  try {
    const response = await apiClient.request<ApiEnvelope<T>>({
      url: path,
      method,
      data: options.body,
      signal: options.signal,
      headers: options.accessToken
        ? {
            Authorization: `Bearer ${options.accessToken}`
          }
        : undefined
    })

    const payload = response.data
    if (!payload || payload.success === false) {
      throw new ApiError(
        payload?.error?.message ?? `Request failed with status ${response.status}`,
        payload?.error?.code ?? 'HTTP_ERROR',
        response.status,
        payload?.error?.details
      )
    }

    reportPerfMetric({
      name: 'api_request',
      value: getNow() - startTime,
      unit: 'ms',
      kind: 'api',
      extra: {
        method,
        path,
        status: response.status,
        success: true
      }
    })

    return {
      data: payload.data,
      meta: payload.meta
    }
  } catch (error) {
    const apiError = error instanceof ApiError ? error : normalizeAxiosError(error)

    reportPerfMetric({
      name: 'api_request',
      value: getNow() - startTime,
      unit: 'ms',
      kind: 'api',
      extra: {
        method,
        path,
        status: apiError.status,
        success: false,
        errorCode: apiError.code,
        errorMessage: apiError.message
      }
    })

    throw apiError
  }
}

function normalizeAxiosError(error: unknown) {
  if (!axios.isAxiosError(error)) {
    return new ApiError('Unexpected request error', 'UNKNOWN_ERROR', 0)
  }

  const status = error.response?.status ?? 0
  const payload = error.response?.data as ApiEnvelope<unknown> | undefined

  return new ApiError(
    payload?.error?.message ?? error.message ?? 'Network request failed',
    payload?.error?.code ?? (status > 0 ? 'HTTP_ERROR' : 'NETWORK_ERROR'),
    status,
    payload?.error?.details
  )
}

function getNow() {
  return typeof performance === 'undefined' ? Date.now() : performance.now()
}
