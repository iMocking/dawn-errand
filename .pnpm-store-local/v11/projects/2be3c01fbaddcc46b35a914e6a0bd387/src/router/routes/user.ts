import type { RouteRecordRaw } from 'vue-router'

/**
 * 用户管理分片路由：
 *  - /users         用户管理列表（主体管理）
 *  - /users/:id     用户详情（hidden，不进菜单）
 */
const routes: RouteRecordRaw[] = [
  {
    path: '/users',
    name: 'users',
    component: () => import('@/views/users/Index.vue'),
    meta: {
      title: '用户管理',
      icon: 'user',
      group: '主体管理',
      permission: 'user:list',
    },
  },
  {
    path: '/users/realname',
    name: 'user-realname',
    component: () => import('@/views/users/RealName.vue'),
    meta: {
      title: '实名认证',
      icon: 'id-card',
      group: '主体管理',
      permission: 'user:audit',
    },
  },
  {
    path: '/users/:id',
    name: 'user-detail',
    component: () => import('@/views/users/Detail.vue'),
    meta: { title: '用户详情', hidden: true, parent: '/users' },
  },
]

export default routes
