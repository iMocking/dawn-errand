<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import DataTable, { type Column } from '@/components/DataTable.vue'
import StatusTag from '@/components/StatusTag.vue'
import Pagination from '@/components/Pagination.vue'
import Modal from '@/components/Modal.vue'
import { useToast } from '@/composables/useToast'
import { formatDate } from '@/utils/format'
import {
  getRealNameVerifications,
  auditRealName,
  type RealNameVerify,
} from '@/api/users'

const toast = useToast()

const loading = ref(false)
const list = ref<RealNameVerify[]>([])
const total = ref(0)

const query = reactive({
  page: 1,
  pageSize: 10,
  status: '' as '' | '待审核' | '已通过' | '已驳回',
})

const STATUS_TABS = [
  { label: '全部', value: '' },
  { label: '待审核', value: '待审核' },
  { label: '已通过', value: '已通过' },
  { label: '已驳回', value: '已驳回' },
] as const

function statusTone(s: string): 'orange' | 'green' | 'red' {
  if (s === '已通过') return 'green'
  if (s === '已驳回') return 'red'
  return 'orange'
}

async function load() {
  loading.value = true
  try {
    const res = await getRealNameVerifications({
      page: query.page,
      pageSize: query.pageSize,
      status: query.status || undefined,
    })
    list.value = res.list
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function pickStatus(value: (typeof STATUS_TABS)[number]['value']) {
  query.status = value
  query.page = 1
  load()
}

function onPage(p: number) {
  query.page = p
  load()
}

/* 审核弹窗（待审核申请） */
const auditOpen = ref(false)
const submitting = ref(false)
const rejectReason = ref('')
const current = ref<RealNameVerify | null>(null)

function askAudit(row: RealNameVerify) {
  current.value = row
  rejectReason.value = ''
  auditOpen.value = true
}

async function onApprove() {
  if (!current.value || submitting.value) return
  submitting.value = true
  try {
    await auditRealName(current.value.id, 'approve')
    toast.success(`已通过 ${current.value.realName} 的实名认证`)
    auditOpen.value = false
    load()
  } finally {
    submitting.value = false
  }
}

async function onReject() {
  if (!current.value || submitting.value) return
  submitting.value = true
  try {
    await auditRealName(current.value.id, 'reject', rejectReason.value)
    toast.info(`已驳回 ${current.value.realName} 的实名认证申请`)
    auditOpen.value = false
    load()
  } finally {
    submitting.value = false
  }
}

const columns: Column[] = [
  { key: 'realName', title: '申请人', width: '140px' },
  { key: 'userId', title: '用户编号', width: '120px' },
  { key: 'idCardMasked', title: '身份证号', width: '200px' },
  { key: 'status', title: '状态', width: '100px' },
  { key: 'appliedAt', title: '提交时间', width: '140px' },
  { key: 'action', title: '操作', align: 'right' },
]

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-4">
    <!-- 标题行 -->
    <div class="flex shrink-0 items-center">
      <div class="flex items-center gap-3">
        <div class="text-[16px] font-bold text-ink-1">实名认证</div>
        <span class="text-[12px] text-ink-3">共 {{ total }} 条申请</span>
      </div>
      <div class="ml-auto flex items-center gap-2 text-[13px] text-ink-2">
        状态
        <button
          v-for="t in STATUS_TABS"
          :key="t.value"
          type="button"
          class="rounded-md px-3.5 py-1.5 text-[12px]"
          :class="query.status === t.value
            ? 'bg-primary-bg font-semibold text-primary'
            : 'bg-head font-normal text-ink-2'"
          @click="pickStatus(t.value)"
        >{{ t.label }}</button>
      </div>
    </div>

    <!-- 说明条 -->
    <div class="flex shrink-0 items-center gap-2 rounded-lg bg-head px-3 py-2 text-[12px] text-ink-3">
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#2B7FFF" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class="shrink-0">
        <circle cx="12" cy="12" r="10" /><path d="M12 16v-4M12 8h.01" />
      </svg>
      用户提交实名申请后在此审核；通过后用户资料同步为「已实名」，身份证号仅以脱敏形式展示。
    </div>

    <!-- 列表 -->
    <div class="rounded-card bg-white">
      <DataTable :columns="columns" :rows="list" :loading="loading" row-key="id">
        <template #col-realName="{ row }">
          <div class="flex items-center gap-3">
            <div
              class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-[14px] font-bold"
              style="background:#E8F4FF;color:#2B7FFF;"
            >
              {{ row.realName.charAt(0) }}
            </div>
            <span class="text-[14px] font-semibold text-ink-1">{{ row.realName }}</span>
          </div>
        </template>
        <template #col-userId="{ row }">
          <span class="text-[13px] text-ink-2">{{ row.userId }}</span>
        </template>
        <template #col-idCardMasked="{ row }">
          <span class="font-mono text-[13px] text-ink-1">{{ row.idCardMasked }}</span>
        </template>
        <template #col-status="{ row }">
          <StatusTag :tone="statusTone(row.status)">{{ row.status }}</StatusTag>
        </template>
        <template #col-appliedAt="{ row }">
          <span class="text-[13px] text-ink-3">{{ formatDate(row.appliedAt) }}</span>
        </template>
        <template #col-action="{ row }">
          <div class="flex items-center justify-end gap-3 text-[13px]">
            <span
              v-if="row.status === '待审核'"
              v-permission="'user:audit'"
              class="cursor-pointer font-semibold text-primary"
              @click="askAudit(row)"
            >
              审核
            </span>
            <span
              v-else-if="row.status === '已驳回' && row.rejectReason"
              class="max-w-[260px] cursor-help truncate text-ink-3"
              :title="row.rejectReason"
            >
              驳回：{{ row.rejectReason }}
            </span>
            <span v-else class="text-ink-4">—</span>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- 分页 -->
    <div class="flex shrink-0 items-center text-[13px] text-ink-3">
      <span>共 {{ total }} 条申请 · 每页 {{ query.pageSize }} 条</span>
      <div class="ml-auto">
        <Pagination :page="query.page" :page-size="query.pageSize" :total="total" @change="onPage" />
      </div>
    </div>

    <!-- 审核弹窗 -->
    <Modal
      v-model:open="auditOpen"
      title="实名认证审核"
      confirm-text="确认通过"
      :loading="submitting"
      :width="460"
      @confirm="onApprove"
    >
      <div v-if="current" class="flex flex-col gap-3.5 py-1">
        <div class="grid grid-cols-2 gap-x-4 gap-y-3 text-[13px]">
          <div class="flex justify-between text-ink-3"><span>申请人</span><span class="font-semibold text-ink-1">{{ current.realName }}</span></div>
          <div class="flex justify-between text-ink-3"><span>身份证号</span><span class="font-mono text-ink-1">{{ current.idCardMasked }}</span></div>
          <div class="flex justify-between text-ink-3"><span>提交时间</span><span class="text-ink-1">{{ formatDate(current.appliedAt) }}</span></div>
          <div class="flex justify-between text-ink-3"><span>当前状态</span><span class="text-[#D48806]">{{ current.status }}</span></div>
        </div>
        <p class="text-[12px] leading-[1.8] text-ink-4">
          确认该用户身份信息真实有效？通过后该账号将标记为「已实名」，实名信息同步到用户资料；驳回后用户可修改后重新提交。
        </p>
        <div class="flex items-center gap-2 rounded-lg bg-head px-3 py-2.5">
          <span class="text-[12px] text-ink-3">或选择驳回</span>
          <input
            v-model="rejectReason"
            placeholder="填写驳回原因（选填）"
            class="flex-1 rounded-lg border border-line bg-white px-2.5 py-1.5 text-[12px] outline-none placeholder:text-ink-4"
          />
          <button
            type="button"
            class="rounded-md px-3 py-1.5 text-[12px] font-semibold"
            style="background:#FFF3E8;color:#FF6B2C;"
            @click="onReject"
          >驳回</button>
        </div>
      </div>
    </Modal>
  </div>
</template>
