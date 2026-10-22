/**
 * 时间 / 金额 / 数字格式化工具。
 * - 时间字段一律按 int64 Unix 秒（UTC）接收，展示转本地时区。
 * - 金额以「元」为单位存储，展示千分位。
 */

function pad(n: number): string {
  return n < 10 ? `0${n}` : String(n)
}

/** Unix 秒 → Date（本地时区） */
function toDate(sec: number): Date {
  return new Date(sec * 1000)
}

/** 'YYYY-MM-DD HH:mm' */
export function formatTime(sec: number): string {
  const d = toDate(sec)
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/** 'YYYY-MM-DD' */
export function formatDate(sec: number): string {
  const d = toDate(sec)
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

/** 'HH:mm' */
export function formatClock(sec: number): string {
  const d = toDate(sec)
  return `${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const WEEK_CN = ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六']

/** 顶栏面包屑用：'YYYY-MM-DD 星期X' */
export function formatToday(): string {
  const d = new Date()
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${WEEK_CN[d.getDay()]}`
}

/** 千分位数字，如 1286 → '1,286' */
export function formatNumber(n: number): string {
  return n.toLocaleString('en-US')
}

/** 金额（元）→ '¥32,540.00'；withSymbol=false 时仅千分位 */
export function formatMoney(yuan: number, withSymbol = true): string {
  const v = yuan.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  return withSymbol ? `¥${v}` : v
}

/** 时段问候：上午好 / 下午好 / 晚上好 */
export function greetByHour(): string {
  const h = new Date().getHours()
  if (h < 6) return '凌晨好'
  if (h < 12) return '上午好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
}
