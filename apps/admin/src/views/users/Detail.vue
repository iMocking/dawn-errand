<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import StatusTag from '@/components/StatusTag.vue'
import Modal from '@/components/Modal.vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import AppButton from '@/components/AppButton.vue'
import Pagination from '@/components/Pagination.vue'
import { useToast } from '@/composables/useToast'
import { formatDate, formatMoney, formatTime } from '@/utils/format'
import {
  getUserDetail,
  getUserOrders,
  getUserComplaints,
  banUser,
  unbanUser,
  updateUserRole,
  auditRealName,
  BACKEND_ROLES,
  type UserDetail,
  type UserRecentOrder,
} from '@/api/users'
import type { ComplaintRow } from '@/api/complaints'

const route = useRoute()
const toast = useToast()

type DetailTab = 'basic' | 'orders' | 'complaints' | 'account'

const detail = ref<UserDetail | null>(null)
const loading = ref(true)
const activeTab = ref<DetailTab>('basic')

/* 封禁 / 解封 */
const confirmOpen = ref(false)
const submitting = ref(false)

const confirmTitle = () => (detail.value?.status === '已封禁' ? '解封账号' : '封禁账号')
const confirmMessage = () =>
  detail.value?.status === '已封禁'
    ? `确认解封用户 ${detail.value?.name}（${detail.value?.userNo}）？解封后该账号可正常登录下单。`
    : `确认封禁用户 ${detail.value?.name}（${detail.value?.userNo}）？封禁后该账号将无法登录下单。`

async function onConfirm() {
  if (!detail.value) return
  submitting.value = true
  try {
    const willBan = detail.value.status === '正常'
    if (willBan) {
      await banUser(detail.value.id)
    } else {
      await unbanUser(detail.value.id)
    }
    detail.value.status = willBan ? '已封禁' : '正常'
    toast.success(willBan ? '已封禁该账号' : '已解封该账号')
    confirmOpen.value = false
  } finally {
    submitting.value = false
  }
}

/* 后台角色（账号权限，集成自原系统设置 → 账号权限） */
const roleOpen = ref(false)
const roleSubmitting = ref(false)
const roleForm = ref('user')

function openRoleModal() {
  roleForm.value = detail.value?.role || 'user'
  roleOpen.value = true
}

async function onRoleSubmit() {
  if (!detail.value || roleSubmitting.value) return
  roleSubmitting.value = true
  try {
    await updateUserRole(detail.value.id, roleForm.value)
    detail.value.role = roleForm.value
    detail.value.roleName = BACKEND_ROLES.find((r) => r.value === roleForm.value)?.label ?? roleForm.value
    toast.success('角色已更新，该账号下次操作即按新权限生效')
    roleOpen.value = false
  } finally {
    roleSubmitting.value = false
  }
}

function roleTone(role: string): 'orange' | 'blue' | 'green' | 'gray' | 'red' | 'yellow' {
  switch (role) {
    case 'super_admin':
      return 'orange'
    case 'cs_manager':
    case 'cs_agent':
    case 'region_station':
      return 'blue'
    case 'merchant':
    case 'driver':
      return 'green'
    default:
      return 'gray'
  }
}

/* 实名认证审核（账号状态 tab） */
const realNameOpen = ref(false)
const realNameSubmitting = ref(false)
const realNameRejectReason = ref('')

function openRealNameModal() {
  realNameRejectReason.value = ''
  realNameOpen.value = true
}

async function onRealNameApprove() {
  const d = detail.value
  const v = d?.realNameVerify
  if (!v || realNameSubmitting.value) return
  realNameSubmitting.value = true
  try {
    await auditRealName(v.id, 'approve')
    d.verified = '已实名'
    d.realName = v.realName
    v.status = '已通过'
    toast.success('已通过实名认证')
    realNameOpen.value = false
  } finally {
    realNameSubmitting.value = false
  }
}

async function onRealNameReject() {
  const d = detail.value
  const v = d?.realNameVerify
  if (!v || realNameSubmitting.value) return
  realNameSubmitting.value = true
  try {
    await auditRealName(v.id, 'reject', realNameRejectReason.value)
    v.status = '已驳回'
    v.rejectReason = realNameRejectReason.value
    toast.info('已驳回实名认证申请')
    realNameOpen.value = false
  } finally {
    realNameSubmitting.value = false
  }
}

