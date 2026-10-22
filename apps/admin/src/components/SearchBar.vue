<script setup lang="ts">
import { ref } from 'vue'
import AppButton from './AppButton.vue'
import AppInput from './AppInput.vue'

withDefaults(
  defineProps<{
    keyword?: string
    placeholder?: string
  }>(),
  { keyword: '', placeholder: '搜索关键字' },
)

const emit = defineEmits<{
  (e: 'search', keyword: string): void
  (e: 'reset'): void
}>()

const kw = ref('')
</script>

<template>
  <div class="flex items-center gap-2.5">
    <AppInput
      v-model="kw"
      variant="plain"
      prefix="search"
      class="w-60"
      :placeholder="placeholder"
      @keyup.enter="emit('search', kw)"
    />
    <AppButton variant="primary" @click="emit('search', kw)">
      查询
    </AppButton>
    <AppButton variant="secondary" @click="((kw = ''), emit('reset'))">
      重置
    </AppButton>
    <slot name="extra" />
  </div>
</template>
