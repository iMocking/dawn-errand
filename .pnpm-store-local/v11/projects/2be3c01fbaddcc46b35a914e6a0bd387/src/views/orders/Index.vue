<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppButton from '@/components/AppButton.vue'
import DataTable, { type Column } from '@/components/DataTable.vue'
import StatusTag from '@/components/StatusTag.vue'
import Pagination from '@/components/Pagination.vue'
import SearchBar from '@/components/SearchBar.vue'
import { getOrders, type OrderRow } from '@/api/orders'
import { formatClock, formatNumber } from '@/utils/format'
import { useToast } from '@/composables/useToast'

type Tone = 'orange' | 'green' | 'red' | 'blue' | 'gray' | 'yellow'

/** 订单状态 → StatusTag tone（底色/字色照设计稿标签） */
const STATUS_TONE: Record<string, Tone> = {
  待支付: 'orange',
  待接单: 'red',
  备餐中: 'yellow',
  待取货: 'orange',
  配送中: 'yellow',
  已取件待送达: 'orange',
  已送达: 'blue',
  待评价: 'blue',
  已完成: 'green',
  已取消: 'gray',
  退款售后: 'red',
  已退款: 'gray',
  出餐超时: 'red',
  异常: 'red',
}

/** 订单类型 → 标签配色 */
const TYPE_TONE: Record<string, string> = {
  外卖: 'background:#FFF3E8;color:#FF6B2C;',
  帮买: 'background:#E8F0FF;color:#2B7FFF;',
  帮送: 'background:#E8F7F0;color:#00B578;',
  帮取: 'background:#FFF8E0;color:#D48806;',
  帮办: 'background:#F0E8FF;color:#722ED1;',
}

const TYPE_TABS = ['全部', '外卖', '帮买', '帮送', '帮取', '帮办']
const STATUS_TABS = [
  '全部',
  '待支付',
  '待接单',
  '备餐中',
  '待取货',
  '配送中',
  '已送达',
  '待评价',
  '已完成',
  '已取消',
  '退款售后',
  '已退款',
]

const toast = useToast()

const loading = ref(false)
const list = ref<OrderRow[]>([])
const total = ref(0)
const todayNew = ref(0)

const query = ref({ page: 1, pageSize: 10, keyword: '', status: '' })
const activeType = ref('全部')
const region = ref('')
const timeRange = ref('')

const columns: Column[] = [
  { key: 'id', title: '订单号', width: '150px' },
  { key: 'route', title: '订单内容' },
  { key: 'user_name', title: '用户 / 骑手' },
  { key: 'type', title: '类型', width: '90px' },
  { key: 'amount', title: '金额', width: '100px' },
  { key: 'status', title: '状态', width: '120px' },
  { key: 'created_at', title: '下单时间', width: '90px' },
  { key: 'action', title: '操作', width: '70px', align: 'right' },
]

function statusTone(s: string): Tone {
  return STATUS_TONE[s] ?? 'gray'
}

