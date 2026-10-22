<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import DataTable, { type Column } from '@/components/DataTable.vue'
import Modal from '@/components/Modal.vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import AppButton from '@/components/AppButton.vue'
import AppInput from '@/components/AppInput.vue'
import { useToast } from '@/composables/useToast'
import { formatDate } from '@/utils/format'
import { getFeeRules, saveFeeRule, type FeeRuleRow } from '@/api/settings'

const toast = useToast()

const loading = ref(false)
const rows = ref<FeeRuleRow[]>([])

const summary = computed<FeeRuleRow | undefined>(() => rows.value.find((r) => r.category === 'summary'))
const basicRows = computed(() => rows.value.filter((r) => r.category === 'basic'))
const ladderRows = computed(() => rows.value.filter((r) => r.category === 'ladder'))
const specialRows = computed(() => rows.value.filter((r) => r.category === 'special'))
const scopeRows = computed(() => rows.value.filter((r) => r.category === 'scope'))

async function load() {
  loading.value = true
  try {
    const res = await getFeeRules({ page: 1, pageSize: 100 })
    rows.value = res.list
  } finally {
    loading.value = false
  }
}
onMounted(load)

/* 启用开关切换 */
const enabled = ref(true)
const toggleSubmitting = ref(false)
async function toggleEnabled() {
  if (toggleSubmitting.value) return
  toggleSubmitting.value = true
  try {
    enabled.value = !enabled.value
    await saveFeeRule({ ...(summary.value ?? {}), enabled: enabled.value })
    toast.success(enabled.value ? '费率规则已启用' : '费率规则已停用')
  } finally {
    toggleSubmitting.value = false
  }
}

/* 编辑规则 */
const editOpen = ref(false)
const editSubmitting = ref(false)
const form = reactive({ startFee: 3, continueFee: 1, farFee: 1.5, cap: 25 })
function openEdit() {
  const basic = basicRows.value
  form.startFee = Number(basic[0]?.value?.replace(/[¥,/km]/g, '')) || 3
  form.continueFee = Number(basic[1]?.value?.replace(/[¥,/km]/g, '')) || 1
  form.farFee = Number(basic[2]?.value?.replace(/[¥,/km]/g, '')) || 1.5
  form.cap = Number(basic[3]?.value?.replace(/[¥,]/g, '')) || 25
  editOpen.value = true
}
async function onEditSubmit() {
  if (editSubmitting.value) return
  editSubmitting.value = true
  try {
    // 后端按单条 cms.FeeRule 更新（无批量/版本化契约），将表单值写回各基础费率行的 value
    const targetRows = [basicRows.value[0], basicRows.value[1], basicRows.value[2], basicRows.value[3]]
    const targetValues = [`${form.startFee} 元`, `${form.continueFee} 元/km`, `${form.farFee} 元/km`, `${form.cap} 元`]
    const targets: Array<[FeeRuleRow, string]> = []
    targetRows.forEach((row, i) => {
      if (row?.id) targets.push([row, targetValues[i]])
    })
    if (targets.length === 0) {
      toast.warning('暂无基础费率规则可编辑')
      return
    }
    await Promise.all(targets.map(([row, value]) => saveFeeRule({ ...row, value })))
    toast.success('基础费率规则已更新')
    editOpen.value = false
    await load()
  } finally {
    editSubmitting.value = false
  }
}

const ladderColumns: Column[] = [
  { key: 'name', title: '距离区间' },
  { key: 'rule', title: '计费规则' },
  { key: 'example', title: '示例订单' },
  { key: 'fee', title: '配送费' },
  { key: 'rider_share', title: '骑手分成', align: 'right' },
]
</script>

