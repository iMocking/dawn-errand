import 'vue-router'

/** 菜单分组固定顺序（与设计稿一致） */
export const MENU_GROUPS = ['概览', '交易管理', '主体管理', '资金管理', '系统管理'] as const

declare module 'vue-router' {
  interface RouteMeta {
    /** 菜单 / 面包屑标题 */
    title?: string
    /** SvgIcon 名称（16px 线框图标） */
    icon?: string
    /** 菜单分组：概览 / 交易管理 / 主体管理 / 资金管理 / 系统管理 */
    group?: (typeof MENU_GROUPS)[number]
    /** 菜单角标数量（不传则无） */
    badge?: number
    /** 角标颜色：red 红 / orange 橙 / yellow 黄，默认橙 */
    badgeColor?: 'red' | 'orange' | 'yellow'
    /** 进入该路由所需权限；不传表示登录即可见 */
    permission?: string | string[]
    /** 不在菜单中显示（详情页等） */
    hidden?: boolean
    /** 是否需要登录布局（默认 true；/login 等置 false） */
    layout?: boolean
  }
}
