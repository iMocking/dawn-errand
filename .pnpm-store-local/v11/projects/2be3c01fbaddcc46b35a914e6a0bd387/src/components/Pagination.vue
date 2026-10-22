<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    page: number
    pageSize: number
    total: number
  }>(),
  {},
)

const emit = defineEmits<{ (e: 'change', page: number): void }>()

const pages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

const items = computed<(number | '…')[]>(() => {
  const total = pages.value
  const cur = props.page
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const set = new Set<number>([1, 2, total - 1, total, cur - 1, cur, cur + 1])
  const arr = [...set].filter((n) => n >= 1 && n <= total).sort((a, b) => a - b)
  const out: (number | '…')[] = []
  let prev = 0
  for (const n of arr) {
    if (n - prev > 1) out.push('…')
    out.push(n)
    prev = n
  }
  return out
})

function go(n: number) {
  if (n < 1 || n > pages.value || n === props.page) return
  emit('change', n)
}
</script>

<template>
  <div class="flex items-center justify-end gap-1.5 text-[13px]">
    <button
      class="h-7 rounded border border-line px-2.5 text-ink-2 disabled:opacity-40"
      :disabled="page <= 1"
      @click="go(page - 1)"
    >
      上一页
    </button>
    <button
      v-for="(it, i) in items"
      :key="i"
      class="h-7 min-w-7 rounded border px-2"
      :class="it === page ? 'border-primary bg-primary-bg text-primary font-semibold' : 'border-line text-ink-2'"
      :disabled="it === '…'"
      @click="typeof it === 'number' && go(it)"
    >
      {{ it }}
    </button>
    <button
      class="h-7 rounded border border-line px-2.5 text-ink-2 disabled:opacity-40"
      :disabled="page >= pages"
      @click="go(page + 1)"
    >
      下一页
    </button>
    <span class="ml-2 text-ink-3">共 {{ total }} 条</span>
  </div>
</template>
