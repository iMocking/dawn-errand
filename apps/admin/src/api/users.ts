import { http } from '@/utils/request'
import type { PageResult, PageQuery } from '@/types'
import type { ComplaintRow } from './complaints'

/** 用户状态（与设计稿 / mock 文案逐字一致） */
export type UserStatus = '正常' | '已封禁'

/** 用户列表行 */
export interface UserRow extends Record<string, unknown> {
  /** 用户编号，如 U86540 */
  id: string
  /** 昵称 */
  name: string
  /** 头像底色 / 头像字色（设计稿逐字色值） */
  avatarBg: string
  avatarColor: string
  /** 手机号，如 139-7155-8823 */
  phone: string
  /** 常驻区域 */
  area: string
  /** 累计下单次数 */
  orderCount: number
  /** 账号状态：正常 / 已封禁 */
  status: UserStatus | string
  /** 注册时间，Unix 秒 */
  registeredAt: number
  /** 后台角色编码（user / super_admin / region_station 等） */
  role: string
  /** 后台角色展示名（系统管理员 / 区域站长 / 正式会员…） */
  roleName: string
}

/** 用户模块顶部概览指标（对应 mock/users-overview.json） */
export interface UsersOverview {
  /** 注册用户总数 */
  total: number
  /** 今日新增 */
  todayNew: number
  /** 本月活跃用户 */
  activeMonth: number
  /** 活跃率 % */
  activeRate: number
  /** 人均下单次数 */
  avgOrders: number
  /** 已封禁数 */
  banned: number
}

/** 用户近期订单行 */
export interface UserRecentOrder extends Record<string, unknown> {
  id: string
  /** 业务类型：外卖 / 帮买 / 帮取 */
  bizType: string
  /** 业务类型小标签底色 / 字色 */
  bizBg: string
  bizColor: string
  /** 商家 / 内容 */
  merchant: string
  /** 金额（元） */
  amount: number
  /** 配送骑手 */
  rider: string
  status: string
  /** 下单时间，Unix 秒 */
  orderedAt: number
}

/** 用户详情（对应 mock/user-detail.json） */
export interface UserDetail {
  id: string
  /** 用户编号，如 U00088642 */
  userNo: string
  name: string
  avatarBg: string
  avatarColor: string
  /** 等级标签，如 Lv.5 乡镇常客 */
  level: string
  status: UserStatus | string
  phone: string
  /** 后台角色编码 */
  role: string
  /** 后台角色展示名 */
  roleName: string
  /** 注册时间，Unix 秒 */
  registeredAt: number
  /** 累计订单数 */
  totalOrders: number
  /** 累计消费（元） */
  totalSpent: number
  /** 本月下单数 */
  monthOrders: number
  /** 实名认证文案，如 已实名 */
  verified: string
  /** 用户昵称 */
  nickname: string
  /** 真实姓名 */
  realName: string
  /** 所在区域 */
  region: string
  /** 常用地址 */
  address: string
  /** 注册渠道 */
  registerChannel: string
  /** 常用业务 */
  bizTypes: string
  /** 账户余额（元） */
  balance: number
  /** 红包余额（元） */
  redPacketBalance: number
  /** 最近活跃时间，Unix 秒 */
  recentActiveAt: number
  /** 订单记录总数（左侧角标） */
  orderRecordCount: number
  /** 投诉售后数（左侧角标） */
  complaintCount: number
  recentOrders: UserRecentOrder[]
  /** 实名认证申请（无申请为 null） */
  realNameVerify: RealNameVerify | null
}

/** 实名认证申请（身份证号脱敏） */
export interface RealNameVerify {
  id: string
  userId: string
  realName: string
  idCardMasked: string
  /** 待审核 / 已通过 / 已驳回 */
  status: string
  rejectReason: string
  appliedAt: number
  reviewedAt: number
}

/** 实名认证审核动作 */
export function auditRealName(id: string, action: 'approve' | 'reject', reason?: string) {
  return http.post<{ id: string; action: string }>(`/users/realname-verifications/${id}/audit`, {
    action,
    reason,
  })
}

/** 实名认证审核列表（分页 + status 过滤：待审核/已通过/已驳回） */
export function getRealNameVerifications(q: PageQuery & { status?: string }) {
  return http.get<PageResult<RealNameVerify>>(
    '/users/realname-verifications',
    q as Record<string, unknown>,
  )
}

export function getUsers(q: PageQuery) {
  return http.get<PageResult<UserRow>>('/users', q as Record<string, unknown>)
}

export function getUsersOverview() {
  return http.get<UsersOverview>('/users/overview')
}

export function getUserDetail(id: string) {
  return http.get<UserDetail>(`/users/${id}`)
}

/** 封禁账号（写操作，mock 中间件回显请求体） */
export function banUser(id: string) {
  return http.post<{ id: string; status: string }>(`/users/${id}/ban`, { id, status: '已封禁' })
}

/** 解封账号（写操作，mock 中间件回显请求体） */
export function unbanUser(id: string) {
  return http.post<{ id: string; status: string }>(`/users/${id}/unban`, { id, status: '正常' })
}

/** 用户订单记录（分页，复用近期订单行结构） */
export function getUserOrders(id: string, q: PageQuery) {
  return http.get<PageResult<UserRecentOrder>>(`/users/${id}/orders`, q as Record<string, unknown>)
}

/** 用户投诉售后（分页，复用投诉工单行结构） */
export function getUserComplaints(id: string, q: PageQuery) {
  return http.get<PageResult<ComplaintRow>>(`/users/${id}/complaints`, q as Record<string, unknown>)
}

/** 后台角色枚举（对齐后端 identity.AllRoles） */
export const BACKEND_ROLES: Array<{ value: string; label: string }> = [
  { value: 'user', label: '正式会员' },
  { value: 'super_admin', label: '系统管理员' },
  { value: 'cs_manager', label: '客服经理' },
  { value: 'cs_agent', label: '平台客服' },
  { value: 'region_station', label: '区域站长' },
  { value: 'merchant', label: '商家' },
  { value: 'driver', label: '司机' },
]

/** 修改用户后台角色（账号权限，仅超管可操作，不能改自己） */
export function updateUserRole(id: string, role: string) {
  return http.post<{ id: string; role: string }>(`/users/${id}/role`, { role })
}
