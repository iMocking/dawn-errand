<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import Modal from '@/components/Modal.vue'
import Confirm from '@/components/Confirm.vue'
import StatusTag from '@/components/StatusTag.vue'
import Empty from '@/components/Empty.vue'
import AppButton from '@/components/AppButton.vue'
import AppInput from '@/components/AppInput.vue'
import SmtpConfig from './SmtpConfig.vue'
import SmsConfig from './SmsConfig.vue'
import { useToast } from '@/composables/useToast'
import { formatNumber } from '@/utils/format'
import { getSettings, saveSettings, type ServiceArea, type SystemSettings } from '@/api/settings'

const toast = useToast()

const loading = ref(false)
const settings = ref<SystemSettings | null>(null)

/**
 * 页签：服务区域 / 邮箱 SMTP / 短信服务在本页内切换；
 * 配送费率、公告管理在左侧菜单已有独立入口（/settings/fee-rules、/settings/announcements），不再重复；
 * 提现规则、账号权限按设计稿占位。
 */
const tabs = [
  { label: '服务区域' },
  { label: '邮箱 SMTP 服务配置' },
  { label: '短信服务配置' },
  { label: '提现规则' },
  { label: '账号权限' },
]
const localTab = ref('服务区域')
const activeTab = computed(() => localTab.value)
function pickTab(t: { label: string }) {
  localTab.value = t.label
}

async function load() {
  loading.value = true
  try {
    settings.value = await getSettings()
  } finally {
    loading.value = false
  }
}
onMounted(load)

function areaStatusTone(status: string) {
  return status === '运营中' ? 'green' : 'gray'
}

/* 新增 / 编辑服务区域 */
const formOpen = ref(false)
const submitting = ref(false)
const editing = ref<ServiceArea | null>(null)
const form = reactive({ name: '', villages: 8, radius_km: 12, start_fee: 4 })

function openCreate() {
  editing.value = null
  form.name = ''
  form.villages = 8
  form.radius_km = 12
  form.start_fee = 4
  formOpen.value = true
}
function openEdit(row: ServiceArea) {
  editing.value = row
  form.name = row.name
  form.villages = row.villages
  form.radius_km = row.radius_km
  form.start_fee = row.start_fee
  formOpen.value = true
}
async function onSubmit() {
  if (!form.name.trim()) {
    toast.warning('请输入区域名称')
    return
  }
  submitting.value = true
  try {
    await saveSettings({
      ...(editing.value ?? {}),
      action: editing.value ? 'update' : 'create',
      name: form.name.trim(),
      villages: Number(form.villages),
      radius_km: Number(form.radius_km),
      start_fee: Number(form.start_fee),
    })
    toast.success(editing.value ? '区域信息已保存' : '新增服务区域成功')
    formOpen.value = false
  } finally {
    submitting.value = false
  }
}

/* 停用 / 删除 */
const confirmOpen = ref(false)
const confirmTitle = ref('')
const confirmMessage = ref('')
const confirmAction = ref<{ row: ServiceArea } | null>(null)
const actionSubmitting = ref(false)

