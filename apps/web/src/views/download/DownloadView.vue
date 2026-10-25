<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { siteApi } from '@/api/site'
import { useSiteStore } from '@/stores/site'
import type { DownloadData } from '@/types'
import SvgIcon from '@/components/SvgIcon.vue'

const site = useSiteStore()
const data = ref<DownloadData | null>(null)
const error = ref(false)
const toastMsg = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

async function load() {
  try {
    data.value = await siteApi.download(site.locale)
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

function showToast(msg: string) {
  toastMsg.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMsg.value = ''
  }, 2400)
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
        <div class="sec-head">
          <h2>{{ data.androidSection.title }}</h2>
          <p>{{ data.androidSection.desc }}</p>
        </div>
        <div class="grid-dl">
          <div v-for="app in data.apps" :key="app.name" class="dl">
            <div class="row">
              <span class="ico" :class="`i-${app.tone}`"><SvgIcon :name="app.icon" :size="24" /></span>
              <span>
                <h3>{{ app.name }}</h3>
                <span class="role">{{ app.role }}</span>
              </span>
              <span class="qrcode" aria-hidden="true">
                <svg width="56" height="56" viewBox="0 0 21 21" fill="none">
                  <path d="M3 3h7v7H3zM11 3h7v7h-7zM3 11h7v7H3z" stroke="#FF6B2C" stroke-width="1.6" />
                  <path d="M5 5h3v3H5zM13 5h3v3h-3zM5 13h3v3H5z" fill="#1D2129" />
                  <path d="M13 13h1.5v1.5H13zM15 15h1.5v1.5H15zM13 17h1.5v1.5H13zM17 13h1.5v1.5H17z" fill="#1D2129" />
                </svg>
              </span>
            </div>
            <p class="desc">{{ app.desc }}</p>
            <div class="meta">
              <span v-for="m in app.meta" :key="m">{{ m }}</span>
            </div>
            <router-link :to="`/download/${app.key}`" class="btn btn-primary">
              <SvgIcon name="download" :size="16" />
              {{ app.linkLabel }}
            </router-link>
          </div>
        </div>

        <div style="margin-top: 28px">
          <div class="sec-head">
            <h2>{{ data.adminSection.title }}</h2>
            <p>{{ data.adminSection.desc }}</p>
          </div>
          <div class="admin-entry">
            <div class="row">
              <span class="ico i-purple"><SvgIcon :name="data.adminSection.entry.icon" :size="24" /></span>
              <span>
                <h3>{{ data.adminSection.entry.name }}</h3>
                <span class="role">{{ data.adminSection.entry.role }}</span>
              </span>
            </div>
            <p>{{ data.adminSection.entry.desc }}</p>
            <button type="button" class="btn btn-primary" @click="showToast(data.adminSection.entry.demo)">
              {{ data.adminSection.entry.btnLabel }}
            </button>
          </div>
        </div>

        <div class="scan-note">
          <b>{{ data.scanNoteTitle }}</b>{{ data.scanNote }}
        </div>
      </div>
    </section>

    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="translate-y-4 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="translate-y-4 opacity-0"
    >
      <div v-if="toastMsg" class="toast" role="status">{{ toastMsg }}</div>
    </transition>
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
.sec-head {
  margin-bottom: 26px;
}
.sec-head h2 {
  font-size: 20px;
  font-weight: 800;
  color: #1d2129;
}
.sec-head p {
  margin-top: 8px;
  font-size: 13px;
  color: #86909c;
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

/* 下载卡 */
.grid-dl {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
}
.dl {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  padding: 22px;
  display: flex;
  flex-direction: column;
}
.dl .row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.dl .ico {
  width: 46px;
  height: 46px;
  border-radius: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.dl h3 {
  font-size: 15px;
  font-weight: 700;
}
.dl .role {
  font-size: 12px;
  color: #86909c;
}
.dl .desc {
  margin: 10px 0 14px;
  font-size: 13px;
  color: #4e5969;
  line-height: 1.7;
  flex: 1;
}
.dl .meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 14px;
}
.dl .meta span {
  font-size: 11px;
  color: #86909c;
  background: #f5f6f8;
  border: 1px solid #e5e6eb;
  padding: 3px 9px;
  border-radius: 6px;
}
.dl .btn {
  width: 100%;
}
.dl .qrcode {
  display: none;
}
.i-orange {
  background: #fff3e8;
  color: #ff6b2c;
}
.i-blue {
  background: #eef4ff;
  color: #2b7fff;
}
.i-green {
  background: #e8f9f0;
  color: #00b578;
}
.i-purple {
  background: #f4eeff;
  color: #722ed1;
}

/* 按钮 */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: none;
  cursor: pointer;
  font-family: inherit;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  text-decoration: none;
  transition:
    filter 0.15s,
    box-shadow 0.15s,
    background 0.15s;
}
.btn:active {
  transform: translateY(1px);
}
.btn-primary {
  background: linear-gradient(135deg, #ff8a3d, #ff6b2c);
  color: #fff;
  box-shadow: 0 6px 16px rgba(255, 107, 44, 0.32);
  height: 46px;
  padding: 0 18px;
}
.btn-primary:hover {
  filter: brightness(1.05);
  box-shadow: 0 8px 20px rgba(255, 107, 44, 0.4);
}

/* 后台入口 */
.admin-entry {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  padding: 22px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.admin-entry .row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.admin-entry p {
  font-size: 13px;
  color: #4e5969;
  line-height: 1.7;
}
.admin-entry .btn {
  width: 100%;
}

/* 扫码说明 */
.scan-note {
  margin-top: 28px;
  background: #fff3e8;
  border: 1px solid #ffd9bd;
  border-radius: 14px;
  padding: 18px 20px;
  font-size: 13px;
  color: #4e5969;
  line-height: 1.8;
}
.scan-note b {
  color: #ff6b2c;
}

/* toast */
.toast {
  position: fixed;
  left: 50%;
  bottom: 28px;
  transform: translateX(-50%);
  background: #1d2129;
  color: #fff;
  font-size: 13px;
  padding: 11px 18px;
  border-radius: 10px;
  z-index: 100;
  max-width: 88vw;
  text-align: center;
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
  .grid-dl {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .grid-dl {
    grid-template-columns: repeat(3, 1fr);
  }
  .admin-entry {
    flex-direction: row;
    align-items: center;
  }
  .admin-entry p {
    flex: 1;
  }
  .admin-entry .btn {
    width: auto;
  }
  .dl .qrcode {
    display: flex;
    width: 78px;
    height: 78px;
    border-radius: 10px;
    background: #fff;
    border: 1px solid #e5e6eb;
    align-items: center;
    justify-content: center;
    margin-left: auto;
    flex-shrink: 0;
  }
}
</style>
