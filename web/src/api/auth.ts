import { apiRequest } from '@/lib/api'

export interface AuthUser {
  id: string
  username: string
  email: string
  role: 'user' | 'admin' | 'root'
}

export interface AuthSessionResponse {
  access_token: string
  expires_in: number
  user: AuthUser
}

interface SignUpPayload {
  username: string
  email: string
  password: string
}

interface SignInPayload {
  identifier: string
  password: string
}

export async function signUp(payload: SignUpPayload) {
  const { data } = await apiRequest<AuthSessionResponse>('/auth/signup', {
    method: 'POST',
    body: payload
  })
  return data
}

export async function signIn(payload: SignInPayload) {
  const { data } = await apiRequest<AuthSessionResponse>('/auth/signin', {
    method: 'POST',
    body: payload
  })
  return data
}

export async function signOut() {
  await apiRequest('/auth/signout', {
    method: 'POST'
  })
}

export async function refreshToken() {
  const { data } = await apiRequest<AuthSessionResponse>('/auth/refresh', {
    method: 'POST'
  })
  return data
}
