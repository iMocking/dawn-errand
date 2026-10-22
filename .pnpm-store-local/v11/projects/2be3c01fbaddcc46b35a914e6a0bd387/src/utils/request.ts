import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
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
  (error) => {
    const toast = useToast()
    const status: number = error?.response?.status ?? 0
    const requestId: string =
      (error?.response?.data?.request_id as string) || error?.config?.headers?.['X-Request-ID'] || '-'
    if (status === 401) {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_user')
      localStorage.removeItem('admin_permissions')
      toast.error('登录已失效，请重新登录')
      if (location.pathname !== '/login') location.href = '/login'
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
