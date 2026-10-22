import { defineStore } from 'pinia'
import { ref } from 'vue'

/** 全局应用级状态（全局 loading、当前面包屑标题等） */
export const useAppStore = defineStore('app', () => {
  const globalLoading = ref(false)
  const crumbTitle = ref('')

  function setCrumb(title: string) {
    crumbTitle.value = title
  }

  return { globalLoading, crumbTitle, setCrumb }
})
