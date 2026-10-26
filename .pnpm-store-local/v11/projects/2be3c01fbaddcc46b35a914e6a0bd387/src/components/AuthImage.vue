<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { fetchPrivateBlob, isPrivateSrc } from '@/utils/privateUpload'

/**
 * 证照图片预览：公开桶 URL 直接渲染；私有桶 URL（身份证/营业执照）必须带 Bearer 取 blob，
 * `<img src>` 无法携带请求头，所以这里换成 object URL。
 */
const props = defineProps<{
  src: string
  alt?: string
}>()

const resolved = ref('')
const failed = ref(false)
let activeObjectUrl = ''

async function load(url: string) {
  if (activeObjectUrl) {
    URL.revokeObjectURL(activeObjectUrl)
    activeObjectUrl = ''
  }
  failed.value = false
  if (!url) {
    resolved.value = ''
    return
  }
  if (!isPrivateSrc(url)) {
    resolved.value = url
    return
  }
  try {
    const blob = await fetchPrivateBlob(url)
    activeObjectUrl = URL.createObjectURL(blob)
    resolved.value = activeObjectUrl
  } catch {
    // 403/401/网络失败：显示占位而不是破图，审核人可据此判断是权限问题而非文件丢失
    failed.value = true
    resolved.value = ''
  }
}

watch(() => props.src, (v) => load(v), { immediate: true })

onBeforeUnmount(() => {
  if (activeObjectUrl) URL.revokeObjectURL(activeObjectUrl)
})
</script>

<template>
  <img v-if="resolved" :src="resolved" :alt="alt || ''" />
  <span
    v-else-if="failed"
    class="flex h-full w-full items-center justify-center bg-head text-[11px] text-ink-4"
    title="无读取权限或文件不可用"
  >
    无权查看
  </span>
</template>
