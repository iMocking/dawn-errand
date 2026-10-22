import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/merchants',
    name: 'merchants',
    component: () => import('@/views/merchants/Index.vue'),
    meta: {
      title: '商家管理',
      icon: 'merchant',
      group: '主体管理',
      permission: 'merchant:list',
    },
  },
  {
    path: '/merchants/audit',
    name: 'merchant-audit',
    component: () => import('@/views/merchants/Audit.vue'),
    meta: {
      title: '商家入驻审核',
      icon: 'shield',
      group: '主体管理',
      badge: 3,
      badgeColor: 'orange',
      permission: 'merchant:audit',
    },
  },
  {
    path: '/merchants/:id',
    name: 'merchant-detail',
    component: () => import('@/views/merchants/Detail.vue'),
    meta: { title: '商家详情', hidden: true, parent: '/merchants' },
  },
]

export default routes
