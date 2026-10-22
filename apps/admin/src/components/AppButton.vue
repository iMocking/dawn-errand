<script setup lang="ts">
import { computed } from 'vue'
import SvgIcon from './SvgIcon/SvgIcon.vue'

export type AppButtonVariant = 'primary' | 'secondary' | 'danger' | 'ghost' | 'text'
export type AppButtonSize = 'sm' | 'md'

const props = withDefaults(
  defineProps<{
    /** 按钮风格：primary 主渐变 / secondary 描边 / danger 危险 / ghost 图标幽灵 / text 文字链接 */
    variant?: AppButtonVariant
    /** sm: 28px 紧凑（分页器）；md: 36px 标准（默认） */
    size?: AppButtonSize
    /** 加载中：显示 spinner 并禁用按钮 */
    loading?: boolean
    /** 禁用（loading 时强制禁用） */
    disabled?: boolean
    /** 前置图标名（见 SvgIcon/icons.ts 注册表） */
    icon?: string
    /** 图标尺寸 */
    iconSize?: number
  }>(),
  { variant: 'primary', size: 'md', loading: false, disabled: false, icon: '', iconSize: 15 },
)

const emit = defineEmits<{ (e: 'click', ev: MouseEvent): void }>()

/** 尺寸类：text/ghost 为行内风格，不占固定高度 */
const sizeClass = computed(() => {
  if (props.variant === 'text' || props.variant === 'ghost') return 'h-auto px-1.5'
  return props.size === 'sm' ? 'h-7 px-2.5' : 'h-9 px-4'
})

/** 风格类（对照 design/admin 设计稿与既有实现） */
const variantClass = computed(() => {
  switch (props.variant) {
    case 'secondary':
      return 'border border-line bg-white font-normal text-ink-2 hover:border-primary hover:text-primary'
    case 'danger':
      return 'bg-sem-red font-semibold text-white'
    case 'ghost':
      return 'rounded-md text-ink-3 hover:text-ink-1'
    case 'text':
      return 'rounded-md px-1 font-semibold text-primary hover:text-primary-dark'
    default:
      return 'bg-[linear-gradient(135deg,#FF8A3D,#FF6B2C)] font-semibold text-white'
  }
})

function onClick(ev: MouseEvent) {
  if (props.loading || props.disabled) return
  emit('click', ev)
}
</script>

<template>
  <button
    type="button"
    class="inline-flex select-none items-center justify-center gap-1.5 rounded-lg text-[13px] leading-none transition-colors disabled:cursor-not-allowed disabled:opacity-50"
    :class="[sizeClass, variantClass]"
    :disabled="disabled || loading"
    @click="onClick"
  >
    <span
      v-if="loading"
      class="h-3.5 w-3.5 shrink-0 animate-spin rounded-full border-2 border-current border-t-transparent opacity-40"
      aria-hidden="true"
    />
    <SvgIcon v-else-if="icon" :name="icon" :size="iconSize" class="shrink-0" />
    <slot />
  </button>
</template>
