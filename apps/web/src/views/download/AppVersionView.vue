<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { siteApi } from '@/api/site'
import type { AppVersionData, DownloadData, PlatformId } from '@/types'
import SvgIcon from '@/components/SvgIcon.vue'

const route = useRoute()
const data = ref<DownloadData | null>(null)
const error = ref(false)
const notFound = ref(false)
const toastMsg = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

const app = computed<AppVersionData | null>(() => {
  if (!data.value) return null
  const key = route.params.appKey as string
  return data.value.appVersions.find((a) => a.key === key) ?? null
})

onMounted(async () => {
  try {
    data.value = await siteApi.download()
    if (!app.value) notFound.value = true
  } catch {
    error.value = true
  }
})

/** 某平台可下载的最新版本号；不可用返回 null */
function latestVersion(appVal: AppVersionData, id: PlatformId): string | null {
  for (const v of appVal.versions) {
    if (v.packages[id]) return v.version
  }
  return null
}

/** 版本行某平台是否可下载 */
function pkgOk(v: { packages: Record<PlatformId, boolean> }, id: PlatformId): boolean {
  return v.packages[id]
}

function showToast(msg: string) {
  toastMsg.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMsg.value = ''
  }, 2400)
}

watch(
  app,
  (val) => {
    if (val) document.title = `${val.name} · 版本下载 · 曙光跑腿 Dawn Errand`
  },
  { immediate: true },
)
</script>

<template>
  <div v-if="error" class="mx-auto max-w-3xl px-4 py-24 text-center">
    <p class="text-base text-ink-2">内容加载失败，请稍后刷新重试。</p>
  </div>

  <div
    v-else-if="notFound"
    class="mx-auto max-w-3xl px-4 py-24 text-center"
  >
    <p class="text-base text-ink-2">未找到对应的客户端，请返回下载中心重新选择。</p>
    <router-link to="/download" class="btn btn-primary btn-lg mt-6 inline-flex">
      返回下载中心
    </router-link>
  </div>

  <template v-else-if="data && app">
    <!-- 页头 -->
    <section class="page-head">
      <div class="wrap">
        <span class="kicker">VERSION HISTORY</span>
        <h1>{{ app.name }} · 历史版本</h1>
        <p>{{ app.sub }}</p>
        <router-link to="/download" class="back">
          <SvgIcon name="arrow" :size="14" />
          返回下载中心
        </router-link>
      </div>
    </section>

    <section class="block">
      <div class="wrap">
        <!-- 平台下载 -->
        <div class="sec-head">
          <h2>选择平台下载</h2>
          <p>支持 Android、iOS 与鸿蒙 HarmonyOS，安装包在演示站点暂以提示代替。</p>
        </div>
        <div class="grid-plat">
          <div v-for="p in app.platforms" :key="p.id" class="plat" :class="p.state">
            <div class="row">
              <span class="p-ico"><SvgIcon :name="p.id" :size="24" /></span>
              <span>
                <h3>{{ p.label }}</h3>
                <span class="p-note">{{ p.note }}</span>
              </span>
            </div>
            <p class="p-latest">
              最新版本
              <b>{{ latestVersion(app, p.id) ?? '—' }}</b>
            </p>
            <button
              v-if="p.state === 'ready'"
              type="button"
              class="btn btn-primary"
              @click="showToast('演示站点：安装包待正式发布')"
            >
              <SvgIcon name="download" :size="16" />
              下载最新版
            </button>
            <button v-else type="button" class="btn btn-disabled" disabled>
              即将上线
            </button>
          </div>
        </div>

        <!-- 历史版本 -->
        <div class="sec-head ver-head">
          <h2>历史版本</h2>
          <p>按发布时间倒序展示，支持 Android / iOS / 鸿蒙 平台安装包。</p>
        </div>
        <div class="ver-list">
          <div v-for="v in app.versions" :key="v.version" class="ver">
            <div class="ver-main">
              <div class="ver-tags">
                <span class="ver-tag">{{ v.version }}</span>
                <span class="ver-date">{{ v.date }}</span>
              </div>
              <p class="ver-note">{{ v.note }}</p>
            </div>
            <div class="ver-pkgs">
              <button
                v-for="p in app.platforms"
                :key="p.id"
                type="button"
                class="pkg"
                :class="{ on: pkgOk(v, p.id), off: !pkgOk(v, p.id) }"
                :disabled="!pkgOk(v, p.id)"
                @click="showToast('演示站点：安装包待正式发布')"
              >
                <SvgIcon :name="p.id" :size="14" />
                {{ pkgOk(v, p.id) ? v.version + ' 下载' : '暂无' }}
              </button>
            </div>
          </div>
        </div>

        <div class="scan-note">
          <b>扫码直达：</b>{{ data.scanNote }}
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
.ver-head {
  margin-top: 44px;
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
  line-height: 1.8;
}
.page-head .back {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 18px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.95);
  background: rgba(255, 255, 255, 0.16);
  padding: 8px 14px;
  border-radius: 8px;
  transition: background 0.15s;
}
.page-head .back:hover {
  background: rgba(255, 255, 255, 0.28);
}
.page-head .back svg {
  transform: rotate(180deg);
}

/* 平台卡 */
.grid-plat {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
}
.plat {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  padding: 20px;
  display: flex;
  flex-direction: column;
}
.plat .row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.plat .p-ico {
  width: 46px;
  height: 46px;
  border-radius: 11px;
  background: #fff3e8;
  color: #ff6b2c;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.plat h3 {
  font-size: 15px;
  font-weight: 700;
}
.plat .p-note {
  font-size: 12px;
  color: #86909c;
}
.plat .p-latest {
  margin: 14px 0;
  font-size: 13px;
  color: #4e5969;
}
.plat .p-latest b {
  margin-left: 6px;
  font-size: 15px;
  color: #1d2129;
  font-weight: 700;
}
.plat .btn {
  width: 100%;
}
.plat.coming .p-ico {
  background: #f5f6f8;
  color: #86909c;
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
.btn-disabled {
  background: #f5f6f8;
  color: #86909c;
  height: 46px;
  padding: 0 18px;
  cursor: not-allowed;
}

/* 历史版本列表 */
.ver-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.ver {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.ver-main {
  flex: 1;
  min-width: 0;
}
.ver-tags {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.ver-tag {
  font-size: 13px;
  font-weight: 700;
  color: #ff6b2c;
  background: #fff3e8;
  border: 1px solid #ffd9bd;
  padding: 3px 10px;
  border-radius: 6px;
  letter-spacing: 0.3px;
}
.ver-date {
  font-size: 12px;
  color: #86909c;
}
.ver-note {
  margin-top: 8px;
  font-size: 13px;
  color: #4e5969;
  line-height: 1.7;
}
.ver-pkgs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.pkg {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  font-weight: 600;
  font-family: inherit;
  border-radius: 6px;
  padding: 6px 11px;
  cursor: pointer;
  transition: background 0.15s, filter 0.15s;
}
.pkg.on {
  background: #fff3e8;
  color: #ff6b2c;
  border: 1px solid #ffd9bd;
}
.pkg.on:hover {
  filter: brightness(0.97);
}
.pkg.off {
  background: #f5f6f8;
  color: #a9aeb8;
  border: 1px solid #e5e6eb;
  cursor: not-allowed;
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
  .grid-plat {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 1024px) {
  .ver {
    flex-direction: row;
    align-items: center;
    gap: 24px;
  }
  .ver-pkgs {
    flex-shrink: 0;
  }
}
</style>
