<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { getWithdrawDetail, auditWithdraw, type WithdrawDetailDto } from '@/api/finance'
import AppButton from '@/components/AppButton.vue'
import Confirm from '@/components/Confirm.vue'
import Modal from '@/components/Modal.vue'
import { formatMoney, formatTime } from '@/utils/format'
import { useToast } from '@/composables/useToast'

const route = useRoute()
const toast = useToast()

const detail = ref<WithdrawDetailDto | null>(null)
const loading = ref(true)
const remark = ref('')
const submitting = ref(false)

const confirmOpen = ref(false)
const rejectOpen = ref(false)

const auditId = computed(() => String(route.params.id ?? ''))

onMounted(async () => {
  try {
    detail.value = await getWithdrawDetail(auditId.value)
  } finally {
    loading.value = false
  }
})

const confirmMessage = computed(() => {
  if (!detail.value) return ''
  return `确认打款 ${formatMoney(detail.value.amount)}？打款后系统将推送通知并生成提现凭证，操作不可撤回。`
})

async function onPass() {
  if (!detail.value || submitting.value) return
  submitting.value = true
  try {
    await auditWithdraw(detail.value.id, { action: 'pass', remark: remark.value })
    detail.value.status = '已通过'
    toast.success('打款成功，提现凭证已生成')
    confirmOpen.value = false
  } finally {
    submitting.value = false
  }
}

