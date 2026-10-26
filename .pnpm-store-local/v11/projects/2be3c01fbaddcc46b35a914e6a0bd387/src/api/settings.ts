import { http } from '@/utils/request'
import type { PageQuery, PageResult } from '@/types'

/** 服务区域卡片行（系统设置 · 服务区域管理） */
export interface ServiceArea extends Record<string, unknown> {
  /** 区域 ID */
  id: string
  /** 区域名称，如 王家湾镇 */
  name: string
  /** 运营中 / 筹备中 */
  status: string
  /** 下辖村组数 */
  villages: number
  /** 服务半径（km） */
  radius_km: number
  /** 在线商家；筹备中区域为 null（页面展示 —） */
  online_merchants: number | null
  /** 活跃骑手；筹备中区域为 null */
  active_riders: number | null
  /** 今日订单；筹备中区域为 null */
  today_orders: number | null
  /** 起步价（元） */
  start_fee: number
  /** 起步含里程（km） */
  free_km: number
  /** 开通提示，如「预计 11 月 1 日开通」；运营中为 null */
  notice: string | null
  /** 底部操作：停用 / 删除 */
  action: '停用' | '删除'
  /** 图标底色 / 图标线色（设计稿四色） */
  icon_bg: string
  icon_color: string
}

/** 系统设置对象：GET /api/settings */
export interface SystemSettings {
  /** 已开通乡镇总数 */
  total_towns: number
  /** 服务区域列表 */
  areas: ServiceArea[]
}

/** 费率规则行：按 category 分组渲染到状态条 / 基础费 / 阶梯表 / 特殊规则 / 适用范围 */
export interface FeeRuleRow extends Record<string, unknown> {
  id: string
  /** summary=状态条 · basic=基础配送费 · ladder=阶梯明细 · special=特殊规则 · scope=适用范围 */
  category: 'summary' | 'basic' | 'ladder' | 'special' | 'scope'
  name: string
  /** 已启用 / 已停用 */
  status: string
  /** 展示值（¥3.00 / +30% / 标准费率 等） */
  value?: string
  desc?: string
  /** 首卡橙色高亮 */
  accent?: boolean
  /** 适用范围偏远乡黄色高亮 */
  highlight?: boolean
  /** 阶梯列：计费规则 */
  rule?: string
  /** 阶梯列：示例订单 */
  example?: string
  /** 阶梯列：配送费 */
  fee?: string
  /** 阶梯列：骑手分成 */
  rider_share?: string
  /** 特殊规则图标线色 */
  icon_color?: string
  /** summary 行：生效版本号 */
  version?: string
  /** summary 行：更新时间（Unix 秒） */
  updated_at?: number
  /** summary 行：下次自动生效时间（Unix 秒） */
  next_effective_at?: number
  /** summary 行：规则是否启用 */
  enabled?: boolean
}

/** 公告推送统计行 */
export interface AnnouncementStat {
  label: string
  value: string
  /** 末行橙色高亮 */
  highlight?: boolean
}

/** 公告列表行 */
export interface AnnouncementRow extends Record<string, unknown> {
  id: string
  title: string
  /** 费率 / 活动 / 提醒 / 协议 */
  type: string
  /** 类型标签配色：orange / blue / yellow / purple */
  type_tone: 'orange' | 'blue' | 'yellow' | 'purple'
  /** 发布对象，如 商家 / 骑手、全部用户 */
  audience: string
  /** 已发布 / 待发布 / 审核中 */
  status: string
  /** 发布时间（Unix 秒）；待发布为空 */
  published_at: number | null
  /** 待发布 / 审核中的时间说明，如「计划 11-01 发布」 */
  planned_text: string
  /** 行内次操作：下线 / 发布 / 撤回 */
  action_name: string
  /** 正文（右侧预览） */
  content: string
  /** 标签 chips */
  tags: string[]
  /** 推送统计（右侧预览） */
  stats: AnnouncementStat[]
}

/** 系统设置：GET /api/settings */
export function getSettings() {
  return http.get<SystemSettings>('/settings')
}

/**
 * 区域新增 / 编辑 / 停用 / 删除：POST /api/settings
 * 后端契约：{ action: 'create' | 'update' | 'disable' | 'delete', area: Region }
 * 前端调用传入扁平 payload，此处归一为后端结构并映射操作名：
 * - action 别名：停用→disable、删除→delete，其余原样透传
 * - 编辑（update）需携带完整区域对象，避免整行覆盖清空 status / free_km / notice
 * 注意：后端已改为白名单 DTO（camelCase json tag），表外的键静默丢弃；
 * created_at / updated_at 由服务端生成，多词字段必须按 camelCase 发送（radiusKM / startFee / freeKM）。
 */
