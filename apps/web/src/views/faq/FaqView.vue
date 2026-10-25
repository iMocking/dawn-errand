<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { siteApi } from '@/api/site'
import { useSiteStore } from '@/stores/site'
import type { FaqData } from '@/types'
import SvgIcon from '@/components/SvgIcon.vue'

const site = useSiteStore()
const data = ref<FaqData | null>(null)
const error = ref(false)
const openIdx = ref(0)

async function load() {
  try {
    data.value = await siteApi.faq(site.locale)
  } catch {
    error.value = true
  }
}

onMounted(load)
watch(
  () => site.locale,
  () => {
    error.value = false
    load()
  },
)

function toggle(idx: number) {
  openIdx.value = openIdx.value === idx ? -1 : idx
}
</script>

<template>
  <div v-if="error" class="mx-auto max-w-3xl px-4 py-24 text-center">
    <p class="text-base text-ink-2">{{ site.common?.ui.loadError ?? '内容加载失败，请稍后刷新重试。' }}</p>
  </div>

  <template v-else-if="data">
    <!-- 页头 -->
    <section class="page-head">
      <div class="wrap">
        <span class="kicker">{{ data.pageHead.kicker }}</span>
        <h1>{{ data.pageHead.title }}</h1>
        <p>{{ data.pageHead.desc }}</p>
      </div>
    </section>

    <section class="block">
      <div class="wrap">
        <div class="faq-col">
          <div
            v-for="(item, i) in data.items"
            :key="item.q"
            class="faq-item"
          >
            <button
              type="button"
              class="faq-q"
              :aria-expanded="openIdx === i"
              @click="toggle(i)"
            >
              <span>{{ item.q }}</span>
              <span class="plus" :class="{ open: openIdx === i }"><SvgIcon name="plus" :size="16" /></span>
            </button>
            <transition
              enter-active-class="transition duration-200 ease-out"
              enter-from-class="-translate-y-1 opacity-0"
              enter-to-class="translate-y-0 opacity-100"
              leave-active-class="transition duration-100 ease-in"
              leave-from-class="translate-y-0 opacity-100"
              leave-to-class="-translate-y-1 opacity-0"
            >
              <div v-if="openIdx === i" class="faq-body">{{ item.a }}</div>
            </transition>
          </div>
        </div>

        <div class="faq-note">
          <b>{{ data.note }}</b>
          <router-link :to="data.noteTo">{{ data.noteAction }}</router-link>
        </div>
      </div>
    </section>
  </template>
</template>

<style scoped>
.wrap {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}
.block {
  padding: 44px 0;
}

/* 页头 */
.page-head {
  background: linear-gradient(150deg, #ff8a3d 0%, #ff6b2c 60%, #f2531c 100%);
  color: #fff;
  padding: 52px 0 48px;
  position: relative;
  overflow: hidden;
}
.page-head::before {
  content: '';
  position: absolute;
  right: -90px;
  top: -90px;
  width: 280px;
  height: 280px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
}
.page-head .kicker {
  position: relative;
  display: inline-block;
  font-size: 12px;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.18);
  padding: 5px 12px;
  border-radius: 20px;
  letter-spacing: 1px;
}
.page-head h1 {
  position: relative;
  margin-top: 14px;
  font-size: 28px;
  font-weight: 800;
  letter-spacing: 0.5px;
}
.page-head p {
  position: relative;
  margin-top: 10px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.9);
  max-width: 560px;
}

/* FAQ 两栏 */
.faq-col {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}
.faq-item {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 12px;
}
.faq-q {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 17px 18px;
  font-size: 14px;
  font-weight: 600;
  color: #1d2129;
  text-align: left;
  background: none;
  border: none;
  cursor: pointer;
  font-family: inherit;
  min-height: 44px;
}
.faq-q .plus {
  color: #ff6b2c;
  flex-shrink: 0;
  transition: transform 0.2s;
}
.faq-q .plus.open {
  transform: rotate(45deg);
}
.faq-body {
  padding: 0 18px 17px;
  font-size: 13px;
  color: #4e5969;
  line-height: 1.8;
}

/* FAQ 提示 */
.faq-note {
  margin-top: 24px;
  background: #fff3e8;
  border: 1px solid #ffd9bd;
  border-radius: 14px;
  padding: 18px 20px;
  font-size: 13px;
  color: #4e5969;
  line-height: 1.8;
}
.faq-note b {
  color: #ff6b2c;
}
.faq-note a {
  color: #ff6b2c;
  font-weight: 600;
}
.faq-note a:hover {
  text-decoration: underline;
}

@media (min-width: 768px) {
  .wrap {
    padding: 0 32px;
  }
  .page-head {
    padding: 64px 0 56px;
  }
  .page-head h1 {
    font-size: 34px;
  }
  .faq-col {
    grid-template-columns: 1fr 1fr;
    column-gap: 16px;
    align-items: start;
  }
}
</style>
