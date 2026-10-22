<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import StatusTag from '@/components/StatusTag.vue'
import Empty from '@/components/Empty.vue'
import Confirm from '@/components/Confirm.vue'
import AppButton from '@/components/AppButton.vue'
import { useToast } from '@/composables/useToast'
import { formatDate, formatNumber, formatClock } from '@/utils/format'
import { getRiderDetail, freezeRider, type RiderDetailDto } from '@/api/riders'

const route = useRoute()
const toast = useToast()

const detail = ref<RiderDetailDto | null>(null)
const loading = ref(true)
const activeTab = ref('basic')

const tabs = [
  { key: 'basic', label: '基本信息', icon: 'order' },
  { key: 'stats', label: '配送统计', icon: 'chart' },
  { key: 'score', label: '绩效评分', icon: 'star' },
  { key: 'cert', label: '证件管理', icon: 'image' },
  { key: 'logs', label: '奖惩记录', icon: 'layers' },
] as const

const lastActionText = computed(() => {
  if (!detail.value) return ''
  return `${formatDate(detail.value.lastActionAt)} ${formatClock(detail.value.lastActionAt)} ${detail.value.lastAction}`
})

async function load() {
  loading.value = true
  try {
    detail.value = await getRiderDetail(String(route.params.id))
  } finally {
    loading.value = false
  }
}

const freezeOpen = ref(false)
const freezing = ref(false)
async function onFreeze() {
  if (!detail.value || freezing.value) return
  freezing.value = true
  try {
    await freezeRider(detail.value.id, 'freeze')
    toast.success(`已冻结骑手 ${detail.value.name}`)
    freezeOpen.value = false
  } finally {
    freezing.value = false
  }
}

onMounted(load)
</script>

<template>
  <div v-if="loading" class="text-ink-3">加载中…</div>
  <div v-else-if="detail" class="flex flex-col gap-4">
    <!-- 骑手头部卡 -->
    <div class="flex items-center gap-4 rounded-card bg-white px-5 py-[18px]">
      <div
        class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full text-[22px] font-bold text-white"
        :style="{ background: detail.avatarBg }"
      >
{{ detail.name.charAt(0) }}
</div>
      <div class="min-w-0 flex-1">
        <div class="flex flex-wrap items-center gap-2.5">
          <span class="text-[18px] font-bold text-ink-1">{{ detail.name }}</span>
          <StatusTag tone="green">{{ detail.status }}</StatusTag>
          <span class="inline-block rounded-[4px] px-2.5 py-[3px] text-[12px] font-semibold" style="background:#FFF3E8;color:#FF6B2C;">{{ detail.serviceType }}</span>
          <span class="text-[12px] text-ink-4">骑手编号 {{ detail.id }} · 入驻 {{ formatDate(detail.joinedAt) }}</span>
        </div>
        <div class="mt-2 flex flex-wrap items-center gap-6 text-[13px] text-ink-3">
          <span class="flex items-center gap-1.5">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#D48806" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
            </svg>
            <b class="text-ink-1">{{ detail.rating }}</b>评分
          </span>
          <span>今日完成 <b class="text-ink-1">{{ detail.todayOrders }}</b> 单</span>
          <span>今日在线 <b class="text-ink-1">{{ detail.onlineHours }}</b> 小时</span>
          <span>本月收入 <b class="text-ink-1">¥{{ formatNumber(detail.monthIncome) }}</b></span>
          <span class="flex items-center gap-1.5">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#00B578" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" /><polyline points="22 4 12 14.01 9 11.01" />
            </svg>
            准时率 {{ detail.onTimeRate }}
          </span>
        </div>
      </div>
      <div class="flex shrink-0 gap-2.5">
        <AppButton variant="secondary" class="!px-[18px]">编辑资料</AppButton>
        <AppButton
          variant="secondary"
          class="!px-[18px] !border-[#FFD9D9] !text-[#FF4D4F]"
          @click="freezeOpen = true"
        >
