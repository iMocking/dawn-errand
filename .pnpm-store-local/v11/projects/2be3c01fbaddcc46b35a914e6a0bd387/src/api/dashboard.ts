import { http } from '@/utils/request'

/** 顶部指标卡 */
export interface StatCardDto {
  key: string
  label: string
  /** 主值，已千分位，如 '1,286' */
  value: string
  /** 副值（小号灰），如 '/104' */
  subValue?: string
  /** 涨跌文案，如 '▲ 12.5%' */
  delta: string
  /** 涨跌颜色语义 */
  deltaType: 'green' | 'yellow' | 'red'
  /** 右侧竖排脚注 */
  foot: string
  /** SvgIcon 名称 */
  icon: string
  /** 图标底色 / 图标色 */
  iconBg: string
  iconColor: string
}

/** 折线图点（viewBox 内坐标，与设计稿一致） */
export interface TrendPoint {
  x: number
  y: number
}

export interface TrendTip {
  x: number
  y: number
  w: number
  text: string
}

export interface TrendDto {
  xLabels: string[]
  orderLine: TrendPoint[]
  gmvLine: TrendPoint[]
  orderTip: TrendTip
  gmvTip: TrendTip
}

export interface DistItemDto {
  label: string
  percent: number
  /** CSS 背景（含渐变） */
  color: string
}

export interface RecentOrderDto {
  id: string
  type: string
  typeBg: string
  typeColor: string
  route: string
  status: string
  statusColor: string
  time: string
}

export interface TodoDto {
  icon: string
  bg: string
  color: string
  label: string
  count: number
}

export interface DashboardDto {
  pendingCount: number
  stats: StatCardDto[]
  trend: TrendDto
  distribution: DistItemDto[]
  onTimeRate: string
  onTimeTarget: string
  recentOrders: RecentOrderDto[]
  todos: TodoDto[]
}

export function getDashboard() {
  return http.get<DashboardDto>('/dashboard')
}
