<script setup lang="ts" generic="T extends Record<string, unknown>">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import DataTable, { type Column } from '@/components/DataTable.vue'
import StatusTag from '@/components/StatusTag.vue'
import Pagination from '@/components/Pagination.vue'
import AppButton from '@/components/AppButton.vue'
import { useToast } from '@/composables/useToast'
import { formatNumber } from '@/utils/format'
import {
  getMerchants,
  getMerchantStats,
  type MerchantRow,
  type MerchantStats,
} from '@/api/merchants'

const router = useRouter()
const toast = useToast()

const stats = ref<MerchantStats | null>(null)
const rows = ref<MerchantRow[]>([])
const total = ref(0)
const loading = ref(false)

const query = ref({
  page: 1,
  pageSize: 10,
  keyword: '',
  status: '' as string,
  category: '' as string,
})

const statusTabs = [
  { label: '全部', value: '' },
  { label: '营业中', value: '营业中' },
  { label: '休息中', value: '休息中' },
  { label: '已封禁', value: '已封禁' },
]
const categoryTabs = ['餐饮', '商超', '药店', '其他']

/** logo 图标 path 映射（页面专属，内联 SVG stroke-width 2） */
const logoPaths: Record<MerchantRow['logoIcon'], string> = {
  food: '<path d="M3 11h18c0 5-4 8-9 8s-9-3-9-8z"/><path d="M8 11V7a4 4 0 0 1 8 0v4"/>',
  shop: '<path d="M4 10V4h16v6"/><path d="M2 10h20l-1.5 10h-17z"/><path d="M5 20v-6h14v6"/>',
  pharmacy: '<rect x="4" y="4" width="16" height="16" rx="3"/><path d="M12 8v8M8 12h8"/>',
  other:
    '<path d="M9 21h6"/><path d="M12 21V10"/><path d="M12 10c-2.5 0-4-2-4-4h8c0 2-1.5 4-4 4z"/><path d="M12 14c-2 0-3.5-1.5-3.5-3.5h7c0 2-1.5 3.5-3.5 3.5z"/>',
}

function statusTone(s: string): 'green' | 'gray' | 'red' {
  if (s === '营业中') return 'green'
  if (s === '休息中') return 'gray'
  return 'red'
}

const columns: Column[] = [
  { key: 'id', title: '商家 ID', width: '90px' },
  { key: 'info', title: '商家信息' },
  { key: 'contact', title: '联系人' },
  { key: 'category', title: '经营类型' },
  { key: 'monthlyOrders', title: '月订单量', width: '100px' },
  { key: 'rating', title: '评分', width: '80px', align: 'center' },
  { key: 'action', title: '操作', width: '120px', align: 'right' },
]

async function load() {
  loading.value = true
  try {
    const [list, stat] = await Promise.all([
      getMerchants({
        page: query.value.page,
        pageSize: query.value.pageSize,
        keyword: query.value.keyword || undefined,
        status: query.value.status || undefined,
      }),
      stats.value ? Promise.resolve(stats.value) : getMerchantStats(),
    ])
    rows.value = list.list
    total.value = list.total
    if (!stats.value) stats.value = stat
  } finally {
    loading.value = false
  }
}

function selectStatus(s: string) {
  query.value.status = s
  query.value.page = 1
  load()
}

function goDetail(id: string) {
  router.push(`/merchants/${id}`)
}

