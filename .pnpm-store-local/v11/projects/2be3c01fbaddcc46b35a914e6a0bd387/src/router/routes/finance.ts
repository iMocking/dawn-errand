import type { RouteRecordRaw } from 'vue-router'

/**
 * 资金管理 · 财务结算分片
 *  - /finance                 财务结算（汇总 + 提现审核列表，角标 8 = 待审核提现 8 条）
 *  - /finance/settlement      结算流水
 *  - /finance/withdraw/:id    提现审核详情（hidden，父级 /finance）
 */
const routes: RouteRecordRaw[] = [
  {
    path: '/finance',
    name: 'finance',
    component: () => import('@/views/finance/Index.vue'),
    meta: {
      title: '财务结算',
      icon: 'wallet',
      group: '资金管理',
      permission: 'finance:list',
    },
  },
  {
    path: '/finance/settlement',
    name: 'finance-settlement',
    component: () => import('@/views/finance/Settlement.vue'),
    meta: {
      title: '结算流水',
      icon: 'receipt',
      group: '资金管理',
      permission: 'finance:settlement',
    },
  },
  {
    path: '/finance/withdraw/:id',
    name: 'finance-withdraw-detail',
    component: () => import('@/views/finance/WithdrawDetail.vue'),
    meta: { title: '提现审核', hidden: true, parent: '/finance' },
  },
]

export default routes
