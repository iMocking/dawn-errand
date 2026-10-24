import { http } from '@/utils/request'
import type { CommonData, ContactData, DownloadData, FaqData, HomeData } from '@/types'

/** 官网数据接口：全部文本数据经 /api/site/* mock 请求渲染 */
export const siteApi = {
  common() {
    return http.get<CommonData>('/api/site/common')
  },
  home() {
    return http.get<HomeData>('/api/site/home')
  },
  download() {
    return http.get<DownloadData>('/api/site/download')
  },
  faq() {
    return http.get<FaqData>('/api/site/faq')
  },
  contact() {
    return http.get<ContactData>('/api/site/contact')
  },
}
