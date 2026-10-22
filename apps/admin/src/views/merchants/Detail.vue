<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import StatusTag from '@/components/StatusTag.vue'
import AppButton from '@/components/AppButton.vue'
import { useToast } from '@/composables/useToast'
import { formatDate, formatTime, formatMoney, formatNumber } from '@/utils/format'
import { getMerchantDetail, type MerchantDetail } from '@/api/merchants'

const route = useRoute()
const toast = useToast()

const detail = ref<MerchantDetail | null>(null)
const loading = ref(true)
const activeTab = ref('基本信息')

const tabs = [
  { label: '基本信息', icon: '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>' },
  { label: '证照信息', icon: '<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="M21 15l-5-5L5 21"/>' },
  { label: '经营数据', icon: '<path d="M3 3v18h18"/><path d="M7 14l4-4 3 3 5-6"/>' },
  { label: '菜单管理', icon: '<path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>' },
  { label: '评价管理', icon: '<path d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3z"/><path d="M7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3"/>', badge: 126 },
  { label: '经营异常', icon: '<circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/>' },
]

const basicFields = computed(() => {
  if (!detail.value) return []
  const b = detail.value.basic
  return [
    { label: '商家名称', value: b.name },
    { label: '店主姓名', value: b.owner },
    { label: '联系电话', value: b.phone },
    { label: '经营类型', value: b.type },
    { label: '门店地址', value: b.address },
    { label: '营业时间', value: b.openHours },
    { label: '配送范围', value: b.deliveryRange },
    { label: '结算费率', value: b.commission },
    { label: '收款账户', value: b.bankAccount },
    { label: '配送方式', value: b.deliveryMethod },
  ]
})

