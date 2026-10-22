<script setup lang="ts">
import { computed } from 'vue'

/**
 * 状态标签（4px 圆角，语义色）。
 * tone: 处理中/待审核→orange；成功/完成/通过→green；失败/拒绝/异常→red；
 *       待处理/待接单→blue；冻结/停用→gray；警告→yellow。
 */
type Tone = 'orange' | 'green' | 'red' | 'blue' | 'gray' | 'yellow'

const props = withDefaults(defineProps<{ tone?: Tone }>(), { tone: 'blue' })

const style = computed(() => {
  const map: Record<Tone, string> = {
    orange: 'background:#FFF3E8;color:#FF6B2C;',
    green: 'background:#E8F7F0;color:#00B578;',
    red: 'background:#FFE8E8;color:#FF4D4F;',
    blue: 'background:#E8F0FF;color:#2B7FFF;',
    gray: 'background:#F2F3F5;color:#86909C;',
    yellow: 'background:#FFF8E0;color:#D48806;',
  }
  return map[props.tone]
})
</script>

<template>
  <span
    class="inline-block rounded-tag px-2 py-0.5 text-xs font-semibold"
    :style="style"
  >
    <slot />
  </span>
</template>
