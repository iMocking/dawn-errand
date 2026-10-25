import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/merchants',
    name: 'merchants',
    component: () => import('@/views/merchants/Index.vue'),
    meta: {
      title: '商家管理',
      icon: 'storefront',
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
      icon: 'storefront-check',
      group: '主体管理',
      permission: 'merchant:audit',
    },
  },
  {
    path: '/merchants/new',
    name: 'merchant-create',
    component: () => import('@/views/merchants/Create.vue'),
    meta: { title: '新增商家', hidden: true, parent: '/merchants', permission: 'merchant:create' },
  },
  {
    path: '/merchants/:id/edit',
    name: 'merchant-edit',
    component: () => import('@/views/merchants/Create.vue'),
    meta: { title: '编辑入驻资料', hidden: true, parent: '/merchants', permission: 'merchant:create' },
  },
  {
    path: '/merchants/:id',
    name: 'merchant-detail',
    component: () => import('@/views/merchants/Detail.vue'),
    meta: { title: '商家详情', hidden: true, parent: '/merchants' },
  },
]

export default routes
