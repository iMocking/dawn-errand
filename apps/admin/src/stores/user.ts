import { defineStore } from 'pinia'
import { ref } from 'vue'

/**
 * 用户偏好类局部状态（侧边栏折叠、最近访问等）。
 * 会话身份相关一律用 stores/auth.ts。
 */
export const useUserStore = defineStore('user', () => {
  const sidebarCollapsed = ref(false)

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  return { sidebarCollapsed, toggleSidebar }
})
