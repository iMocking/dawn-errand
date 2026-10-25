<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import Sidebar from './Sidebar.vue'
import Topbar from './Topbar.vue'
import { useWorkspaceStore } from '@/stores/workspace'

const workspace = useWorkspaceStore()
let timer: ReturnType<typeof setInterval> | undefined

onMounted(() => {
  workspace.refresh()
  timer = setInterval(() => workspace.refresh(), 60_000)
})
onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="flex h-screen w-screen bg-page">
    <Sidebar />
    <section class="flex min-w-0 flex-1 flex-col">
      <Topbar />
      <main class="flex-1 overflow-y-auto p-2">
        <router-view />
      </main>
    </section>
  </div>
</template>
