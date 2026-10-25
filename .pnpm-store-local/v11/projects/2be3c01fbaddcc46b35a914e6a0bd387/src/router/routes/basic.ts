import type { RouteRecordRaw } from 'vue-router'

/**
 * 基础路由（骨架 Agent 维护）：
 *  - /login        无布局
 *  - /dashboard    概览
 *  - /403 /404     错误页
 * 其余业务模块路由由后续 Agent 在 router/routes/ 下新增 .ts 文件，自动聚合，勿改本文件。
 */
const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/login/Index.vue'),
    meta: { title: '登录', layout: false, hidden: true },
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: () => import('@/views/dashboard/Index.vue'),
    meta: { title: '数据总览', icon: 'dashboard', group: '概览' },
  },
  {
    path: '/403',
    name: 'forbidden',
    component: () => import('@/views/error/403.vue'),
    meta: { title: '无权限', layout: false, hidden: true },
  },
  {
    path: '/404',
    name: 'notfound',
    component: () => import('@/views/error/404.vue'),
    meta: { title: '页面不存在', layout: false, hidden: true },
  },
]

export default routes
