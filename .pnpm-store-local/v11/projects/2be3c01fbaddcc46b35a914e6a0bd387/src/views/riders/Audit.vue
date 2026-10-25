<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Modal from '@/components/Modal.vue'
import StatusTag from '@/components/StatusTag.vue'
import AppButton from '@/components/AppButton.vue'
import { useToast } from '@/composables/useToast'
import { useWorkspaceStore } from '@/stores/workspace'
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
const workspace = useWorkspaceStore()

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

/** 身份证核验状态语义色 */
function idCheckTone(s: string): 'green' | 'orange' | 'red' {
  if (s === '已通过') return 'green'
  if (s === '已驳回') return 'red'
  return 'orange'
}

/** 证件/身份信息查看弹窗 */
const infoTarget = ref<RiderAuditRow | null>(null)
const infoOpen = ref(false)
function askInfo(row: RiderAuditRow) {
  infoTarget.value = row
  infoOpen.value = true
}

/** 通过：确认弹窗 → POST（联动身份证核验状态） */
const approveTarget = ref<RiderAuditRow | null>(null)
const approveOpen = ref(false)
const submitting = ref(false)
const accountOpen = ref(false)
const accountResult = ref<{ email: string; phone: string; initialPassword: string } | null>(null)

function askApprove(row: RiderAuditRow) {
  approveTarget.value = row
  approveOpen.value = true
}
async function onApprove() {
  const row = approveTarget.value
  if (!row || submitting.value) return
  submitting.value = true
  try {
    const res = await approveRider(row.id)
    toast.success(`已通过 ${row.name} 的入驻申请，账号已开通`)
    if (res.account?.initialPassword) {
      accountResult.value = res.account
      accountOpen.value = true
    }
    list.value = list.value.filter((r) => r.id !== row.id)
    approveOpen.value = false
    workspace.refresh()
    loadStats()
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
  if (!reason.value.trim()) {
    toast.warning('请填写驳回原因，便于骑手补件')
    return
  }
  submitting.value = true
  try {
    await rejectRider(row.id, reason.value.trim())
    toast.success(`已驳回 ${row.name} 的入驻申请`)
    list.value = list.value.filter((r) => r.id !== row.id)
    rejectOpen.value = false
    workspace.refresh()
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
            <StatusTag :tone="idCheckTone(row.idCheck)">{{ row.idCheck }}</StatusTag>
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
          <span class="ml-auto cursor-pointer text-[12px] font-semibold" style="color:#2B7FFF;" @click="askInfo(row)">查看</span>
        </div>

        <div class="mt-2 flex gap-2">
          <AppButton
            v-permission="'rider:audit'"
            class="!h-[30px] flex-1"
            :loading="submitting"
            @click="askApprove(row)"
          >
通过入驻
</AppButton>
          <AppButton
            v-permission="'rider:audit'"
            variant="secondary"
            class="!h-[30px] flex-1"
            :disabled="submitting"
            @click="askReject(row)"
          >
驳回
</AppButton>
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
          本月已通过 {{ stats?.audit.monthApproved ?? 0 }} 人，较上月 {{ stats?.audit.momRate || '—' }}
        </div>
      </div>
    </div>

    <!-- 身份信息查看 -->
    <Modal v-model:open="infoOpen" title="身份信息核验" :width="460">
      <div v-if="infoTarget" class="flex flex-col gap-3 py-1">
        <div class="grid grid-cols-2 gap-x-4 gap-y-3 text-[13px]">
          <div class="flex justify-between text-ink-3"><span>姓名</span><span class="font-semibold text-ink-1">{{ infoTarget.name }}</span></div>
          <div class="flex justify-between text-ink-3"><span>身份证号</span><span class="text-ink-1">{{ infoTarget.idCardMasked || '未填写' }}</span></div>
          <div class="flex justify-between text-ink-3"><span>联系电话</span><span class="text-ink-1">{{ infoTarget.phone }}</span></div>
          <div class="flex justify-between text-ink-3"><span>核验状态</span><StatusTag :tone="idCheckTone(infoTarget.idCheck)">{{ infoTarget.idCheck }}</StatusTag></div>
        </div>
        <div class="rounded-lg bg-head px-3 py-2.5 text-[12px] leading-[1.8] text-ink-2">
          <div class="mb-1 font-semibold text-ink-1">已上传材料（{{ infoTarget.docCount }} 张）</div>
          <div>{{ infoTarget.docs || '—' }}</div>
          <div class="mt-1.5 text-ink-4">通过入驻后，身份证核验状态将自动更新为「已通过」；驳回后为「已驳回」，骑手重新提交资料后恢复「待补充」。</div>
        </div>
      </div>
    </Modal>

    <!-- 通过确认 -->
    <Modal
      v-model:open="approveOpen"
      title="审核通过"
      confirm-text="确认通过"
      :loading="submitting"
      @confirm="onApprove"
    >
      <p class="text-sm text-ink-2">确认通过骑手 {{ approveTarget?.name }}（{{ approveTarget?.applyNo }}）的入驻申请？通过后将自动开通骑手端账号。</p>
      <div
        v-if="approveTarget?.idCheck !== '已通过'"
        class="mt-3 rounded-lg border px-3 py-2.5 text-[12px] leading-[1.8]"
        style="background:#FFF7E6;border-color:#FFE7BA;color:#D48806;"
      >
        该骑手身份证核验状态为「{{ approveTarget?.idCheck }}」。通过入驻将同时把身份证核验状态置为「已通过」，请确认证件真实有效。
      </div>
    </Modal>

    <!-- 驳回 -->
    <Modal
      v-model:open="rejectOpen"
      title="驳回申请"
      confirm-text="确认驳回"
      :loading="submitting"
      @confirm="onReject"
    >
      <p class="text-sm text-ink-2">确认驳回骑手 {{ rejectTarget?.name }}（{{ rejectTarget?.applyNo }}）的入驻申请？驳回后身份证核验状态将置为「已驳回」，骑手重新提交资料后恢复「待补充」。</p>
      <textarea
        v-model="reason"
        rows="3"
        placeholder="请填写驳回原因（必填，骑手据此补件）"
        class="mt-3 w-full rounded-lg border border-line bg-head px-3 py-2 text-[13px] outline-none placeholder:text-ink-4"
      ></textarea>
    </Modal>

    <Modal v-model:open="accountOpen" title="骑手账号已开通" :width="500" confirm-text="我已记录" @confirm="accountOpen = false">
      <div class="rounded-lg bg-[#E8F7F0] px-3 py-2.5 text-[12px]" style="color:#00B578;">
        审核已通过，骑手端账号已创建并绑定；初始密码仅展示一次，请线下交付骑手。
      </div>
      <div class="mt-4 space-y-3">
        <div class="flex items-center justify-between rounded-lg border border-line px-3 py-2.5">
          <div>
            <div class="text-[11px] text-ink-4">登录手机号</div>
            <div class="mt-0.5 font-mono text-[13px] font-semibold text-ink-1">{{ accountResult?.phone || '—' }}</div>
          </div>
        </div>
        <div class="flex items-center justify-between rounded-lg border border-line px-3 py-2.5">
          <div>
            <div class="text-[11px] text-ink-4">初始密码</div>
            <div class="mt-0.5 font-mono text-[13px] font-semibold text-ink-1">{{ accountResult?.initialPassword || '—' }}</div>
          </div>
        </div>
      </div>
      <div class="mt-3 text-[11px] text-ink-4">骑手登录后请在骑手端修改密码；密码遗失可在骑手详情重置。</div>
    </Modal>
  </div>
</template>
