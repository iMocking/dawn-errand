import type { RouteRecordRaw } from 'vue-router'

/**
 * 自动聚合 router/routes/ 下所有分模块路由文件。
 * 后续 Agent 只需新增 routes/xxx.ts 并 export default RouteRecordRaw[]，无需改动本文件。
 *
 * 拆分规则：
 *  - meta.layout === false（/login、/403、/404）→ constantRoutes，启动即注册；
 *  - 其余 → asyncRoutes，登录后按权限 addRoute 挂到 Layout 下。
 */
const modules = import.meta.glob('./*.ts', { eager: true }) as Record<
  string,
  { default: RouteRecordRaw[] }
>

const all: RouteRecordRaw[] = []
for (const [path, mod] of Object.entries(modules)) {
  if (path.endsWith('/index.ts')) continue
  if (!mod?.default) continue
  all.push(...mod.default)
}

export const constantRoutes: RouteRecordRaw[] = all.filter((r) => r.meta?.layout === false)
export const asyncRoutes: RouteRecordRaw[] = all.filter((r) => r.meta?.layout !== false)
