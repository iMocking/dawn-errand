import { http } from '@/utils/request'
import type { PageQuery, PageResult } from '@/types'

/** 投诉工单列表行（卡片式） */
export interface ComplaintRow {
  /** 工单号，如 GD20261022006 */
  id: string
  /** 分类标签：配送异常 / 商家投诉 / 退款申请 */
  category: string
  /** 工单标题 */
  title: string
  /** 关联订单号（不含 #），如 202610220083 */
  order_id: string
  /** 状态：待处理 / 处理中 / 已完结 */
  status: string
  /** 投诉人原文，如「周桂花（139-7155-8811）」 */
  complainant: string
  /** 责任方原文，如「骑手赵铁柱」」「村口老陈家常菜」 */
  target: string
  /** 副信息（金额 / 等待时长） */
  extra: string
  /** 提交时间（Unix 秒） */
  created_at: number
  /** 用户/骑手描述原文（待处理工单展示） */
  description?: string
  /** 描述块配色变体：red / yellow */
  desc_tone?: 'red' | 'yellow'
  /** 处理时限原文，如「处理时限剩余 1 小时 20 分」 */
  deadline?: string
  /** 当前处理人（处理中工单） */
  handler?: string
  /** 处理进度原文（处理中工单） */
  progress?: string
  /** 进度更新时间（Unix 秒） */
  updated_at?: number
}

/** 处理流程时间线节点 */
export interface ComplaintFlowStep {
  title: string
  desc: string
  /** done=已完成勾选；current=当前（橙色数字）；pending=后续（黄色数字） */
  state: 'done' | 'current' | 'pending'
  /** 节点序号（current/pending 显示） */
  step?: number
}

/** 沟通记录条目 */
export interface ComplaintNote {
  /** 头像首字 */
  avatar: string
  avatar_bg: string
  avatar_color: string
  name: string
  time: number
  content: string
  /** 气泡底色变体：gray / yellow */
  tone: 'gray' | 'yellow'
}

/** 关联订单信息键值对 */
export interface ComplaintOrderItem {
  label: string
  value: string
  /** 高亮配色 */
  tone?: 'blue' | 'primary' | 'normal'
}

/** 投诉工单详情 */
export interface ComplaintDetail {
  id: string
  title: string
  status: string
  priority: string
  created_at: number
  /** 发起人姓名 */
  initiator: string
  initiator_extra: string
  target: string
  order_id: string
  flow: ComplaintFlowStep[]
  notes: ComplaintNote[]
  order: ComplaintOrderItem[]
  pay_amount: number
}

/** 投诉工单列表：GET /api/complaints?page&pageSize&keyword&status */
export function getComplaints(q: PageQuery = {}) {
  return http.get<PageResult<ComplaintRow>>('/complaints', q as Record<string, unknown>)
}

/** 投诉工单详情：GET /api/complaints/:id */
export function getComplaintDetail(id: string) {
  return http.get<ComplaintDetail>(`/complaints/${id}`)
}
