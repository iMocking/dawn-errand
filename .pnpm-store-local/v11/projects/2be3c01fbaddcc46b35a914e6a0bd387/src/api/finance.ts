import { http } from '@/utils/request'
import type { PageQuery, PageResult } from '@/types'

/** 财务结算页 · 提现审核列表行（卡片式列表） */
export type WithdrawRow = {
  /** 提现单号（string） */
  id: string
  /** 商家 / 骑手名称 */
  name: string
  /** 头像底色（设计稿色值） */
  avatarBg: string
  /** 头像字色 */
  avatarColor: string
  /** 类型标签：骑手提现 / 商家结算 / 异常 */
  kind: string
  kindBg: string
  kindColor: string
  /** 提现至：xxx */
  bankInfo: string
  /** 风险提示（异常行红字） */
  riskTip?: string
  /** 驳回原因（已驳回行） */
  rejectReason?: string
  /** 金额（元） */
  amount: number
  /** 异常行金额标红 */
  amountColor?: string
  /** 申请时间 Unix 秒 */
  created_at: number
  /** 待审核 / 已通过 / 已驳回 */
  status: string
}

/** 财务结算页 · 顶部汇总指标 */
export type FinanceStats = {
  todayIncome: number
  todayIncomeDelta: string
  pendingBalance: number
  pendingBalanceBreakdown: string
  withdrawPendingCount: number
  abnormalBills: number
  abnormalBillsAmount: number
}

/** 结算流水页 · 顶部汇总指标 */
export type SettlementStats = {
  todaySettled: number
  todaySettledDelta: string
  weekTotal: number
  weekBreakdown: string
  monthTotal: number
  monthBreakdown: string
  pendingAmount: number
  pendingCountText: string
  dateRange: string
  exportedText: string
}

export type FinanceDto = {
  stats: FinanceStats
  settlementStats: SettlementStats
  withdraws: WithdrawRow[]
}

/** 结算流水页 · 列表行 */
export type SettlementRow = {
  /** 流水号（string） */
  id: string
  /** 结算对象 */
  party: string
  /** 类型：商家结算 / 骑手结算 / 用户退款 */
  type: string
  typeBg: string
  typeColor: string
  /** 金额（元，退款为负） */
  amount: number
  /** 已到账 / 处理中 */
  status: string
  /** 结算周期（展示串） */
  period: string
  /** 结算时间 Unix 秒 */
  settled_at: number
  /** 操作文案 */
  action: string
}

/** 提现审核详情 */
export type WithdrawDetailDto = {
  id: string
  amount: number
  fee: number
  feeNote: string
  status: string
  kind: string
  applicant: string
  applicantNo: string
  created_at: number
  withdrawType: string
  bankInfo: string
  accountHolder: string
  arriveExpect: string
  purpose: string
  balance: number
  balanceNote: string
  usedPercentText: string
  remainBalance: number
  remainNote: string
  monthOrders: number
  monthIncome: number
  lastMonthNote: string
  riskItems: string[]
  riskWarning: string
  timeline: { time: number; text: string; operator: string }[]
}

/** 提现审核提交体 */
export interface AuditPayload {
  action: 'pass' | 'reject'
  remark?: string
}

/** GET /api/finance 财务结算汇总 + 提现审核列表 */
export function getFinance() {
  return http.get<FinanceDto>('/finance')
}

/** GET /api/finance/settlement 结算流水列表 */
export function getSettlements(params: PageQuery) {
  return http.get<PageResult<SettlementRow>>('/finance/settlement', params as Record<string, unknown>)
}

/** GET /api/finance/withdraw/:id 提现审核详情 */
export function getWithdrawDetail(id: string) {
  return http.get<WithdrawDetailDto>(`/finance/withdraw/${id}`)
}

/** POST /api/finance/withdraw/:id/audit 审核打款 / 驳回（dev mock 回显请求体） */
export function auditWithdraw(id: string, payload: AuditPayload) {
  return http.post<Record<string, unknown>>(`/finance/withdraw/${id}/audit`, payload)
}
