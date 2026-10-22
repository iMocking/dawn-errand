import type { RouteRecordRaw } from 'vue-router'

/**
 * 骑手管理分片路由：
 *  - /riders        骑手管理列表
 *  - /riders/audit  骑手入驻审核（角标 5）
 *  - /riders/:id    骑手详情（hidden，不进菜单）
 */
const routes: RouteRecordRaw[] = [
  {
    path: '/riders',
    name: 'riders',
    component: () => import('@/views/riders/Index.vue'),
    meta: {
      title: '骑手管理',
      icon: 'grid',
      group: '主体管理',
      permission: 'rider:list',
    },
  },
  {
    path: '/riders/audit',
    name: 'rider-audit',
    component: () => import('@/views/riders/Audit.vue'),
    meta: {
      title: '骑手入驻审核',
      icon: 'rider-audit',
      group: '主体管理',
      badge: 5,
      badgeColor: 'orange',
      permission: 'rider:audit',
    },
  },
  {
    path: '/riders/:id',
    name: 'rider-detail',
    component: () => import('@/views/riders/Detail.vue'),
    meta: { title: '骑手详情', hidden: true, parent: '/riders' },
  },
]

export default routes
