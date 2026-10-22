<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { formatToday } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const auth = useAuthStore()
const crumb = computed(() => (route.meta?.title as string) ?? '')
</script>

<template>
  <header class="flex h-14 shrink-0 items-center border-b border-line bg-white px-6">
    <div class="flex items-center gap-2 text-[15px] font-semibold text-ink-1">
      {{ crumb }}
      <SvgIcon name="chevron-right" :size="14" class="text-ink-4" />
      <span class="text-[13px] font-normal text-ink-3">{{ formatToday() }}</span>
    </div>

    <div class="ml-auto flex items-center gap-4">
      <div class="flex h-[34px] w-[220px] items-center gap-2 rounded-lg bg-head px-3">
        <SvgIcon name="search" :size="15" class="text-ink-3" />
        <span class="text-[13px] text-ink-4">搜索订单 / 商家 / 骑手</span>
      </div>

      <div class="relative cursor-pointer">
        <SvgIcon name="bell" :size="20" class="text-ink-2" />
        <span
          class="absolute -right-1.5 -top-1 flex h-[15px] min-w-[15px] items-center justify-center rounded-full px-1 text-[10px] font-semibold text-white"
          style="background:#FF4D4F;"
        >
          3
        </span>
      </div>

      <div
        class="flex h-[34px] w-[34px] items-center justify-center rounded-full text-[14px] font-bold text-white"
        style="background:linear-gradient(135deg,#FF8A3D,#FF6B2C);"
      >
        {{ auth.user?.name?.charAt(0) ?? '陈' }}
      </div>
    </div>
  </header>
</template>
