<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import StatusTag from '@/components/StatusTag.vue'
import Modal from '@/components/Modal.vue'
import AppButton from '@/components/AppButton.vue'
import { useToast } from '@/composables/useToast'
import { useWorkspaceStore } from '@/stores/workspace'
import { formatTime } from '@/utils/format'
import {
  getMerchantAuditList,
  auditMerchant,
  type MerchantAuditRow,
} from '@/api/merchants'

const toast = useToast()
const router = useRouter()
const workspace = useWorkspaceStore()

const rows = ref<MerchantAuditRow[]>([])
const loading = ref(false)
const activeTab = ref('待审核')
const tabs = ['待审核', '已通过', '已驳回']

const submittingId = ref('')
const confirmState = ref<{ open: boolean; type: 'approve' | 'reject'; row: MerchantAuditRow | null }>({
  open: false,
  type: 'approve',
  row: null,
})
const rejectReason = ref('')
const accountOpen = ref(false)
const accountResult = ref<{ userId: string; email: string; phone: string; initialPassword: string } | null>(null)

const logoPaths: Record<MerchantAuditRow['logoIcon'], string> = {
  food: '<path d="M3 11h18c0 5-4 8-9 8s-9-3-9-8z"/><path d="M8 11V7a4 4 0 0 1 8 0v4"/>',
  shop: '<path d="M6 8h12l-1.5 13h-9z"/><path d="M8 8a4 4 0 0 1 8 0"/>',
  pharmacy: '<rect x="4" y="4" width="16" height="16" rx="3"/><path d="M12 8v8M8 12h8"/>',
}

const certIconPaths: Record<string, string> = {
  license: '<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="M21 15l-5-5L5 21"/>',
  idcard: '<circle cx="12" cy="8" r="4"/><path d="M4 21v-1a8 8 0 0 1 16 0v1"/>',
  storefront: '<path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><path d="M9 22V12h6v10"/>',
  foodlicense: '<path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5M2 12l10 5 10-5"/>',
  pharmacylicense: '<path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5M2 12l10 5 10-5"/>',
}

function statusTone(s: string): 'yellow' | 'green' | 'red' {
  if (s === '待审核') return 'yellow'
  if (s === '已通过') return 'green'
  return 'red'
}

const filteredRows = computed(() => rows.value.filter((r) => r.status === activeTab.value))
const pendingCount = computed(() => rows.value.filter((r) => r.status === '待审核').length)

async function load() {
  loading.value = true
  try {
    const res = await getMerchantAuditList({ page: 1, pageSize: 50 })
    rows.value = res.list
  } finally {
    loading.value = false
  }
}

function openConfirm(row: MerchantAuditRow, type: 'approve' | 'reject') {
  confirmState.value = { open: true, type, row }
  rejectReason.value = ''
}

