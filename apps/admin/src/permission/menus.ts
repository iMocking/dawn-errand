import { computed, type ComputedRef } from 'vue'
import { useRoute } from 'vue-router'
import { asyncRoutes } from '@/router/routes'
import { useAuthStore } from '@/stores/auth'
import { MENU_GROUPS } from '@/router/types'

export interface MenuEntry {
  path: string
  title: string
  icon?: string
  badge?: number
  badgeColor?: 'red' | 'orange' | 'yellow'
}

/**
 * 从路由表生成分组菜单（顺序：概览/交易管理/主体管理/资金管理/系统管理）。
 * 与权限路由一致：无权限的路由不会出现在这里。
 */
export function useMenus(): ComputedRef<{ group: string; items: MenuEntry[] }[]> {
  const auth = useAuthStore()

  return computed(() => {
    const buckets = new Map<string, MenuEntry[]>()
    for (const g of MENU_GROUPS) buckets.set(g, [])

    for (const r of asyncRoutes) {
      if (r.meta?.hidden) continue
      const perm = r.meta?.permission
      if (perm && !auth.hasPerm(perm)) continue
      const group = (r.meta?.group as string) || '概览'
      if (!buckets.has(group)) buckets.set(group, [])
      buckets.get(group)!.push({
        path: r.path,
        title: r.meta?.title ?? r.path,
        icon: r.meta?.icon,
        badge: r.meta?.badge,
        badgeColor: r.meta?.badgeColor,
      })
    }

    return MENU_GROUPS.map((g) => ({ group: g, items: buckets.get(g) ?? [] })).filter(
      (b) => b.items.length > 0,
    )
  })
}

export function useActiveMenu(): ComputedRef<string> {
  const route = useRoute()
  return computed(() => route.path)
}
