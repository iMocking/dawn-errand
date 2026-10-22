import type { RouteRecordRaw } from 'vue-router'
import { hasPermission } from '@/utils/permission'

/** 按 permissions 过滤异步路由（保留路由结构，去掉无权限项） */
export function filterAsyncRoutes(routes: RouteRecordRaw[], permissions: string[]): RouteRecordRaw[] {
  const result: RouteRecordRaw[] = []
  for (const r of routes) {
    if (!hasPermission(permissions, r.meta?.permission)) continue
    const copy: RouteRecordRaw = { ...r }
    if (r.children) copy.children = filterAsyncRoutes(r.children, permissions)
    result.push(copy)
  }
  return result
}

export interface MenuItem {
  path: string
  title: string
  icon?: string
  badge?: number
  badgeColor?: 'red' | 'orange' | 'yellow'
}

/** 从有权限的路由表生成扁平菜单项（按 meta.group 分组，顺序见 Layout） */
export function buildMenus(routes: RouteRecordRaw[]): MenuItem[] {
  return routes
    .filter((r) => !r.meta?.hidden && r.meta?.title)
    .map((r) => ({
      path: r.path,
      title: r.meta!.title as string,
      icon: r.meta?.icon,
      badge: r.meta?.badge,
      badgeColor: r.meta?.badgeColor,
    }))
}
