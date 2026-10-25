import type { RouteRecordRaw } from 'vue-router'

/**
 * 交易管理：订单管理 + 投诉售后。
 * 详情页（:id）hidden 不进菜单，parent 指回列表页。
 */
const routes: RouteRecordRaw[] = [
  {
    path: '/orders',
    name: 'orders',
    component: () => import('@/views/orders/Index.vue'),
    meta: {
      title: '订单管理',
      icon: 'clipboard',
      group: '交易管理',
      permission: 'order:list',
    },
  },
  {
    path: '/orders/:id',
    name: 'order-detail',
    component: () => import('@/views/orders/Detail.vue'),
    meta: { title: '订单详情', hidden: true, parent: '/orders' },
  },
  {
    path: '/complaints',
    name: 'complaints',
    component: () => import('@/views/complaints/Index.vue'),
    meta: {
      title: '投诉售后',
      icon: 'headset',
      group: '交易管理',
      permission: 'complaint:list',
    },
  },
  {
    path: '/complaints/:id',
    name: 'complaint-detail',
    component: () => import('@/views/complaints/Detail.vue'),
    meta: { title: '工单详情', hidden: true, parent: '/complaints' },
  },
]

export default routes
