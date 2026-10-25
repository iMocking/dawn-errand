import { http } from '@/utils/request'

/** 左侧菜单待办角标（真实接口数据）。 */
export interface BadgeCounts {
  complaints: number
  merchantAudits: number
  riderAudits: number
  withdrawals: number
}

/** 消息中心条目。 */
export interface WorkspaceMessage {
  id: string
  type: 'complaint' | 'merchant' | 'rider' | 'withdraw' | 'system'
  title: string
  content: string
  link: string
  level: 'warning' | 'info'
  createdAt: number
}

export interface MessagesResult {
  list: WorkspaceMessage[]
  badges: BadgeCounts
  total: number
}

/** GET /api/messages：待办消息 + 角标。 */
export function getMessages() {
  return http.get<MessagesResult>('/messages')
}
