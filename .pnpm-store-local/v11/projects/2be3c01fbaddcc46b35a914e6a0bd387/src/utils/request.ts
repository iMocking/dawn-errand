import axios, { type AxiosInstance, type AxiosRequestConfig, type InternalAxiosRequestConfig } from 'axios'
import type { ApiResponse } from '@/types'
import { genRequestId } from './ulid'
import { useToast } from '@/composables/useToast'

/** 业务错误：携带 code / request_id，便于排查 */
export class ApiError extends Error {
  code: number
  requestId: string
  constructor(code: number, message: string, requestId: string) {
    super(message)
    this.code = code
    this.requestId = requestId
  }
}

const instance: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

instance.interceptors.request.use((config) => {
  config.headers.set('X-Request-ID', genRequestId())
  const token = localStorage.getItem('admin_token')
  if (token) config.headers.set('Authorization', `Bearer ${token}`)
  return config
})

/**
 * 401 静默续期（单飞）：access token 过期时用 refresh_token 换新并重放原请求。
 * 多个并发请求同时 401 时只发起一次刷新，其余排队等待同一个 Promise。
 */
let refreshing: Promise<string> | null = null

async function doRefresh(): Promise<string> {
  const rt = localStorage.getItem('admin_refresh_token')
  if (!rt) throw new Error('缺少 refresh_token')
  const resp = await axios.post<ApiResponse<{
    token: string
    refresh_token: string
    user: { id: string; name: string; avatar: string; role: string }
    permissions: string[]
  }>>('/api/auth/refresh', { refresh_token: rt })
  const body = resp.data?.body
  if (!body?.token || !body?.refresh_token) throw new Error('刷新响应不完整')
  localStorage.setItem('admin_token', body.token)
  localStorage.setItem('admin_refresh_token', body.refresh_token)
  localStorage.setItem('admin_user', JSON.stringify(body.user))
  localStorage.setItem('admin_permissions', JSON.stringify(body.permissions))
  return body.token
}

function clearSession(toast: ReturnType<typeof useToast>) {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_refresh_token')
  localStorage.removeItem('admin_user')
  localStorage.removeItem('admin_permissions')
  if (location.pathname !== '/login') location.href = '/login'
}

instance.interceptors.response.use(
  ((resp: { data: ApiResponse }) => {
    const data = resp.data
    if (data && typeof data === 'object' && 'code' in data) {
      if (data.code === 0) return data.body
      const err = new ApiError(data.code, data.message || `业务错误 ${data.code}`, data.request_id || '-')
      useToast().error(`${err.message}（${err.requestId}）`)
      return Promise.reject(err)
    }
    return data
    // 拦截器统一解包 body，返回类型与 axios 默认 AxiosResponse 不同，用 never 断言绕过
  }) as (resp: unknown) => never,
  async (error) => {
    const toast = useToast()
    const status: number = error?.response?.status ?? 0
    const cfg = error?.config as InternalAxiosRequestConfig | undefined
    const requestId: string =
      (error?.response?.data?.request_id as string) || error?.config?.headers?.['X-Request-ID'] || '-'

    if (status === 401 && cfg && !String(cfg.url).includes('/auth/refresh')) {
      try {
        refreshing = refreshing ?? doRefresh().finally(() => {
          refreshing = null
        })
        const newToken = await refreshing
        cfg.headers.set('Authorization', `Bearer ${newToken}`)
        return instance(cfg)
      } catch {
        toast.error('登录已失效，请重新登录')
        clearSession(toast)
        return Promise.reject(error)
      }
    }

    if (status === 401) {
      toast.error('登录已失效，请重新登录')
      clearSession(toast)
    } else if (status === 403) {
      toast.error('无权限访问')
      location.href = '/403'
    } else if (status === 404) {
      toast.error(`接口不存在（${requestId}）`)
    } else if (status >= 500) {
      toast.error(`服务器异常（${requestId}）`)
    } else {
      toast.error(error?.message || '网络异常')
    }
    return Promise.reject(error)
  },
)

/** 统一请求：成功返回 body，失败抛 ApiError */
export function request<T = unknown>(config: AxiosRequestConfig): Promise<T> {
  return instance.request(config) as Promise<T>
}

export const http = {
  get<T = unknown>(url: string, params?: Record<string, unknown>) {
    return request<T>({ url, method: 'GET', params })
  },
  post<T = unknown>(url: string, data?: unknown) {
    return request<T>({ url, method: 'POST', data })
  },
  put<T = unknown>(url: string, data?: unknown) {
    return request<T>({ url, method: 'PUT', data })
  },
  delete<T = unknown>(url: string, params?: Record<string, unknown>) {
    return request<T>({ url, method: 'DELETE', params })
  },
}
