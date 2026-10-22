<script setup lang="ts">
import SvgIcon from './SvgIcon/SvgIcon.vue'

withDefaults(
  defineProps<{
    open: boolean
    title?: string
    width?: number
    confirmText?: string
    cancelText?: string
    loading?: boolean
  }>(),
  { title: '', width: 480, confirmText: '确定', cancelText: '取消', loading: false },
)

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()

function close() {
  emit('update:open', false)
  emit('cancel')
}
</script>

<template>
  <teleport to="body">
    <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center" style="background:rgba(0,0,0,.4);">
      <div
        class="rounded-card bg-white shadow-xl"
        :style="{ width: `${width}px`, maxWidth: '92vw' }"
      >
        <div class="flex items-center justify-between border-b border-line px-5 py-3.5">
          <div class="text-[15px] font-semibold text-ink-1">{{ title }}</div>
          <button class="text-ink-3 hover:text-ink-1" @click="close">
            <SvgIcon name="close" :size="16" />
          </button>
        </div>
        <div class="px-5 py-4">
          <slot />
        </div>
        <div class="flex justify-end gap-2.5 border-t border-line px-5 py-3">
          <button
            class="h-8 rounded-lg border border-line bg-white px-4 text-[13px] text-ink-2 hover:border-primary hover:text-primary"
            @click="close"
          >
            {{ cancelText }}
          </button>
          <button
            class="h-8 rounded-lg border-none px-4 text-[13px] font-semibold text-white disabled:opacity-60"
            :style="{ background: 'linear-gradient(135deg,#FF8A3D,#FF6B2C)' }"
            :disabled="loading"
            @click="emit('confirm')"
          >
            {{ loading ? '提交中…' : confirmText }}
          </button>
        </div>
      </div>
    </div>
  </teleport>
</template>