/** 近期订单时间列：'MM-DD HH:mm'（formatTime 去掉年份前缀） */
function shortTime(sec: number) {
  return formatTime(sec).slice(5)
}

/** ---- 订单记录 tab ---- */
const ordersRows = ref<UserRecentOrder[]>([])
const ordersTotal = ref(0)
const ordersPage = ref(1)
const ordersLoading = ref(false)
const ordersLoaded = ref(false)
const ordersRange = ref<'30d' | 'all'>('all')

const ordersShown = computed(() => {
  if (ordersRange.value !== '30d') return ordersRows.value
  const cutoff = Math.floor(Date.now() / 1000) - 30 * 86400
  return ordersRows.value.filter((o) => o.orderedAt >= cutoff)
})

async function loadOrders(page: number) {
  ordersLoading.value = true
  try {
    const res = await getUserOrders(String(route.params.id), { page, pageSize: 10 })
    ordersRows.value = res.list
    ordersTotal.value = res.total
    ordersPage.value = page
    ordersLoaded.value = true
  } finally {
    ordersLoading.value = false
  }
}

/** ---- 投诉售后 tab ---- */
const complaintRows = ref<ComplaintRow[]>([])
const complaintTotal = ref(0)
const complaintPage = ref(1)
const complaintLoading = ref(false)
const complaintLoaded = ref(false)

async function loadComplaints(page: number) {
  complaintLoading.value = true
  try {
    const res = await getUserComplaints(String(route.params.id), { page, pageSize: 10 })
    complaintRows.value = res.list
    complaintTotal.value = res.total
    complaintPage.value = page
    complaintLoaded.value = true
  } finally {
    complaintLoading.value = false
  }
}

/* 左侧菜单 */
const tabs: { key: DetailTab; label: string; icon: string; badge?: () => number; badgeRed?: boolean }[] = [
  { key: 'basic', label: '基本信息', icon: 'order' },
  { key: 'orders', label: '订单记录', icon: 'order', badge: () => detail.value?.orderRecordCount ?? 0 },
  {
    key: 'complaints',
    label: '投诉售后',
    icon: 'complaint',
    badge: () => detail.value?.complaintCount ?? 0,
    badgeRed: true,
  },
  { key: 'account', label: '账号状态', icon: 'shield' },
]

watch(activeTab, (t) => {
  if (t === 'orders' && !ordersLoaded.value) loadOrders(1)
  if (t === 'complaints' && !complaintLoaded.value) loadComplaints(1)
})

/* 状态语义色 */
function orderTone(s: string): 'green' | 'blue' | 'orange' | 'gray' | 'red' {
  if (s === '已完成') return 'green'
  if (s === '待接单') return 'orange'
  if (s === '配送中' || s === '已取件待送达') return 'blue'
  if (s === '已取消') return 'gray'
  return 'red'
}
function complaintTone(s: string): 'blue' | 'orange' | 'green' {
  if (s === '已完结') return 'green'
  if (s === '处理中') return 'orange'
  return 'blue'
}

