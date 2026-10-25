import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { siteApi } from '@/api/site'
import type { CommonData, Locale } from '@/types'

const LOCALE_KEY = 'dawn-site-locale'
const LOCALES: Locale[] = ['zh', 'en', 'es']

/** 读取持久化语言（localStorage），非法值回退中文 */
function initialLocale(): Locale {
  const saved = localStorage.getItem(LOCALE_KEY)
  return LOCALES.includes(saved as Locale) ? (saved as Locale) : 'zh'
}

/** 官网公共数据（导航 / 页脚）与多语言状态：全局共享 */
export const useSiteStore = defineStore('site', () => {
  const common = ref<CommonData | null>(null)
  const loaded = ref(false)
  const locale = ref<Locale>(initialLocale())
  document.documentElement.lang = locale.value

  const siteName = computed(() => common.value?.siteName ?? '曙光跑腿')

  async function loadCommon(lang?: Locale): Promise<CommonData> {
    const want = lang ?? locale.value
    if (loaded.value && common.value && want === locale.value) return common.value
    const data = await siteApi.common(want)
    common.value = data
    loaded.value = true
    locale.value = want
    document.documentElement.lang = want
    return data
  }

  /** 切换语言：持久化 + 更新 <html lang> + 重取公共数据（页面数据由各视图响应 locale 变化重新请求） */
  async function setLocale(lang: Locale) {
    if (lang === locale.value) return
    locale.value = lang
    localStorage.setItem(LOCALE_KEY, lang)
    document.documentElement.lang = lang
    common.value = null
    loaded.value = false
    await loadCommon()
  }

  return { common, loaded, locale, siteName, loadCommon, setLocale }
})
