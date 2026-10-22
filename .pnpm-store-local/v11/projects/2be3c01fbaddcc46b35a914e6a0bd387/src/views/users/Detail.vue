<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import StatusTag from '@/components/StatusTag.vue'
import Modal from '@/components/Modal.vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { useToast } from '@/composables/useToast'
import { formatDate, formatMoney, formatTime } from '@/utils/format'
import { getUserDetail, banUser, unbanUser, type UserDetail } from '@/api/users'

const route = useRoute()
const toast = useToast()

const detail = ref<UserDetail | null>(null)
const loading = ref(true)

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

/** 近期订单时间列：'MM-DD HH:mm'（formatTime 去掉年份前缀） */
function shortTime(sec: number) {
  return formatTime(sec).slice(5)
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
  <div v-if="loading" class="p-6 text-ink-3">加载中…</div>

  <div v-else-if="detail" class="flex flex-col gap-2 px-6 py-3">
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
        <button
          class="h-[30px] rounded-lg border border-line bg-white px-[18px] text-[13px] text-ink-2"
          @click="toast.info('发送站内信')"
        >
          发送站内信
        </button>
        <button
          v-permission="'user:ban'"
          class="h-[30px] rounded-lg border bg-white px-[18px] text-[13px]"
          :class="detail.status === '已封禁' ? 'border-line text-ink-2' : ''"
          :style="detail.status === '已封禁' ? {} : { borderColor: '#FFD9D9', color: '#FF4D4F' }"
          @click="confirmOpen = true"
        >
          {{ detail.status === '已封禁' ? '解封账号' : '封禁账号' }}
        </button>
      </div>
    </div>

    <!-- 主体区 -->
    <div class="flex min-h-0 flex-1 gap-3">
      <!-- 左侧 tab -->
      <div class="flex w-[168px] shrink-0 flex-col rounded-card bg-white py-2">
        <div
          class="mx-2 mb-0.5 mt-0.5 flex items-center gap-2 rounded-lg px-4 py-2.5 text-[14px] font-semibold"
          style="background:#FFF3E8;color:#FF6B2C;"
        >
          <SvgIcon name="order" :size="17" />
          基本信息
        </div>
        <div class="mx-2 mb-0.5 flex items-center gap-2 rounded-lg px-4 py-2.5 text-[14px] text-ink-2">
          <SvgIcon name="order" :size="17" />
          订单记录
          <span class="ml-auto flex h-[18px] min-w-[18px] items-center justify-center rounded-[9px] bg-head px-1 text-[11px] font-semibold text-ink-3">
            {{ detail.orderRecordCount }}
          </span>
        </div>
        <div class="mx-2 mb-0.5 flex items-center gap-2 rounded-lg px-4 py-2.5 text-[14px] text-ink-2">
          <SvgIcon name="complaint" :size="17" />
          投诉售后
          <span class="ml-auto flex h-[18px] min-w-[18px] items-center justify-center rounded-[9px] bg-sem-red px-1 text-[11px] font-semibold text-white">
            {{ detail.complaintCount }}
          </span>
        </div>
        <div class="mx-2 flex items-center gap-2 rounded-lg px-4 py-2.5 text-[14px] text-ink-2">
          <SvgIcon name="shield" :size="17" />
          账号状态
        </div>
        <div class="flex-1" />
        <div class="border-t border-line px-4 py-2.5 text-[12px] text-ink-4">
          最近活跃：{{ formatTime(detail.recentActiveAt) }} 下单
        </div>
      </div>

      <!-- 右侧内容 -->
      <div class="flex min-w-0 flex-1 flex-col gap-3">
        <!-- 基本信息卡 -->
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
                  <td class="px-3.5 py-2 font-semibold" style="color:#00B578;">{{ o.status }}</td>
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
  </div>
</template>
