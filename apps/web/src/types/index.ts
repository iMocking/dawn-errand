/** 统一后端响应包：{message, body, request_id, code}，code===0 成功 */
export interface ApiResponse<T = unknown> {
  message: string
  body: T
  request_id: string
  code: number
}

/** 通用链接 */
export interface SiteLink {
  label: string
  to: string
  external?: boolean
}

/** 导航 CTA 样式 */
export type CtaStyle = 'primary' | 'ghost' | 'outline' | 'white'

/** 公共数据（导航 / 页脚） */
export interface CommonData {
  siteName: string
  siteSub: string
  navLinks: SiteLink[]
  navCta: Array<SiteLink & { style: CtaStyle }>
  footer: {
    about: string
    columns: Array<{ title: string; links: SiteLink[] }>
    bottomLeft: string
    bottomRight: string
  }
}

/** 图标名称（SvgIcon 字典 key） */
export type IconName =
  | 'clock'
  | 'pin'
  | 'order'
  | 'delivery'
  | 'track'
  | 'finance'
  | 'role'
  | 'region'
  | 'store'
  | 'rider'
  | 'chart'
  | 'van'
  | 'food'
  | 'check'
  | 'box'
  | 'card'
  | 'shield'
  | 'lock'
  | 'phone'
  | 'shop'
  | 'bike'
  | 'console'
  | 'mail'
  | 'tel'
  | 'chat'
  | 'handshake'
  | 'download'
  | 'arrow'
  | 'plus'
  | 'menu'
  | 'android'
  | 'ios'
  | 'harmony'

/** 首页 */
export interface HomeData {
  hero: {
    titleLines: string[]
    lead: string
    stats: Array<{ value: string; label: string }>
    devicesLabel: string
  }
  capabilities: {
    kicker: string
    title: string
    desc: string
    items: Array<{ icon: IconName; title: string; desc: string }>
  }
  scenarios: {
    kicker: string
    title: string
    desc: string
    items: Array<{ num: string; icon: IconName; title: string; desc: string }>
  }
  features: {
    kicker: string
    title: string
    desc: string
    items: Array<{ icon: IconName; tone: string; title: string; desc: string }>
  }
  pipeline: {
    title: string
    nodes: Array<{ name: string; tag: string; strong?: boolean }>
  }
  ends: {
    kicker: string
    title: string
    desc: string
    items: Array<{
      icon: IconName
      tone: string
      name: string
      role: string
      desc: string
      linkLabel: string
      to: string
    }>
  }
  roadmap: {
    kicker: string
    title: string
    desc: string
    items: Array<{
      time: string
      title: string
      desc: string
      state: 'done' | 'plan'
      stateLabel: string
    }>
  }
  faqPreview: {
    kicker: string
    title: string
    desc: string
    items: Array<{ q: string; a: string }>
    moreLabel: string
    moreTo: string
  }
  cta: {
    title: string
    desc: string
    primary: SiteLink
    secondary: SiteLink
  }
}

/** 下载中心 */
export interface DownloadData {
  pageHead: { kicker: string; title: string; desc: string }
  androidSection: { title: string; desc: string }
  apps: Array<{
    key: string
    icon: IconName
    tone: string
    name: string
    role: string
    desc: string
    meta: string[]
    btnLabel: string
    linkLabel: string
  }>
  adminSection: {
    title: string
    desc: string
    entry: {
      icon: IconName
      name: string
      role: string
      desc: string
      btnLabel: string
      demo: string
    }
  }
  scanNote: string
  appVersions: AppVersionData[]
}

/** 平台标识：安卓 / iOS / 鸿蒙 */
export type PlatformId = 'android' | 'ios' | 'harmony'

/** 平台卡信息 */
export interface AppPlatform {
  id: PlatformId
  label: string
  note: string
  state: 'ready' | 'coming'
}

/** 单条历史版本（packages 标记各平台是否可用） */
export interface AppVersion {
  version: string
  date: string
  note: string
  packages: Record<PlatformId, boolean>
}

/** 单端版本子页数据 */
export interface AppVersionData {
  key: string
  name: string
  sub: string
  icon: IconName
  tone: string
  platforms: AppPlatform[]
  versions: AppVersion[]
}

/** 常见问题 */
export interface FaqData {
  pageHead: { kicker: string; title: string; desc: string }
  items: Array<{ q: string; a: string }>
  note: string
  noteAction: string
  noteTo: string
}

/** 商务合作 */
export interface ContactData {
  pageHead: { kicker: string; title: string; desc: string }
  modes: {
    title: string
    desc: string
    items: Array<{
      id?: string
      icon: IconName
      tone: string
      title: string
      desc: string
      tag: string
    }>
  }
  steps: {
    title: string
    desc: string
    items: Array<{ num: string; title: string; desc: string }>
  }
  contactSection: {
    title: string
    desc: string
    cards: Array<{ icon: IconName; label: string; value: string }>
    form: {
      fields: Array<{
        name: string
        label: string
        type: 'text' | 'select' | 'textarea'
        placeholder?: string
        options?: string[]
      }>
      submitLabel: string
      submitTip: string
      demo: string
    }
  }
  notice: string
}
