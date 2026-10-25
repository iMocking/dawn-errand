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
/** 商家入驻完整资料（资质/结算/账号） */
export interface MerchantApplication {
  licenseNo: string
  licenseImage: string
  legalName: string
  legalIdCard: string
  idCardFront: string
  idCardBack: string
  foodLicenseNo: string
  foodLicenseImage: string
  storefrontImages: string[]
  bankName: string
  bankAccountName: string
  bankAccountNo: string
  accountEmail: string
  accountPhone: string
  notes: string
}

export interface MerchantAuditRow {
  id: string
  applyNo: string
  name: string
  operator: string
  type: string
  address: string
  contact: string
  phone: string
  appliedAt: number
  submittedAt: number
  logoBg: string
  logoColor: string
  logoIcon: 'food' | 'shop' | 'pharmacy'
  certs: { label: string; icon: string; url?: string }[]
  application: MerchantApplication
  status: string
  rejectReason: string
  verifiedAt: number
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
  auditStatus: string
  rejectReason: string
  submittedAt: number
  verifiedAt: number
  ownerUserId: string
  application: MerchantApplication
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

export interface MerchantOrderBrief {
  id: string
  orderNo: string
  status: string
  type: string
  payAmount: number
  createdAt: number
  userName: string
  route: string
}

export interface MerchantWorkspace {
  application: MerchantApplication
  certs: { label: string; icon: string; url?: string }[]
  auditStatus: string
  verifiedAt: number
  orders: MerchantOrderBrief[]
  abnormal: MerchantOrderBrief[]
  trend: { date: string; orders: number; amount: number }[]
  menu: {
    categories: { id: string; name: string; sort: number }[]
    dishes: {
      id: string
      categoryId: string
      name: string
      description: string
      imageUrl: string
      price: number
      stock: number
      onSale: boolean
      soldOut: boolean
      monthlySales: number
    }[]
  }
  reviews: {
    list: {
      id: string
      orderId: string
      userId: string
      userName: string
      rating: number
      tags: string[]
      content: string
      images: string[]
      reply: string
      repliedAt: number
      createdAt: number
    }[]
    total: number
    rating: number
  }
}

/** 商家详情页各 Tab 数据（证照/经营/菜单/评价/异常） */
export function getMerchantWorkspace(id: string) {
  return http.get<MerchantWorkspace>(`/merchants/${id}/workspace`)
}

export function auditMerchant(id: string, action: 'approve' | 'reject', reason?: string) {
  return http.post<{
    id: string
    action: string
    account?: { userId: string; email: string; phone: string; initialPassword: string }
  }>(`/merchants/audit/${id}`, { id, action, reason })
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
  lat?: number
  lng?: number
  autoAccept?: boolean
  minPrice?: number
  deliveryFee?: number
  packingFee?: number
  announcement?: string
  servicePhone?: string
  application: MerchantApplication
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

/** 驳回后补件重提（仅未开通账号、未通过审核的申请） */
export function updateMerchantApplication(id: string, data: CreateMerchantPayload) {
  return http.put<{ id: string; auditStatus: string }>(`/merchants/${id}/application`, data)
}

/** 重置商家登录密码（旧会话失效，返回一次性新密码） */
export function resetMerchantAccount(id: string) {
  return http.post<{ account: { userId: string; email: string; phone: string; initialPassword: string } }>(
    `/merchants/${id}/account/reset`,
  )
}

/** 管理端上传图片（营业执照/身份证/门头照） */
export function uploadMerchantImage(file: File) {
  const form = new FormData()
  form.append('file', file)
  return http.post<{ url: string; name: string; size: number; mime: string }>('/uploads', form)
}
