<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { siteApi } from '@/api/site'
import { useSiteStore } from '@/stores/site'
import type { HomeData } from '@/types'
import SvgIcon from '@/components/SvgIcon.vue'

const site = useSiteStore()
const data = ref<HomeData | null>(null)
const error = ref(false)
const faqOpen = ref<Record<number, boolean>>({})

async function load() {
  try {
    data.value = await siteApi.home(site.locale)
  } catch {
    error.value = true
  }
}

onMounted(load)
// 语言切换后按新 locale 重新请求页面数据
watch(
  () => site.locale,
  () => {
    error.value = false
    load()
  },
)

function toggleFaq(idx: number) {
  faqOpen.value[idx] = !faqOpen.value[idx]
}
</script>

<template>
  <div v-if="error" class="mx-auto max-w-3xl px-4 py-24 text-center">
    <p class="text-base text-ink-2">{{ site.common?.ui.loadError ?? '内容加载失败，请稍后刷新重试。' }}</p>
  </div>

  <template v-else-if="data">
    <!-- ===== Hero ===== -->
    <section class="hero">
      <div class="hero-inner">
        <div class="hero-left">
          <h1>
            <template v-for="(line, i) in data.hero.titleLines" :key="i">
              {{ line }}<br v-if="i < data.hero.titleLines.length - 1" />
            </template>
          </h1>
          <p class="lead">{{ data.hero.lead }}</p>
          <div class="hero-stats">
            <div v-for="stat in data.hero.stats" :key="stat.label" class="stat">
              <b>{{ stat.value }}</b>
              <span>{{ stat.label }}</span>
            </div>
          </div>
        </div>
        <div class="hero-devices" aria-hidden="true">
          <div class="pc rear-right">
            <div class="p-screen">
              <div class="bar w90"></div>
              <div class="p-cols">
                <div class="p-chip">
                  <span class="c-ico"><SvgIcon name="console" :size="14" /></span>
                  <span class="c-line"></span>
                </div>
                <div class="p-chip">
                  <span class="c-ico"><SvgIcon name="chart" :size="14" /></span>
                  <span class="c-line"></span>
                </div>
              </div>
            </div>
          </div>
          <div class="pad rear-left">
            <div class="p-screen">
              <div class="bar w70"></div>
              <div class="bar w90"></div>
              <div class="p-chip">
                <span class="c-ico"><SvgIcon name="shop" :size="14" /></span>
                <span class="c-line"></span>
              </div>
            </div>
          </div>
          <div class="phone front">
            <div class="p-screen">
              <div class="bar w90"></div>
              <div class="bar w70"></div>
              <div class="bar w55"></div>
              <div class="p-chip">
                <span class="c-ico"><SvgIcon name="bike" :size="12" /></span>
                <span class="c-line"></span>
              </div>
            </div>
          </div>
          <div class="lbl">{{ data.hero.devicesLabel }}</div>
        </div>
      </div>
    </section>

    <!-- ===== 能力全景 ===== -->
    <section class="block">
      <div class="wrap">
        <div class="sec-head">
          <span class="kicker">{{ data.capabilities.kicker }}</span>
          <h2>{{ data.capabilities.title }}</h2>
          <p>{{ data.capabilities.desc }}</p>
        </div>
        <div class="grid-cap">
          <div v-for="item in data.capabilities.items" :key="item.title" class="cap">
            <div class="cap-ico"><SvgIcon :name="item.icon" :size="22" /></div>
            <div>
              <h3>{{ item.title }}</h3>
              <p>{{ item.desc }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== 适用场景 ===== -->
    <section class="block bg-soft">
      <div class="wrap">
        <div class="sec-head">
          <span class="kicker">{{ data.scenarios.kicker }}</span>
          <h2>{{ data.scenarios.title }}</h2>
          <p>{{ data.scenarios.desc }}</p>
        </div>
        <div class="grid-scene">
          <div v-for="item in data.scenarios.items" :key="item.num" class="scene">
            <span class="num">{{ item.num }}</span>
            <div class="scene-ico"><SvgIcon :name="item.icon" :size="26" /></div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== 核心功能 ===== -->
    <section class="block">
      <div class="wrap">
        <div class="sec-head">
          <span class="kicker">{{ data.features.kicker }}</span>
          <h2>{{ data.features.title }}</h2>
          <p>{{ data.features.desc }}</p>
        </div>
        <div class="grid-feat">
          <div v-for="item in data.features.items" :key="item.title" class="feat">
            <div class="ico" :class="`i-${item.tone}`"><SvgIcon :name="item.icon" :size="20" /></div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== 架构链路 ===== -->
    <section class="block pt-0">
      <div class="wrap">
        <div class="pipeline">
          <div class="t">{{ data.pipeline.title }}</div>
          <div class="pipeline-flow">
            <template v-for="(node, i) in data.pipeline.nodes" :key="node.name">
              <div class="pipe-node" :class="{ strong: node.strong }">
                <b>{{ node.name }}</b>
                <span class="tag">{{ node.tag }}</span>
              </div>
              <span v-if="i < data.pipeline.nodes.length - 1" class="pipe-arrow">
                <SvgIcon name="arrow" :size="16" />
              </span>
            </template>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== 多端矩阵 ===== -->
    <section class="block bg-soft">
      <div class="wrap">
        <div class="sec-head">
          <span class="kicker">{{ data.ends.kicker }}</span>
          <h2>{{ data.ends.title }}</h2>
          <p>{{ data.ends.desc }}</p>
        </div>
        <div class="grid-end">
          <div v-for="item in data.ends.items" :key="item.name" class="end">
            <div class="row">
              <div class="ico" :class="`i-${item.tone}`"><SvgIcon :name="item.icon" :size="22" /></div>
              <div>
                <h3>{{ item.name }}</h3>
                <div class="role">{{ item.role }}</div>
              </div>
            </div>
            <p>{{ item.desc }}</p>
            <router-link :to="item.to" class="link">
              {{ item.linkLabel }}
              <SvgIcon name="arrow" :size="14" />
            </router-link>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== 版本动态 ===== -->
    <section id="roadmap" class="block">
      <div class="wrap">
        <div class="sec-head">
          <span class="kicker">{{ data.roadmap.kicker }}</span>
          <h2>{{ data.roadmap.title }}</h2>
          <p>{{ data.roadmap.desc }}</p>
        </div>
        <div class="grid-ver">
          <div v-for="item in data.roadmap.items" :key="item.title" class="ver">
            <div class="dot" :class="item.state">
              <SvgIcon :name="item.state === 'done' ? 'check' : 'clock'" :size="13" />
            </div>
            <div>
              <b class="time">{{ item.time }}</b>
              <h3>{{ item.title }}</h3>
              <p>{{ item.desc }}</p>
              <span class="state" :class="item.state === 'done' ? 'state-done' : 'state-plan'">
                {{ item.stateLabel }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== FAQ 摘要 ===== -->
    <section class="block bg-soft">
      <div class="wrap">
        <div class="sec-head">
          <span class="kicker">{{ data.faqPreview.kicker }}</span>
          <h2>{{ data.faqPreview.title }}</h2>
          <p>{{ data.faqPreview.desc }}</p>
        </div>
        <div class="mx-auto max-w-3xl">
          <div v-for="(item, i) in data.faqPreview.items" :key="item.q" class="faq-item">
            <button type="button" class="faq-q" @click="toggleFaq(i)">
              {{ item.q }}
              <span class="plus" :class="{ open: faqOpen[i] }"><SvgIcon name="plus" :size="16" /></span>
            </button>
            <div v-if="faqOpen[i]" class="faq-body">{{ item.a }}</div>
          </div>
          <div class="faq-more">
            <router-link :to="data.faqPreview.moreTo" class="btn btn-ghost btn-sm">
              {{ data.faqPreview.moreLabel }}
              <SvgIcon name="arrow" :size="14" />
            </router-link>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== 合作 CTA ===== -->
    <section class="block pt-0">
      <div class="wrap">
        <div class="cta">
          <h2>{{ data.cta.title }}</h2>
          <p>{{ data.cta.desc }}</p>
          <div class="row">
            <router-link :to="data.cta.primary.to" class="btn btn-white btn-lg">
              {{ data.cta.primary.label }}
            </router-link>
            <router-link :to="data.cta.secondary.to" class="btn btn-outline btn-lg">
              {{ data.cta.secondary.label }}
            </router-link>
          </div>
        </div>
      </div>
    </section>
  </template>
</template>

<style scoped>
/* ===== 复刻设计稿版式：wrap / hero / 区段 / 能力 / 场景 / 功能 / 架构 / 多端 / 版本 / FAQ / CTA ===== */
.wrap {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}
.block {
  padding: 56px 0;
}
.bg-soft {
  background: #fff;
}

/* Hero */
.hero {
  position: relative;
  overflow: hidden;
  background: linear-gradient(150deg, #ff8a3d 0%, #ff6b2c 55%, #f2531c 100%);
  color: #fff;
}
.hero::before {
  content: '';
  position: absolute;
  right: -120px;
  top: -120px;
  width: 380px;
  height: 380px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
}
.hero::after {
  content: '';
  position: absolute;
  left: -70px;
  bottom: -110px;
  width: 300px;
  height: 300px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.08);
}
.hero-inner {
  position: relative;
  z-index: 1;
  padding: 64px 20px 56px;
  display: flex;
  flex-direction: column;
  gap: 36px;
}
.hero h1 {
  font-size: 32px;
  font-weight: 800;
  line-height: 1.3;
  letter-spacing: 1px;
}
.hero p.lead {
  margin-top: 16px;
  font-size: 15px;
  line-height: 1.8;
  color: rgba(255, 255, 255, 0.92);
  max-width: 560px;
}
.hero-stats {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 36px;
}
.stat {
  background: rgba(255, 255, 255, 0.13);
  border: 1px solid rgba(255, 255, 255, 0.22);
  border-radius: 12px;
  padding: 14px 18px;
  flex: 1 1 30%;
  min-width: 140px;
}
.stat b {
  display: block;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: 0.5px;
}
.stat span {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.85);
}
.hero-devices {
  display: none;
}

/* 区段头 */
.sec-head {
  text-align: center;
  max-width: 640px;
  margin: 0 auto 36px;
}
.sec-head .kicker {
  display: inline-block;
  font-size: 12px;
  font-weight: 600;
  color: #ff6b2c;
  background: #fff3e8;
  padding: 5px 12px;
  border-radius: 20px;
  letter-spacing: 1px;
}
.sec-head h2 {
  margin-top: 14px;
  font-size: 24px;
  font-weight: 800;
  color: #1d2129;
  letter-spacing: 0.5px;
}
.sec-head p {
  margin-top: 10px;
  font-size: 14px;
  color: #86909c;
}

/* 按钮（与设计稿 .btn 系列一致） */
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
}
.btn-primary:hover {
  filter: brightness(1.05);
  box-shadow: 0 8px 20px rgba(255, 107, 44, 0.4);
}
.btn-ghost {
  background: #fff3e8;
  color: #ff6b2c;
}
.btn-ghost:hover {
  background: #ffe9d5;
}
.btn-lg {
  height: 52px;
  padding: 0 26px;
  font-size: 15px;
  border-radius: 10px;
}
.btn-sm {
  height: 38px;
  padding: 0 16px;
}
.btn-ghost-invert {
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
}
.btn-ghost-invert:hover {
  background: rgba(255, 255, 255, 0.26);
}
.btn-white {
  background: #fff;
  color: #ff6b2c;
}
.btn-white:hover {
  background: #fff3e8;
}
.btn-outline {
  background: transparent;
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.55);
}
.btn-outline:hover {
  background: rgba(255, 255, 255, 0.14);
}

