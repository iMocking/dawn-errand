import { http } from '@/utils/request'
import type { CommonData, ContactData, DownloadData, FaqData, HomeData, Locale } from '@/types'

/**
 * 官网数据接口：全部文本数据经 /api/site/* mock 请求渲染。
 * 多语言：lang ∈ zh / en / es，经 query 参数下发给 mock 中间件（缺失回退默认语言文件）。
 */
export const siteApi = {
  common(lang?: Locale) {
    return http.get<CommonData>('/api/site/common', { lang })
  },
  home(lang?: Locale) {
    return http.get<HomeData>('/api/site/home', { lang })
  },
  download(lang?: Locale) {
    return http.get<DownloadData>('/api/site/download', { lang })
  },
  faq(lang?: Locale) {
    return http.get<FaqData>('/api/site/faq', { lang })
  },
  contact(lang?: Locale) {
    return http.get<ContactData>('/api/site/contact', { lang })
  },
}
