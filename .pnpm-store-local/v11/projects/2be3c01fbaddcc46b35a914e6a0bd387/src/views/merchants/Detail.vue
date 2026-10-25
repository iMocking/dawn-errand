<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import StatusTag from '@/components/StatusTag.vue'
import AppButton from '@/components/AppButton.vue'
import Modal from '@/components/Modal.vue'
import { useToast } from '@/composables/useToast'
import { formatDate, formatTime, formatMoney, formatNumber } from '@/utils/format'
import {
  getMerchantDetail,
  getMerchantWorkspace,
  resetMerchantAccount,
  type MerchantDetail,
  type MerchantWorkspace,
} from '@/api/merchants'

const route = useRoute()
const toast = useToast()

const detail = ref<MerchantDetail | null>(null)
const workspace = ref<MerchantWorkspace | null>(null)
const loading = ref(true)
const activeTab = ref('基本信息')
const resetOpen = ref(false)
const resetting = ref(false)
const resetAccount = ref<{ email: string; phone: string; initialPassword: string } | null>(null)

const tabs = [
  { label: '基本信息', icon: '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>' },
  { label: '证照信息', icon: '<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="M21 15l-5-5L5 21"/>' },
  { label: '经营数据', icon: '<path d="M3 3v18h18"/><path d="M7 14l4-4 3 3 5-6"/>' },
  { label: '菜单管理', icon: '<path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>' },
  { label: '评价管理', icon: '<path d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3z"/><path d="M7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3"/>' },
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

const menuGroups = computed(() => {
  const menu = workspace.value?.menu
  if (!menu) return []
  const groups = menu.categories.map((c) => ({
    ...c,
    dishes: menu.dishes.filter((d) => d.categoryId === c.id),
  }))
  const other = menu.dishes.filter((d) => !menu.categories.some((c) => c.id === d.categoryId))
  if (other.length) groups.push({ id: '__other__', name: '未分类', sort: 999, dishes: other })
  return groups
})

const trendMax = computed(() => {
  const max = Math.max(1, ...(workspace.value?.trend ?? []).map((t) => t.orders))
  return max
})

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
  const n = Math.max(0, Math.min(5, rating))
  return '★'.repeat(n) + '☆'.repeat(5 - n)
}

