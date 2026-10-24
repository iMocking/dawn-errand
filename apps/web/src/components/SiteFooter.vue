<script setup lang="ts">
import { useSiteStore } from '@/stores/site'

const site = useSiteStore()
</script>

<template>
  <footer class="border-t border-line bg-white">
    <div class="mx-auto max-w-6xl px-4 py-10 sm:px-6 md:py-14">
      <div class="grid gap-10 md:grid-cols-4 md:gap-8">
        <!-- 品牌简介 -->
        <div class="md:col-span-1">
          <div class="flex items-center gap-2.5">
            <span
              class="flex h-9 w-9 items-center justify-center rounded-lg bg-brand-gradient text-white"
            >
              <svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true">
                <path d="M10 20c4.5 0 7-2 7-6.5M20 12.5a4 4 0 1 0-8 0 4 4 0 0 0 8 0z" />
              </svg>
            </span>
            <span class="text-[15px] font-semibold text-ink-1">
              {{ site.common?.siteName }}<span class="ml-2 text-xs font-normal text-ink-3">{{ site.common?.siteSub }}</span>
            </span>
          </div>
          <p class="mt-4 text-sm leading-relaxed text-ink-3">{{ site.common?.footer.about }}</p>
        </div>

        <!-- 链接列 -->
        <div
          v-for="col in site.common?.footer.columns"
          :key="col.title"
          class="md:col-span-1"
        >
          <h3 class="text-sm font-semibold text-ink-1">{{ col.title }}</h3>
          <ul class="mt-4 flex flex-col gap-2.5">
            <li v-for="link in col.links" :key="link.label">
              <a
                v-if="link.external"
                :href="link.to"
                target="_blank"
                rel="noopener noreferrer"
                class="text-sm text-ink-2 transition-colors hover:text-primary"
              >
                {{ link.label }}
              </a>
              <router-link
                v-else
                :to="link.to"
                class="text-sm text-ink-2 transition-colors hover:text-primary"
              >
                {{ link.label }}
              </router-link>
            </li>
          </ul>
        </div>
      </div>

      <div class="mt-10 flex flex-col gap-2 border-t border-line pt-6 text-xs text-ink-3 md:flex-row md:items-center md:justify-between">
        <span>{{ site.common?.footer.bottomLeft }}</span>
        <span class="md:text-right">{{ site.common?.footer.bottomRight }}</span>
      </div>
    </div>
  </footer>
</template>
