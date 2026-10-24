import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import type { ApiResponse } from '@/types'
import { genRequestId } from './ulid'

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
  return config
})

instance.interceptors.response.use(
  ((resp: { data: ApiResponse }) => {
    const data = resp.data
    if (data && typeof data === 'object' && 'code' in data) {
      if (data.code === 0) return data.body
      const err = new ApiError(
        data.code,
        data.message || `业务错误 ${data.code}`,
        data.request_id || '-',
      )
      return Promise.reject(err)
    }
    return data
  }) as (resp: unknown) => never,
  (error) => {
    const status: number = error?.response?.status ?? 0
    const requestId: string =
      (error?.response?.data?.request_id as string) || error?.config?.headers?.['X-Request-ID'] || '-'
    if (status === 404) {
      return Promise.reject(new ApiError(404, `接口不存在（${requestId}）`, requestId))
    }
    if (status >= 500) {
      return Promise.reject(new ApiError(status, `服务器异常（${requestId}）`, requestId))
    }
    return Promise.reject(new ApiError(status, error?.message || '网络异常', requestId))
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
}
