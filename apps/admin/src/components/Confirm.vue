<script setup lang="ts">
import Modal from './Modal.vue'

withDefaults(
  defineProps<{
    open: boolean
    title?: string
    message?: string
    confirmText?: string
  }>(),
  { title: '操作确认', message: '确认执行该操作吗？', confirmText: '确认' },
)

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()
</script>

<template>
  <Modal
    :open="open"
    :title="title"
    :confirm-text="confirmText"
    :width="400"
    @update:open="emit('update:open', $event)"
    @confirm="emit('confirm')"
    @cancel="emit('cancel')"
  >
    <div class="text-sm text-ink-2">{{ message }}</div>
  </Modal>
</template>
