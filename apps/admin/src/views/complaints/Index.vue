<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import StatusTag from '@/components/StatusTag.vue'
import { getComplaints, type ComplaintRow } from '@/api/complaints'
import { formatTime } from '@/utils/format'

type Tone = 'orange' | 'green' | 'red' | 'blue' | 'gray' | 'yellow'

/** 分类标签配色（照设计稿） */
const CATEGORY_TONE: Record<string, Tone> = {
  配送异常: 'red',
  商家投诉: 'yellow',
  退款申请: 'blue',
}

/** 工单状态 → tone */
const STATUS_TONE: Record<string, Tone> = {
  待处理: 'yellow',
  处理中: 'blue',
  已完结: 'green',
}

const TABS = ['待处理', '处理中', '已完结']

const router = useRouter()
const loading = ref(false)
const all = ref<ComplaintRow[]>([])
const activeTab = ref('待处理')

const list = computed(() =>
  activeTab.value === '全部' ? all.value : all.value.filter((r) => r.status === activeTab.value),
)

const pendingCount = computed(() => all.value.filter((r) => r.status === '待处理').length)
const todayNew = computed(() => {
  const start = new Date()
  start.setHours(0, 0, 0, 0)
  const sec = Math.floor(start.getTime() / 1000)
  return all.value.filter((r) => r.created_at >= sec).length
})
const closeRate = computed(() => {
  if (all.value.length === 0) return '0.0'
  const done = all.value.filter((r) => r.status === '已完结').length
  return ((done / all.value.length) * 100).toFixed(1)
})

async function load() {
  loading.value = true
  try {
    const res = await getComplaints({ page: 1, pageSize: 100 })
    all.value = res.list
  } finally {
    loading.value = false
  }
}

function openDetail(row: ComplaintRow) {
  router.push(`/complaints/${row.id}`)
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-4 p-5">
    <!-- 标题行 -->
    <div class="flex items-center">
      <div>
        <div class="text-[18px] font-bold text-ink-1">投诉与售后工单</div>
        <div class="mt-1 text-[12px] text-ink-3">
          共 <span class="font-semibold" style="color:#FF4D4F;">{{ pendingCount }}</span> 条待处理 ·
          今日新增 {{ todayNew }} 条
        </div>
      </div>
      <div class="ml-auto flex items-center gap-2 text-[13px] text-ink-2">
        <span
          v-for="t in TABS"
          :key="t"
          class="cursor-pointer rounded-md px-3.5 py-1.5 text-[13px]"
          :class="activeTab === t ? 'bg-primary-bg font-semibold text-primary' : 'bg-head text-ink-2'"
          @click="activeTab = t"
        >
          {{ t }}
        </span>
      </div>
    </div>

    <!-- 统计小卡 -->
    <div class="grid grid-cols-3 gap-4">
      <div class="flex items-center gap-3 rounded-card bg-white p-[14px]">
        <div class="flex h-10 w-10 items-center justify-center rounded-[10px]" style="background:#FFE8E8;">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#FF4D4F" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10" /><path d="M12 8v4M12 16h.01" /></svg>
        </div>
        <div class="flex-1">
          <div class="text-[12px] text-ink-3">待处理工单</div>
          <div class="mt-0.5 text-[20px] font-bold" style="color:#FF4D4F;">{{ pendingCount }}</div>
        </div>
      </div>
      <div class="flex items-center gap-3 rounded-card bg-white p-[14px]">
        <div class="flex h-10 w-10 items-center justify-center rounded-[10px]" style="background:#FFF8E0;">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#D48806" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10" /><path d="M12 6v6l4 2" /></svg>
        </div>
        <div class="flex-1">
          <div class="text-[12px] text-ink-3">今日新增</div>
          <div class="mt-0.5 text-[20px] font-bold" style="color:#D48806;">{{ todayNew }}</div>
        </div>
      </div>
      <div class="flex items-center gap-3 rounded-card bg-white p-[14px]">
        <div class="flex h-10 w-10 items-center justify-center rounded-[10px]" style="background:#E8F7F0;">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#00B578" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" /><polyline points="22 4 12 14.01 9 11.01" /></svg>
        </div>
        <div class="flex-1">
          <div class="text-[12px] text-ink-3">本月完结率</div>
          <div class="mt-0.5 text-[20px] font-bold" style="color:#00B578;">{{ closeRate }}%</div>
        </div>
      </div>
    </div>

    <!-- 工单列表 -->
    <div v-if="loading" class="py-10 text-center text-ink-3">加载中…</div>
    <div v-else-if="list.length === 0" class="rounded-card bg-white py-10 text-center text-ink-4">暂无工单</div>
    <div v-else class="flex flex-col gap-3 pb-1">
      <div
        v-for="row in list"
        :key="row.id"
        class="rounded-card bg-white p-5"
        :class="row.status === '处理中' ? 'opacity-90' : ''"
      >
        <!-- 头部行 -->
        <div class="flex items-center gap-2.5">
          <StatusTag :tone="CATEGORY_TONE[row.category] ?? 'blue'">{{ row.category }}</StatusTag>
          <span class="text-[14px] font-bold text-ink-1">{{ row.title }}</span>
          <span class="text-[12px] text-ink-4">工单号 {{ row.id }} · 关联订单 #{{ row.order_id }}</span>
          <span class="ml-auto">
            <StatusTag :tone="STATUS_TONE[row.status] ?? 'gray'">{{ row.status }}</StatusTag>
          </span>
        </div>
        <!-- meta 行 -->
        <div class="mt-2.5 flex flex-wrap items-center gap-4 text-[12px] text-ink-3">
          <span>投诉人：{{ row.complainant }}</span>
          <span v-if="row.target">责任方：{{ row.target }}</span>
          <span>{{ row.extra }}</span>
          <span>提交于 {{ formatTime(row.created_at) }}</span>
        </div>
        <!-- 描述块（待处理/处理中且有描述时） -->
        <div
          v-if="row.description"
          class="mt-2.5 rounded-lg p-2.5 text-[13px] leading-[1.7] text-ink-2"
          :style="
            row.desc_tone === 'yellow'
              ? 'background:#FFFBF0;border:1px solid #FFE9C2;'
              : 'background:#FFF7F7;border:1px solid #FFE0E0;'
          "
        >
          {{ row.description }}
        </div>
        <!-- 处理中行进度 -->
        <div v-if="row.status === '处理中'" class="mt-3 flex items-center gap-2.5">
          <span class="flex items-center gap-1.5 text-[12px]" style="color:#00B578;">
            <span class="inline-block h-1.5 w-1.5 rounded-full" style="background:#00B578;"></span>
            {{ row.progress }}
          </span>
          <span class="ml-auto text-[12px] text-ink-4">
            更新于 {{ row.updated_at ? formatTime(row.updated_at).slice(5) : '' }}<template v-if="row.handler"> · 当前处理人：{{ row.handler }}</template>
          </span>
        </div>
        <!-- 操作行 -->
        <div v-if="row.status === '待处理'" class="mt-3 flex items-center gap-2.5">
          <button
            v-permission="'complaint:handle'"
            class="h-[34px] rounded-lg border-none px-5 text-[13px] font-semibold text-white"
            style="background:linear-gradient(135deg,#FF8A3D,#FF6B2C);"
            @click="openDetail(row)"
          >
            开始处理
          </button>
          <button
            class="h-[34px] rounded-lg border border-line bg-white px-5 text-[13px] text-ink-2"
            @click="router.push(`/orders/${row.order_id}`)"
          >
            查看订单
          </button>
          <span class="ml-auto text-[12px] text-ink-4">{{ row.deadline }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
