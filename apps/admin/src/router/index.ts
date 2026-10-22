import { createRouter, createWebHistory, type Router } from 'vue-router'
import './types'
import { constantRoutes, asyncRoutes } from './routes'
import { filterAsyncRoutes } from '@/permission/filterRoutes'
import { useAuthStore } from '@/stores/auth'

/** 白名单：无需登录即可访问 */
const WHITE_LIST = ['/login', '/403', '/404']

let routesRegistered = false

/** 登录/刷新后，按权限注册动态路由到 Layout 根下 */
export function registerDynamicRoutes(router: Router): void {
  const auth = useAuthStore()
  if (routesRegistered) return
  const accessible = filterAsyncRoutes(asyncRoutes, auth.permissions)
  for (const r of accessible) {
    router.addRoute('layout', r)
  }
  // 404 兜底必须最后注册
  router.addRoute({ path: '/:pathMatch(.*)*', redirect: '/404' })
  routesRegistered = true
}

/** 退出登录时重置：清掉动态路由，恢复初始路由表 */
export function resetRouter(router: Router): void {
  const all = router.getRoutes()
  for (const r of all) {
    if (r.name && !WHITE_LIST.includes(r.path) && r.path !== '/') {
      router.removeRoute(r.name)
    }
  }
  routesRegistered = false
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'layout',
      component: () => import('@/layouts/Index.vue'),
      children: constantRoutes.filter((r) => r.meta?.layout !== false),
      redirect: '/dashboard',
    },
    ...constantRoutes.filter((r) => r.meta?.layout === false),
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()

  if (WHITE_LIST.includes(to.path)) {
    // 已登录访问登录页 → 去首页
    if (to.path === '/login' && auth.isLoggedIn) return '/dashboard'
    return true
  }

  if (!auth.isLoggedIn) {
    return { path: '/login', query: to.fullPath !== '/dashboard' ? { redirect: to.fullPath } : {} }
  }

  if (!routesRegistered) {
    registerDynamicRoutes(router)
    return to.fullPath
  }
  return true
})

export default router
