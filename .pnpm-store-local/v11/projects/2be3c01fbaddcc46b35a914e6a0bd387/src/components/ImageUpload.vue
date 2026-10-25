<script setup lang="ts">
import { ref } from 'vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { uploadMerchantImage } from '@/api/merchants'
import { useToast } from '@/composables/useToast'

defineProps<{
  modelValue: string
  label: string
  hint?: string
  ratio?: 'wide' | 'square'
}>()
const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()
const toast = useToast()
const input = ref<HTMLInputElement | null>(null)
const uploading = ref(false)

async function pick(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (!file.type.startsWith('image/')) {
    toast.warning('请选择图片文件')
    return
  }
  if (file.size > 10 * 1024 * 1024) {
    toast.warning('图片不能超过 10MB')
    return
  }
  uploading.value = true
  try {
    const res = await uploadMerchantImage(file)
    emit('update:modelValue', res.url)
  } finally {
    uploading.value = false
    if (input.value) input.value.value = ''
  }
}
</script>

<template>
  <div>
    <input ref="input" type="file" accept="image/*" class="hidden" @change="pick" />
    <div
      v-if="modelValue"
      class="group relative overflow-hidden rounded-lg border border-line bg-head"
      :class="ratio === 'wide' ? 'aspect-[16/9]' : 'aspect-[4/3]'"
    >
      <img :src="modelValue" alt="" class="h-full w-full object-cover" />
      <button
        type="button"
        class="absolute right-1.5 top-1.5 flex h-6 w-6 items-center justify-center rounded-full bg-black/45 text-white opacity-0 transition-opacity group-hover:opacity-100"
        title="重新上传"
        @click="input?.click()"
      >
        <SvgIcon name="plus" :size="12" />
      </button>
      <button
        type="button"
        class="absolute bottom-1.5 right-1.5 rounded bg-black/45 px-1.5 py-0.5 text-[10px] text-white"
        @click="emit('update:modelValue', '')"
      >
        移除
      </button>
    </div>
    <button
      v-else
      type="button"
      class="flex w-full flex-col items-center justify-center gap-1 rounded-lg border border-dashed border-[#D0D3D9] bg-head text-ink-3 transition-colors hover:border-primary hover:text-primary"
      :class="ratio === 'wide' ? 'h-20' : 'h-28'"
      :disabled="uploading"
      @click="input?.click()"
    >
      <SvgIcon name="plus" :size="18" />
      <span class="text-[12px]">{{ uploading ? '上传中…' : `上传${label}` }}</span>
    </button>
    <div v-if="hint" class="mt-1 text-[11px] text-ink-4">{{ hint }}</div>
  </div>
</template>
