<script setup lang="ts">
import { computed } from 'vue'
import SvgIcon from './SvgIcon/SvgIcon.vue'
import AppButton from './AppButton.vue'
export type AppInputSize = 'sm' | 'md'
export type AppInputVariant = 'default' | 'plain'

const props = withDefaults(
  defineProps<{
    modelValue?: string | number
    /** 输入类型，默认 text，密码场景传 password */
    type?: string
    placeholder?: string
    /** sm: 32px 紧凑；md: 36px 标准（默认，对照设计稿筛选框） */
    size?: AppInputSize
    /** default: 白底描边聚焦高亮；plain: 浅灰底无边框（如列表搜索框 bg-head） */
    variant?: AppInputVariant
    /** 有内容时显示清空按钮 */
    clearable?: boolean
    disabled?: boolean
    /** 前置图标名（见 SvgIcon/icons.ts 注册表） */
    prefix?: string
    /** 透传给内部 input 的额外类（字号/字距微调用，如需覆盖默认字号请带 ! 前缀） */
    inputClass?: string
  }>(),
  {
    modelValue: '',
    type: 'text',
    placeholder: '',
    size: 'md',
    variant: 'default',
    clearable: false,
    disabled: false,
    prefix: '',
    inputClass: '',
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', v: string): void
  (e: 'clear'): void
  /** 输入框内任意按键（供父级 @keyup.enter 等修饰符监听） */
  (e: 'keyup', ev: KeyboardEvent): void
}>()

const wrapperClass = computed(() => {
  const height = props.size === 'sm' ? 'h-8' : 'h-9'
  const bg =
    props.variant === 'plain'
      ? 'border-transparent bg-head'
      : 'border-line bg-white focus-within:border-primary'
  return `${height} ${bg}`
})

const hasValue = computed(
  () => props.modelValue !== '' && props.modelValue !== null && props.modelValue !== undefined,
)

function onInput(ev: Event) {
  emit('update:modelValue', (ev.target as HTMLInputElement).value)
}

function onKeyup(ev: KeyboardEvent) {
  emit('keyup', ev)
}

function onClear() {
  emit('update:modelValue', '')
  emit('clear')
}
</script>

<template>
  <div
    class="flex items-center gap-2 rounded-lg border px-3 transition-colors"
    :class="wrapperClass"
  >
    <slot name="prefix">
      <SvgIcon v-if="prefix" :name="prefix" :size="15" class="shrink-0 text-ink-4" />
    </slot>
    <input
      :type="type"
      class="h-full min-w-0 flex-1 bg-transparent text-[13px] text-ink-1 outline-none placeholder:text-ink-4"
      :class="inputClass"
      :value="String(modelValue)"
      :placeholder="placeholder"
      :disabled="disabled"
      @input="onInput"
      @keyup="onKeyup"
    />
    <AppButton
      v-if="clearable && hasValue && !disabled"
      variant="ghost"
      icon="close"
      :icon-size="14"
      @click="onClear"
    />
    <slot name="suffix" />
  </div>
</template>
