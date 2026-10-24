<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useSiteStore } from '@/stores/site'
import SvgIcon from './SvgIcon.vue'

const route = useRoute()
const site = useSiteStore()

const menuOpen = ref(false)
watch(
  () => route.path,
  () => {
    menuOpen.value = false
  },
)

function isActive(to: string): boolean {
  if (to === '/') return route.path === '/'
  return route.path.startsWith(to)
}

const logoMark = 'M10 20c4.5 0 7-2 7-6.5M20 12.5a4 4 0 1 0-8 0 4 4 0 0 0 8 0z'
</script>

<template>
  <header
    class="sticky top-0 z-50 border-b border-line bg-white/90 backdrop-blur transition-shadow"
    :class="menuOpen ? 'shadow-card' : ''"
  >
    <div class="mx-auto flex h-16 max-w-6xl items-center justify-between px-4 sm:px-6">
      <!-- Logo -->
      <router-link to="/" class="flex min-w-0 items-center gap-2.5">
        <span
          class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-brand-gradient text-white"
        >
          <svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true">
            <path :d="logoMark" />
          </svg>
        </span>
        <span class="flex min-w-0 flex-col leading-tight">
          <span class="truncate text-[15px] font-semibold text-ink-1">{{ site.common?.siteName }}</span>
          <span class="truncate text-[11px] tracking-wide text-ink-3">{{ site.common?.siteSub }}</span>
        </span>
      </router-link>

      <!-- 桌面导航 -->
      <nav class="hidden items-center gap-1 md:flex">
        <router-link
          v-for="link in site.common?.navLinks"
          :key="link.to"
          :to="link.to"
          class="rounded-lg px-3.5 py-2 text-sm text-ink-2 transition-colors hover:bg-primary-bg hover:text-primary"
          :class="isActive(link.to) ? 'bg-primary-bg font-medium text-primary' : ''"
        >
          {{ link.label }}
        </router-link>
      </nav>

      <div class="hidden items-center gap-3 md:flex">
        <router-link
          v-for="cta in site.common?.navCta"
          :key="cta.to"
          :to="cta.to"
          class="rounded-lg px-4 py-2 text-sm font-medium transition-colors"
          :class="
            cta.style === 'primary'
              ? 'bg-brand-gradient text-white shadow-brand hover:shadow-brand-hover'
              : 'border border-primary bg-white text-primary hover:bg-primary-bg'
          "
        >
          {{ cta.label }}
        </router-link>
      </div>

      <!-- 移动端汉堡按钮（热区 44px） -->
      <button
        class="flex h-11 w-11 items-center justify-center rounded-lg text-ink-2 transition-colors hover:bg-head md:hidden"
        :aria-expanded="menuOpen"
        aria-label="打开导航菜单"
        @click="menuOpen = !menuOpen"
      >
        <SvgIcon :name="menuOpen ? 'plus' : 'menu'" :size="22" />
      </button>
    </div>

    <!-- 移动端下拉菜单 -->
    <transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="-translate-y-2 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="-translate-y-2 opacity-0"
    >
      <div v-if="menuOpen" class="border-t border-line bg-white px-4 py-3 md:hidden">
        <nav class="flex flex-col gap-1">
          <router-link
            v-for="link in site.common?.navLinks"
            :key="link.to"
            :to="link.to"
            class="rounded-lg px-3 py-3 text-[15px] text-ink-1 transition-colors hover:bg-primary-bg hover:text-primary"
            :class="isActive(link.to) ? 'bg-primary-bg font-medium text-primary' : ''"
          >
            {{ link.label }}
          </router-link>
        </nav>
        <div class="mt-2 flex flex-col gap-2 border-t border-line pt-3">
          <router-link
            v-for="cta in site.common?.navCta"
            :key="cta.to"
            :to="cta.to"
            class="flex min-h-11 items-center justify-center rounded-lg px-4 text-[15px] font-medium"
            :class="
              cta.style === 'primary'
                ? 'bg-brand-gradient text-white shadow-brand'
                : 'border border-primary bg-white text-primary'
            "
          >
            {{ cta.label }}
          </router-link>
        </div>
      </div>
    </transition>
  </header>
</template>