冻结账号
</AppButton>
      </div>
    </div>

    <!-- 主体区 -->
    <div class="flex gap-4">
      <!-- 左侧 tab -->
      <div class="flex w-[168px] shrink-0 flex-col rounded-card bg-white py-2">
        <div
          v-for="t in tabs"
          :key="t.key"
          class="mx-2 my-[2px] flex cursor-pointer items-center gap-2 rounded-lg px-4 py-3 text-[14px]"
          :class="activeTab === t.key ? 'bg-primary-bg font-semibold text-primary' : 'text-ink-2'"
          @click="activeTab = t.key"
        >
          <svg v-if="t.icon === 'star'" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
          </svg>
          <svg v-else-if="t.icon === 'image'" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="3" width="18" height="18" rx="2" /><circle cx="9" cy="9" r="2" /><path d="M21 15l-5-5L5 21" />
          </svg>
          <svg v-else-if="t.icon === 'layers'" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="12 2 2 7 12 12 22 7 12 2" /><polyline points="2 17 12 22 22 17" /><polyline points="2 12 12 17 22 12" />
          </svg>
          <svg v-else width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <template v-if="t.icon === 'order'">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
              <polyline points="14 2 14 8 20 8" />
            </template>
            <template v-else>
              <path d="M3 3v18h18" />
              <path d="M7 14l4-4 3 3 5-6" />
            </template>
          </svg>
          {{ t.label }}
        </div>
        <div class="flex-1" />
        <div class="border-t border-line px-4 py-3 text-[12px] text-ink-4">最近操作：{{ lastActionText }}</div>
      </div>

      <!-- 右侧内容 -->
      <div class="flex min-w-0 flex-1 flex-col gap-4">
        <template v-if="activeTab === 'basic'">
          <!-- 基本信息表单卡 -->
          <div class="rounded-card bg-white p-5">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">基本信息</div>
              <div class="ml-auto rounded-[4px] px-2.5 py-[3px] text-[12px] font-semibold" style="background:#E8F7F0;color:#00B578;">资质审核已通过</div>
            </div>
            <div class="mt-[18px] grid grid-cols-2 gap-x-7 gap-y-4">
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">姓名</div>
                <div class="flex h-[38px] items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ detail.profile.name }}</div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">联系电话</div>
                <div class="flex h-[38px] items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ detail.profile.phone }}</div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">身份证号</div>
                <div class="flex h-[38px] items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ detail.profile.idCard }}</div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">常驻区域</div>
                <div class="flex h-[38px] items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ detail.profile.area }}</div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">配送工具</div>
                <div class="flex h-[38px] items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ detail.profile.vehicle }}</div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">可接单类型</div>
                <div class="flex h-[38px] items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ detail.profile.orderTypes }}</div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">收款账户</div>
                <div class="flex h-[38px] items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ detail.profile.account }}</div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">加入时间</div>
                <div class="flex h-[38px] items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ detail.profile.joinedAt }} · 服务 {{ detail.profile.serviceMonths }} 个月</div>
              </div>
            </div>
          </div>

          <!-- 今日配送速览 -->
          <div class="rounded-card bg-white p-5">
            <div class="text-[15px] font-semibold text-ink-1">今日配送速览</div>
            <div class="mt-4 grid grid-cols-4 gap-4">
              <div class="rounded-[10px] bg-head p-3.5">
                <div class="text-[12px] text-ink-3">今日完成</div>
                <div class="mt-1.5 text-2xl font-bold text-ink-1">{{ detail.todayOverview.todayOrders }}</div>
                <div class="mt-1 text-[11px]" style="color:#00B578;">{{ detail.todayOverview.todayOrdersDelta }}</div>
              </div>
              <div class="rounded-[10px] bg-head p-3.5">
                <div class="text-[12px] text-ink-3">准时率</div>
                <div class="mt-1.5 text-2xl font-bold text-ink-1">{{ detail.todayOverview.onTimeRate }}</div>
                <div class="mt-1 text-[11px] text-ink-3">{{ detail.todayOverview.onTimeTarget }}</div>
              </div>
              <div class="rounded-[10px] bg-head p-3.5">
                <div class="text-[12px] text-ink-3">平均配送时长</div>
                <div class="mt-1.5 text-2xl font-bold text-ink-1">{{ detail.todayOverview.avgMinutes }}</div>
                <div class="mt-1 text-[11px] text-ink-3">{{ detail.todayOverview.platformAvg }}</div>
              </div>
              <div class="rounded-[10px] bg-head p-3.5">
                <div class="text-[12px] text-ink-3">本月投诉</div>
                <div class="mt-1.5 text-2xl font-bold text-ink-1">{{ detail.todayOverview.complaints }}</div>
                <div class="mt-1 text-[11px]" style="color:#00B578;">{{ detail.todayOverview.complaintNote }}</div>
              </div>
            </div>
          </div>
        </template>
        <div v-else class="rounded-card bg-white p-10">
          <Empty text="该模块建设中" />
        </div>
      </div>
    </div>

    <Confirm
      :open="freezeOpen"
      title="冻结账号"
      :message="`确认冻结骑手 ${detail.name}（${detail.id}）的账号吗？冻结后将停止派单。`"
      confirm-text="确认冻结"
      @update:open="(v) => (freezeOpen = v)"
      @confirm="onFreeze"
    />
  </div>
</template>