onMounted(async () => {
  const id = route.params.id as string
  try {
    detail.value = await getMerchantDetail(id)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-if="loading" class="text-ink-3">加载中…</div>
  <div v-else-if="detail" class="flex flex-col gap-3">
    <!-- 商家头部卡 -->
    <div class="flex items-center gap-3 rounded-card bg-white p-[12px_20px]">
      <div
        class="flex h-[38px] w-[56px] shrink-0 items-center justify-center rounded-[14px]"
        :style="{ background: detail.logoBg }"
      >
        <svg width="30" height="30" viewBox="0 0 24 24" fill="none" :stroke="detail.logoColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><path d="M3 11h18c0 5-4 8-9 8s-9-3-9-8z" /><path d="M8 11V7a4 4 0 0 1 8 0v4" /></svg>
      </div>
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <span class="text-[18px] font-bold text-ink-1">{{ detail.name }}</span>
          <StatusTag tone="green">{{ detail.status }}</StatusTag>
          <StatusTag tone="orange">{{ detail.category }}</StatusTag>
          <span class="text-[12px] text-ink-4">商家编号 {{ detail.id }} · 入驻 {{ formatDate(detail.joinedAt) }}</span>
        </div>
        <div class="mt-2 flex flex-wrap items-center gap-6 text-[13px] text-ink-3">
          <span class="flex items-center gap-1.5">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#D48806" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" /></svg>
            <b class="text-ink-1">{{ detail.rating }}</b>评分
          </span>
          <span>月订单 <b class="text-ink-1">{{ formatNumber(detail.monthlyOrders) }}</b></span>
          <span>月营业额 <b class="text-ink-1">¥{{ formatNumber(detail.monthlyRevenue) }}</b></span>
          <span>客单价 <b class="text-ink-1">{{ formatMoney(detail.avgOrder) }}</b></span>
          <span class="flex items-center gap-1.5">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#00B578" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" /><polyline points="22 4 12 14.01 9 11.01" /></svg>
            好评率 {{ detail.goodRate }}
          </span>
        </div>
      </div>
      <div class="flex shrink-0 gap-2">
        <AppButton variant="secondary" class="!h-[30px] !px-[18px]">编辑资料</AppButton>
        <AppButton variant="secondary" class="!h-[30px] !px-[18px] !border-[#FFD9D9] !text-[#FF4D4F]">暂停营业</AppButton>
      </div>
    </div>

    <!-- 主体区 -->
    <div class="flex gap-3">
      <!-- 左侧 tab -->
      <div class="flex w-[168px] shrink-0 flex-col rounded-card bg-white py-2">
        <div
          v-for="t in tabs"
          :key="t.label"
          class="mx-2 mb-0.5 flex cursor-pointer items-center gap-2 rounded-lg px-4 py-2.5 text-[14px]"
          :class="activeTab === t.label ? 'bg-primary-bg font-semibold text-primary' : 'text-ink-2 hover:bg-head'"
          @click="activeTab = t.label"
        >
          <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" v-html="t.icon" />
          {{ t.label }}
          <span
            v-if="t.badge"
            class="ml-auto flex h-[18px] min-w-[18px] items-center justify-center rounded-full px-1 text-[11px] font-semibold"
            style="background:#F2F3F5;color:#86909C;"
          >{{ t.badge }}</span>
        </div>
        <div class="flex-1"></div>
        <div class="border-t border-line px-4 py-2.5 text-[12px] text-ink-4">最近操作：{{ detail.lastOperation }}</div>
      </div>

      <!-- 右侧内容 -->
      <div class="flex min-w-0 flex-1 flex-col gap-3">
        <!-- 基本信息卡 -->
        <div class="rounded-card bg-white p-3">
          <div class="flex items-center">
            <div class="text-[15px] font-semibold text-ink-1">基本信息</div>
            <span class="ml-3 text-[12px] text-ink-4">最后更新：{{ formatTime(detail.lastUpdatedAt) }}</span>
            <span class="ml-auto rounded-tag px-2.5 py-1 text-[12px] font-semibold" style="background:#E8F7F0;color:#00B578;">
              资料完整度 {{ detail.completeness }}%
            </span>
          </div>
          <div class="mt-2.5 grid grid-cols-2 gap-2 gap-x-3.5">
            <div v-for="f in basicFields" :key="f.label">
              <div class="mb-1.5 text-[12px] text-ink-3">{{ f.label }}</div>
              <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ f.value }}</div>
            </div>
          </div>
          <div class="mt-2 flex justify-end gap-2 border-t border-line pt-4">
            <AppButton variant="secondary" class="!h-[30px] !px-5">取消</AppButton>
            <AppButton class="!h-[30px] !px-5" @click="toast.success('资料已保存')">保存修改</AppButton>
          </div>
        </div>

        <!-- 今日经营速览 -->
        <div class="rounded-card bg-white p-3">
          <div class="text-[15px] font-semibold text-ink-1">今日经营速览</div>
          <div class="mt-2 grid grid-cols-4 gap-3">
            <div class="rounded-[10px] bg-head p-3.5">
              <div class="text-[12px] text-ink-3">今日订单</div>
              <div class="mt-1 text-[22px] font-bold text-ink-1">{{ detail.today.orders }}</div>
              <div class="mt-1 text-[11px]" style="color:#00B578;">{{ detail.today.ordersDelta }}</div>
            </div>
            <div class="rounded-[10px] bg-head p-3.5">
              <div class="text-[12px] text-ink-3">今日营业额</div>
              <div class="mt-1 text-[22px] font-bold text-ink-1">¥{{ formatNumber(detail.today.revenue) }}</div>
              <div class="mt-1 text-[11px]" style="color:#00B578;">{{ detail.today.revenueDelta }}</div>
            </div>
            <div class="rounded-[10px] bg-head p-3.5">
              <div class="text-[12px] text-ink-3">出餐准时率</div>
              <div class="mt-1 text-[22px] font-bold text-ink-1">{{ detail.today.onTimeRate }}</div>
              <div class="mt-1 text-[11px] text-ink-3">{{ detail.today.onTimeTarget }}</div>
            </div>
            <div class="rounded-[10px] bg-head p-3.5">
              <div class="text-[12px] text-ink-3">待处理异常</div>
              <div class="mt-1 text-[22px] font-bold" style="color:#FF4D4F;">{{ detail.today.pendingAbnormal }}</div>
              <div class="mt-1 cursor-pointer text-[11px] text-primary">查看明细</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