function askToggle(row: ServiceArea) {
  confirmAction.value = { row }
  if (row.action === '删除') {
    confirmTitle.value = '删除区域'
    confirmMessage.value = `确认删除区域「${row.name}」？删除后不可恢复。`
  } else {
    confirmTitle.value = '停用区域'
    confirmMessage.value = `确认停用区域「${row.name}」？停用后该区域暂停接单。`
  }
  confirmOpen.value = true
}
async function onConfirmAction() {
  if (!confirmAction.value || actionSubmitting.value) return
  actionSubmitting.value = true
  try {
    const row = confirmAction.value.row
    await saveSettings({ action: row.action, id: row.id })
    toast.success(`已${row.action}区域「${row.name}」`)
    confirmOpen.value = false
  } finally {
    actionSubmitting.value = false
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <!-- 设置页签（照设计稿） -->
    <div class="flex shrink-0 items-center gap-5">
      <span
        v-for="t in tabs"
        :key="t.label"
        class="cursor-pointer pb-2 text-[14px]"
        :class="
          activeTab === t.label
            ? 'border-b-2 border-primary text-[15px] font-bold text-primary'
            : 'text-ink-3'
        "
        @click="pickTab(t)"
      >
        {{ t.label }}
      </span>
    </div>

    <!-- 服务区域管理 -->
    <div v-if="activeTab === '服务区域'" class="flex flex-1 flex-col rounded-card bg-white">
      <div class="flex items-center border-b border-line px-5 py-4">
        <div>
          <div class="text-[15px] font-semibold text-ink-1">服务区域管理</div>
          <div class="mt-1 text-[12px] text-ink-3">
            共开通
            <span class="font-semibold text-primary">
              {{ settings ? formatNumber(settings.total_towns) : '—' }}
            </span>
            个乡镇 · 平台当前覆盖王家湾镇及周边 4 个乡镇
          </div>
        </div>
        <div class="ml-auto flex gap-3">
          <AppButton icon="download" @click="toast.info('导入区域')">导入区域</AppButton>
          <AppButton
            class="shadow-[0_4px_12px_rgba(255,107,44,.3)]"
            @click="openCreate"
          >
            + 新增服务区域
          </AppButton>
        </div>
      </div>

      <!-- 区域卡列表 -->
      <div v-if="loading" class="py-16 text-center text-[13px] text-ink-3">加载中…</div>
      <div v-else class="grid flex-1 grid-cols-2 content-start gap-3 overflow-y-auto p-5">
        <div
          v-for="row in settings?.areas ?? []"
          :key="row.id"
          class="rounded-[10px] border border-line p-[14px_16px]"
        >
          <div class="flex items-center gap-[10px]">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-[10px]"
              :style="{ background: row.icon_bg }"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" :stroke="row.icon_color" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 10c0 7-9 12-9 12s-9-5-9-12a9 9 0 0 1 18 0z" />
                <circle cx="12" cy="10" r="3" />
              </svg>
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2 text-[14px] font-semibold text-ink-1">
                {{ row.name }}
                <StatusTag :tone="areaStatusTone(row.status)">{{ row.status }}</StatusTag>
              </div>
              <div class="mt-[3px] text-[12px] text-ink-3">
                下辖 {{ row.villages }} 个村组 · 服务半径 {{ row.radius_km }}km
              </div>
            </div>
          </div>

          <div class="mt-3 grid grid-cols-3 gap-2">
            <div class="rounded-lg bg-head p-2 text-center">
              <div class="text-[16px] font-bold" :style="{ color: row.online_merchants === null ? '#C9CDD4' : '#1D2129' }">
                {{ row.online_merchants ?? '—' }}
              </div>
              <div class="mt-[2px] text-[11px] text-ink-3">在线商家</div>
            </div>
            <div class="rounded-lg bg-head p-2 text-center">
              <div class="text-[16px] font-bold" :style="{ color: row.active_riders === null ? '#C9CDD4' : '#1D2129' }">
                {{ row.active_riders ?? '—' }}
              </div>
              <div class="mt-[2px] text-[11px] text-ink-3">活跃骑手</div>
            </div>
            <div class="rounded-lg bg-head p-2 text-center">
              <div class="text-[16px] font-bold" :style="{ color: row.today_orders === null ? '#C9CDD4' : '#FF6B2C' }">
                {{ row.today_orders === null ? '—' : formatNumber(row.today_orders) }}
              </div>
              <div class="mt-[2px] text-[11px] text-ink-3">今日订单</div>
            </div>
          </div>

          <div class="mt-3 flex items-center gap-[10px] border-t border-[#F2F3F5] pt-2.5">
            <span v-if="row.notice" class="text-[12px] font-medium" style="color:#D48806;">{{ row.notice }}</span>
            <span v-else class="text-[12px] text-ink-2">配送费：起步 ¥{{ row.start_fee }} / {{ row.free_km }}km</span>
            <span class="ml-auto cursor-pointer text-[13px] font-semibold text-primary" @click="openEdit(row)">编辑</span>
            <span class="cursor-pointer text-[13px]" style="color:#FF4D4F;" @click="askToggle(row)">{{ row.action }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 邮箱 SMTP 服务配置 -->
    <SmtpConfig v-else-if="activeTab === '邮箱 SMTP 服务配置'" class="min-h-0 flex-1" />

    <!-- 短信服务配置 -->
    <SmsConfig v-else-if="activeTab === '短信服务配置'" class="min-h-0 flex-1" />

    <!-- 提现规则 / 账号权限：设计稿仅占位，无内容页 -->
    <div v-else class="flex flex-1 items-center justify-center rounded-card bg-white py-20">
      <Empty :text="`「${activeTab}」模块建设中，敬请期待`" />
    </div>

    <!-- 新增 / 编辑区域弹窗 -->
    <Modal
      v-model:open="formOpen"
      :title="editing ? '编辑服务区域' : '新增服务区域'"
      :confirm-text="editing ? '保存' : '新增'"
      :loading="submitting"
      :width="440"
      @confirm="onSubmit"
    >
      <div class="flex flex-col gap-3.5 py-1">
        <label class="flex items-center gap-2 text-[13px] text-ink-2">
          <span class="w-20 shrink-0 text-right">区域名称</span>
          <AppInput v-model="form.name" class="flex-1" placeholder="如：王家湾镇" />
        </label>
        <label class="flex items-center gap-2 text-[13px] text-ink-2">
          <span class="w-20 shrink-0 text-right">下辖村组</span>
          <AppInput v-model.number="form.villages" type="number" class="flex-1" />
        </label>
        <label class="flex items-center gap-2 text-[13px] text-ink-2">
          <span class="w-20 shrink-0 text-right">服务半径</span>
          <AppInput v-model.number="form.radius_km" type="number" class="flex-1" />
          <span class="text-[12px] text-ink-4">km</span>
        </label>
        <label class="flex items-center gap-2 text-[13px] text-ink-2">
          <span class="w-20 shrink-0 text-right">起步价</span>
          <AppInput v-model.number="form.start_fee" type="number" class="flex-1" />
          <span class="text-[12px] text-ink-4">元</span>
        </label>
      </div>
    </Modal>

    <!-- 停用 / 删除确认 -->
    <Confirm
      v-model:open="confirmOpen"
      :title="confirmTitle"
      :message="confirmMessage"
      :confirm-text="confirmAction?.row.action ?? '确认'"
      @confirm="onConfirmAction"
    />
  </div>
</template>
