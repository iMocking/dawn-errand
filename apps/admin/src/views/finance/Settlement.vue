<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  getFinance,
  getSettlements,
  type FinanceDto,
  type SettlementRow,
} from '@/api/finance'
import AppButton from '@/components/AppButton.vue'
import AppInput from '@/components/AppInput.vue'
import DataTable, { type Column } from '@/components/DataTable.vue'
import Pagination from '@/components/Pagination.vue'
import { formatClock, formatDate, formatMoney } from '@/utils/format'
import { useToast } from '@/composables/useToast'

const toast = useToast()

const stats = ref<FinanceDto['settlementStats'] | null>(null)
const rows = ref<SettlementRow[]>([])
const loading = ref(true)

const TYPES = ['全部', '商家结算', '骑手结算', '用户退款'] as const
const activeType = ref<(typeof TYPES)[number]>('全部')
const keyword = ref('')
const page = ref(1)
const pageSize = 10

const filtered = computed(() => {
  let list = rows.value
  if (activeType.value !== '全部') list = list.filter((r) => r.type === activeType.value)
  const kw = keyword.value.trim().toLowerCase()
  if (kw) list = list.filter((r) => r.id.toLowerCase().includes(kw) || r.party.toLowerCase().includes(kw))
  return list
})

const paged = computed(() => filtered.value.slice((page.value - 1) * pageSize, page.value * pageSize))

const columns: Column[] = [
  { key: 'id', title: '流水号', width: '150px' },
  { key: 'party', title: '结算对象' },
  { key: 'type', title: '类型', width: '100px' },
  { key: 'amount', title: '金额', width: '120px' },
  { key: 'status', title: '状态', width: '90px' },
  { key: 'period', title: '结算周期', width: '150px' },
  { key: 'settled_at', title: '结算时间', width: '130px' },
  { key: 'action', title: '操作', align: 'right' },
]

function shortTime(sec: number) {
  return `${formatDate(sec).slice(5)} ${formatClock(sec)}`
}

function onTypeChange(t: (typeof TYPES)[number]) {
  activeType.value = t
  page.value = 1
}

function onSearch(kw: string) {
  keyword.value = kw
  page.value = 1
}

onMounted(async () => {
  try {
    const [finance, result] = await Promise.all([getFinance(), getSettlements({ page: 1, pageSize: 50 })])
    stats.value = finance.settlementStats
    rows.value = result.list
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="flex flex-col gap-4">
    <div v-if="loading" class="py-16 text-center text-ink-3">加载中…</div>
    <template v-else-if="stats">
      <!-- 汇总指标卡 -->
      <div class="grid shrink-0 grid-cols-4 gap-4">
        <div class="rounded-card bg-white px-[18px] py-4">
          <div class="text-[12px] text-ink-3">今日结算</div>
          <div class="mt-1.5 text-[24px] font-bold text-ink-1">{{ formatMoney(stats.todaySettled) }}</div>
          <div class="mt-1 text-[11px]" style="color:#00B578;">{{ stats.todaySettledDelta }}</div>
        </div>
        <div class="rounded-card bg-white px-[18px] py-4">
          <div class="text-[12px] text-ink-3">本周累计</div>
          <div class="mt-1.5 text-[24px] font-bold text-ink-1">{{ formatMoney(stats.weekTotal) }}</div>
          <div class="mt-1 text-[11px] text-ink-3">{{ stats.weekBreakdown }}</div>
        </div>
        <div class="rounded-card bg-white px-[18px] py-4">
          <div class="text-[12px] text-ink-3">本月累计</div>
          <div class="mt-1.5 text-[24px] font-bold text-ink-1">{{ formatMoney(stats.monthTotal) }}</div>
          <div class="mt-1 text-[11px] text-ink-3">{{ stats.monthBreakdown }}</div>
        </div>
        <div class="rounded-card bg-white px-[18px] py-4">
          <div class="text-[12px] text-ink-3">待结算金额</div>
          <div class="mt-1.5 text-[24px] font-bold" style="color:#D48806;">{{ formatMoney(stats.pendingAmount) }}</div>
          <AppButton variant="text" class="mt-1 cursor-pointer !text-[11px]">{{ stats.pendingCountText }}</AppButton>
        </div>
      </div>

      <!-- 筛选条 -->
      <div class="flex shrink-0 items-center gap-3 rounded-card bg-white px-[18px] py-[14px]">
        <span class="shrink-0 text-[12px] text-ink-3">结算类型</span>
        <AppButton
          v-for="t in TYPES"
          :key="t"
          variant="ghost"
          class="!rounded-md !px-3 !py-[5px] !text-[12px]"
          :class="
            activeType === t
              ? '!bg-[#FFF3E8] !font-semibold !text-[#FF6B2C] hover:!text-[#FF6B2C]'
              : '!bg-[#F7F8FA] !text-ink-2 hover:!text-ink-2'
          "
          @click="onTypeChange(t)"
        >
          {{ t }}
        </AppButton>
        <div class="ml-auto flex items-center gap-2">
          <div class="flex h-8 items-center gap-2 rounded-lg bg-head px-3 text-[12px] text-ink-2">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#86909C" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="4" width="18" height="18" rx="2" />
              <line x1="16" y1="2" x2="16" y2="6" />
              <line x1="8" y1="2" x2="8" y2="6" />
              <line x1="3" y1="10" x2="21" y2="10" />
            </svg>
            {{ stats.dateRange }}
          </div>
          <AppInput
            v-model="keyword"
            size="sm"
            variant="plain"
            prefix="search"
            placeholder="流水号搜索"
            class="w-[126px]"
            input-class="!text-[12px] !text-ink-2"
            @keyup.enter="onSearch(keyword)"
          />
          <AppButton
            variant="primary"
            class="!h-8 !text-[12px]"
            @click="toast.info('对账单导出任务已提交')"
          >
            导出对账单
          </AppButton>
        </div>
      </div>

      <!-- 流水表格 -->
      <div class="flex flex-col gap-4">
        <div class="overflow-hidden rounded-card bg-white">
          <DataTable :columns="columns" :rows="paged" row-key="id">
            <template #col-id="{ row }">
              <span class="font-semibold" style="color:#2B7FFF;">{{ row.id }}</span>
            </template>
            <template #col-type="{ row }">
              <span class="rounded-tag px-2 py-0.5 text-[12px] font-semibold" :style="{ background: row.typeBg, color: row.typeColor }">{{ row.type }}</span>
            </template>
            <template #col-amount="{ row }">
              <span class="font-bold text-ink-1">{{ formatMoney(row.amount) }}</span>
            </template>
            <template #col-status="{ row }">
              <span class="font-semibold" :style="row.status === '已到账' ? 'color:#00B578;' : 'color:#D48806;'">{{ row.status }}</span>
            </template>
            <template #col-settled_at="{ row }">
              <span class="text-ink-3">{{ shortTime(row.settled_at) }}</span>
            </template>
            <template #col-action="{ row }">
              <AppButton
                variant="text"
                class="!text-[12px] !text-[#2B7FFF]"
                @click="toast.info(`${row.id} 明细查看功能待接入`)"
              >
                {{ row.action }}
              </AppButton>
            </template>
          </DataTable>
        </div>
        <div class="flex items-center justify-between rounded-card bg-white px-[18px] py-2.5">
          <span class="text-[12px] text-ink-3">共 {{ filtered.length }} 条记录 · {{ stats.exportedText }}</span>
          <Pagination :page="page" :page-size="pageSize" :total="filtered.length" @change="page = $event" />
        </div>
      </div>
    </template>
  </div>
</template>
