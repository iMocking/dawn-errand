<script setup lang="ts">
import { useRouter } from 'vue-router'
import AppButton from '@/components/AppButton.vue'
const router = useRouter()
</script>

<template>
  <div class="flex h-screen flex-col items-center justify-center bg-page">
    <div class="text-6xl font-bold text-primary">404</div>
    <div class="mt-3 text-ink-2">抱歉，页面不存在</div>
    <AppButton
      variant="primary"
      class="mt-6 !px-5 !text-sm"
      @click="router.push('/dashboard')"
    >
      返回首页
    </AppButton>
  </div>
</template>
