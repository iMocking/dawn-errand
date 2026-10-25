<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import StatusTag from '@/components/StatusTag.vue'
import Confirm from '@/components/Confirm.vue'
import Modal from '@/components/Modal.vue'
import AppButton from '@/components/AppButton.vue'
import { useToast } from '@/composables/useToast'
import { formatDate, formatMoney, formatTime } from '@/utils/format'
import {
  getRiderDetail,
  getRiderWorkspace,
  freezeRider,
  resetRiderAccount,
  type RiderDetailDto,
  type RiderWorkspace,
} from '@/api/riders'

const route = useRoute()
const toast = useToast()

const detail = ref<RiderDetailDto | null>(null)
const workspace = ref<RiderWorkspace | null>(null)
const loading = ref(true)
const activeTab = ref('basic')
const resetOpen = ref(false)
const resetting = ref(false)
const resetAccount = ref<{ email: string; phone: string; initialPassword: string } | null>(null)

const tabs = [
  { key: 'basic', label: '基本信息', icon: 'order' },
  { key: 'stats', label: '配送统计', icon: 'chart' },
  { key: 'score', label: '绩效评分', icon: 'star' },
  { key: 'cert', label: '证件管理', icon: 'image' },
  { key: 'logs', label: '奖惩记录', icon: 'layers' },
] as const

const lastActionText = computed(() => {
  if (!detail.value) return ''
  return `${formatDate(detail.value.lastActionAt)} ${detail.value.lastAction}`
})

const trendMax = computed(() => Math.max(1, ...(workspace.value?.trend ?? []).map((t) => t.orders)))
const recentOrders = computed(() => (workspace.value?.orders ?? []).slice(0, 10))

function orderTone(status: string): 'green' | 'red' | 'gray' | 'yellow' | 'blue' | 'orange' {
  if (status === '已完成') return 'green'
  if (status === '已取消' || status === '已退款') return 'gray'
  if (status === '异常' || status === '退款售后') return 'red'
  if (status === '待支付') return 'orange'
  if (status === '待评价' || status === '已送达') return 'blue'
  return 'yellow'
}

function stars(rating: number): string {
  const n = Math.max(0, Math.min(5, Math.round(rating)))
  return '★'.repeat(n) + '☆'.repeat(5 - n)
}

async function load() {
  loading.value = true
  try {
    const id = String(route.params.id)
    detail.value = await getRiderDetail(id)
    workspace.value = await getRiderWorkspace(id)
  } finally {
    loading.value = false
  }
}

/* 冻结 / 解冻 */
const freezeOpen = ref(false)
const freezing = ref(false)
const willFreeze = computed(() => detail.value?.status !== '已冻结')
async function onFreeze() {
  if (!detail.value || freezing.value) return
  freezing.value = true
  try {
    const action = willFreeze.value ? 'freeze' : 'unfreeze'
    await freezeRider(detail.value.id, action)
    detail.value.status = willFreeze.value ? '已冻结' : '在线'
    toast.success(willFreeze.value ? '已冻结该骑手' : '已解冻该骑手')
    freezeOpen.value = false
  } finally {
    freezing.value = false
  }
}

async function onResetPassword() {
  if (!detail.value || resetting.value) return
  resetting.value = true
  try {
    const res = await resetRiderAccount(detail.value.id)
    resetAccount.value = res.account
    resetOpen.value = true
  } finally {
    resetting.value = false
  }
}

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    toast.success('已复制')
  } catch {
    toast.warning('复制失败，请手动选择')
  }
}

onMounted(load)
</script>

