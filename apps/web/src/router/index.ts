import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ left: 0, top: 0 , behavior: 'smooth' }),
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/WebLayout.vue'),
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('@/views/home/HomeView.vue'),
          meta: { title: '首页' },
        },
        {
          path: 'download',
          name: 'download',
          component: () => import('@/views/download/DownloadView.vue'),
          meta: { title: '下载中心' },
        },
        {
          path: 'download/:appKey',
          name: 'download-app',
          component: () => import('@/views/download/AppVersionView.vue'),
          meta: { title: '版本下载' },
        },
        {
          path: 'faq',
          name: 'faq',
          component: () => import('@/views/faq/FaqView.vue'),
          meta: { title: '常见问题' },
        },
        {
          path: 'contact',
          name: 'contact',
          component: () => import('@/views/contact/ContactView.vue'),
          meta: { title: '商务合作' },
        },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.afterEach((to) => {
  const title = to.meta?.title as string | undefined
  document.title = title ? `${title} · 曙光跑腿 Dawn Errand` : '曙光跑腿 · Dawn Errand'
})

export default router
