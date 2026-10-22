<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getFinance, auditWithdraw, type FinanceDto, type WithdrawRow } from '@/api/finance'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import StatusTag from '@/components/StatusTag.vue'
import Confirm from '@/components/Confirm.vue'
import Empty from '@/components/Empty.vue'
import { formatMoney, formatTime } from '@/utils/format'
import { useToast } from '@/composables/useToast'

const toast = useToast()

const data = ref<FinanceDto | null>(null)
const loading = ref(true)

const TABS = ['全部', '待审核', '已通过', '已驳回'] as const
type Tab = (typeof TABS)[number]
const activeTab = ref<Tab>('全部')

const filtered = computed<WithdrawRow[]>(() => {
  const list = data.value?.withdraws ?? []
  if (activeTab.value === '全部') return list
  return list.filter((w) => w.status === activeTab.value)
})

const countOf = (status: string) => (data.value?.withdraws ?? []).filter((w) => w.status === status).length

onMounted(async () => {
  try {
    data.value = await getFinance()
  } finally {
    loading.value = false
  }
})

/** 审核确认（打款 / 驳回） */
const confirmOpen = ref(false)
const submitting = ref(false)
const auditTarget = ref<WithdrawRow | null>(null)
const auditAction = ref<'pass' | 'reject'>('pass')

const confirmMessage = computed(() => {
  if (!auditTarget.value) return ''
  return auditAction.value === 'pass'
    ? `确认打款 ${formatMoney(auditTarget.value.amount)} 给 ${auditTarget.value.name}？打款后操作不可撤回。`
    : `确认驳回「${auditTarget.value.name}」的提现申请？驳回后金额将退回账户余额。`
})

function askAudit(row: WithdrawRow, action: 'pass' | 'reject') {
  auditTarget.value = row
  auditAction.value = action
  confirmOpen.value = true
}