<template>
  <div v-if="loading" class="text-ink-3">加载中…</div>
  <div v-else-if="detail" class="flex flex-col gap-4">
    <!-- 头部 -->
    <div class="flex items-center gap-4 rounded-card bg-white px-5 py-[18px]">
      <div class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full text-[22px] font-bold text-white" :style="{ background: detail.avatarBg }">
        {{ detail.name.charAt(0) }}
      </div>
      <div class="min-w-0 flex-1">
        <div class="flex flex-wrap items-center gap-2.5">
          <span class="text-[18px] font-bold text-ink-1">{{ detail.name }}</span>
          <StatusTag :tone="detail.status === '已冻结' ? 'red' : detail.status === '在线' ? 'green' : 'gray'">{{ detail.status }}</StatusTag>
          <span class="inline-block rounded-[4px] px-2.5 py-[3px] text-[12px] font-semibold" style="background:#FFF3E8;color:#FF6B2C;">{{ detail.serviceType }}</span>
          <span class="text-[12px] text-ink-4">骑手编号 {{ detail.id }} · 入驻 {{ formatDate(detail.joinedAt) }}</span>
        </div>
        <div class="mt-2 flex flex-wrap items-center gap-6 text-[13px] text-ink-3">
          <span class="flex items-center gap-1.5">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#D48806" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" /></svg>
            <b class="text-ink-1">{{ detail.rating }}</b>评分
          </span>
          <span>今日完成 <b class="text-ink-1">{{ detail.todayOrders }}</b> 单</span>
          <span>今日在线 <b class="text-ink-1">{{ detail.onlineHours }}</b> 小时</span>
          <span>本月收入 <b class="text-ink-1">{{ formatMoney(detail.monthIncome) }}</b></span>
          <span>准时率 <b class="text-ink-1">{{ detail.onTimeRate }}</b></span>
        </div>
      </div>
      <div class="flex shrink-0 gap-2.5">
        <AppButton variant="secondary" class="!px-[18px]" @click="toast.info('骑手资料变更审批流待接入')">编辑资料</AppButton>
        <AppButton
          variant="secondary"
          class="!px-[18px]"
          :class="willFreeze ? '!border-[#FFD9D9] !text-[#FF4D4F]' : '!border-[#B7E8D0] !text-[#00B578]'"
          @click="freezeOpen = true"
        >
          {{ willFreeze ? '冻结账号' : '解冻账号' }}
        </AppButton>
      </div>
    </div>

    <div class="flex gap-4">
      <!-- 左侧 tab -->
      <div class="flex w-[168px] shrink-0 flex-col rounded-card bg-white py-2">
        <div
          v-for="t in tabs"
          :key="t.key"
          class="mx-2 my-[2px] flex cursor-pointer items-center gap-2 rounded-lg px-4 py-3 text-[14px]"
          :class="activeTab === t.key ? 'bg-primary-bg font-semibold text-primary' : 'text-ink-2 hover:bg-head'"
          @click="activeTab = t.key"
        >
          <svg v-if="t.icon === 'star'" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" /></svg>
          <svg v-else-if="t.icon === 'image'" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" /><circle cx="9" cy="9" r="2" /><path d="M21 15l-5-5L5 21" /></svg>
          <svg v-else-if="t.icon === 'layers'" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 2 7 12 12 22 7 12 2" /><polyline points="2 17 12 22 22 17" /><polyline points="2 12 12 17 22 12" /></svg>
          <svg v-else-if="t.icon === 'chart'" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v18h18" /><path d="M7 14l4-4 3 3 5-6" /></svg>
          <svg v-else width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><polyline points="14 2 14 8 20 8" /></svg>
          {{ t.label }}
          <span v-if="t.key === 'logs' && workspace?.records.length" class="ml-auto flex h-[18px] min-w-[18px] items-center justify-center rounded-full bg-head px-1 text-[11px] font-semibold text-ink-3">{{ workspace.records.length }}</span>
        </div>
        <div class="flex-1" />
        <div class="border-t border-line px-4 py-3 text-[12px] text-ink-4">最近操作：{{ lastActionText }}</div>
      </div>

      <!-- 右侧内容 -->
      <div class="flex min-w-0 flex-1 flex-col gap-4">
        <!-- 基本信息 -->
        <template v-if="activeTab === 'basic'">
          <div class="rounded-card bg-white p-5">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">基本信息</div>
              <div class="ml-auto rounded-[4px] px-2.5 py-[3px] text-[12px] font-semibold" style="background:#E8F7F0;color:#00B578;">
                {{ workspace?.profile.auditStatus || '—' }}
              </div>
            </div>
            <div class="mt-[18px] grid grid-cols-2 gap-x-7 gap-y-4">
              <div v-for="f in [
                { label: '姓名', value: detail.profile.name },
                { label: '联系电话', value: detail.profile.phone },
                { label: '身份证号', value: detail.profile.idCard },
                { label: '常驻区域', value: detail.profile.area },
                { label: '配送工具', value: detail.profile.vehicle },
                { label: '可接单类型', value: detail.profile.orderTypes },
                { label: '收款账户', value: detail.profile.account },
                { label: '加入时间', value: `${detail.profile.joinedAt} · 服务 ${detail.profile.serviceMonths} 个月` },
              ]" :key="f.label">
                <div class="mb-1.5 text-[12px] text-ink-3">{{ f.label }}</div>
                <div class="flex h-[38px] items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ f.value || '—' }}</div>
              </div>
            </div>
            <div class="mt-3 flex justify-end gap-2 border-t border-line pt-4">
              <AppButton v-if="workspace?.profile.auditStatus === '已通过'" variant="secondary" class="!h-[30px] !px-5" :loading="resetting" @click="onResetPassword">重置登录密码</AppButton>
            </div>
          </div>
          <div class="rounded-card bg-white p-5">
            <div class="text-[15px] font-semibold text-ink-1">今日配送速览</div>
            <div class="mt-4 grid grid-cols-4 gap-4">
              <div class="rounded-[10px] bg-head p-3.5">
                <div class="text-[12px] text-ink-3">今日完成</div>
                <div class="mt-1.5 text-2xl font-bold text-ink-1">{{ detail.todayOverview.todayOrders }}</div>
              </div>
              <div class="rounded-[10px] bg-head p-3.5">
                <div class="text-[12px] text-ink-3">准时率</div>
                <div class="mt-1.5 text-2xl font-bold text-ink-1">{{ detail.todayOverview.onTimeRate }}</div>
              </div>
              <div class="rounded-[10px] bg-head p-3.5">
                <div class="text-[12px] text-ink-3">平均配送时长</div>
                <div class="mt-1.5 text-2xl font-bold text-ink-1">{{ detail.todayOverview.avgMinutes }}</div>
              </div>
              <div class="rounded-[10px] bg-head p-3.5">
                <div class="text-[12px] text-ink-3">本月投诉</div>
                <div class="mt-1.5 text-2xl font-bold text-ink-1">{{ detail.todayOverview.complaints }}</div>
              </div>
            </div>
          </div>
        </template>

        <!-- 配送统计 -->
        <template v-else-if="activeTab === 'stats'">
          <div class="grid grid-cols-4 gap-3">
            <div class="rounded-card bg-white p-4">
              <div class="text-[12px] text-ink-3">今日完成</div>
              <div class="mt-1 text-[24px] font-bold text-ink-1">{{ workspace?.performance.todayOrders ?? detail.todayOrders }}</div>
              <div class="mt-1 text-[11px] text-ink-4">今日在线 {{ workspace?.performance.onlineHours ?? detail.onlineHours }} 小时</div>
            </div>
            <div class="rounded-card bg-white p-4">
              <div class="text-[12px] text-ink-3">本月收入</div>
              <div class="mt-1 text-[24px] font-bold text-ink-1">{{ formatMoney(workspace?.performance.monthIncome ?? detail.monthIncome) }}</div>
            </div>
            <div class="rounded-card bg-white p-4">
              <div class="text-[12px] text-ink-3">准时率</div>
              <div class="mt-1 text-[24px] font-bold text-ink-1">{{ workspace?.performance.onTimeRate ?? detail.onTimeRate }}</div>
            </div>
            <div class="rounded-card bg-white p-4">
              <div class="text-[12px] text-ink-3">平均时长</div>
              <div class="mt-1 text-[24px] font-bold text-ink-1">{{ workspace?.performance.avgMinutes ?? 0 }} 分钟</div>
            </div>
          </div>
          <div class="rounded-card bg-white p-4">
            <div class="text-[15px] font-semibold text-ink-1">近 7 日配送量</div>
            <div class="mt-4 flex h-[160px] items-end gap-3">
              <div v-for="p in workspace?.trend ?? []" :key="p.date" class="flex flex-1 flex-col items-center gap-2">
                <div class="w-full rounded-t-md bg-[linear-gradient(180deg,#FF8A3D,#FF6B2C)]" :style="{ height: `${Math.max(4, (p.orders / trendMax) * 120)}px`, opacity: p.orders ? 1 : 0.25 }" />
                <div class="text-[11px] text-ink-3">{{ p.date.slice(5) }}</div>
                <div class="text-[11px] font-semibold text-ink-1">{{ p.orders }} 单</div>
              </div>
            </div>
          </div>
          <div class="rounded-card bg-white p-4">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">最近配送订单</div>
              <span class="ml-auto text-[12px] text-ink-4">共 {{ workspace?.orders.length ?? 0 }} 条</span>
            </div>
            <div v-if="recentOrders.length" class="mt-3 overflow-hidden rounded-lg border border-line">
              <div class="grid grid-cols-[1.4fr_1fr_0.8fr_0.8fr_1fr] bg-head px-3 py-2 text-[12px] text-ink-3">
                <span>订单号</span><span>用户</span><span>金额</span><span>状态</span><span>时间</span>
              </div>
              <div v-for="o in recentOrders" :key="o.id" class="grid grid-cols-[1.4fr_1fr_0.8fr_0.8fr_1fr] items-center border-t border-line px-3 py-2.5 text-[12px]">
                <span class="font-mono text-ink-2">{{ o.orderNo }}</span>
                <span class="text-ink-2">{{ o.userName || '—' }}</span>
                <span class="font-semibold text-ink-1">{{ formatMoney(o.payAmount) }}</span>
                <StatusTag :tone="orderTone(o.status)">{{ o.status }}</StatusTag>
                <span class="text-ink-3">{{ formatTime(o.createdAt) }}</span>
              </div>
            </div>
            <div v-else class="py-10 text-center text-[12px] text-ink-4">暂无配送订单</div>
          </div>
        </template>

        <!-- 绩效评分 -->
        <template v-else-if="activeTab === 'score'">
          <div class="rounded-card bg-white p-5">
            <div class="flex items-center gap-6">
              <div class="text-center">
                <div class="text-[36px] font-bold text-[#D48806]">{{ workspace?.reviews.rating ?? detail.rating }}</div>
                <div class="text-[13px] tracking-[2px]" style="color:#D48806;">{{ stars(workspace?.reviews.rating ?? detail.rating) }}</div>
                <div class="mt-1 text-[12px] text-ink-4">{{ workspace?.reviews.total ?? 0 }} 条评价</div>
              </div>
              <div class="grid flex-1 grid-cols-3 gap-3">
                <div class="rounded-lg bg-head p-3">
                  <div class="text-[12px] text-ink-3">准时率</div>
                  <div class="mt-1 text-[18px] font-bold text-ink-1">{{ workspace?.performance.onTimeRate ?? detail.onTimeRate }}</div>
                </div>
                <div class="rounded-lg bg-head p-3">
                  <div class="text-[12px] text-ink-3">平均时长</div>
                  <div class="mt-1 text-[18px] font-bold text-ink-1">{{ workspace?.performance.avgMinutes ?? 0 }} 分钟</div>
                </div>
                <div class="rounded-lg bg-head p-3">
                  <div class="text-[12px] text-ink-3">今日投诉</div>
                  <div class="mt-1 text-[18px] font-bold text-ink-1">{{ workspace?.performance.complaintsToday ?? detail.todayOverview.complaints }}</div>
                </div>
              </div>
            </div>
          </div>
          <div class="rounded-card bg-white p-4">
            <div class="text-[15px] font-semibold text-ink-1">评价明细</div>
            <div v-if="workspace?.reviews.list.length" class="mt-3 flex flex-col gap-2.5">
              <div v-for="rv in workspace.reviews.list" :key="rv.id" class="rounded-lg border border-line p-3">
                <div class="flex items-center gap-2">
                  <span class="text-[13px] font-semibold text-ink-1">{{ rv.userName || '匿名用户' }}</span>
                  <span class="text-[13px] tracking-[1px]" style="color:#D48806;">{{ stars(rv.rating) }}</span>
                  <span class="ml-auto text-[11px] text-ink-4">{{ formatTime(rv.createdAt) }}</span>
                </div>
                <div class="mt-1.5 text-[13px] text-ink-2">{{ rv.content || '（未填写评价内容）' }}</div>
                <div v-if="rv.tags.length" class="mt-2 flex flex-wrap gap-1.5">
                  <span v-for="tag in rv.tags" :key="tag" class="rounded bg-head px-2 py-0.5 text-[11px] text-ink-3">{{ tag }}</span>
                </div>
              </div>
            </div>
            <div v-else class="py-10 text-center text-[12px] text-ink-4">暂无评价</div>
          </div>
        </template>

        <!-- 证件管理 -->
        <template v-else-if="activeTab === 'cert'">
          <div class="rounded-card bg-white p-5">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">证件与账号</div>
              <span class="ml-2 rounded-tag bg-[#E8F7F0] px-2 py-0.5 text-[11px] font-semibold text-[#00B578]">{{ workspace?.profile.auditStatus || '—' }}</span>
              <span class="ml-auto text-[12px] text-ink-4">认证于 {{ workspace?.profile.joinedAt ? formatTime(workspace.profile.joinedAt) : '—' }}</span>
            </div>
            <div class="mt-3 grid grid-cols-2 gap-2 gap-x-3.5">
              <div v-for="f in [
                { label: '身份证号', value: workspace?.profile.idCard },
                { label: '身份核验状态', value: workspace?.profile.idCheck },
                { label: '配送工具', value: workspace?.profile.tool },
                { label: '车辆信息', value: workspace?.profile.vehicle },
                { label: '收款账户', value: workspace?.profile.account },
                { label: '已上传材料', value: `${workspace?.profile.docCount ?? 0} 张 · ${workspace?.profile.docs || '—'}` },
              ]" :key="f.label">
                <div class="mb-1.5 text-[12px] text-ink-3">{{ f.label }}</div>
                <div class="flex min-h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ f.value || '—' }}</div>
              </div>
            </div>
            <div class="mt-3 rounded-lg bg-[#FFF8E0] px-3 py-2.5 text-[12px]" style="color:#D48806;">
              当前为人工核验 + 材料留档；身份证/健康证/车辆证件影像可后续接入对象存储与 OCR。
            </div>
          </div>
        </template>

        <!-- 奖惩记录 -->
        <template v-else-if="activeTab === 'logs'">
          <div class="rounded-card bg-white p-4">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">奖惩与操作记录</div>
              <span class="ml-auto text-[12px] text-ink-4">{{ workspace?.records.length ?? 0 }} 条</span>
            </div>
            <div v-if="workspace?.records.length" class="mt-3 flex flex-col gap-2">
              <div v-for="(rec, i) in workspace.records" :key="i" class="flex items-start gap-3 rounded-lg border border-line px-3 py-2.5">
                <span class="mt-1 h-2 w-2 shrink-0 rounded-full" :style="{ background: rec.tone === 'warning' ? '#FF4D4F' : rec.tone === 'success' ? '#00B578' : '#2B7FFF' }" />
                <div class="min-w-0 flex-1">
                  <div class="text-[13px] font-semibold text-ink-1">{{ rec.title }}</div>
                  <div class="mt-0.5 truncate text-[11px] text-ink-4">{{ rec.detail }}</div>
                </div>
                <span class="text-[11px] text-ink-4">{{ rec.time ? formatTime(rec.time) : '—' }}</span>
              </div>
            </div>
            <div v-else class="py-10 text-center text-[12px] text-ink-4">暂无奖惩记录</div>
          </div>
          <div class="rounded-card bg-white p-4">
            <div class="text-[15px] font-semibold text-ink-1">异常配送</div>
            <div v-if="workspace?.abnormal.length" class="mt-3 flex flex-col gap-2">
              <div v-for="o in workspace.abnormal" :key="o.id" class="flex items-center gap-3 rounded-lg border border-line px-3 py-2.5">
                <span class="font-mono text-[12px] text-ink-2">{{ o.orderNo }}</span>
                <StatusTag :tone="orderTone(o.status)">{{ o.status }}</StatusTag>
                <span class="text-[11px] text-ink-4">{{ o.userName }}</span>
                <span class="ml-auto text-[11px] text-ink-4">{{ formatTime(o.createdAt) }}</span>
              </div>
            </div>
            <div v-else class="py-8 text-center text-[12px] text-ink-4">暂无异常配送</div>
          </div>
        </template>
      </div>
    </div>

    <Confirm
      :open="freezeOpen"
      :title="willFreeze ? '冻结账号' : '解冻账号'"
      :message="`确认${willFreeze ? '冻结' : '解冻'}骑手 ${detail.name}（${detail.id}）的账号吗？${willFreeze ? '冻结后将停止派单。' : ''}`"
      :confirm-text="willFreeze ? '确认冻结' : '确认解冻'"
      @update:open="(v) => (freezeOpen = v)"
      @confirm="onFreeze"
    />

    <Modal v-model:open="resetOpen" title="骑手密码已重置" :width="460" confirm-text="我已记录" @confirm="resetOpen = false">
      <div class="rounded-lg bg-[#FFF8E0] px-3 py-2.5 text-[12px]" style="color:#D48806;">旧密码与旧会话已失效，请将新密码线下交付骑手。</div>
      <div class="mt-4 space-y-3">
        <div class="flex items-center justify-between rounded-lg border border-line px-3 py-2.5">
          <div>
            <div class="text-[11px] text-ink-4">登录手机号</div>
            <div class="mt-0.5 font-mono text-[13px] font-semibold text-ink-1">{{ resetAccount?.phone || '—' }}</div>
          </div>
          <button v-if="resetAccount?.phone" type="button" class="text-[12px] text-primary" @click="copyText(resetAccount.phone)">复制</button>
        </div>
        <div class="flex items-center justify-between rounded-lg border border-line px-3 py-2.5">
          <div>
            <div class="text-[11px] text-ink-4">新初始密码</div>
            <div class="mt-0.5 font-mono text-[13px] font-semibold text-ink-1">{{ resetAccount?.initialPassword || '—' }}</div>
          </div>
          <button v-if="resetAccount?.initialPassword" type="button" class="text-[12px] text-primary" @click="copyText(resetAccount.initialPassword)">复制</button>
        </div>
      </div>
    </Modal>
  </div>
</template>
