import { http } from '@/utils/request'
import type { PageQuery, PageResult } from '@/types'

/** 订单列表行 */
export interface OrderRow {
  /** 订单号（不含 #），如 202610220086 */
  id: string
  /** 订单内容（起 → 终） */
  route: string
  /** 下单用户姓名 */
  user_name: string
  /** 用户 / 骑手副行原文，如「骑手：赵铁柱」「待分配骑手」「已取消」 */
  rider: string
  /** 订单类型：外卖 / 帮买 / 帮送 / 帮取 / 帮办 */
  type: string
  /** 金额（元） */
  amount: number
  /** 订单状态：待接单 / 配送中 / 已取件待送达 / 已完成 / 已取消 / 出餐超时 / 异常 */
  status: string
  /** 下单时间（Unix 秒） */
  created_at: number
}

/** 订单状态步骤条节点 */
export interface OrderStep {
  title: string
  /** 完成时间（Unix 秒）；当前进行中 / 未开始节点为空 */
  time?: number
  /** done=已完成勾选；active=当前进行中；pending=未开始 */
  state: 'done' | 'active' | 'pending'
  /** 节点副文案（覆盖时间），如「进行中」「待完成」 */
  sub?: string
}

/** 商品明细行 */
export interface OrderGoods {
  name: string
  spec: string
  price: number
}

/** 费用明细行 */
export interface OrderFee {
  label: string
  value: number
  /** 负值（如优惠券） */
  negative?: boolean
}

/** 订单详情 */
export interface OrderDetail {
  id: string
  /** 平台内部单号 SG202610220086 */
  order_no: string
  status: string
  type: string
  created_at: number
  /** 下单渠道 */
  channel: string
  /** 预计送达（Unix 秒） */
  estimated_arrival: number
  steps: OrderStep[]
  merchant: { name: string; address: string; phone: string; status: string }
  user: { name: string; address: string; phone: string }
  rider: { name: string; vehicle: string; phone: string; service: string; score: string }
  goods: OrderGoods[]
  remark: string
  fees: OrderFee[]
  /** 实付金额（元） */
  pay_amount: number
  pay_method: string
  fund_flow: string
}

/** 订单列表：GET /api/orders?page&pageSize&keyword&status */
export function getOrders(q: PageQuery = {}) {
  return http.get<PageResult<OrderRow>>('/orders', q as Record<string, unknown>)
}

/** 订单详情：GET /api/orders/:id */
export function getOrderDetail(id: string) {
  return http.get<OrderDetail>(`/orders/${id}`)
}
