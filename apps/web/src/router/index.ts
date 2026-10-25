import { createRouter, createWebHistory } from 'vue-router'
import { useSiteStore } from '@/stores/site'

/** 页面标题映射 key（对应 common.ui.titleMap，随语言切换） */
export type TitleKey = 'home' | 'download' | 'downloadApp' | 'faq' | 'contact'

/** 按当前语言与路由刷新文档标题 */
function applyTitle(to: { meta?: unknown }) {
  const site = useSiteStore()
  const meta = to.meta as { titleKey?: TitleKey } | undefined
  const key = meta?.titleKey ?? 'home'
  const title = site.common?.ui.titleMap?.[key] ?? ''
  const siteName = site.common?.siteName ?? '曙光跑腿'
  document.title = title ? `${title} · ${siteName} Dawn Errand` : `曙光跑腿 · Dawn Errand`
}

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ left: 0, top: 0, behavior: 'smooth' }),
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/WebLayout.vue'),
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('@/views/home/HomeView.vue'),
          meta: { titleKey: 'home' },
        },
        {
          path: 'download',
          name: 'download',
          component: () => import('@/views/download/DownloadView.vue'),
          meta: { titleKey: 'download' },
        },
        {
          path: 'download/:appKey',
          name: 'download-app',
          component: () => import('@/views/download/AppVersionView.vue'),
          meta: { titleKey: 'downloadApp' },
        },
        {
          path: 'faq',
          name: 'faq',
          component: () => import('@/views/faq/FaqView.vue'),
          meta: { titleKey: 'faq' },
        },
        {
          path: 'contact',
          name: 'contact',
          component: () => import('@/views/contact/ContactView.vue'),
          meta: { titleKey: 'contact' },
        },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.afterEach((to) => {
  applyTitle(to)
})

/** 按当前路由语言刷新标题（语言切换后由入口调用） */
export function refreshTitle() {
  applyTitle(router.currentRoute.value)
}

export default router
