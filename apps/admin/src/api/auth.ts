import { http } from '@/utils/request'

export interface AuthUser {
  id: string
  name: string
  avatar: string
  role: string
}

export interface LoginResult {
  token: string
  refreshToken: string
  user: AuthUser
  permissions: string[]
}

/** 后端登录/刷新原始响应（refresh_token 为 snake_case，需映射为 camelCase） */
interface LoginResultRaw {
  token: string
  refresh_token: string
  user: AuthUser
  permissions: string[]
}

export interface LoginBody {
  account: string
  password: string
  captcha: string
}

function mapLoginResult(res: LoginResultRaw): LoginResult {
  return {
    token: res.token,
    refreshToken: res.refresh_token,
    user: res.user,
    permissions: res.permissions,
  }
}

export async function login(body: LoginBody): Promise<LoginResult> {
  const res = await http.post<LoginResultRaw>('/auth/login', body)
  return mapLoginResult(res)
}

export function logout() {
  return http.post<Record<string, never>>('/auth/logout')
}

/** 用 refresh_token 换新令牌（access + refresh 同时轮换，返回与登录一致） */
export async function refreshToken(refresh_token: string): Promise<LoginResult> {
  const res = await http.post<LoginResultRaw>('/auth/refresh', { refresh_token })
  return mapLoginResult(res)
}

export function getUserInfo() {
  return http.get<{ user: AuthUser; permissions: string[] }>('/auth/me')
}