<template>
  <div class="flex flex-col gap-3">
    <!-- 费率状态条 -->
    <div class="flex shrink-0 items-center gap-3 rounded-card bg-white px-5 py-3">
      <div class="flex h-[34px] w-10 items-center justify-center rounded-[10px]" style="background:#FFF3E8;">
        <SvgIcon name="money" :size="22" :stroke-width="1.9" class="text-primary" />
      </div>
      <div>
        <div class="text-[15px] font-bold text-ink-1">配送费率规则</div>
        <div class="mt-[2px] text-[12px] text-ink-3">
          当前生效版本 {{ summary?.version ?? '—' }} ·
          {{ summary?.updated_at ? formatDate(summary.updated_at) : '—' }} 更新 ·
          下次自动生效 {{ summary?.next_effective_at ? formatDate(summary.next_effective_at) : '—' }}
        </div>
      </div>
      <div class="ml-auto flex items-center gap-3">
        <AppButton
          variant="ghost"
          class="!rounded-lg !px-3.5 !py-2 !bg-[#E8F7F0] !text-[#00B578]"
          :disabled="toggleSubmitting"
          @click="toggleEnabled"
        >
          <span class="relative h-[18px] w-[34px] rounded-full" :style="{ background: enabled ? '#00B578' : '#C9CDD4' }">
            <span
              class="absolute top-[2px] h-[14px] w-[14px] rounded-full bg-white transition-all"
              :style="enabled ? { right: '2px' } : { left: '2px' }"
            />
          </span>
          <span class="text-[12px] font-semibold" style="color:#00B578;">已启用</span>
        </AppButton>
        <AppButton
          class="!h-[30px] !px-[18px]"
          @click="openEdit"
        >
          编辑规则
        </AppButton>
      </div>
    </div>

    <!-- 基础配送费 -->
    <div class="shrink-0 rounded-card bg-white p-3">
      <div class="flex items-center">
        <div class="text-[15px] font-semibold text-ink-1">基础配送费</div>
        <span class="ml-3 text-[12px] text-ink-4">按距离阶梯计费，适用外卖 / 帮送 / 帮取</span>
      </div>
      <div class="mt-2 grid grid-cols-4 gap-3">
        <div v-for="item in basicRows" :key="item.id" class="rounded-[10px] bg-head p-3.5">
          <div class="text-[12px] text-ink-3">{{ item.name }}</div>
          <div
            class="mt-1 text-[22px] font-bold"
            :class="item.accent ? 'text-primary' : 'text-ink-1'"
          >
            {{ item.value }}
          </div>
          <div class="mt-1 text-[11px] text-ink-3">{{ item.desc }}</div>
        </div>
      </div>
    </div>

    <!-- 阶梯费率明细 -->
    <div class="shrink-0 rounded-card bg-white p-3">
      <div class="text-[15px] font-semibold text-ink-1">阶梯费率明细</div>
      <div class="mt-2 overflow-hidden rounded-[10px] border border-line">
        <DataTable :columns="ladderColumns" :rows="ladderRows" :loading="loading" row-key="id">
          <template #col-name="{ row }">
            <span class="font-semibold text-ink-1">{{ row.name }}</span>
          </template>
          <template #col-rule="{ row }">
            <span class="text-ink-2">{{ row.rule }}</span>
          </template>
          <template #col-example="{ row }">
            <span class="text-ink-2">{{ row.example }}</span>
          </template>
          <template #col-fee="{ row }">
            <span class="font-bold text-primary">{{ row.fee }}</span>
          </template>
          <template #col-rider_share="{ row }">
            <span class="text-ink-2">{{ row.rider_share }}</span>
          </template>
        </DataTable>
      </div>
    </div>

    <!-- 特殊规则 -->
    <div class="grid shrink-0 grid-cols-3 gap-3">
      <div v-for="item in specialRows" :key="item.id" class="rounded-card bg-white p-[18px]">
        <div class="flex items-center gap-2">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" :stroke="item.icon_color" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round">
            <path v-if="item.id === 'fee-special-1'" d="M4 10c0-5 16-5 16 0 0 3-4 4-4 8H8c0-4-4-5-4-8z" />
            <template v-else-if="item.id === 'fee-special-2'">
              <path d="M12 2L2 7l10 5 10-5-10-5z" />
              <polyline points="2 17 12 22 22 17" />
              <polyline points="2 12 12 17 22 12" />
            </template>
            <path v-else d="M20.38 3.46L16 2a4 4 0 0 1-8 0L3.62 3.46a2 2 0 0 0-1.34 2.23l.58 3.47a1 1 0 0 0 .99.84H6v10c0 1.1.9 2 2 2h8a2 2 0 0 0 2-2V10h2.15a1 1 0 0 0 .99-.84l.58-3.47a2 2 0 0 0-1.34-2.23z" />
          </svg>
          <span class="text-[14px] font-semibold text-ink-1">{{ item.name }}</span>
          <span class="ml-auto rounded-tag bg-[#E8F7F0] px-2 py-0.5 text-[11px] font-semibold" style="color:#00B578;">已启用</span>
        </div>
        <div class="mt-2 text-[12px] leading-[1.8] text-ink-3">{{ item.desc }}</div>
        <div class="mt-2 text-[22px] font-bold text-ink-1">{{ item.value }}</div>
      </div>
    </div>

    <!-- 规则适用范围 -->
    <div class="shrink-0 rounded-card bg-white p-3">
      <div class="text-[15px] font-semibold text-ink-1">规则适用范围</div>
      <div class="mt-2 grid grid-cols-4 gap-3">
        <div
          v-for="item in scopeRows"
          :key="item.id"
          class="rounded-[10px] p-3"
          :style="item.highlight ? 'background:#FFF7E6;border:1px solid #FFE7BA;' : 'background:#F7F8FA;'"
        >
          <div class="text-[12px]" :style="item.highlight ? 'color:#D48806;' : 'color:#86909C;'">{{ item.name }}</div>
          <div class="mt-1 text-[13px] font-semibold" :style="item.highlight ? 'color:#D48806;' : 'color:#1D2129;'">{{ item.value }}</div>
        </div>
      </div>
      <div class="mt-2 flex items-center gap-1.5 text-[12px] text-ink-4">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#C9CDD4" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10" />
          <path d="M12 16v-4M12 8h.01" />
        </svg>
        修改规则后需平台审核，次日 00:00 生效并全端同步
      </div>
    </div>

    <!-- 编辑规则弹窗 -->
    <Modal
      v-model:open="editOpen"
      title="编辑配送费率规则"
      confirm-text="提交审核"
      :loading="editSubmitting"
      :width="440"
      @confirm="onEditSubmit"
    >
      <div class="flex flex-col gap-3.5 py-1">
        <label class="flex items-center gap-2 text-[13px] text-ink-2">
          <span class="w-24 shrink-0 text-right">起步价（0-3km）</span>
          <AppInput v-model.number="form.startFee" type="number" class="flex-1" />
          <span class="text-[12px] text-ink-4">元</span>
        </label>
        <label class="flex items-center gap-2 text-[13px] text-ink-2">
          <span class="w-24 shrink-0 text-right">续程单价</span>
          <AppInput v-model.number="form.continueFee" type="number" class="flex-1" />
          <span class="text-[12px] text-ink-4">元/km</span>
        </label>
        <label class="flex items-center gap-2 text-[13px] text-ink-2">
          <span class="w-24 shrink-0 text-right">远途单价</span>
          <AppInput v-model.number="form.farFee" type="number" class="flex-1" />
          <span class="text-[12px] text-ink-4">元/km</span>
        </label>
        <label class="flex items-center gap-2 text-[13px] text-ink-2">
          <span class="w-24 shrink-0 text-right">配送费上限</span>
          <AppInput v-model.number="form.cap" type="number" class="flex-1" />
          <span class="text-[12px] text-ink-4">元</span>
        </label>
      </div>
    </Modal>
  </div>
</template>
