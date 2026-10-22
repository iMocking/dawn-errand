<script setup lang="ts">
import Sidebar from './Sidebar.vue'
import Topbar from './Topbar.vue'
</script>

<template>
  <div class="flex h-screen min-w-[1440px] bg-page">
    <Sidebar />
    <section class="flex min-w-0 flex-1 flex-col">
      <Topbar />
      <main class="flex-1 overflow-y-auto">
        <router-view />
      </main>
    </section>
  </div>
</template>
