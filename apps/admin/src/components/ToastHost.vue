<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const { state } = useToast()

const bar: Record<string, string> = {
  success: 'background:#00B578;',
  error: 'background:#FF4D4F;',
  warning: 'background:#D48806;',
  info: 'background:#4E5969;',
}
</script>

<template>
  <div class="fixed left-1/2 top-6 z-[9999] -translate-x-1/2 space-y-2">
    <div
      v-for="t in state.list"
      :key="t.id"
      class="rounded-lg bg-white px-4 py-2.5 text-sm text-ink-1 shadow-lg ring-1 ring-line"
      style="min-width:240px;"
    >
      <span class="mr-2 inline-block h-2 w-2 rounded-full" :style="bar[t.type]" />
      {{ t.message }}
    </div>
  </div>
</template>
