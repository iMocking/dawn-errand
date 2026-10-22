<script setup lang="ts">
import SvgIcon from './SvgIcon/SvgIcon.vue'

withDefaults(
  defineProps<{
    label: string
    value: string
    subValue?: string
    delta?: string
    deltaColor?: string
    foot: string
    icon: string
    iconBg: string
    iconColor: string
  }>(),
  { subValue: '', delta: '', deltaColor: '#00B578' },
)
</script>

<template>
  <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
    <div
      class="flex h-[46px] w-[46px] shrink-0 items-center justify-center rounded-xl"
      :style="{ background: iconBg }"
    >
      <SvgIcon :name="icon" :size="22" :stroke-width="1.9" :style="{ color: iconColor }" />
    </div>
    <div class="min-w-0 flex-1">
      <div class="text-xs text-ink-3">{{ label }}</div>
      <div class="mt-1 flex items-baseline gap-2">
        <span class="text-2xl font-bold text-ink-1">
          {{ value }}<span v-if="subValue" class="text-[13px] font-normal text-ink-3">{{ subValue }}</span>
        </span>
        <span v-if="delta" class="text-[11px] font-semibold" :style="{ color: deltaColor }">{{ delta }}</span>
      </div>
    </div>
    <div class="shrink-0 text-[11px] text-ink-4" style="writing-mode: vertical-lr;">{{ foot }}</div>
  </div>
</template>
