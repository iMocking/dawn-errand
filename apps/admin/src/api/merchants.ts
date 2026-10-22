import { http } from '@/utils/request'
import type { PageResult, PageQuery } from '@/types'

/** 商家列表行 */
export interface MerchantRow {
  id: string
  name: string
  /** logo 底色 / 图标描边色 */
  logoBg: string
  logoColor: string
  /** logo 图标类型（餐饮 cup / 商超 store / 药店 cross / 其他 sprout） */
  logoIcon: 'food' | 'shop' | 'pharmacy' | 'other'
  address: string
  addressSub: string
  contact: string
  phone: string
  category: string
  categoryBg: string
  categoryColor: string
  /** 营业中 / 休息中 / 已封禁 */
  status: string
  /** 月订单量（封禁商家为 null） */
  monthlyOrders: number | null
  /** 评分（封禁商家为 null） */
  rating: number | null
}

/** 商家列表顶部聚合统计 */
export interface MerchantStats {
  total: number
  weekNew: number
  open: number
  openRate: string
  pending: number
  closed: number
  closedDetail: string
}

/** 入驻审核申请行 */
export interface MerchantAuditRow {
  id: string
  /** 申请单号 SQ 开头 */
  applyNo: string
  name: string
  operator: string
  type: string
  address: string
  /** 申请时间 Unix 秒 */
  appliedAt: number
  logoBg: string
  logoColor: string
  logoIcon: 'food' | 'shop' | 'pharmacy'
  /** 证照缩略项 */
  certs: { label: string; icon: 'license' | 'idcard' | 'storefront' | 'foodlicense' | 'pharmacylicense' }[]
  /** 待审核 / 已通过 / 已驳回 */
  status: string
}

/** 商家详情 */
export interface MerchantDetail {
  id: string
  name: string
  status: string
  category: string
  categoryBg: string
  categoryColor: string
  /** 入驻时间 Unix 秒 */
  joinedAt: number
  rating: number
  monthlyOrders: number
  monthlyRevenue: number
  avgOrder: number
  goodRate: string
  logoBg: string
  logoColor: string
  basic: {
    name: string
    owner: string
    phone: string
    type: string
    address: string
    openHours: string
    deliveryRange: string
    commission: string
    bankAccount: string
    deliveryMethod: string
  }
  /** 最后更新时间 Unix 秒 */
  lastUpdatedAt: number
  completeness: number
  today: {
    orders: number
    ordersDelta: string
    revenue: number
    revenueDelta: string
    onTimeRate: string
    onTimeTarget: string
    pendingAbnormal: number
  }
  reviewCount: number
  lastOperation: string
}

export function getMerchants(q: PageQuery) {
  return http.get<PageResult<MerchantRow>>('/merchants', q as Record<string, unknown>)
}

export function getMerchantStats() {
  return http.get<MerchantStats>('/merchants/stats')
}

export function getMerchantAuditList(q: PageQuery) {
  return http.get<PageResult<MerchantAuditRow>>('/merchants/audit', q as Record<string, unknown>)
}

export function getMerchantDetail(id: string) {
  return http.get<MerchantDetail>(`/merchants/${id}`)
}

export function auditMerchant(id: string, action: 'approve' | 'reject', reason?: string) {
  return http.post<unknown>(`/merchants/audit/${id}`, { id, action, reason })
}

/** 新增商家（创建即提交入驻申请，进入待审核） */
export interface CreateMerchantPayload {
  name: string
  category?: string
  contact?: string
  phone: string
  address?: string
  addressSub?: string
  openHours?: string
  deliveryRange?: string
  commission?: number
  deliveryMethod?: string
  regionId?: string
}

/** 新增商家返回：创建成功后回传 id / 申请单号 / 审核状态 */
export interface CreateMerchantResult {
  id: string
  applyNo: string
  auditStatus: string
}

export function createMerchant(data: CreateMerchantPayload) {
  return http.post<CreateMerchantResult>('/merchants', data)
}
