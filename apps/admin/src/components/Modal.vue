<script setup lang="ts">
import AppButton from './AppButton.vue'

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
          <AppButton variant="ghost" icon="close" :icon-size="16" class="!p-0" @click="close" />
        </div>
        <div class="px-5 py-4">
          <slot />
        </div>
        <div class="flex justify-end gap-2.5 border-t border-line px-5 py-3">
          <AppButton variant="secondary" class="!h-8" @click="close">
            {{ cancelText }}
          </AppButton>
          <AppButton variant="primary" class="!h-8" :loading="loading" @click="emit('confirm')">
            {{ loading ? '提交中…' : confirmText }}
          </AppButton>
        </div>
      </div>
    </div>
  </teleport>
</template>
