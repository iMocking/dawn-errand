<script setup lang="ts">
import { onMounted } from 'vue'
import { useSiteStore } from '@/stores/site'
import SiteNav from '@/components/SiteNav.vue'
import SiteFooter from '@/components/SiteFooter.vue'

const site = useSiteStore()

onMounted(() => {
  site.loadCommon().catch(() => {
    // 公共数据加载失败不阻断页面；导航/页脚为空态由组件容忍
  })
})
</script>

<template>
  <div class="flex min-h-screen flex-col bg-page">
    <SiteNav />
    <main class="flex-1">
      <router-view />
    </main>
    <SiteFooter />
  </div>
</template>