async function load() {
  loading.value = true
  try {
    const res = await getOrders({
      page: query.value.page,
      pageSize: query.value.pageSize,
      keyword: query.value.keyword || undefined,
      status: query.value.status || undefined,
    })
    list.value = res.list
    total.value = res.total
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  // 今日新增：拉全量后按创建时间统计当日数量（后端列表暂无日期参数）
  const res = await getOrders({ page: 1, pageSize: 100 })
  const start = new Date()
  start.setHours(0, 0, 0, 0)
  const sec = Math.floor(start.getTime() / 1000)
  todayNew.value = res.list.filter((r) => r.created_at >= sec).length
}

function onSearch(kw: string) {
  query.value.keyword = kw
  query.value.page = 1
  load()
}

function onReset() {
  query.value.keyword = ''
  query.value.status = ''
  query.value.page = 1
  activeType.value = '全部'
  region.value = ''
  timeRange.value = ''
  load()
}

function pickStatus(s: string) {
  query.value.status = s === '全部' ? '' : s
  query.value.page = 1
  load()
}

function pickType(t: string) {
  activeType.value = t
}

function onPage(p: number) {
  query.value.page = p
  load()
}

function onExport() {
  toast.info('导出任务已提交，完成后将在下载中心通知')
}

onMounted(() => {
  load()
  loadStats()
})
</script>

<template>
  <div class="flex flex-col gap-4">
    <!-- 标题行 -->
    <div class="flex items-center">
      <div>
        <div class="text-[18px] font-bold text-ink-1">订单列表</div>
        <div class="mt-1 text-[12px] text-ink-3">
          共 <span class="font-semibold text-primary">{{ formatNumber(total) }}</span> 条订单 ·
          今日新增 {{ todayNew }} 条
        </div>
      </div>
      <div class="ml-auto">
        <AppButton
          variant="secondary"
          icon="download"
          :icon-size="15"
          class="!gap-2"
          @click="onExport"
        >
          导出订单
        </AppButton>
      </div>
    </div>

    <!-- 筛选卡 -->
    <div class="flex flex-col gap-3.5 rounded-card bg-white p-4">
      <div class="flex flex-wrap items-center gap-3">
        <div class="flex items-center gap-2 text-[13px] text-ink-2">
          订单类型
          <span
            v-for="t in TYPE_TABS"
            :key="t"
            class="cursor-pointer rounded-md px-3.5 py-1.5 text-[13px]"
            :class="activeType === t ? 'bg-primary-bg font-semibold text-primary' : 'bg-head text-ink-2'"
            @click="pickType(t)"
          >
            {{ t }}
          </span>
        </div>
        <div class="h-6 w-px bg-line"></div>
        <div class="flex items-center gap-2 text-[13px] text-ink-2">
          订单状态
          <span
            v-for="s in STATUS_TABS"
            :key="s"
            class="cursor-pointer rounded-md px-3.5 py-1.5 text-[13px]"
            :class="query.status === (s === '全部' ? '' : s) ? 'bg-primary-bg font-semibold text-primary' : 'bg-head text-ink-2'"
            @click="pickStatus(s)"
          >
            {{ s }}
          </span>
        </div>
      </div>
      <SearchBar placeholder="输入订单号搜索" @search="onSearch" @reset="onReset">
        <template #extra>
          <select
            v-model="region"
            class="h-[34px] w-[200px] rounded-lg bg-head px-3 text-[13px] text-ink-3 outline-none"
          >
            <option value="">配送区域：全部</option>
            <option value="王家湾村">王家湾村</option>
            <option value="李庄村">李庄村</option>
            <option value="柳树沟村">柳树沟村</option>
            <option value="三岔口村">三岔口村</option>
          </select>
          <select
            v-model="timeRange"
            class="h-[34px] w-[220px] rounded-lg bg-head px-3 text-[13px] text-ink-3 outline-none"
          >
            <option value="">下单时间：近 7 天</option>
            <option value="today">今天</option>
            <option value="3d">近 3 天</option>
            <option value="30d">近 30 天</option>
          </select>
        </template>
      </SearchBar>
    </div>

    <!-- 表格 -->
    <div class="rounded-card bg-white">
      <DataTable :columns="columns" :rows="list" :loading="loading" row-key="id">
        <template #col-id="{ row }">
          <router-link :to="`/orders/${row.id}`" class="font-semibold" style="color:#2B7FFF">
            #{{ row.id }}
          </router-link>
        </template>
        <template #col-route="{ row }">
          <div class="max-w-[260px] truncate text-ink-2">{{ row.route }}</div>
        </template>
        <template #col-user_name="{ row }">
          <div class="text-ink-2">
            {{ row.user_name }}
            <div class="mt-0.5 text-[12px] text-ink-4">{{ row.rider }}</div>
          </div>
        </template>
        <template #col-type="{ row }">
          <span
            class="inline-block rounded-tag px-2 py-0.5 text-[12px] font-semibold"
            :style="TYPE_TONE[row.type]"
          >
            {{ row.type }}
          </span>
        </template>
        <template #col-amount="{ row }">
          <span class="font-semibold text-ink-1">¥{{ row.amount.toFixed(2) }}</span>
        </template>
        <template #col-status="{ row }">
          <StatusTag :tone="statusTone(row.status)">{{ row.status }}</StatusTag>
        </template>
        <template #col-created_at="{ row }">
          <span class="text-ink-3">{{ formatClock(row.created_at) }}</span>
        </template>
        <template #col-action="{ row }">
          <router-link :to="`/orders/${row.id}`" class="font-semibold" style="color:#FF6B2C">
            详情
          </router-link>
        </template>
      </DataTable>
    </div>

    <!-- 分页 -->
    <Pagination
      :page="query.page"
      :page-size="query.pageSize"
      :total="total"
      @change="onPage"
    />
  </div>
</template>
