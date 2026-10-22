/** 统一后端响应包：{message, body, request_id, code}，code===0 成功 */
export interface ApiResponse<T = unknown> {
  message: string
  body: T
  request_id: string
  code: number
}

/** 分页返回结构 */
export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

/** 分页查询参数 */
export interface PageQuery {
  page?: number
  pageSize?: number
  keyword?: string
  status?: string
}

/** 通用字典 */
export type Dictionary<T = unknown> = Record<string, T>
