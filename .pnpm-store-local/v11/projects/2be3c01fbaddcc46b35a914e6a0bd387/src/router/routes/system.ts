import type { RouteRecordRaw } from 'vue-router'

/**
 * 系统管理 · 系统设置分片
 *  - /settings                   系统设置（服务区域管理）
 *  - /settings/fee-rules         配送费率
 *  - /settings/announcements     公告管理
 */
const routes: RouteRecordRaw[] = [
  {
    path: '/settings',
    name: 'settings',
    component: () => import('@/views/settings/Index.vue'),
    meta: {
      title: '系统设置',
      icon: 'settings',
      group: '系统管理',
      permission: 'settings:view',
    },
  },
  {
    path: '/settings/fee-rules',
    name: 'settings-fee-rules',
    component: () => import('@/views/settings/FeeRules.vue'),
    meta: {
      title: '配送费率',
      icon: 'tag',
      group: '系统管理',
      permission: 'settings:fee',
    },
  },
  {
    path: '/settings/announcements',
    name: 'settings-announcements',
    component: () => import('@/views/settings/Announcements.vue'),
    meta: {
      title: '公告管理',
      icon: 'bell',
      group: '系统管理',
      permission: 'settings:announcement',
    },
  },
]

export default routes