async function onReject() {
  if (!detail.value || submitting.value) return
  submitting.value = true
  try {
    await auditWithdraw(detail.value.id, { action: 'reject', remark: remark.value || '审核驳回' })
    detail.value.status = '已驳回'
    toast.success('已驳回该提现申请')
    rejectOpen.value = false
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <div v-if="loading" class="py-16 text-center text-ink-3">加载中…</div>
    <template v-else-if="detail">
      <!-- 提现头部卡 -->
      <div class="flex shrink-0 items-center gap-4 rounded-card bg-white px-5 py-[18px]">
        <div class="flex h-14 w-14 shrink-0 items-center justify-center rounded-[14px]" style="background:#FFF7E6;">
          <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="#D48806" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="1" x2="12" y2="23" />
            <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
          </svg>
        </div>
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2.5">
            <span class="text-[18px] font-bold text-ink-1">
              提现申请 · <span style="color:#FF6B2C;">{{ formatMoney(detail.amount) }}</span>
            </span>
            <span class="rounded-tag px-2.5 py-[3px] text-[12px] font-semibold" style="background:#FFF7E6;color:#D48806;">{{ detail.status }}</span>
            <span class="rounded-tag px-2.5 py-[3px] text-[12px] font-semibold" style="background:#FFF0E8;color:#FF6B2C;">{{ detail.kind }}</span>
          </div>
          <div class="mt-2 flex flex-wrap items-center gap-6 text-[13px] text-ink-3">
            <span>提现单号 <b class="font-semibold text-ink-1">{{ detail.id }}</b></span>
            <span>申请人 <b class="font-semibold text-ink-1">{{ detail.applicant }}</b>（骑手 {{ detail.applicantNo }}）</span>
            <span>申请时间 <b class="font-semibold text-ink-1">{{ formatTime(detail.created_at) }}</b></span>
            <span>提现类型 <b class="font-semibold text-ink-1">{{ detail.withdrawType }}</b></span>
          </div>
        </div>
        <div class="flex shrink-0 gap-2.5">
          <AppButton
            v-permission="'withdraw:audit'"
            variant="secondary"
            class="!border-[#FFD9D9] !text-[#FF4D4F] !px-[18px]"
            :disabled="submitting || detail.status !== '待审核'"
            @click="rejectOpen = true"
          >
            驳回申请
          </AppButton>
          <AppButton
            v-permission="'withdraw:audit'"
            variant="primary"
            class="!px-[18px]"
            :disabled="submitting || detail.status !== '待审核'"
            @click="confirmOpen = true"
          >
            {{ detail.status === '待审核' ? '确认打款' : '已打款' }}
          </AppButton>
        </div>
      </div>

      <!-- 主体区 -->
      <div class="flex gap-4">
        <!-- 左侧 -->
        <div class="flex flex-1 flex-col gap-4" style="flex:1.1;">
          <!-- 申请信息 -->
          <div class="rounded-card bg-white p-5">
            <div class="text-[15px] font-semibold text-ink-1">申请信息</div>
            <div class="mt-3.5 flex flex-col gap-2.5 text-[13px]">
              <div class="flex items-center justify-between border-b border-[#F2F3F5] pb-2.5">
                <span class="text-ink-3">提现单号</span><span class="font-semibold text-ink-1">{{ detail.id }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-ink-3">提现金额</span>
                <span class="text-[16px] font-bold" style="color:#FF6B2C;">{{ formatMoney(detail.amount) }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-ink-3">到账账户</span><span class="text-ink-2">{{ detail.bankInfo }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-ink-3">开户人</span><span class="text-ink-2">{{ detail.accountHolder }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-ink-3">申请时间</span><span class="text-ink-2">{{ formatTime(detail.created_at) }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-ink-3">预计到账</span><span class="text-ink-2">{{ detail.arriveExpect }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-ink-3">提现用途</span><span class="text-ink-2">{{ detail.purpose }}</span>
              </div>
            </div>
          </div>

          <!-- 账单核对 -->
          <div class="rounded-card bg-white p-5">
            <div class="text-[15px] font-semibold text-ink-1">账单核对</div>
            <div class="mt-4 grid grid-cols-3 gap-3.5">
              <div class="rounded-[10px] bg-head p-3.5">
                <div class="text-[12px] text-ink-3">可提现余额</div>
                <div class="mt-1.5 text-[22px] font-bold text-ink-1">{{ formatMoney(detail.balance, false) }}</div>
                <div class="mt-1 text-[11px] text-ink-3">{{ detail.balanceNote }}</div>
              </div>
              <div class="rounded-[10px] bg-head p-3.5">
                <div class="text-[12px] text-ink-3">本次提现</div>
                <div class="mt-1.5 text-[22px] font-bold" style="color:#FF6B2C;">{{ formatMoney(detail.amount, false) }}</div>
                <div class="mt-1 text-[11px] text-ink-3">{{ detail.usedPercentText }}</div>
              </div>
              <div class="rounded-[10px] bg-head p-3.5">
                <div class="text-[12px] text-ink-3">提现后余额</div>
                <div class="mt-1.5 text-[22px] font-bold text-ink-1">{{ formatMoney(detail.remainBalance, false) }}</div>
                <div class="mt-1 text-[11px] text-ink-3">{{ detail.remainNote }}</div>
              </div>
            </div>
            <div class="mt-4 flex flex-col gap-2.5 text-[13px]">
              <div class="flex items-center justify-between rounded-lg bg-head px-3 py-2.5">
                <span class="text-ink-3">本月已完成订单</span><span class="font-semibold text-ink-1">{{ detail.monthOrders }} 单</span>
              </div>
              <div class="flex items-center justify-between rounded-lg bg-head px-3 py-2.5">
                <span class="text-ink-3">本月配送收入</span><span class="font-semibold text-ink-1">{{ formatMoney(detail.monthIncome) }}</span>
              </div>
              <div class="flex items-center justify-between rounded-lg bg-head px-3 py-2.5">
                <span class="text-ink-3">上月同期提现次数</span><span class="font-semibold text-ink-1">{{ detail.lastMonthNote }}</span>
              </div>
            </div>
          </div>

          <!-- 审核时间线 -->
          <div class="rounded-card bg-white p-5">
            <div class="text-[15px] font-semibold text-ink-1">审核时间线</div>
            <div class="mt-4 flex flex-col">
              <div v-for="(t, i) in detail.timeline" :key="i" class="flex gap-3">
                <div class="flex flex-col items-center">
                  <div
                    class="mt-1 h-2.5 w-2.5 rounded-full"
                    :style="{ background: i === detail.timeline.length - 1 ? '#FF6B2C' : '#00B578' }"
                  />
                  <div v-if="i < detail.timeline.length - 1" class="w-px flex-1 bg-line" style="min-height:28px;" />
                </div>
                <div class="pb-5">
                  <div class="text-[13px] font-semibold text-ink-1">{{ t.text }}<span class="ml-2 font-normal text-ink-3">{{ t.operator }}</span></div>
                  <div class="mt-0.5 text-[12px] text-ink-3">{{ formatTime(t.time) }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 右侧 -->
        <div class="flex flex-1 flex-col gap-4">
          <!-- 风控检查 -->
          <div class="rounded-card bg-white p-5">
            <div class="flex items-center">
              <div class="text-[15px] font-semibold text-ink-1">风控检查</div>
              <span class="ml-auto rounded-tag px-2.5 py-[3px] text-[12px] font-semibold" style="background:#E8F7F0;color:#00B578;">全部通过</span>
            </div>
            <div class="mt-3.5 flex flex-col gap-2.5">
              <div v-for="(item, i) in detail.riskItems" :key="i" class="flex items-center gap-2 rounded-lg bg-head px-3 py-2.5">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#00B578" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="shrink-0">
                  <circle cx="12" cy="12" r="10" />
                  <polyline points="9 12 11 14 15 10" />
                </svg>
                <span class="text-[13px] text-ink-2">{{ item }}</span>
              </div>
            </div>
            <div class="mt-3 flex items-center gap-2 rounded-lg border px-3 py-2.5" style="background:#FFF7E6;border-color:#FFE7BA;">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#D48806" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class="shrink-0">
                <path d="M12 9v4M12 17h.01" />
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
              </svg>
              <span class="text-[12px]" style="color:#D48806;">{{ detail.riskWarning }}</span>
            </div>
          </div>

          <!-- 打款确认 -->
          <div class="rounded-card bg-white p-5">
            <div class="text-[15px] font-semibold text-ink-1">打款确认</div>
            <div class="mt-3.5">
              <div class="mb-1.5 text-[12px] text-ink-3">打款金额</div>
              <div class="flex h-[42px] items-center rounded-lg border px-3.5 text-[15px] font-bold" style="background:#FFF3E8;border-color:#FFD8C0;color:#FF6B2C;">
                {{ formatMoney(detail.amount) }}（含手续费 {{ formatMoney(detail.fee) }}，{{ detail.feeNote }}）
              </div>
              <div class="mb-1.5 mt-3.5 text-[12px] text-ink-3">审核备注（选填）</div>
              <textarea
                v-model="remark"
                class="h-14 w-full resize-none rounded-lg border border-line bg-head p-2.5 text-[13px] outline-none placeholder:text-ink-4"
                placeholder="填写审核备注，将同步给骑手…"
              />
              <div class="mt-4 flex gap-2.5">
                <AppButton
                  v-permission="'withdraw:audit'"
                  variant="primary"
                  class="!h-10 flex-1"
                  :disabled="submitting || detail.status !== '待审核'"
                  @click="confirmOpen = true"
                >
                  确认打款 {{ formatMoney(detail.amount) }}
                </AppButton>
                <AppButton
                  variant="secondary"
                  class="!h-10 flex-1"
                  :disabled="submitting"
                  @click="toast.info('已加入待复核队列，稍后处理')"
                >
                  加入待复核
                </AppButton>
              </div>
              <div class="mt-3 flex items-center gap-1.5 text-[12px] text-ink-4">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#C9CDD4" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10" />
                  <path d="M12 16v-4M12 8h.01" />
                </svg>
                打款后系统将推送通知并生成提现凭证，操作不可撤回
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <Confirm
      :open="confirmOpen"
      title="确认打款"
      :message="confirmMessage"
      confirm-text="确认打款"
      :loading="submitting"
      @update:open="confirmOpen = $event"
      @confirm="onPass"
    />
    <Modal
      :open="rejectOpen"
      title="驳回申请"
      confirm-text="确认驳回"
      :loading="submitting"
      @update:open="rejectOpen = $event"
      @confirm="onReject"
    >
      <div class="text-sm text-ink-2">确认驳回「{{ detail?.applicant }}」的提现申请？驳回后金额将退回账户余额。</div>
      <textarea
        v-model="remark"
        class="mt-3 h-20 w-full resize-none rounded-lg border border-line bg-head p-2.5 text-[13px] outline-none placeholder:text-ink-4"
        placeholder="请填写驳回原因（选填，将同步给骑手）…"
      />
    </Modal>
  </div>
</template>
