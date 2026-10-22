import { http } from '@/utils/request'

export interface AuthUser {
  id: string
  name: string
  avatar: string
  role: string
}

export interface LoginResult {
  token: string
  user: AuthUser
  permissions: string[]
}

export interface LoginBody {
  account: string
  password: string
  captcha: string
}

export function login(body: LoginBody) {
  return http.post<LoginResult>('/auth/login', body)
}

export function logout() {
  return http.post<Record<string, never>>('/auth/logout')
}

export function getUserInfo() {
  return http.get<{ user: AuthUser; permissions: string[] }>('/auth/me')
}