async function onConfirm() {
  const { row, type } = confirmState.value
  if (!row) return
  if (type === 'reject' && !rejectReason.value.trim()) {
    toast.warning('请填写驳回原因，便于业务员补件')
    return
  }
  submittingId.value = row.id
  try {
    const res = await auditMerchant(row.id, type, rejectReason.value.trim() || undefined)
    if (type === 'approve' && res.account?.initialPassword) {
      accountResult.value = res.account
      accountOpen.value = true
    }
    toast.success(type === 'approve' ? `已通过 ${row.name} 的入驻申请，账号已开通` : `已驳回 ${row.name} 的入驻申请`)
    confirmState.value.open = false
    activeTab.value = type === 'approve' ? '已通过' : '已驳回'
    await load()
    workspace.refresh()
  } finally {
    submittingId.value = ''
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
  <div class="flex flex-col gap-4">
    <!-- 标题行 -->
    <div class="flex items-center">
      <div>
        <div class="text-[18px] font-bold text-ink-1">入驻申请审核</div>
        <div class="mt-1 text-[12px] text-ink-3">
          共 <span class="font-semibold text-primary">{{ pendingCount }}</span> 条待审核申请
        </div>
      </div>
      <div class="ml-auto flex items-center gap-2">
        <AppButton
          v-for="t in tabs"
          :key="t"
          variant="text"
          class="!rounded-md !px-3.5 !py-1.5 !leading-normal"
          :class="activeTab === t ? '!bg-primary-bg !font-semibold !text-primary' : '!bg-head !font-normal !text-ink-2 hover:!text-ink-1'"
          @click="activeTab = t"
        >
{{ t }}
</AppButton>
      </div>
    </div>

    <!-- 申请卡片列表 -->
    <div v-if="loading" class="py-10 text-center text-ink-3">加载中…</div>
    <div v-else-if="filteredRows.length === 0" class="rounded-card bg-white py-10 text-center text-ink-4">暂无{{ activeTab }}申请</div>
    <div v-else class="flex flex-col gap-4">
      <div
        v-for="row in filteredRows"
        :key="row.id"
        class="flex gap-5 rounded-card bg-white p-[18px_20px]"
      >
        <!-- logo -->
        <div
          class="flex h-16 w-16 shrink-0 items-center justify-center rounded-xl"
          :style="{ background: row.logoBg }"
        >
          <svg width="30" height="30" viewBox="0 0 24 24" fill="none" :stroke="row.logoColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" v-html="logoPaths[row.logoIcon]" />
        </div>

        <!-- 信息区 -->
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2.5">
            <span class="text-[16px] font-bold text-ink-1">{{ row.name }}</span>
            <StatusTag :tone="statusTone(row.status)">{{ row.status }}</StatusTag>
            <span class="text-[12px] text-ink-4">申请编号：{{ row.applyNo }}</span>
          </div>
          <div class="mt-2.5 flex flex-wrap gap-x-6 gap-y-1 text-[13px] text-ink-2">
            <span>经营者：{{ row.operator }}</span>
            <span>类型：{{ row.type }}</span>
            <span>地址：{{ row.address }}</span>
            <span>申请时间：{{ formatTime(row.appliedAt) }}</span>
          </div>
          <!-- 证照 -->
          <div class="mt-3 flex flex-wrap gap-2">
            <component
              :is="cert.url ? 'a' : 'div'"
              v-for="cert in row.certs"
              :key="cert.label"
              :href="cert.url || undefined"
              :target="cert.url ? '_blank' : undefined"
              class="flex h-16 w-16 flex-col items-center justify-center gap-0.5 overflow-hidden rounded-lg bg-head"
              style="border:1px dashed #D0D3D9;"
            >
              <img v-if="cert.url" :src="cert.url" alt="" class="h-full w-full object-cover" />
              <template v-else>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#86909C" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" v-html="certIconPaths[cert.icon]" />
                <span class="text-[10px] text-ink-3">{{ cert.label }}</span>
              </template>
            </component>
          </div>
          <div class="mt-3 grid grid-cols-2 gap-x-6 gap-y-1 rounded-lg bg-head px-3 py-2 text-[12px] text-ink-3">
            <span>执照号：{{ row.application?.licenseNo || '—' }}</span>
            <span>法人：{{ row.application?.legalName || '—' }} {{ row.application?.legalIdCard || '' }}</span>
            <span>结算：{{ row.application?.bankName || '—' }} · {{ row.application?.bankAccountName || '—' }}</span>
            <span>登录账号：{{ row.application?.accountPhone || row.phone }}</span>
            <span v-if="row.application?.foodLicenseNo">食品许可证：{{ row.application.foodLicenseNo }}</span>
            <span v-if="row.rejectReason" class="col-span-2 text-[#FF4D4F]">驳回原因：{{ row.rejectReason }}</span>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div v-if="row.status === '待审核'" class="flex shrink-0 flex-col justify-center gap-2.5">
          <AppButton
            v-permission="'merchant:audit'"
            class="!px-6"
            :loading="submittingId === row.id"
            @click="openConfirm(row, 'approve')"
          >
{{ submittingId === row.id ? '提交中…' : '通过入驻' }}
</AppButton>
          <AppButton
            v-permission="'merchant:audit'"
            variant="secondary"
            class="!px-6 !text-[#FF4D4F]"
            :disabled="submittingId === row.id"
            @click="openConfirm(row, 'reject')"
          >
{{ submittingId === row.id ? '提交中…' : '驳回申请' }}
</AppButton>
        </div>
        <div v-else-if="row.status === '已驳回'" class="flex shrink-0 flex-col justify-center gap-2.5">
          <AppButton variant="secondary" class="!px-5" @click="router.push(`/merchants/${row.id}/edit`)">
            修改资料
          </AppButton>
        </div>
      </div>
    </div>

    <!-- 审核确认弹窗 -->
    <Modal
      v-model:open="confirmState.open"
      :title="confirmState.type === 'approve' ? '通过入驻审核' : '驳回入驻申请'"
      :loading="!!submittingId"
      :confirm-text="confirmState.type === 'approve' ? '确认通过' : '确认驳回'"
      @confirm="onConfirm"
    >
      <p class="text-sm text-ink-2">
        {{ confirmState.type === 'approve' ? '确认通过' : '确认驳回' }}
        <span class="font-semibold text-ink-1">{{ confirmState.row?.name }}</span>
        的入驻申请？
      </p>
      <div v-if="confirmState.type === 'approve'" class="mt-3 rounded-lg bg-head px-3 py-2 text-[12px] text-ink-3">
        通过后将自动开通商家端账号，并生成一次性初始密码；商家可立即登录，但需自行开店后才会接单。
      </div>
      <div v-if="confirmState.type === 'reject'" class="mt-3">
        <textarea
          v-model="rejectReason"
          rows="3"
          placeholder="请填写驳回原因（必填，业务员据此补件）"
          class="w-full resize-none rounded-lg border border-line bg-head px-3 py-2 text-[13px] outline-none focus:border-primary"
        ></textarea>
      </div>
    </Modal>

    <!-- 账号开通结果：初始密码仅展示一次 -->
    <Modal
      v-model:open="accountOpen"
      title="商家账号已开通"
      :width="520"
      confirm-text="我已记录"
      @confirm="accountOpen = false"
    >
      <div class="rounded-lg bg-[#E8F7F0] px-3 py-2.5 text-[12px]" style="color:#00B578;">
        审核已通过，商家端账号已创建并绑定。请将以下信息线下交付给商家；初始密码仅本次展示。
      </div>
      <div class="mt-4 space-y-3">
        <div v-for="item in [
          { label: '登录手机号', value: accountResult?.phone || '—' },
          { label: '登录邮箱', value: accountResult?.email || '—' },
          { label: '初始密码', value: accountResult?.initialPassword || '—' },
        ]" :key="item.label" class="flex items-center justify-between rounded-lg border border-line px-3 py-2.5">
          <div>
            <div class="text-[11px] text-ink-4">{{ item.label }}</div>
            <div class="mt-0.5 font-mono text-[13px] font-semibold text-ink-1">{{ item.value }}</div>
          </div>
          <button
            v-if="item.value !== '—'"
            type="button"
            class="text-[12px] text-primary"
            @click="copyText(item.value)"
          >复制</button>
        </div>
      </div>
      <div class="mt-3 text-[11px] text-ink-4">
        商家登录后请尽快修改密码；如密码遗失，可由平台在商家详情重新开通/重置。
      </div>
    </Modal>
  </div>
</template>
