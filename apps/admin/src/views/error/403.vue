<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()
</script>

<template>
  <div class="flex h-screen flex-col items-center justify-center bg-page">
    <div class="text-6xl font-bold text-primary">403</div>
    <div class="mt-3 text-ink-2">抱歉，你没有权限访问该页面</div>
    <button
      class="mt-6 h-9 rounded-lg border-none px-5 text-sm font-semibold text-white"
      style="background:linear-gradient(135deg,#FF8A3D,#FF6B2C);"
      @click="router.push('/dashboard')"
    >
      返回首页
    </button>
  </div>
</template>
