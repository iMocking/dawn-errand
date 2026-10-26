import axios from 'axios'

/** 私有桶对象的路径标记（后端 storage 生成的 URL 形如 <base>/api/uploads/private/<主体>/<ID>/<ulid>.<ext>） */
const PRIVATE_MARK = '/api/uploads/private/'

/** 是否私有桶对象（证件类） */
export function isPrivateSrc(url: string): boolean {
  return url.includes(PRIVATE_MARK)
}

/**
 * 剥成同源路径：STORAGE_PUBLIC_BASE_URL 默认带 http://localhost:8080，
 * 直接请求会绕开前端 /api 代理并触发跨域预检，取标记起的同源路径最稳。
 */
export function privatePath(url: string): string {
  const i = url.indexOf(PRIVATE_MARK)
  return i === -1 ? url : url.slice(i)
}

/**
 * 带 Bearer 拉取私有文件。这里刻意不走 utils/request 的 http 封装：
 * 它的响应拦截器把业务错误统一 toast，且 403 会 location.href='/403' 整页跳转——
 * 一张图取不到不该把审核人踢出后台，图片侧只需要占位提示。
 */
export async function fetchPrivateBlob(url: string): Promise<Blob> {
  const token = localStorage.getItem('admin_token')
  const resp = await axios.get<Blob>(privatePath(url), {
    responseType: 'blob',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  return resp.data
}

/** 在新标签页打开私有文件（对象 URL 延迟释放，留够查看时间） */
export async function openPrivateFile(url: string): Promise<void> {
  const blob = await fetchPrivateBlob(url)
  const objectUrl = URL.createObjectURL(blob)
  window.open(objectUrl, '_blank', 'noopener')
  setTimeout(() => URL.revokeObjectURL(objectUrl), 60_000)
}
