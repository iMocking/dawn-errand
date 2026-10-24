import { defineStore } from 'pinia'
import { ref } from 'vue'
import { siteApi } from '@/api/site'
import type { CommonData } from '@/types'

/** 官网公共数据（导航 / 页脚）：一次请求，全局共享 */
export const useSiteStore = defineStore('site', () => {
  const common = ref<CommonData | null>(null)
  const loaded = ref(false)

  async function loadCommon(): Promise<CommonData> {
    if (loaded.value && common.value) return common.value
    const data = await siteApi.common()
    common.value = data
    loaded.value = true
    return data
  }

  return { common, loaded, loadCommon }
})
