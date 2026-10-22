import type { Directive, DirectiveBinding } from 'vue'
import { useAuthStore } from '@/stores/auth'

/**
 * v-permission 指令：
 *   v-permission="'merchant:audit'"
 *   v-permission="['merchant:audit','rider:audit']"
 * 无权限时直接从 DOM 移除该元素（不只是 hidden）。
 */
function apply(el: HTMLElement, binding: DirectiveBinding<string | string[] | undefined>) {
  const auth = useAuthStore()
  if (!auth.hasPerm(binding.value)) {
    el.parentNode?.removeChild(el)
  }
}

export const permission: Directive<HTMLElement, string | string[] | undefined> = {
  mounted: apply,
}
