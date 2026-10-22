import { reactive } from 'vue'

export type ToastType = 'success' | 'error' | 'info' | 'warning'

export interface ToastItem {
  id: number
  type: ToastType
  message: string
}

const state = reactive<{ list: ToastItem[] }>({ list: [] })
let seq = 0

function push(type: ToastType, message: string, duration = 2600) {
  const id = ++seq
  state.list.push({ id, type, message })
  window.setTimeout(() => {
    const idx = state.list.findIndex((t) => t.id === id)
    if (idx >= 0) state.list.splice(idx, 1)
  }, duration)
}

/** 全局轻提示（不引入组件库）：在 App.vue 挂载 <ToastHost /> */
export function useToast() {
  return {
    state,
    success: (m: string) => push('success', m),
    error: (m: string) => push('error', m),
    info: (m: string) => push('info', m),
    warning: (m: string) => push('warning', m),
  }
}