const AREA_BODY_FIELDS: Record<string, string> = {
  id: 'id',
  name: 'name',
  status: 'status',
  villages: 'villages',
  radius_km: 'radiusKM',
  start_fee: 'startFee',
  free_km: 'freeKM',
  notice: 'notice',
}

const AREA_ACTION_ALIAS: Record<string, string> = {
  停用: 'disable',
  删除: 'delete',
}

export function saveSettings(payload: Record<string, unknown>) {
  const rawAction = String(payload.action ?? '')
  const action = AREA_ACTION_ALIAS[rawAction] ?? rawAction
  const area: Record<string, unknown> = {}
  for (const [from, to] of Object.entries(AREA_BODY_FIELDS)) {
    if (payload[from] !== undefined) area[to] = payload[from]
  }
  return http.post<Record<string, unknown>>('/settings', { action, area })
}

/** 费率规则：GET /api/settings/fee-rules */
export function getFeeRules(q: PageQuery = {}) {
  return http.get<PageResult<FeeRuleRow>>('/settings/fee-rules', q as Record<string, unknown>)
}

/**
 * 编辑规则 / 启用切换：POST /api/settings/fee-rules
 * 后端契约：白名单 DTO feeRuleBody（camelCase 多词字段 riderShare / iconColor / nextEffectiveAt）。
 * - 启用切换只传 { id, enabled } 时补 status（后端以 status 持久化，enabled 由 status 派生，请求体的 enabled 被忽略）
 * - updatedAt 由服务端写入，请求体不再生效；前端表单遗留字段（action / start_fee / continue_fee / far_fee / cap 等）后端不存在，丢弃
 */
const FEE_BODY_FIELDS: Record<string, string> = {
  id: 'id',
  category: 'category',
  name: 'name',
  status: 'status',
  value: 'value',
  desc: 'desc',
  accent: 'accent',
  highlight: 'highlight',
  rule: 'rule',
  example: 'example',
  fee: 'fee',
  rider_share: 'riderShare',
  icon_color: 'iconColor',
  version: 'version',
  updated_at: 'updatedAt',
  next_effective_at: 'nextEffectiveAt',
  enabled: 'enabled',
}

export function saveFeeRule(payload: Record<string, unknown>) {
  const body: Record<string, unknown> = {}
  for (const [from, to] of Object.entries(FEE_BODY_FIELDS)) {
    if (payload[from] !== undefined) body[to] = payload[from]
  }
  if (payload.enabled !== undefined) {
    body.status = payload.enabled ? '已启用' : '已停用'
  }
  return http.post<Record<string, unknown>>('/settings/fee-rules', body)
}

/** 公告列表：GET /api/settings/announcements */
export function getAnnouncements(q: PageQuery = {}) {
  return http.get<PageResult<AnnouncementRow>>('/settings/announcements', q as Record<string, unknown>)
}

/**
 * 发布 / 编辑 / 下线 / 撤回公告：POST /api/settings/announcements
 * 后端契约：{ action: 'save' | 'publish' | 'offline' | 'review' | 'withdraw', announcement: Announcement }
 * - action 别名：create/update→save、发布→publish、下线→offline、撤回→withdraw
 * - 编辑（save）需携带完整公告对象，避免整行覆盖清空 status / type_tone / 计划时间等
 * - 后端为白名单 DTO（announcementBody）：createdBy / createdAt / actionName / stats 不在其中，
 *   编辑时由服务端从原行继承，请求体传了也不生效
 */
const ANN_ACTION_ALIAS: Record<string, string> = {
  create: 'save',
  update: 'save',
  发布: 'publish',
  下线: 'offline',
  撤回: 'withdraw',
}

const ANN_BODY_FIELDS: Record<string, string> = {
  id: 'id',
  title: 'title',
  type: 'type',
  type_tone: 'typeTone',
  audience: 'audience',
  status: 'status',
  content: 'content',
  tags: 'tags',
  planned_at: 'plannedAt',
  published_at: 'publishedAt',
}

export function saveAnnouncement(payload: Record<string, unknown>) {
  const rawAction = String(payload.action ?? 'save')
  const action = ANN_ACTION_ALIAS[rawAction] ?? rawAction
  const announcement: Record<string, unknown> = {}
  for (const [from, to] of Object.entries(ANN_BODY_FIELDS)) {
    if (payload[from] !== undefined) announcement[to] = payload[from]
  }
  return http.post<Record<string, unknown>>('/settings/announcements', { action, announcement })
}
