<script setup lang="ts">
import { computed } from 'vue'
import type { IconName } from '@/types'

const props = withDefaults(
  defineProps<{
    name: IconName
    size?: number
  }>(),
  { size: 20 },
)

/** 内嵌 SVG 线框图标字典（1.5 描边，与设计稿图标语言一致） */
const ICONS: Record<IconName, string> = {
  clock: '<circle cx="12" cy="12" r="8.5"/><path d="M12 7.5V12l3 2"/>',
  pin: '<path d="M12 21s-6.5-5.2-6.5-10a6.5 6.5 0 1 1 13 0c0 4.8-6.5 10-6.5 10z"/><circle cx="12" cy="11" r="2.4"/>',
  order: '<path d="M6 3.5h12l1.5 17h-15L6 3.5z"/><path d="M9 8.5h6M9 12.5h6M9 16.5h4"/>',
  delivery: '<circle cx="6" cy="17.5" r="2.4"/><circle cx="17" cy="17.5" r="2.4"/><path d="M8.4 17.5h6.2M4.5 8.5h7v6h-9zM11.5 8.5h3.6l3 4.5H15"/><path d="M2.5 11h2"/>',
  track: '<path d="M4 6.5h10.5v8H4z"/><path d="M14.5 9.5H17l3 3v2h-5.5z"/><circle cx="8" cy="11.5" r="1.4"/><circle cx="15.5" cy="16.5" r="1.4"/>',
  finance: '<rect x="3.5" y="5.5" width="17" height="13" rx="2"/><path d="M3.5 9.5h17"/><path d="M7 14h4"/>',
  role: '<circle cx="12" cy="8" r="3.4"/><path d="M5 19.5c.8-3.4 3.6-5.2 7-5.2s6.2 1.8 7 5.2"/>',
  region: '<path d="M3.5 12h5l2.5-7 3 11 2-4h4.5"/>',
  store: '<path d="M4.5 10.5h15l-1 9h-13z"/><path d="M4.5 10.5l1.5-5h12l1.5 5"/><path d="M9.5 19.5v-5h5v5"/>',
  rider: '<circle cx="8" cy="6.5" r="3"/><path d="M5.5 12h5l2 4 3-3h4M13 10.5l-1.5 1.5-1-3.5M6 12.5l2 3.5h2.5M12.5 13h5M14 13v3.5"/>',
  chart: '<path d="M4 19.5h16"/><path d="M6.5 15.5v-4M11 15.5V8.5M15.5 15.5v-6M20 15.5V6.5"/>',
  van: '<circle cx="6" cy="16" r="2.2"/><circle cx="17.5" cy="16" r="2.2"/><path d="M8.2 16h7.1M3 9h8v5.5H3zM11 9h3.5L17 13h-6zM3 11.5h8"/>',
  food: '<path d="M5 3.5v7.5h3.5v-7.5M6.75 3.5V1M5 13.5h3.5V21M16 3.5c-2 1-3.5 3.5-3.5 6 0 2.5 1.5 4.5 3.5 4.5s3.5-2 3.5-4.5c0-2.5-1.5-5-3.5-6z"/><path d="M16 3.5V21"/>',
  check: '<path d="M4.5 12.5l5 5 10-11"/>',
  box: '<path d="M12 3.5l8 4v9l-8 4-8-4v-9l8-4z"/><path d="M4.5 7.5l7.5 4 7.5-4M12 11.5v9"/>',
  card: '<rect x="3.5" y="6" width="17" height="12" rx="2"/><path d="M3.5 10h17M7 14h4"/>',
  shield: '<path d="M12 3.5l7 2.5v5.5c0 4.5-3.5 7.8-7 9-3.5-1.2-7-4.5-7-9V6l7-2.5z"/><path d="M9 12l2 2 4-4.5"/>',
  lock: '<rect x="5" y="10.5" width="14" height="9.5" rx="2"/><path d="M8.5 10.5V8a3.5 3.5 0 0 1 7 0v2.5"/><circle cx="12" cy="15" r="1.4"/>',
  phone: '<rect x="7" y="3.5" width="10" height="17" rx="2.4"/><path d="M11 18.5h2"/>',
  shop: '<path d="M4.5 9.5h15l-1 8.5h-13z"/><path d="M4.5 9.5l1.5-4.5h12l1.5 4.5"/><path d="M9.5 18v-4.5h5V18"/>',
  bike: '<circle cx="7.5" cy="16.5" r="3"/><circle cx="16.5" cy="16.5" r="3"/><path d="M10 16.5h3.6L17 12.5H13.8L10 16.5zM13.8 12.5l-1.3-3h-2.6"/><path d="M8.5 6.5h2.5"/>',
  console: '<rect x="3.5" y="5" width="17" height="10" rx="2"/><path d="M7 11.5l3 2.5-3 2.5M12 16.5h4"/><path d="M3.5 19.5h17"/>',
  mail: '<rect x="3.5" y="5.5" width="17" height="13" rx="2"/><path d="M4.5 8l7.5 5.5L19.5 8"/>',
  tel: '<path d="M7.5 4h3l1.5 4.5-2 1.5a12 12 0 0 0 4.5 4.5l1.5-2 4.5 1.5v3a2 2 0 0 1-2 2A15.5 15.5 0 0 1 5.5 6a2 2 0 0 1 2-2z"/>',
  chat: '<path d="M4 6.5h16v9.5H9.5L5.5 19v-3H4V6.5z"/><path d="M8 10.5h8M8 13.5h5"/>',
  handshake: '<path d="M7 7.5l4.5 3.5 4-3 3.5 2.5v4.5H15.5l-3 3-3-3H6v-4z"/><path d="M3.5 11.5h2.5v4h-2.5zM17 14h3.5v4H17zM11.5 11l-1.5-2 2-1.5 1.5 2"/>',
  download: '<path d="M12 4v9.5M7.5 10.5L12 15l4.5-4.5"/><path d="M4.5 19h15"/>',
  arrow: '<path d="M4.5 12h15M13 6l6 6-6 6"/>',
  plus: '<path d="M12 5v14M5 12h14"/>',
  menu: '<path d="M4.5 6.5h15M4.5 12h15M4.5 17.5h15"/>',
  android: '<circle cx="8" cy="10.5" r="1.3"/><circle cx="16" cy="10.5" r="1.3"/><path d="M12 9.5c-3.5 0-6.5 1.4-6.5 3.6v3.4h13v-3.4c0-2.2-3-3.6-6.5-3.6z"/><path d="M6.5 5.5L5 3.8M17.5 5.5L19 3.8"/><path d="M5.5 19.5h13"/>',
  ios: '<path d="M12 7.2c1.8-1.6 4.2-1.8 5.5-.6.8.8.6 2.3-.2 3.8-1.3 2.2-4.1 2.8-5.3 2.8s-4-.6-5.3-2.8c-.8-1.5-1-3-.2-3.8 1.3-1.2 3.7-1 5.5.6z"/><path d="M12 7.2v-1.7c0-1.4 1-2.2 2.2-2.2"/>',
  harmony: '<circle cx="12" cy="12" r="8.5"/><path d="M9 7.5v9M15 7.5v9M9 12h6"/>',
}

const inner = computed(() => ICONS[props.name])
</script>

<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.5"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
    v-html="inner"
  />
</template>
