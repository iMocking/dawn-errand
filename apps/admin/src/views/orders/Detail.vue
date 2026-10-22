<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import StatusTag from '@/components/StatusTag.vue'
import AppButton from '@/components/AppButton.vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { getOrderDetail, type OrderDetail } from '@/api/orders'
import { formatClock, formatTime } from '@/utils/format'
import { useToast } from '@/composables/useToast'

const route = useRoute()
const toast = useToast()
const loading = ref(true)
const detail = ref<OrderDetail | null>(null)

function lineColor(i: number): string {
  const next = detail.value?.steps[i + 1]
  if (!next) return '#E5E6EB'
  return next.state === 'pending' ? '#E5E6EB' : '#FF6B2C'
}

async function load() {
  loading.value = true
  try {
    detail.value = await getOrderDetail(String(route.params.id))
  } finally {
    loading.value = false
  }
}

function onContactRider() {
  toast.success('正在为您接通骑手…')
}

onMounted(load)
</script>

<template>
  <div v-if="loading" class="text-ink-3">加载中…</div>
  <div v-else-if="detail" class="flex flex-col gap-4">
    <!-- 订单概要 -->
    <div class="flex items-center rounded-card bg-white px-5 py-4">
      <div>
        <div class="flex items-center gap-2.5">
          <span class="text-[18px] font-bold text-ink-1">订单 #{{ detail.id }}</span>
          <StatusTag tone="yellow">{{ detail.status }}</StatusTag>
          <StatusTag tone="orange">{{ detail.type }}</StatusTag>
        </div>
        <div class="mt-1.5 text-[12px] text-ink-3">
          下单时间：{{ formatTime(detail.created_at) }} · 下单渠道：{{ detail.channel }} ·
          订单号：{{ detail.order_no }}
        </div>
      </div>
      <div class="ml-auto flex items-center gap-3">
        <AppButton
          variant="secondary"
          icon="download"
          :icon-size="15"
          class="!gap-2"
          @click="toast.info('打印任务已发送')"
        >
          打印小票
        </AppButton>
        <AppButton
          class="shadow-[0_4px_12px_rgba(255,107,44,.3)]"
          @click="onContactRider"
        >
          联系骑手
        </AppButton>
      </div>
    </div>

    <!-- 状态时间线 -->
    <div class="flex items-center rounded-card bg-white px-6 py-[18px]">
      <div class="flex flex-1 items-start">
        <div
          v-for="(s, i) in detail.steps"
          :key="s.title"
          class="relative flex flex-1 flex-col items-center"
        >
          <div
            v-if="i < detail.steps.length - 1"
            class="absolute left-1/2 top-[13px] z-[1] h-0.5 w-full"
            :style="{ background: lineColor(i) }"
          ></div>
          <!-- done -->
          <div
            v-if="s.state === 'done'"
            class="relative z-[2] flex h-7 w-7 items-center justify-center rounded-full"
            style="background:linear-gradient(135deg,#FF8A3D,#FF6B2C);"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5" /></svg>
          </div>
          <!-- active -->
          <div
            v-else-if="s.state === 'active'"
            class="relative z-[2] flex h-7 w-7 items-center justify-center rounded-full"
            style="background:#F7F8FA;border:2px solid #FF6B2C;"
          >
            <div class="h-[10px] w-[10px] rounded-full" style="background:#FF6B2C;"></div>
          </div>
          <!-- pending -->
          <div
            v-else
            class="relative z-[2] h-7 w-7 rounded-full"
            style="background:#F7F8FA;border:2px solid #E5E6EB;"
          ></div>
          <div
            class="mt-2 text-[13px] font-semibold"
            :class="s.state === 'active' ? 'text-primary' : s.state === 'pending' ? 'text-ink-4' : 'text-ink-1'"
          >
            {{ s.title }}
          </div>
          <div
            class="mt-0.5 text-[11px]"
            :class="s.state === 'active' ? 'text-primary' : s.state === 'pending' ? 'text-ink-4' : 'text-ink-3'"
          >
            {{ s.sub ?? (s.time ? formatClock(s.time) : '') }}
          </div>
        </div>
      </div>
      <div class="ml-6 w-[120px] shrink-0 rounded-lg px-[14px] py-2.5 text-center" style="background:#FFF3E8;">
        <div class="text-[11px] text-primary">预计送达</div>
        <div class="mt-0.5 text-[16px] font-bold text-primary">{{ formatClock(detail.estimated_arrival) }}</div>
      </div>
    </div>

    <!-- 三栏 -->
    <div class="grid flex-1 gap-4" style="grid-template-columns:1.2fr 1fr 1fr;">
      <!-- 配送链路 -->
      <div class="flex flex-col rounded-card bg-white px-5 py-4">
        <div class="flex items-center gap-2 text-[15px] font-semibold text-ink-1">
          <span class="flex h-[26px] w-[26px] items-center justify-center rounded-lg" style="background:#FFF3E8;">
            <SvgIcon name="merchant" :size="14" class="text-primary" />
          </span>
          配送链路
        </div>
        <!-- 商家 -->
        <div class="mt-4 flex gap-3">
          <div class="flex flex-col items-center">
            <div class="flex h-[34px] w-[34px] shrink-0 items-center justify-center rounded-full" style="background:#FFF3E8;">
              <SvgIcon name="merchant" :size="16" class="text-primary" />
            </div>
            <div class="my-1 w-0.5 flex-1" style="background:#FF6B2C;"></div>
          </div>
          <div class="flex-1 pb-2">
            <div class="text-[13px] font-semibold text-ink-1">{{ detail.merchant.name }}</div>
            <div class="mt-1 text-[12px] text-ink-3">{{ detail.merchant.address }}</div>
            <div class="mt-0.5 text-[12px] text-ink-3">{{ detail.merchant.phone }}</div>
            <div class="mt-1.5 text-[12px]" style="color:#00B578;">{{ detail.merchant.status }}</div>
          </div>
        </div>
        <!-- 用户 -->
        <div class="flex gap-3">
          <div class="flex flex-col items-center">
            <div class="flex h-[34px] w-[34px] shrink-0 items-center justify-center rounded-full" style="background:#E8F0FF;">
              <SvgIcon name="user" :size="16" style="color:#2B7FFF;" />
            </div>
            <div class="my-1 w-0.5 flex-1" style="background:#E8F0FF;"></div>
          </div>
          <div class="flex-1 pb-2">
            <div class="text-[13px] font-semibold text-ink-1">{{ detail.user.name }}</div>
            <div class="mt-1 text-[12px] text-ink-3">{{ detail.user.address }}</div>
            <div class="mt-0.5 text-[12px] text-ink-3">{{ detail.user.phone }}</div>
          </div>
        </div>
        <!-- 骑手 -->
        <div class="flex gap-3">
          <div class="flex flex-col items-center">
            <div class="flex h-[34px] w-[34px] shrink-0 items-center justify-center rounded-full" style="background:#E8F7F0;">
              <SvgIcon name="grid" :size="16" style="color:#00B578;" />
            </div>
          </div>
          <div class="flex-1">
            <div class="text-[13px] font-semibold text-ink-1">{{ detail.rider.name }}</div>
            <div class="mt-1 text-[12px] text-ink-3">{{ detail.rider.phone }}</div>
            <div class="mt-1.5 flex gap-2">
              <span class="rounded-tag px-2.5 py-[3px] text-[12px] font-semibold" style="background:#E8F7F0;color:#00B578;">{{ detail.rider.service }}</span>
              <span class="rounded-tag px-2.5 py-[3px] text-[12px]" style="background:#F7F8FA;color:#4E5969;">{{ detail.rider.score }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 商品明细 -->
      <div class="flex flex-col rounded-card bg-white px-5 py-4">
        <div class="flex items-center gap-2 text-[15px] font-semibold text-ink-1">
          <span class="flex h-[26px] w-[26px] items-center justify-center rounded-lg" style="background:#E8F0FF;">
            <SvgIcon name="order" :size="14" style="color:#2B7FFF;" />
          </span>
          商品明细
        </div>
        <div class="mt-3.5 flex flex-1 flex-col gap-2.5">
          <div
            v-for="g in detail.goods"
            :key="g.name"
            class="flex items-center gap-2.5"
            :class="g !== detail.goods[detail.goods.length - 1] ? 'border-b border-line pb-2.5' : ''"
          >
            <div class="flex h-[38px] w-[38px] shrink-0 items-center justify-center rounded-lg bg-head">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#86909C" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 11h18c0 5-4 8-9 8s-9-3-9-8z" /><path d="M8 11V7a4 4 0 0 1 8 0v4" /></svg>
            </div>
            <div class="flex-1">
              <div class="text-[13px] font-semibold text-ink-1">{{ g.name }}</div>
              <div class="mt-0.5 text-[12px] text-ink-3">{{ g.spec }}</div>
            </div>
            <div class="text-[13px] font-semibold text-ink-1">¥{{ g.price.toFixed(2) }}</div>
          </div>
        </div>
        <div class="border-t border-line pt-2.5 text-[12px] leading-[1.9] text-ink-3">
          备注：{{ detail.remark }}
        </div>
      </div>

      <!-- 费用明细 -->
      <div class="flex flex-col rounded-card bg-white px-5 py-4">
        <div class="flex items-center gap-2 text-[15px] font-semibold text-ink-1">
          <span class="flex h-[26px] w-[26px] items-center justify-center rounded-lg" style="background:#FFF8E0;">
            <SvgIcon name="money" :size="14" style="color:#D48806;" />
          </span>
          费用明细
        </div>
        <div class="mt-3.5 flex flex-1 flex-col gap-3 text-[13px] text-ink-2">
          <div v-for="f in detail.fees" :key="f.label" class="flex justify-between">
            <span>{{ f.label }}</span>
            <span :class="f.negative ? 'font-medium' : 'text-ink-1'" :style="f.negative ? 'color:#00B578;' : ''">
              {{ f.negative ? '-' : '' }}¥{{ f.value.toFixed(2) }}
            </span>
          </div>
        </div>
        <div class="flex items-center justify-between border-t border-line pt-3">
          <span class="text-[13px] text-ink-2">实付金额</span>
          <span class="text-[20px] font-bold text-primary">¥{{ detail.pay_amount.toFixed(2) }}</span>
        </div>
        <div class="mt-3 rounded-lg bg-head px-3 py-2.5 text-[12px] leading-[1.8] text-ink-3">
          支付方式：{{ detail.pay_method }}<br />
          资金流向：{{ detail.fund_flow }}
        </div>
      </div>
    </div>
  </div>
</template>
