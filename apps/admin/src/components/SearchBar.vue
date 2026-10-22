<script setup lang="ts">
import SvgIcon from './SvgIcon/SvgIcon.vue'

withDefaults(
  defineProps<{
    keyword?: string
    placeholder?: string
  }>(),
  { keyword: '', placeholder: '搜索关键字' },
)

const emit = defineEmits<{
  (e: 'search', keyword: string): void
  (e: 'reset'): void
}>()

import { ref } from 'vue'
const kw = ref('')
</script>

<template>
  <div class="flex items-center gap-2.5">
    <div class="flex h-[34px] w-60 items-center gap-2 rounded-lg bg-head px-3">
      <SvgIcon name="search" :size="14" class="text-ink-4" />
      <input
        v-model="kw"
        class="flex-1 bg-transparent text-[13px] outline-none placeholder:text-ink-4"
        :placeholder="placeholder"
        @keyup.enter="emit('search', kw)"
      />
    </div>
    <button
      class="h-[34px] rounded-lg border-none px-4 text-[13px] font-semibold text-white"
      :style="{ background: 'linear-gradient(135deg,#FF8A3D,#FF6B2C)' }"
      @click="emit('search', kw)"
    >
      查询
    </button>
    <button
      class="h-[34px] rounded-lg border border-line bg-white px-4 text-[13px] text-ink-2"
      @click="((kw = ''), emit('reset'))"
    >
      重置
    </button>
    <slot name="extra" />
  </div>
</template>