onMounted(async () => {
  try {
    detail.value = await getUserDetail(String(route.params.id))
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-if="loading" class="text-ink-3">加载中…</div>

  <div v-else-if="detail" class="flex flex-col gap-2">
    <!-- 用户头部卡 -->
    <div class="flex shrink-0 items-center gap-3 rounded-card bg-white px-5 py-3">
      <div
        class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full text-[22px] font-bold"
        :style="{ background: detail.avatarBg, color: detail.avatarColor }"
      >
        {{ detail.name.charAt(0) }}
      </div>
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <span class="text-[18px] font-bold text-ink-1">{{ detail.name }}</span>
          <StatusTag :tone="detail.status === '已封禁' ? 'red' : 'green'">{{ detail.status }}</StatusTag>
          <span class="inline-block rounded-tag bg-primary-bg px-2.5 py-0.5 text-xs font-semibold text-primary">
            {{ detail.level }}
          </span>
          <span class="text-[12px] text-ink-4">用户编号 {{ detail.userNo }}</span>
        </div>
        <div class="mt-2 flex flex-wrap items-center gap-6 text-[13px] text-ink-3">
          <span>手机 <b class="font-medium text-ink-1">{{ detail.phone }}</b></span>
          <span>注册于 <b class="font-medium text-ink-1">{{ formatDate(detail.registeredAt) }}</b></span>
          <span>累计订单 <b class="font-medium text-ink-1">{{ detail.totalOrders }}</b> 单</span>
          <span>累计消费 <b class="font-medium text-ink-1">{{ formatMoney(detail.totalSpent) }}</b></span>
          <span>本月下单 <b class="font-medium text-ink-1">{{ detail.monthOrders }}</b> 单</span>
        </div>
      </div>
      <div class="flex shrink-0 gap-2">
        <AppButton
          variant="secondary"
          class="!h-[30px] !px-[18px]"
          @click="toast.info('发送站内信')"
        >
发送站内信
</AppButton>
        <AppButton
          v-permission="'user:ban'"
          variant="secondary"
          class="!h-[30px] !px-[18px]"
          :class="detail.status === '已封禁' ? '' : '!border-[#FFD9D9] !text-[#FF4D4F]'"
          @click="confirmOpen = true"
        >
{{ detail.status === '已封禁' ? '解封账号' : '封禁账号' }}
</AppButton>
      </div>
    </div>

    <!-- 主体区 -->
    <div class="flex min-h-0 flex-1 gap-3">
      <!-- 左侧 tab -->
      <div class="flex w-[168px] shrink-0 flex-col rounded-card bg-white py-2">
        <div
          v-for="t in tabs"
          :key="t.key"
          class="mx-2 mb-0.5 flex cursor-pointer items-center gap-2 rounded-lg px-4 py-2.5 text-[14px]"
          :class="
            activeTab === t.key
              ? 'mt-0.5 font-semibold'
              : 'text-ink-2 hover:bg-head hover:text-ink-1'
          "
          :style="activeTab === t.key ? 'background:#FFF3E8;color:#FF6B2C;' : ''"
          @click="activeTab = t.key"
        >
          <SvgIcon :name="t.icon" :size="17" />
          {{ t.label }}
          <span
            v-if="t.badge"
            class="ml-auto flex h-[18px] min-w-[18px] items-center justify-center rounded-[9px] px-1 text-[11px] font-semibold"
            :class="t.badgeRed ? 'bg-sem-red text-white' : 'bg-head text-ink-3'"
          >{{ t.badge() }}</span>
        </div>
        <div class="flex-1" />
        <div class="border-t border-line px-4 py-2.5 text-[12px] text-ink-4">
          最近活跃：{{ formatTime(detail.recentActiveAt) }} 下单
        </div>
      </div>

      <!-- 右侧内容 -->
      <div class="flex min-w-0 flex-1 flex-col gap-3">
        <!-- 基本信息 tab -->
        <template v-if="activeTab === 'basic'">
          <div class="shrink-0 rounded-card bg-white p-3">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">基本信息</div>
              <div class="ml-auto text-[12px] text-ink-3">
                实名认证：<b class="font-semibold" style="color:#00B578;">{{ detail.verified }}</b> · 手机号已绑定
              </div>
            </div>
            <div class="mt-2.5 grid grid-cols-2 gap-2 gap-x-3.5">
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">用户昵称</div>
                <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                  {{ detail.nickname }}
                </div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">真实姓名</div>
                <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                  {{ detail.realName }}
                </div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">联系电话</div>
                <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                  {{ detail.phone }}
                </div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">所在区域</div>
                <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                  {{ detail.region }}
                </div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">常用地址</div>
                <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                  {{ detail.address }}
                </div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">注册渠道</div>
                <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                  {{ detail.registerChannel }}
                </div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">常用业务</div>
                <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                  {{ detail.bizTypes }}
                </div>
              </div>
              <div>
                <div class="mb-1.5 text-[12px] text-ink-3">账户余额</div>
                <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                  {{ formatMoney(detail.balance) }}（红包余额 {{ formatMoney(detail.redPacketBalance) }}）
                </div>
              </div>
            </div>
          </div>

          <!-- 近期订单卡 -->
          <div class="shrink-0 rounded-card bg-white p-3">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">近期订单</div>
              <div class="ml-auto flex gap-2">
                <span class="cursor-pointer rounded-md bg-primary-bg px-3 py-1.5 text-[12px] font-semibold text-primary">近 30 天</span>
                <span class="cursor-pointer rounded-md bg-head px-3 py-1.5 text-[12px] text-ink-2">全部</span>
              </div>
            </div>

            <div class="mt-2 overflow-hidden rounded-[10px] border border-line">
              <table class="w-full border-collapse text-[13px]">
                <thead>
                  <tr class="bg-head text-[12px] text-ink-3">
                    <th class="px-3.5 py-2.5 text-left font-normal">订单号</th>
                    <th class="px-3.5 py-2.5 text-left font-normal">业务类型</th>
                    <th class="px-3.5 py-2.5 text-left font-normal">商家 / 内容</th>
                    <th class="px-3.5 py-2.5 text-left font-normal">金额</th>
                    <th class="px-3.5 py-2.5 text-left font-normal">配送骑手</th>
                    <th class="px-3.5 py-2.5 text-left font-normal">状态</th>
                    <th class="px-3.5 py-2.5 text-left font-normal">下单时间</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="o in detail.recentOrders" :key="o.id" class="border-t border-line">
                    <td class="px-3.5 py-2 font-semibold" style="color:#2B7FFF;">#{{ o.id }}</td>
                    <td class="px-3.5 py-2">
                      <span
                        class="inline-block rounded-tag px-2 py-0.5 text-[12px] font-semibold"
                        :style="{ background: o.bizBg, color: o.bizColor }"
                      >{{ o.bizType }}</span>
                    </td>
                    <td class="px-3.5 py-2 text-ink-2">{{ o.merchant }}</td>
                    <td class="px-3.5 py-2 font-semibold text-ink-1">{{ formatMoney(o.amount) }}</td>
                    <td class="px-3.5 py-2 text-ink-2">{{ o.rider }}</td>
                    <td class="px-3.5 py-2"><StatusTag :tone="orderTone(o.status)">{{ o.status }}</StatusTag></td>
                    <td class="px-3.5 py-2 text-ink-3">{{ shortTime(o.orderedAt) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="mt-2 flex items-center justify-end gap-3 text-[12px] text-ink-3">
              <span>共 {{ detail.orderRecordCount }} 条记录</span>
              <span class="flex items-center gap-1.5">
                <span class="flex h-[26px] w-[26px] cursor-not-allowed items-center justify-center rounded-md border border-line text-ink-4">&lt;</span>
                <span class="flex h-[26px] w-[26px] items-center justify-center rounded-md text-[12px] font-semibold text-white" style="background:#FF6B2C;">1</span>
                <span class="flex h-[26px] w-[26px] cursor-pointer items-center justify-center rounded-md border border-line text-ink-2">2</span>
                <span class="flex h-[26px] w-[26px] cursor-pointer items-center justify-center rounded-md border border-line text-ink-2">&gt;</span>
              </span>
            </div>
          </div>
        </template>

        <!-- 订单记录 tab -->
        <div v-if="activeTab === 'orders'" class="shrink-0 rounded-card bg-white p-3">
          <div class="flex items-center">
            <div class="text-[15px] font-semibold text-ink-1">订单记录</div>
            <div class="ml-auto flex gap-2">
              <span
                class="cursor-pointer rounded-md px-3 py-1.5 text-[12px] font-semibold"
                :class="ordersRange === '30d' ? 'bg-primary-bg text-primary' : 'bg-head text-ink-2'"
                @click="ordersRange = '30d'"
              >近 30 天</span>
              <span
                class="cursor-pointer rounded-md px-3 py-1.5 text-[12px]"
                :class="ordersRange === 'all' ? 'bg-primary-bg font-semibold text-primary' : 'bg-head text-ink-2'"
                @click="ordersRange = 'all'"
              >全部</span>
            </div>
          </div>

          <div class="mt-2 overflow-hidden rounded-[10px] border border-line">
            <table class="w-full border-collapse text-[13px]">
              <thead>
                <tr class="bg-head text-[12px] text-ink-3">
                  <th class="px-3.5 py-2.5 text-left font-normal">订单号</th>
                  <th class="px-3.5 py-2.5 text-left font-normal">业务类型</th>
                  <th class="px-3.5 py-2.5 text-left font-normal">商家 / 内容</th>
                  <th class="px-3.5 py-2.5 text-left font-normal">金额</th>
                  <th class="px-3.5 py-2.5 text-left font-normal">配送骑手</th>
                  <th class="px-3.5 py-2.5 text-left font-normal">状态</th>
                  <th class="px-3.5 py-2.5 text-left font-normal">下单时间</th>
                </tr>
              </thead>
              <tbody v-if="ordersShown.length">
                <tr v-for="o in ordersShown" :key="o.id" class="border-t border-line">
                  <td class="px-3.5 py-2 font-semibold" style="color:#2B7FFF;">#{{ o.id }}</td>
                  <td class="px-3.5 py-2">
                    <span
                      class="inline-block rounded-tag px-2 py-0.5 text-[12px] font-semibold"
                      :style="{ background: o.bizBg, color: o.bizColor }"
                    >{{ o.bizType }}</span>
                  </td>
                  <td class="px-3.5 py-2 text-ink-2">{{ o.merchant }}</td>
                  <td class="px-3.5 py-2 font-semibold text-ink-1">{{ formatMoney(o.amount) }}</td>
                  <td class="px-3.5 py-2 text-ink-2">{{ o.rider }}</td>
                  <td class="px-3.5 py-2"><StatusTag :tone="orderTone(o.status)">{{ o.status }}</StatusTag></td>
                  <td class="px-3.5 py-2 text-ink-3">{{ shortTime(o.orderedAt) }}</td>
                </tr>
              </tbody>
              <tbody v-else>
                <tr>
                  <td colspan="7" class="px-3.5 py-8 text-center text-[13px] text-ink-4">
                    {{ ordersLoading ? '加载中…' : '暂无订单记录' }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="mt-2.5 flex items-center justify-end gap-3 text-[12px] text-ink-3">
            <span>共 {{ ordersTotal }} 条记录</span>
            <Pagination
              :page="ordersPage"
              :page-size="10"
              :total="ordersTotal"
              @change="loadOrders"
            />
          </div>
        </div>

        <!-- 投诉售后 tab -->
        <div v-if="activeTab === 'complaints'" class="shrink-0 rounded-card bg-white p-3">
          <div class="flex items-center">
            <div class="text-[15px] font-semibold text-ink-1">投诉售后</div>
            <div class="ml-auto text-[12px] text-ink-3">共 {{ complaintTotal }} 条记录</div>
          </div>

          <div class="mt-2 overflow-hidden rounded-[10px] border border-line">
            <table class="w-full border-collapse text-[13px]">
              <thead>
                <tr class="bg-head text-[12px] text-ink-3">
                  <th class="px-3.5 py-2.5 text-left font-normal">工单号</th>
                  <th class="px-3.5 py-2.5 text-left font-normal">分类</th>
                  <th class="px-3.5 py-2.5 text-left font-normal">工单标题</th>
                  <th class="px-3.5 py-2.5 text-left font-normal">关联订单</th>
                  <th class="px-3.5 py-2.5 text-left font-normal">责任方</th>
                  <th class="px-3.5 py-2.5 text-left font-normal">状态</th>
                  <th class="px-3.5 py-2.5 text-left font-normal">提交时间</th>
                </tr>
              </thead>
              <tbody v-if="complaintRows.length">
                <tr v-for="c in complaintRows" :key="c.id" class="border-t border-line">
                  <td class="px-3.5 py-2 font-semibold" style="color:#2B7FFF;">{{ c.id }}</td>
                  <td class="px-3.5 py-2">
                    <span
                      class="inline-block rounded-tag px-2 py-0.5 text-[12px] font-semibold"
                      :style="c.category.includes('配送') ? 'background:#FFF3E8;color:#FF6B2C;' : c.category.includes('商家') ? 'background:#FFF8E0;color:#D48806;' : 'background:#E8F0FF;color:#2B7FFF;'"
                    >{{ c.category }}</span>
                  </td>
                  <td class="max-w-[220px] truncate px-3.5 py-2 text-ink-2">{{ c.title }}</td>
                  <td class="px-3.5 py-2 text-ink-3">{{ c.order_id }}</td>
                  <td class="px-3.5 py-2 text-ink-2">{{ c.target }}</td>
                  <td class="px-3.5 py-2"><StatusTag :tone="complaintTone(c.status)">{{ c.status }}</StatusTag></td>
                  <td class="px-3.5 py-2 text-ink-3">{{ shortTime(c.created_at) }}</td>
                </tr>
              </tbody>
              <tbody v-else>
                <tr>
                  <td colspan="7" class="px-3.5 py-8 text-center text-[13px] text-ink-4">
                    {{ complaintLoading ? '加载中…' : '暂无投诉售后记录' }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="complaintTotal > 10" class="mt-2.5 flex items-center justify-end gap-3 text-[12px] text-ink-3">
            <Pagination
              :page="complaintPage"
              :page-size="10"
              :total="complaintTotal"
              @change="loadComplaints"
            />
          </div>
        </div>

        <!-- 账号状态 tab -->
        <div v-if="activeTab === 'account'" class="shrink-0 rounded-card bg-white p-3">
          <div class="flex items-center">
            <div class="text-[15px] font-semibold text-ink-1">账号状态</div>
            <div class="ml-auto flex items-center gap-2">
              <StatusTag :tone="detail.status === '已封禁' ? 'red' : 'green'">{{ detail.status }}</StatusTag>
              <AppButton
                v-permission="'user:ban'"
                variant="secondary"
                class="!h-[30px] !px-[16px]"
                :class="detail.status === '已封禁' ? '' : '!border-[#FFD9D9] !text-[#FF4D4F]'"
                @click="confirmOpen = true"
              >
{{ detail.status === '已封禁' ? '解封账号' : '封禁账号' }}
</AppButton>
            </div>
          </div>

          <div class="mt-2.5 grid grid-cols-2 gap-2 gap-x-3.5">
            <div>
              <div class="mb-1.5 text-[12px] text-ink-3">用户编号</div>
              <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                {{ detail.userNo }}
              </div>
            </div>
            <div>
              <div class="mb-1.5 text-[12px] text-ink-3">手机号</div>
              <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                {{ detail.phone }}
              </div>
            </div>
            <div>
              <div class="mb-1.5 text-[12px] text-ink-3">注册时间</div>
              <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                {{ formatDate(detail.registeredAt) }}
              </div>
            </div>
            <div>
              <div class="mb-1.5 text-[12px] text-ink-3">当前等级</div>
              <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                {{ detail.level }}
              </div>
            </div>
            <div>
              <div class="mb-1.5 text-[12px] text-ink-3">后台角色</div>
              <div class="flex h-8 items-center gap-2 rounded-lg border border-line bg-head px-3">
                <StatusTag :tone="roleTone(detail.role)">{{ detail.roleName || '普通用户' }}</StatusTag>
                <span
                  v-permission="'user:role'"
                  class="cursor-pointer text-[12px] font-semibold text-primary"
                  @click="openRoleModal"
                >
                  修改
                </span>
              </div>
            </div>
            <div>
              <div class="mb-1.5 text-[12px] text-ink-3">实名认证</div>
              <div class="flex h-8 items-center gap-2 rounded-lg border border-line bg-head px-3">
                <span class="text-[13px] text-ink-1">{{ detail.verified }}</span>
                <span
                  v-if="detail.realNameVerify?.status === '待审核' && detail.verified === '未实名'"
                  v-permission="'user:audit'"
                  class="cursor-pointer text-[12px] font-semibold text-primary"
                  @click="openRealNameModal"
                >
                  审核
                </span>
              </div>
              <div
                v-if="detail.realNameVerify?.status === '已驳回' && detail.verified === '未实名'"
                class="mt-1.5 text-[12px]"
                style="color:#FF4D4F;"
              >
                已驳回：{{ detail.realNameVerify.rejectReason || '材料不通过，请重新提交' }}
              </div>
            </div>
            <div>
              <div class="mb-1.5 text-[12px] text-ink-3">账户余额</div>
              <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                {{ formatMoney(detail.balance) }}（红包 {{ formatMoney(detail.redPacketBalance) }}）
              </div>
            </div>
            <div>
              <div class="mb-1.5 text-[12px] text-ink-3">最近活跃</div>
              <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                {{ formatTime(detail.recentActiveAt) }}
              </div>
            </div>
            <div>
              <div class="mb-1.5 text-[12px] text-ink-3">累计订单 / 消费</div>
              <div class="flex h-8 items-center rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
                {{ detail.totalOrders }} 单 / {{ formatMoney(detail.totalSpent) }}
              </div>
            </div>
          </div>

          <div
            class="mt-3 rounded-lg px-3 py-2.5 text-[12px]"
            :class="detail.status === '已封禁' ? 'bg-[#FFE8E8] text-[#FF4D4F]' : 'bg-head text-ink-3'"
          >
            {{ detail.status === '已封禁'
              ? '该账号已被封禁，无法登录与下单。如需恢复，请点击「解封账号」。'
              : '该账号状态正常，可正常登录下单。如需限制，请点击「封禁账号」，封禁后账号将无法登录。' }}
          </div>
        </div>
      </div>
    </div>

    <!-- 封禁 / 解封确认 -->
    <Modal
      v-model:open="confirmOpen"
      :title="confirmTitle()"
      :confirm-text="detail?.status === '已封禁' ? '确认解封' : '确认封禁'"
      :loading="submitting"
      @confirm="onConfirm"
    >
      <p class="text-sm text-ink-2">{{ confirmMessage() }}</p>
    </Modal>

    <!-- 实名认证审核 -->
    <Modal
      v-model:open="realNameOpen"
      title="实名认证审核"
      confirm-text="确认通过"
      :loading="realNameSubmitting"
      :width="460"
      @confirm="onRealNameApprove"
    >
      <div v-if="detail?.realNameVerify" class="flex flex-col gap-3.5 py-1">
        <div class="grid grid-cols-2 gap-x-4 gap-y-3 text-[13px]">
          <div class="flex justify-between text-ink-3"><span>申请人</span><span class="font-semibold text-ink-1">{{ detail.realNameVerify.realName }}</span></div>
          <div class="flex justify-between text-ink-3"><span>身份证号</span><span class="text-ink-1">{{ detail.realNameVerify.idCardMasked }}</span></div>
          <div class="flex justify-between text-ink-3"><span>提交时间</span><span class="text-ink-1">{{ formatTime(detail.realNameVerify.appliedAt) }}</span></div>
          <div class="flex justify-between text-ink-3"><span>当前状态</span><span class="text-[#D48806]">{{ detail.realNameVerify.status }}</span></div>
        </div>
        <p class="text-[12px] leading-[1.8] text-ink-4">
          确认该用户身份信息真实有效？通过后该账号将标记为「已实名」，实名信息同步到用户资料；驳回后用户可修改后重新提交。
        </p>
        <div class="flex items-center gap-2 rounded-lg bg-head px-3 py-2.5">
          <span class="text-[12px] text-ink-3">或选择驳回</span>
          <input
            v-model="realNameRejectReason"
            placeholder="填写驳回原因（选填）"
            class="flex-1 rounded-lg border border-line bg-white px-2.5 py-1.5 text-[12px] outline-none placeholder:text-ink-4"
          />
          <button
            type="button"
            class="rounded-md px-3 py-1.5 text-[12px] font-semibold"
            style="background:#FFF3E8;color:#FF6B2C;"
            @click="onRealNameReject"
          >驳回</button>
        </div>
      </div>
    </Modal>

    <!-- 修改后台角色 -->
    <Modal
      v-model:open="roleOpen"
      title="修改后台角色"
      confirm-text="确认修改"
      :loading="roleSubmitting"
      :width="440"
      @confirm="onRoleSubmit"
    >
      <div class="flex flex-col gap-3.5 py-1">
        <div class="flex flex-wrap gap-2">
          <button
            v-for="r in BACKEND_ROLES.filter((x) => x.value !== 'system')"
            :key="r.value"
            type="button"
            class="rounded-lg border px-3.5 py-2 text-[13px] transition-colors"
            :class="
              roleForm === r.value
                ? 'border-primary bg-primary-bg font-semibold text-primary'
                : 'border-line bg-white text-ink-2 hover:text-ink-1'
            "
            @click="roleForm = r.value"
          >
            {{ r.label }}
          </button>
        </div>
        <p class="text-[12px] text-ink-4">
          角色决定该账号可访问的后台菜单与操作权限；系统级服务身份不可修改。角色变更后，该账号的下一次请求即按新权限生效。
        </p>
      </div>
    </Modal>
  </div>
</template>