function onToggleStatus(row: MerchantRow) {
  if (row.status === '已封禁') {
    toast.info(`已申请解封 ${row.name}，等待复审`)
  } else {
    toast.info(`已对 ${row.name} 发起停业操作`)
  }
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-4">
    <!-- 指标行 -->
    <div class="grid grid-cols-4 gap-4">
      <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary-bg">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#FF6B2C" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" /><path d="M9 22V12h6v10" /></svg>
        </div>
        <div class="flex-1">
          <div class="text-xs text-ink-3">商家总数</div>
          <div class="mt-1 text-[22px] font-bold text-ink-1">{{ stats?.total ?? '-' }}</div>
        </div>
        <div class="text-[11px] font-semibold" style="color:#00B578;">+{{ stats?.weekNew ?? 0 }} 本周</div>
      </div>

      <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl" style="background:#E8F7F0;">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#00B578" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10" /><path d="M9 12l2 2 4-4" /></svg>
        </div>
        <div class="flex-1">
          <div class="text-xs text-ink-3">营业中</div>
          <div class="mt-1 text-[22px] font-bold text-ink-1">{{ stats?.open ?? '-' }}</div>
        </div>
        <div class="text-[11px] font-semibold" style="color:#00B578;">{{ stats?.openRate ?? '' }}</div>
      </div>

      <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl" style="background:#FFF8E0;">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#D48806" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="12" /><line x1="12" y1="16" x2="12.01" y2="16" /></svg>
        </div>
        <div class="flex-1">
          <div class="text-xs text-ink-3">待审核</div>
          <div class="mt-1 text-[22px] font-bold" style="color:#D48806;">{{ stats?.pending ?? '-' }}</div>
        </div>
        <div class="cursor-pointer text-[11px] font-semibold text-primary" @click="router.push('/merchants/audit')">去审核</div>
      </div>

      <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl" style="background:#FFE8E8;">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#FF4D4F" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10" /><path d="M4.9 4.9l14.2 14.2" /></svg>
        </div>
        <div class="flex-1">
          <div class="text-xs text-ink-3">已停业 / 封禁</div>
          <div class="mt-1 text-[22px] font-bold" style="color:#FF4D4F;">{{ stats?.closed ?? '-' }}</div>
        </div>
        <div class="text-[11px] text-ink-4">{{ stats?.closedDetail ?? '' }}</div>
      </div>
    </div>

    <!-- 标题行 -->
    <div class="flex items-center">
      <div class="text-[16px] font-bold text-ink-1">商家列表</div>
      <div class="ml-auto flex gap-3">
        <AppButton variant="secondary" icon="download">导出列表</AppButton>
        <AppButton class="shadow-[0_4px_12px_rgba(255,107,44,.3)]">+ 新增商家</AppButton>
      </div>
    </div>

    <!-- 筛选行 -->
    <div class="flex items-center gap-3">
      <span class="text-[13px] text-ink-2">状态</span>
      <AppButton
        v-for="t in statusTabs"
        :key="t.value"
        variant="text"
        class="!rounded-md !px-3.5 !py-1.5 !leading-normal"
        :class="query.status === t.value ? '!bg-primary-bg !font-semibold !text-primary' : '!bg-head !font-normal !text-ink-2 hover:!text-ink-1'"
        @click="selectStatus(t.value)"
      >
{{ t.label }}
</AppButton>
      <div class="h-6 w-px bg-line"></div>
      <span class="text-[13px] text-ink-2">经营类型</span>
      <AppButton
        v-for="c in categoryTabs"
        :key="c"
        variant="text"
        class="!rounded-md !px-3.5 !py-1.5 !leading-normal"
        :class="query.category === c ? '!bg-primary-bg !font-semibold !text-primary' : '!bg-head !font-normal !text-ink-2 hover:!text-ink-1'"
        @click="query.category = query.category === c ? '' : c"
      >
{{ c }}
</AppButton>
      <div class="ml-auto flex h-9 w-[220px] items-center gap-2 rounded-lg bg-head px-3">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#86909C" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v18h18" /><path d="M7 14l4-4 3 3 5-6" /></svg>
        <span class="text-[13px] text-ink-4">配送区域：全部</span>
        <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="#86909C" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9" /></svg>
      </div>
    </div>

    <!-- 表格 -->
    <div class="rounded-card bg-white">
      <DataTable :columns="columns" :rows="rows" :loading="loading" row-key="id">
        <template #col-info="{ row }">
          <div class="flex items-center gap-3">
            <div
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-[10px]"
              :style="{ background: row.logoBg }"
            >
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" :stroke="row.logoColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" v-html="logoPaths[row.logoIcon]" />
            </div>
            <div>
              <div class="flex items-center gap-2 text-[14px] font-semibold text-ink-1">
                {{ row.name }}
                <StatusTag :tone="statusTone(row.status)">{{ row.status }}</StatusTag>
              </div>
              <div class="mt-0.5 text-[12px] text-ink-3">{{ row.address }} · {{ row.addressSub }}</div>
            </div>
          </div>
        </template>
        <template #col-contact="{ row }">
          <div class="text-[13px] text-ink-2">
            {{ row.contact }}
            <div class="text-[12px] text-ink-4">{{ row.phone }}</div>
          </div>
        </template>
        <template #col-category="{ row }">
          <span
            class="inline-block rounded-tag px-2 py-0.5 text-[12px] font-semibold"
            :style="{ background: row.categoryBg, color: row.categoryColor }"
          >{{ row.category }}</span>
        </template>
        <template #col-monthlyOrders="{ row }">
          <span v-if="row.monthlyOrders === null" class="text-[13px]" style="color:#4E5969;">—</span>
          <span v-else class="text-[13px] text-ink-2">{{ formatNumber(row.monthlyOrders) }}</span>
        </template>
        <template #col-rating="{ row }">
          <span v-if="row.rating === null" class="text-[13px]" style="color:#C9CDD4;">—</span>
          <span v-else class="text-[13px] font-semibold" style="color:#D48806;">{{ row.rating }}</span>
        </template>
        <template #col-action="{ row }">
          <div class="flex items-center justify-end gap-3 text-[13px]">
            <span class="cursor-pointer font-semibold text-primary" @click="goDetail(row.id)">详情</span>
            <span
              v-if="row.status === '已封禁'"
              class="cursor-pointer text-primary"
              @click="onToggleStatus(row)"
            >解封</span>
            <span
              v-else
              class="cursor-pointer"
              style="color:#FF4D4F;"
              @click="onToggleStatus(row)"
            >停业</span>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- 分页 -->
    <div class="flex items-center">
      <span class="text-[13px] text-ink-3">共 {{ total }} 家 · 每页 {{ query.pageSize }} 家</span>
      <div class="ml-auto">
        <Pagination
          :page="query.page"
          :page-size="query.pageSize"
          :total="total"
          @change="(p: number) => { query.page = p; load() }"
        />
      </div>
    </div>
  </div>
</template>
