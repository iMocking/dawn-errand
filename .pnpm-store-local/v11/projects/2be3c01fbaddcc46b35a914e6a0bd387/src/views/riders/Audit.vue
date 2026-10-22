<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Modal from '@/components/Modal.vue'
import StatusTag from '@/components/StatusTag.vue'
import { useToast } from '@/composables/useToast'
import { formatTime } from '@/utils/format'
import {
  getRiderAudit,
  getRiderStats,
  approveRider,
  rejectRider,
  type RiderAuditRow,
  type RiderStatsDto,
} from '@/api/riders'

const toast = useToast()

const list = ref<RiderAuditRow[]>([])
const loading = ref(false)
const stats = ref<RiderStatsDto | null>(null)

async function load() {
  loading.value = true
  try {
    const res = await getRiderAudit()
    list.value = res.list
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  stats.value = await getRiderStats()
}

/** 通过：确认弹窗 → POST */
const approveTarget = ref<RiderAuditRow | null>(null)
const approveOpen = ref(false)
const submitting = ref(false)

function askApprove(row: RiderAuditRow) {
  approveTarget.value = row
  approveOpen.value = true
}
async function onApprove() {
  const row = approveTarget.value
  if (!row || submitting.value) return
  submitting.value = true
  try {
    await approveRider(row.id)
    toast.success(`已通过 ${row.name} 的入驻申请`)
    list.value = list.value.filter((r) => r.id !== row.id)
    approveOpen.value = false
  } finally {
    submitting.value = false
  }
}

/** 驳回：理由弹窗 → POST */
const rejectTarget = ref<RiderAuditRow | null>(null)
const rejectOpen = ref(false)
const reason = ref('')
function askReject(row: RiderAuditRow) {
  rejectTarget.value = row
  reason.value = ''
  rejectOpen.value = true
}
async function onReject() {
  const row = rejectTarget.value
  if (!row || submitting.value) return
  submitting.value = true
  try {
    await rejectRider(row.id, reason.value)
    toast.success(`已驳回 ${row.name} 的入驻申请`)
    list.value = list.value.filter((r) => r.id !== row.id)
    rejectOpen.value = false
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  load()
  loadStats()
})
</script>

<template>
  <div class="flex flex-col gap-3">
    <!-- 统计与筛选条 -->
    <div class="flex items-center gap-8 rounded-card bg-white px-5 py-3">
      <div class="flex items-center gap-2">
        <span class="text-[14px] font-semibold text-ink-1">骑手入驻审核</span>
        <span class="text-[12px] text-ink-3">共 {{ list.length }} 条待审核申请</span>
      </div>
      <div class="ml-auto flex gap-2">
        <span class="cursor-pointer rounded-md px-3.5 py-1.5 text-[12px] font-semibold" style="background:#FFF3E8;color:#FF6B2C;">待审核 {{ list.length }}</span>
        <span class="cursor-pointer rounded-md bg-head px-3.5 py-1.5 text-[12px] text-ink-2">今日新增 {{ stats?.audit.todayNew }}</span>
        <span class="cursor-pointer rounded-md bg-head px-3.5 py-1.5 text-[12px] text-ink-2">已通过 {{ stats?.audit.approved }}</span>
        <span class="cursor-pointer rounded-md bg-head px-3.5 py-1.5 text-[12px] text-ink-2">已驳回 {{ stats?.audit.rejected }}</span>
      </div>
    </div>

    <!-- 申请卡片列表 -->
    <div v-if="loading" class="rounded-card bg-white p-10 text-center text-ink-3">加载中…</div>
    <div v-else class="grid grid-cols-2 gap-3">
      <div
        v-for="row in list"
        :key="row.id"
        class="rounded-card bg-white px-5 py-3"
      >
        <div class="flex items-center gap-3">
          <div
            class="flex h-[38px] w-11 shrink-0 items-center justify-center rounded-full text-[17px] font-bold"
            :style="{ background: row.avatarBg, color: row.avatarColor }"
          >
{{ row.name.charAt(0) }}
</div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="text-[15px] font-semibold text-ink-1">{{ row.name }}</span>
              <StatusTag tone="orange">申请中</StatusTag>
            </div>
            <div class="mt-[3px] text-[12px] text-ink-3">
              申请编号 {{ row.applyNo }} · 提交 {{ formatTime(row.submittedAt) }}
            </div>
          </div>
          <div class="shrink-0 rounded-md bg-head px-2.5 py-1.5 text-[12px] text-ink-2">{{ row.area }}</div>
        </div>

        <div class="mt-2 grid grid-cols-2 gap-x-3.5 gap-y-2">
          <div class="flex justify-between text-[12px] text-ink-3"><span>联系电话</span><span class="text-ink-1">{{ row.phone }}</span></div>
          <div class="flex justify-between text-[12px] text-ink-3"><span>配送工具</span><span class="text-ink-1">{{ row.tool }}</span></div>
          <div class="flex justify-between text-[12px] text-ink-3"><span>接单类型</span><span class="text-ink-1">{{ row.orderTypes }}</span></div>
          <div class="flex justify-between text-[12px] text-ink-3">
            <span>身份证核验</span>
            <span :style="row.idCheck === '已通过' ? 'color:#00B578;' : 'color:#D48806;'">{{ row.idCheck }}</span>
          </div>
        </div>

        <div
          v-if="row.noticeType === 'warning'"
          class="mt-2 flex items-center gap-2 rounded-lg border px-3 py-2"
          style="background:#FFF7E6;border-color:#FFE7BA;"
        >
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#D48806" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class="shrink-0">
            <path d="M12 9v4M12 17h.01" />
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
          </svg>
          <span class="text-[12px]" style="color:#D48806;">{{ row.notice }}</span>
        </div>
        <div v-else class="mt-2 flex items-center gap-2 rounded-lg bg-head px-3 py-2">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#86909C" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class="shrink-0">
            <rect x="3" y="3" width="18" height="18" rx="2" /><circle cx="9" cy="9" r="2" /><path d="M21 15l-5-5L5 21" />
          </svg>
          <span class="text-[12px] text-ink-2">证件已上传 {{ row.docCount }} 张：{{ row.docs }}</span>
          <span class="ml-auto cursor-pointer text-[12px] font-semibold" style="color:#2B7FFF;">查看</span>
        </div>

        <div class="mt-2 flex gap-2">
          <button
            v-permission="'rider:audit'"
            class="h-[30px] flex-1 rounded-lg text-[13px] font-semibold text-white disabled:opacity-60"
            style="background:linear-gradient(135deg,#FF8A3D,#FF6B2C);"
            :disabled="submitting"
            @click="askApprove(row)"
          >
通过入驻
</button>
          <button
            v-permission="'rider:audit'"
            class="h-[30px] flex-1 rounded-lg border border-line bg-white text-[13px] text-ink-2 disabled:opacity-60"
            :disabled="submitting"
            @click="askReject(row)"
          >
驳回
</button>
        </div>
      </div>

      <!-- 审核标准提醒卡 -->
      <div class="flex flex-col justify-center rounded-card bg-white px-5 py-3">
        <div class="text-[14px] font-semibold text-ink-1">审核标准提醒</div>
        <div class="mt-2 text-[12px] leading-[1.9] text-ink-3">
          · 身份证、健康证须为原件拍摄，且在有效期内<br />
          · 无犯罪记录证明开具时间不超过 30 天<br />
          · 配送工具须为合规的电动二轮 / 三轮车<br />
          · 同一申请人 30 天内最多提交 2 次
        </div>
        <div class="mt-2 flex items-center gap-2 rounded-lg bg-head px-3 py-2 text-[12px] text-ink-2">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#2B7FFF" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class="shrink-0">
            <circle cx="12" cy="12" r="10" /><path d="M12 16v-4M12 8h.01" />
          </svg>
          本月已通过 {{ stats?.audit.monthApproved ?? 78 }} 人，较上月 {{ stats?.audit.momRate ?? '+12.4%' }}
        </div>
      </div>
    </div>

    <!-- 通过确认 -->
    <Modal
      v-model:open="approveOpen"
      title="审核通过"
      confirm-text="确认通过"
      :loading="submitting"
      @confirm="onApprove"
    >
      <p class="text-sm text-ink-2">确认通过骑手 {{ approveTarget?.name }}（{{ approveTarget?.applyNo }}）的入驻申请？通过后该骑手将自动上线接单。</p>
    </Modal>

    <!-- 驳回 -->
    <Modal
      v-model:open="rejectOpen"
      title="驳回申请"
      confirm-text="确认驳回"
      :loading="submitting"
      @confirm="onReject"
    >
      <p class="text-sm text-ink-2">确认驳回骑手 {{ rejectTarget?.name }}（{{ rejectTarget?.applyNo }}）的入驻申请？</p>
      <textarea
        v-model="reason"
        rows="3"
        placeholder="请填写驳回原因（选填）"
        class="mt-3 w-full rounded-lg border border-line bg-head px-3 py-2 text-[13px] outline-none placeholder:text-ink-4"
      ></textarea>
    </Modal>
  </div>
</template>
