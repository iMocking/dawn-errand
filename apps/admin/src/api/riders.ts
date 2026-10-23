import { http } from '@/utils/request'
import type { PageResult, PageQuery } from '@/types'

/** 骑手列表行（type 别名：DataTable 约束 T extends Record<string, unknown>） */
export type RiderRow = {
  id: string
  /** 姓名 */
  name: string
  phone: string
  /** 服务区域，如「王家湾村 / 李庄村」 */
  area: string
  /** 配送工具：电动车 / 摩托车 / 步行 */
  tool: string
  /** 等级，如 Lv.5 */
  level: string
  /** 在线状态：在线 / 忙碌 / 离线 / 已冻结（中间件 status 精确筛选用） */
  status: string
  /** 今日单量（已冻结为 null，展示 —） */
  todayOrders: number | null
  /** 评分（已冻结为 null，展示 —） */
  rating: number | null
  /** 头像底色 / 头像字色（设计稿指定色） */
  avatarBg: string
  avatarColor: string
}

/** 顶部统计行 */
export interface RiderStatsDto {
  total: number
  online: number
  onlineRate: string
  pendingAudit: number
  frozen: number
  frozenNote: string
  weekDelta: string
  /** 入驻审核页 tab 汇总 */
  audit: {
    todayNew: number
    approved: number
    rejected: number
    monthApproved: number
    momRate: string
  }
}

/** 入驻申请行 */
export interface RiderAuditRow {
  id: string
  /** 状态：待审核（中间件 status 精确筛选用） */
  status: string
  name: string
  avatarBg: string
  avatarColor: string
  /** 申请编号，如 SQ20261022007 */
  applyNo: string
  /** 提交时间，Unix 秒 */
  submittedAt: number
  /** 区域角标，如「王家湾镇 · 一组」 */
  area: string
  phone: string
  /** 配送工具 */
  tool: string
  /** 接单类型 */
  orderTypes: string
  /** 身份证核验：已通过 / 待补充 / 已驳回 */
  idCheck: string
  /** 身份证号（脱敏，如 110101********7733） */
  idCardMasked: string
  /** 材料提示条：null 无 / warning 黄色警告 */
  noticeType: 'none' | 'warning'
  notice: string
  /** 已上传材料张数 */
  docCount: number
  docs: string
}

/** 骑手详情 */
export interface RiderDetailDto {
  id: string
  name: string
  status: string
  /** 服务标签，如 外卖/帮送 */
  serviceType: string
  /** 入驻日期（展示用，如 2025-06-18） */
  joinedAt: number
  rating: number
  todayOrders: number
  /** 今日在线小时，如 8.5 */
  onlineHours: number
  /** 本月收入（元） */
  monthIncome: number
  onTimeRate: string
  avatarBg: string
  /** 基本信息表单 */
  profile: {
    name: string
    phone: string
    idCard: string
    area: string
    vehicle: string
    orderTypes: string
    account: string
    /** 加入时间展示前缀 */
    joinedAt: string
    serviceMonths: number
  }
  /** 今日配送速览 4 格 */
  todayOverview: {
    todayOrders: number
    todayOrdersDelta: string
    onTimeRate: string
    onTimeTarget: string
    avgMinutes: string
    platformAvg: string
    complaints: number
    complaintNote: string
  }
  /** 最近操作描述，如「2026-10-22 10:26 上线接单」 */
  lastAction: string
  lastActionAt: number
}

export function getRiders(q: PageQuery) {
  return http.get<PageResult<RiderRow>>('/riders', q as Record<string, unknown>)
}

export function getRiderStats() {
  return http.get<RiderStatsDto>('/riders/stats')
}

export function getRiderAudit() {
  return http.get<PageResult<RiderAuditRow>>('/riders/audit')
}

export function getRiderDetail(id: string) {
  return http.get<RiderDetailDto>(`/riders/${id}`)
}

/** 审核通过（mock 中间件回显请求体） */
export function approveRider(id: string) {
  return http.post<{ id: string; action: string }>('/riders/audit', { id, action: 'approve' })
}

/** 审核驳回 */
export function rejectRider(id: string, reason?: string) {
  return http.post<{ id: string; action: string; reason?: string }>('/riders/audit', {
    id,
    action: 'reject',
    reason,
  })
}

/** 冻结 / 解冻骑手（mock 中间件回显请求体） */
export function freezeRider(id: string, action: 'freeze' | 'unfreeze') {
  return http.post<{ id: string; action: string }>('/riders/freeze', { id, action })
}

/** 新增骑手（创建即提交入驻申请，进入待审核） */
export interface CreateRiderPayload {
  name: string
  phone: string
  area?: string
  tool?: string
  orderTypes?: string
  idCard?: string
  serviceType?: string
  regionId?: string
}

/** 新增骑手返回：创建成功后回传 id / 申请单号 / 审核状态 */
export interface CreateRiderResult {
  id: string
  applyNo: string
  auditStatus: string
}

export function createRider(data: CreateRiderPayload) {
  return http.post<CreateRiderResult>('/riders', data)
}