async function onConfirm() {
  if (!auditTarget.value || submitting.value) return
  submitting.value = true
  try {
    await auditWithdraw(auditTarget.value.id, { action: auditAction.value })
    auditTarget.value.status = auditAction.value === 'pass' ? '已通过' : '已驳回'
    toast.success(auditAction.value === 'pass' ? '打款成功，凭证已生成' : '已驳回该提现申请')
    confirmOpen.value = false
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="flex flex-col gap-4 p-5">
    <div v-if="loading" class="py-16 text-center text-ink-3">加载中…</div>
    <template v-else-if="data">
      <!-- 汇总指标卡 -->
      <div class="grid shrink-0 grid-cols-4 gap-4">
        <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
          <div class="flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-xl" style="background:#FFF3E8;">
            <SvgIcon name="money" :size="20" :stroke-width="1.9" style="color:#FF6B2C;" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="text-[12px] text-ink-3">今日平台收入</div>
            <div class="mt-1 text-[22px] font-bold leading-7 text-ink-1">{{ formatMoney(data.stats.todayIncome) }}</div>
          </div>
          <div class="text-[11px] font-semibold" style="color:#00B578;">{{ data.stats.todayIncomeDelta }}</div>
        </div>

        <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
          <div class="flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-xl" style="background:#E8F0FF;">
            <SvgIcon name="user" :size="20" :stroke-width="1.9" style="color:#2B7FFF;" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="text-[12px] text-ink-3">待结算余额</div>
            <div class="mt-1 text-[22px] font-bold leading-7 text-ink-1">{{ formatMoney(data.stats.pendingBalance) }}</div>
          </div>
          <div class="text-[11px] text-ink-3">{{ data.stats.pendingBalanceBreakdown }}</div>
        </div>

        <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
          <div class="flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-xl" style="background:#FFF8E0;">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#D48806" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="12" />
              <line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
          </div>
          <div class="min-w-0 flex-1">
            <div class="text-[12px] text-ink-3">提现申请待处理</div>
            <div class="mt-1 text-[22px] font-bold leading-7" style="color:#D48806;">{{ data.stats.withdrawPendingCount }}</div>
          </div>
          <button class="cursor-pointer text-[11px] font-semibold" style="color:#FF6B2C;" @click="activeTab = '待审核'">去审核</button>
        </div>

        <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
          <div class="flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-xl" style="background:#FFE8E8;">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#FF4D4F" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <path d="M12 8v4M12 16h.01" />
            </svg>
          </div>
          <div class="min-w-0 flex-1">
            <div class="text-[12px] text-ink-3">异常账单</div>
            <div class="mt-1 text-[22px] font-bold leading-7" style="color:#FF4D4F;">{{ data.stats.abnormalBills }}</div>
          </div>
          <div class="text-[11px] text-ink-4">金额 {{ formatMoney(data.stats.abnormalBillsAmount) }}</div>
        </div>
      </div>

      <!-- 标题行：页面 tabs + 导出 -->
      <div class="flex shrink-0 items-center">
        <div class="flex items-center gap-5">
          <span class="cursor-pointer border-b-2 pb-[6px] text-[16px] font-bold" style="border-color:#FF6B2C;color:#1D2129;">提现审核</span>
          <router-link to="/finance/settlement" class="cursor-pointer text-[14px] text-ink-3 hover:text-ink-1">结算流水</router-link>
          <span class="cursor-pointer text-[14px] text-ink-3 hover:text-ink-1">对账管理</span>
        </div>
        <div class="ml-auto">
          <button
            class="flex h-9 items-center gap-2 rounded-lg border border-line bg-white px-4 text-[13px] text-ink-2"
            @click="toast.info('导出任务已提交，完成后可在下载中心查看')"
          >
            <SvgIcon name="download" :size="15" />
            导出流水
          </button>
        </div>
      </div>

      <!-- 状态筛选 tabs -->
      <div class="flex shrink-0 items-center gap-2">
        <button
          v-for="t in TABS"
          :key="t"
          class="rounded-md px-3 py-[5px] text-[12px]"
          :class="activeTab === t ? 'font-semibold' : 'text-ink-2'"
          :style="activeTab === t ? 'background:#FFF3E8;color:#FF6B2C;' : 'background:#F7F8FA;'"
          @click="activeTab = t"
        >
          {{ t }}<span class="ml-1 text-[11px]" :style="activeTab === t ? 'color:#FF6B2C;' : 'color:#C9CDD4;'">
            {{ t === '全部' ? (data.withdraws?.length ?? 0) : countOf(t) }}
          </span>
        </button>
      </div>

      <!-- 提现审核列表 -->
      <div class="flex flex-col gap-3">
        <Empty v-if="filtered.length === 0" text="暂无相关提现申请" />
        <div
          v-for="row in filtered"
          :key="row.id"
          class="flex items-center gap-4 rounded-card bg-white px-5 py-4"
          :style="row.kind === '异常' ? 'border:1.5px solid #FFE0E0;' : ''"
        >
          <div
            class="flex h-[42px] w-[42px] shrink-0 items-center justify-center rounded-full text-[15px] font-bold"
            :style="{ background: row.avatarBg, color: row.avatarColor }"
          >
            {{ row.name.slice(0, 1) }}
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2.5">
              <span class="text-[14px] font-semibold text-ink-1">{{ row.name }}</span>
              <span class="rounded-tag px-2 py-0.5 text-[11px] font-semibold" :style="{ background: row.kindBg, color: row.kindColor }">{{ row.kind }}</span>
              <router-link class="cursor-pointer text-[12px] text-ink-4 hover:text-primary" :to="`/finance/withdraw/${row.id}`">
                提现单号 {{ row.id }}
              </router-link>
            </div>
            <div class="mt-1 text-[12px]" :style="row.riskTip ? 'color:#FF4D4F;' : 'color:#86909C;'">
              <span v-if="!row.riskTip">提现至：{{ row.bankInfo }}</span>
              <span v-else>{{ row.riskTip }}</span>
            </div>
            <div v-if="row.rejectReason" class="mt-1 text-[12px]" style="color:#86909C;">{{ row.rejectReason }}</div>
          </div>
          <div class="shrink-0 text-right">
            <div class="text-[18px] font-bold" :style="{ color: row.amountColor ?? '#1D2129' }">{{ formatMoney(row.amount) }}</div>
            <div class="mt-0.5 text-[11px] text-ink-3">申请于 {{ formatTime(row.created_at) }}</div>
          </div>
          <div class="flex shrink-0 gap-2.5">
            <template v-if="row.status === '待审核'">
              <button
                v-permission="'withdraw:audit'"
                class="h-[34px] rounded-lg border-none px-5 text-[13px] font-semibold text-white"
                style="background:linear-gradient(135deg,#FF8A3D,#FF6B2C);"
                @click="askAudit(row, 'pass')"
              >
                确认打款
              </button>
              <button
                v-permission="'withdraw:audit'"
                class="h-[34px] rounded-lg border border-line bg-white px-5 text-[13px]"
                style="color:#FF4D4F;"
                @click="askAudit(row, 'reject')"
              >
                驳回
              </button>
            </template>
            <StatusTag v-else :tone="row.status === '已通过' ? 'green' : 'red'">{{ row.status }}</StatusTag>
          </div>
        </div>
      </div>
    </template>

    <Confirm
      :open="confirmOpen"
      :title="auditAction === 'pass' ? '确认打款' : '驳回申请'"
      :message="confirmMessage"
      :confirm-text="auditAction === 'pass' ? '确认打款' : '确认驳回'"
      :loading="submitting"
      @update:open="confirmOpen = $event"
      @confirm="onConfirm"
    />
  </div>
</template>