/* 能力全景 */
.grid-cap {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
}
.cap {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  padding: 20px;
  display: flex;
  gap: 14px;
  align-items: flex-start;
}
.cap-ico {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  background: #fff3e8;
  color: #ff6b2c;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.cap h3 {
  font-size: 15px;
  font-weight: 700;
  color: #1d2129;
}
.cap p {
  margin-top: 5px;
  font-size: 13px;
  color: #86909c;
  line-height: 1.7;
}

/* 适用场景 */
.grid-scene {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
}
.scene {
  background: #fff3e8;
  border-radius: 14px;
  padding: 24px 20px;
  position: relative;
  overflow: hidden;
}
.scene .num {
  position: absolute;
  right: 16px;
  top: 10px;
  font-size: 40px;
  font-weight: 800;
  color: rgba(255, 107, 44, 0.16);
}
.scene-ico {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  background: #fff;
  color: #ff6b2c;
  display: flex;
  align-items: center;
  justify-content: center;
}
.scene h3 {
  margin-top: 12px;
  font-size: 16px;
  font-weight: 700;
  color: #1d2129;
}
.scene p {
  margin-top: 8px;
  font-size: 13px;
  color: #4e5969;
  line-height: 1.7;
}

/* 核心功能 */
.grid-feat {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
}
.feat {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  padding: 20px;
}
.feat .ico {
  width: 40px;
  height: 40px;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.feat h3 {
  font-size: 15px;
  font-weight: 700;
  margin-top: 12px;
}
.feat p {
  margin-top: 6px;
  font-size: 13px;
  color: #86909c;
  line-height: 1.7;
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
.i-yellow {
  background: #fff8e6;
  color: #d48806;
}

/* 架构链路 */
.pipeline {
  background: #1d2129;
  border-radius: 16px;
  padding: 24px 18px;
  margin-top: 26px;
}
.pipeline .t {
  font-size: 13px;
  color: #a9b0bc;
  letter-spacing: 1px;
  text-align: center;
}
.pipeline-flow {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 16px;
  align-items: center;
}
.pipe-node {
  background: #2c323c;
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: #e5e6eb;
  border-radius: 10px;
  padding: 12px 16px;
  font-size: 13px;
  text-align: center;
  width: 100%;
}
.pipe-node b {
  color: #fff;
}
.pipe-node .tag {
  display: inline-block;
  margin-top: 4px;
  font-size: 11px;
  color: #ffb98f;
}
.pipe-arrow {
  color: #ff8a3d;
  display: flex;
  align-items: center;
  transform: rotate(90deg);
}

/* 多端矩阵 */
.grid-end {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
}
.end {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  padding: 22px 20px;
  display: flex;
  flex-direction: column;
}
.end .row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.end .ico {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.end h3 {
  font-size: 15px;
  font-weight: 700;
}
.end .role {
  font-size: 12px;
  color: #86909c;
}
.end p {
  margin: 10px 0 14px;
  font-size: 13px;
  color: #4e5969;
  line-height: 1.7;
  flex: 1;
}
.end .link {
  font-size: 13px;
  font-weight: 600;
  color: #ff6b2c;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.end .link:hover {
  text-decoration: underline;
}

/* 版本动态 */
.grid-ver {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0;
}
.ver {
  display: flex;
  gap: 16px;
  position: relative;
  padding-bottom: 20px;
}
.ver::before {
  content: '';
  position: absolute;
  left: 13px;
  top: 26px;
  bottom: 0;
  width: 2px;
  background: #e5e6eb;
}
.ver:last-child::before {
  display: none;
}
.ver .dot {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #fff;
  border: 2px solid #ff6b2c;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 1;
}
.ver .dot.done {
  background: #ff6b2c;
  color: #fff;
}
.ver .dot.plan {
  border-color: #c9cdd4;
  color: #86909c;
}
.ver b.time {
  display: block;
  font-size: 12px;
  color: #86909c;
  font-weight: 500;
}
.ver h3 {
  font-size: 14px;
  font-weight: 700;
  color: #1d2129;
  margin-top: 3px;
}
.ver p {
  font-size: 13px;
  color: #4e5969;
  margin-top: 5px;
  line-height: 1.7;
}
.ver .state {
  display: inline-block;
  font-size: 11px;
  padding: 2px 9px;
  border-radius: 10px;
  margin-top: 7px;
  font-weight: 600;
}
.state-done {
  background: #e8f9f0;
  color: #00b578;
}
.state-plan {
  background: #fff3e8;
  color: #ff6b2c;
}

/* FAQ 摘要 */
.faq-item {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 12px;
  margin-bottom: 12px;
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
.faq-more {
  text-align: center;
  margin-top: 20px;
}

/* 合作 CTA */
.cta {
  background: linear-gradient(150deg, #ff8a3d, #ff6b2c);
  border-radius: 20px;
  padding: 36px 24px;
  color: #fff;
  text-align: center;
}
.cta h2 {
  font-size: 22px;
  font-weight: 800;
  letter-spacing: 0.5px;
}
.cta p {
  margin-top: 10px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.9);
}
.cta .row {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 22px;
  align-items: center;
  justify-content: center;
}

/* ===== 断点：≥768px ===== */
@media (min-width: 768px) {
  .wrap {
    padding: 0 32px;
  }
  .hero-inner {
    padding: 88px 32px 72px;
  }
  .hero h1 {
    font-size: 44px;
  }
  .hero p.lead {
    font-size: 16px;
  }
  .hero-stats .stat {
    flex: 1;
  }
  .grid-cap {
    grid-template-columns: repeat(3, 1fr);
  }
  .grid-scene {
    grid-template-columns: repeat(3, 1fr);
  }
  .grid-feat {
    grid-template-columns: repeat(2, 1fr);
  }
  .grid-end {
    grid-template-columns: repeat(2, 1fr);
  }
  .pipeline-flow {
    flex-direction: row;
    justify-content: center;
  }
  .pipe-node {
    width: auto;
    flex: 0 1 auto;
    padding: 12px 22px;
  }
  .pipe-arrow {
    transform: none;
  }
  .cta .row {
    flex-direction: row;
    gap: 14px;
  }
}

/* ===== 断点：≥1024px ===== */
@media (min-width: 1024px) {
  .hero-inner {
    flex-direction: row;
    align-items: center;
    gap: 48px;
    padding: 96px 32px 80px;
  }
  .hero-left {
    flex: 1.15;
  }
  .hero h1 {
    font-size: 50px;
  }
  .hero-devices {
    display: block;
    flex: 1;
    position: relative;
  }
  .hero-devices .pc {
    position: absolute;
    right: 0;
    top: -55px;
    width: 300px;
    background: #fff;
    border-radius: 12px;
    padding: 12px;
    border: 6px solid #1d2129;
    box-shadow: 0 24px 50px rgba(0, 0, 0, 0.28);
    transform: rotate(3deg);
    z-index: 0;
    opacity: 0.95;
  }
  .hero-devices .pc::after {
    content: '';
    position: absolute;
    left: 50%;
    bottom: -14px;
    width: 52px;
    height: 7px;
    background: #1d2129;
    border-radius: 3px;
    transform: translateX(-50%) rotate(-3deg);
  }
  .hero-devices .pad {
    position: absolute;
    left: 10px;
    top: -15px;
    width: 240px;
    background: #fff;
    border-radius: 20px;
    padding: 14px;
    border: 6px solid #1d2129;
    box-shadow: 0 24px 50px rgba(0, 0, 0, 0.28);
    transform: rotate(-3deg);
    z-index: 1;
  }
  .hero-devices .phone {
    position: relative;
    left: 150px;
    width: 118px;
    background: #fff;
    border-radius: 26px;
    padding: 10px;
    border: 5px solid #1d2129;
    box-shadow: 0 24px 50px rgba(0, 0, 0, 0.28);
    transform: rotate(6deg);
    z-index: 2;
  }
  .hero-devices .p-screen {
    background: linear-gradient(160deg, #ff8a3d, #f2531c);
    color: #fff;
  }
  .hero-devices .pc .p-screen {
    border-radius: 8px;
    padding: 12px;
    height: 150px;
  }
  .hero-devices .pad .p-screen {
    border-radius: 12px;
    padding: 14px;
    height: 153px;
  }
  .hero-devices .phone .p-screen {
    border-radius: 16px;
    padding: 10px;
    height: 149px;
  }
  .hero-devices .p-screen .bar {
    height: 12px;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.35);
    margin-bottom: 8px;
  }
  .hero-devices .pc .bar {
    height: 14px;
  }
  .hero-devices .phone .bar {
    height: 10px;
  }
  .hero-devices .p-screen .bar.w70 {
    width: 70%;
  }
  .hero-devices .p-screen .bar.w90 {
    width: 90%;
  }
  .hero-devices .p-screen .bar.w55 {
    width: 55%;
  }
  .hero-devices .p-cols {
    display: flex;
    gap: 8px;
    margin-top: 10px;
  }
  .hero-devices .p-chip {
    background: #fff;
    border-radius: 10px;
    padding: 10px;
    margin-top: 10px;
    display: flex;
    gap: 8px;
    align-items: center;
  }
  .hero-devices .p-cols .p-chip {
    margin-top: 0;
    flex: 1;
  }
  .hero-devices .phone .p-chip {
    padding: 8px;
    margin-top: 8px;
  }
  .hero-devices .p-chip .c-ico {
    width: 30px;
    height: 30px;
    border-radius: 8px;
    background: #fff3e8;
    color: #ff6b2c;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .hero-devices .phone .c-ico {
    width: 24px;
    height: 24px;
  }
  .hero-devices .p-chip .c-line {
    height: 8px;
    border-radius: 4px;
    background: #e5e6eb;
    flex: 1;
  }
  .hero-devices .lbl {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.85);
    margin-top: 12px;
    text-align: center;
  }
  .grid-end {
    grid-template-columns: repeat(4, 1fr);
  }
  .grid-ver {
    grid-template-columns: 1fr 1fr;
    column-gap: 40px;
  }
}
</style>