async function onResetPassword() {
  if (!detail.value || resetting.value) return
  resetting.value = true
  try {
    const res = await resetMerchantAccount(detail.value.id)
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

onMounted(async () => {
  const id = route.params.id as string
  try {
    detail.value = await getMerchantDetail(id)
    workspace.value = await getMerchantWorkspace(id)
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
          <StatusTag :tone="detail.status === '营业中' ? 'green' : detail.status === '已封禁' ? 'red' : 'gray'">{{ detail.status }}</StatusTag>
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
        <AppButton variant="secondary" class="!h-[30px] !px-[18px]" @click="toast.info('资料变更审批流待接入')">编辑资料</AppButton>
        <AppButton variant="secondary" class="!h-[30px] !px-[18px] !border-[#FFD9D9] !text-[#FF4D4F]" @click="toast.info('营业状态由商家端控制')">暂停营业</AppButton>
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
            v-if="t.label === '评价管理' && workspace?.reviews.total"
            class="ml-auto flex h-[18px] min-w-[18px] items-center justify-center rounded-full bg-head px-1 text-[11px] font-semibold text-[#86909C]"
          >{{ workspace.reviews.total }}</span>
          <span
            v-else-if="t.label === '经营异常' && workspace?.abnormal.length"
            class="ml-auto flex h-[18px] min-w-[18px] items-center justify-center rounded-full px-1 text-[11px] font-semibold"
            style="background:#FFF1F0;color:#FF4D4F;"
          >{{ workspace.abnormal.length }}</span>
        </div>
        <div class="flex-1"></div>
        <div class="border-t border-line px-4 py-2.5 text-[12px] text-ink-4">最近操作：{{ detail.lastOperation }}</div>
      </div>

      <!-- 右侧内容 -->
      <div class="flex min-w-0 flex-1 flex-col gap-3">
        <!-- 基本信息 -->
        <template v-if="activeTab === '基本信息'">
          <div class="rounded-card bg-white p-3">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">基本信息</div>
              <span class="ml-3 text-[12px] text-ink-4">最后更新：{{ formatTime(detail.lastUpdatedAt) }}</span>
              <span class="ml-auto rounded-tag bg-[#E8F7F0] px-2.5 py-1 text-[12px] font-semibold text-[#00B578]">
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
              <AppButton
                v-if="detail.ownerUserId"
                variant="secondary"
                class="!h-[30px] !px-5"
                :loading="resetting"
                @click="onResetPassword"
              >重置登录密码</AppButton>
              <AppButton class="!h-[30px] !px-5" @click="toast.info('资料变更审批流待接入')">申请变更</AppButton>
            </div>
          </div>
        </template>

        <!-- 证照信息 -->
        <template v-else-if="activeTab === '证照信息'">
          <div class="rounded-card bg-white p-4">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">入驻认证</div>
              <span class="ml-2 rounded-tag bg-[#E8F7F0] px-2 py-0.5 text-[11px] font-semibold text-[#00B578]">
                {{ workspace?.auditStatus || detail.auditStatus || '—' }}
              </span>
              <span class="ml-auto text-[12px] text-ink-4">
                {{ detail.ownerUserId ? '商家端账号已开通' : '未开通商家端账号' }} ·
                {{ workspace?.verifiedAt ? `认证于 ${formatTime(workspace.verifiedAt)}` : '未认证' }}
              </span>
            </div>
            <div class="mt-3 grid grid-cols-2 gap-2 gap-x-3.5">
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">营业执照号</div>
                <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ workspace?.application.licenseNo || '—' }}</div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">法人 / 身份证号</div>
                <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ workspace?.application.legalName || '—' }} {{ workspace?.application.legalIdCard || '' }}</div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">食品经营许可证</div>
                <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ workspace?.application.foodLicenseNo || '—' }}</div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">结算账户</div>
                <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">{{ workspace?.application.bankName || '—' }} · {{ workspace?.application.bankAccountName || '—' }} · {{ workspace?.application.bankAccountNo || '—' }}</div>
              </div>
            </div>
            <div class="mt-4 text-[13px] font-semibold text-ink-1">证照影像</div>
            <div v-if="workspace?.certs.length" class="mt-2 flex flex-wrap gap-2.5">
              <a
                v-for="cert in workspace.certs"
                :key="cert.label"
                :href="cert.url || undefined"
                :target="cert.url ? '_blank' : undefined"
                class="w-32 overflow-hidden rounded-lg border border-line bg-white"
              >
                <div class="flex h-20 w-full items-center justify-center bg-head">
                  <img v-if="cert.url" :src="cert.url" alt="" class="h-full w-full object-cover" />
                  <div v-else class="flex flex-col items-center gap-1 text-ink-4">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="M21 15l-5-5L5 21"/></svg>
                    <span class="text-[10px]">未上传</span>
                  </div>
                </div>
                <div class="px-2 py-1 text-[11px] text-ink-3">{{ cert.label }}</div>
              </a>
              <div
                v-for="(img, i) in workspace.application.storefrontImages"
                :key="img"
                class="w-32 overflow-hidden rounded-lg border border-line bg-white"
              >
                <a :href="img" target="_blank"><img :src="img" alt="" class="h-20 w-full object-cover" /></a>
                <div class="px-2 py-1 text-[11px] text-ink-3">门头照 {{ i + 1 }}</div>
              </div>
            </div>
            <div v-else class="mt-2 rounded-lg bg-head py-8 text-center text-[12px] text-ink-4">暂无证照资料（历史商家未补录）</div>
            <div v-if="workspace?.application.notes" class="mt-3 rounded-lg bg-head px-3 py-2 text-[12px] text-ink-3">
              补充说明：{{ workspace.application.notes }}
            </div>
          </div>
        </template>

        <!-- 经营数据 -->
        <template v-else-if="activeTab === '经营数据'">
          <div class="grid grid-cols-4 gap-3">
            <div class="rounded-card bg-white p-4">
              <div class="text-[12px] text-ink-3">今日订单</div>
              <div class="mt-1 text-[24px] font-bold text-ink-1">{{ detail.today.orders }}</div>
              <div class="mt-1 text-[11px] text-ink-4">累计 {{ detail.monthlyOrders }} 单/月</div>
            </div>
            <div class="rounded-card bg-white p-4">
              <div class="text-[12px] text-ink-3">今日营业额</div>
              <div class="mt-1 text-[24px] font-bold text-ink-1">¥{{ formatNumber(detail.today.revenue) }}</div>
              <div class="mt-1 text-[11px] text-ink-4">客单价 {{ formatMoney(detail.avgOrder) }}</div>
            </div>
            <div class="rounded-card bg-white p-4">
              <div class="text-[12px] text-ink-3">出餐准时率</div>
              <div class="mt-1 text-[24px] font-bold text-ink-1">{{ detail.today.onTimeRate }}</div>
              <div class="mt-1 text-[11px] text-ink-4">目标 {{ detail.today.onTimeTarget }}</div>
            </div>
            <div class="rounded-card bg-white p-4">
              <div class="text-[12px] text-ink-3">好评率</div>
              <div class="mt-1 text-[24px] font-bold text-ink-1">{{ detail.goodRate }}</div>
              <div class="mt-1 text-[11px] text-ink-4">{{ detail.reviewCount }} 条评价</div>
            </div>
          </div>

          <div class="rounded-card bg-white p-4">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">近 7 日订单趋势</div>
              <span class="ml-2 text-[12px] text-ink-4">按订单量</span>
            </div>
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
              <div class="text-[15px] font-semibold text-ink-1">最近订单</div>
              <span class="ml-auto text-[12px] text-ink-4">共 {{ workspace?.orders.length ?? 0 }} 条</span>
            </div>
            <div v-if="recentOrders.length" class="mt-3 overflow-hidden rounded-lg border border-line">
              <div class="grid grid-cols-[1.4fr_1fr_0.8fr_0.8fr_1fr] bg-head px-3 py-2 text-[12px] text-ink-3">
                <span>订单号</span><span>用户</span><span>金额</span><span>状态</span><span>下单时间</span>
              </div>
              <div v-for="o in recentOrders" :key="o.id" class="grid grid-cols-[1.4fr_1fr_0.8fr_0.8fr_1fr] items-center border-t border-line px-3 py-2.5 text-[12px]">
                <span class="font-mono text-ink-2">{{ o.orderNo }}</span>
                <span class="text-ink-2">{{ o.userName || '—' }}</span>
                <span class="font-semibold text-ink-1">{{ formatMoney(o.payAmount) }}</span>
                <StatusTag :tone="orderTone(o.status)">{{ o.status }}</StatusTag>
                <span class="text-ink-3">{{ formatTime(o.createdAt) }}</span>
              </div>
            </div>
            <div v-else class="py-10 text-center text-[12px] text-ink-4">暂无订单</div>
          </div>
        </template>

        <!-- 菜单管理 -->
        <template v-else-if="activeTab === '菜单管理'">
          <div class="rounded-card bg-white p-4">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">菜单与商品</div>
              <span class="ml-2 text-[12px] text-ink-4">{{ workspace?.menu.categories.length ?? 0 }} 个分类 · {{ workspace?.menu.dishes.length ?? 0 }} 个商品</span>
              <span class="ml-auto text-[12px] text-ink-4">商家端维护，平台侧只读</span>
            </div>
            <div v-if="menuGroups.length" class="mt-3 flex flex-col gap-3">
              <div v-for="g in menuGroups" :key="g.id" class="overflow-hidden rounded-lg border border-line">
                <div class="flex items-center bg-head px-3 py-2">
                  <span class="text-[13px] font-semibold text-ink-1">{{ g.name }}</span>
                  <span class="ml-2 text-[11px] text-ink-4">{{ g.dishes.length }} 个商品</span>
                </div>
                <div v-if="g.dishes.length" class="divide-y divide-line">
                  <div v-for="d in g.dishes" :key="d.id" class="flex items-center gap-3 px-3 py-2.5">
                    <div class="h-10 w-10 shrink-0 overflow-hidden rounded-md bg-head">
                      <img v-if="d.imageUrl" :src="d.imageUrl" alt="" class="h-full w-full object-cover" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <div class="flex items-center gap-2">
                        <span class="text-[13px] font-semibold text-ink-1">{{ d.name }}</span>
                        <span v-if="!d.onSale" class="rounded bg-head px-1.5 py-0.5 text-[10px] text-ink-4">已下架</span>
                        <span v-else-if="d.soldOut" class="rounded bg-[#FFF8E0] px-1.5 py-0.5 text-[10px] text-[#D48806]">售罄</span>
                      </div>
                      <div class="mt-0.5 truncate text-[11px] text-ink-4">{{ d.description || '暂无描述' }}</div>
                    </div>
                    <div class="text-right">
                      <div class="text-[13px] font-semibold text-ink-1">{{ formatMoney(d.price) }}</div>
                      <div class="text-[11px] text-ink-4">库存 {{ d.stock }} · 月售 {{ d.monthlySales }}</div>
                    </div>
                  </div>
                </div>
                <div v-else class="px-3 py-4 text-[12px] text-ink-4">该分类暂无商品</div>
              </div>
            </div>
            <div v-else class="py-12 text-center text-[12px] text-ink-4">尚未创建菜单</div>
          </div>
        </template>

        <!-- 评价管理 -->
        <template v-else-if="activeTab === '评价管理'">
          <div class="rounded-card bg-white p-4">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">用户评价</div>
              <span class="ml-2 text-[12px] text-ink-4">共 {{ workspace?.reviews.total ?? 0 }} 条 · 综合 {{ workspace?.reviews.rating ?? detail.rating }} 分</span>
              <span class="ml-auto text-[12px] text-ink-4">商家可回复，平台侧只读</span>
            </div>
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
                <div v-if="rv.reply" class="mt-2 rounded-lg bg-head px-3 py-2 text-[12px] text-ink-2">
                  <span class="font-semibold text-ink-1">商家回复：</span>{{ rv.reply }}
                </div>
                <div v-else class="mt-2 text-[11px] text-ink-4">商家未回复</div>
              </div>
            </div>
            <div v-else class="py-12 text-center text-[12px] text-ink-4">暂无评价</div>
          </div>
        </template>

        <!-- 经营异常 -->
        <template v-else-if="activeTab === '经营异常'">
          <div class="rounded-card bg-white p-4">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">异常与退款订单</div>
              <span class="ml-2 text-[12px] text-ink-4">含异常、出餐超时、已取消、退款售后、已退款</span>
              <span class="ml-auto rounded-full bg-[#FFF1F0] px-2.5 py-1 text-[12px] font-semibold text-[#FF4D4F]">
                {{ workspace?.abnormal.length ?? 0 }} 条
              </span>
            </div>
            <div v-if="workspace?.abnormal.length" class="mt-3 flex flex-col gap-2">
              <div v-for="o in workspace.abnormal" :key="o.id" class="flex items-center gap-3 rounded-lg border border-line px-3 py-2.5">
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="font-mono text-[12px] text-ink-2">{{ o.orderNo }}</span>
                    <StatusTag :tone="orderTone(o.status)">{{ o.status }}</StatusTag>
                  </div>
                  <div class="mt-1 truncate text-[11px] text-ink-4">{{ o.userName }} · {{ formatMoney(o.payAmount) }}</div>
                </div>
                <span class="text-[11px] text-ink-4">{{ formatTime(o.createdAt) }}</span>
              </div>
            </div>
            <div v-else class="py-12 text-center text-[12px] text-ink-4">暂无经营异常</div>
          </div>
        </template>
      </div>
    </div>

    <Modal v-model:open="resetOpen" title="商家密码已重置" :width="480" confirm-text="我已记录" @confirm="resetOpen = false">
      <div class="rounded-lg bg-[#FFF8E0] px-3 py-2.5 text-[12px]" style="color:#D48806;">
        旧密码与旧会话已失效，请将新密码线下交付给商家。
      </div>
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
