<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import DataTable, { type Column } from '@/components/DataTable.vue'
import StatusTag from '@/components/StatusTag.vue'
import Pagination from '@/components/Pagination.vue'
import SearchBar from '@/components/SearchBar.vue'
import Confirm from '@/components/Confirm.vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import AppButton from '@/components/AppButton.vue'
import { useToast } from '@/composables/useToast'
import {
  getRiders,
  getRiderStats,
  freezeRider,
  type RiderRow,
  type RiderStatsDto,
} from '@/api/riders'

const router = useRouter()
const toast = useToast()

const STATUS_TABS = ['', '在线', '离线', '忙碌', '已冻结'] as const
const TOOL_TABS = ['', '电动车', '摩托车', '步行'] as const

const statusTone: Record<string, 'green' | 'yellow' | 'gray' | 'red'> = {
  在线: 'green',
  忙碌: 'yellow',
  离线: 'gray',
  已冻结: 'red',
}

const query = reactive({ page: 1, pageSize: 10, keyword: '', status: '' as string })
const vehicle = ref('')

const list = ref<RiderRow[]>([])
const total = ref(0)
const loading = ref(false)
const stats = ref<RiderStatsDto | null>(null)

const columns: Column[] = [
  { key: 'id', title: '骑手 ID', width: '90px' },
  { key: 'info', title: '骑手信息' },
  { key: 'phone', title: '联系方式', width: '140px' },
  { key: 'area', title: '服务区域', width: '180px' },
  { key: 'todayOrders', title: '今日单量', width: '90px', align: 'center' },
  { key: 'rating', title: '评分', width: '80px', align: 'center' },
  { key: 'action', title: '操作', width: '120px', align: 'right' },
]

/** 配送工具为本地筛选（中间件仅支持 status/keyword），作用于当前页行 */
const rows = computed(() => {
  if (!vehicle.value) return list.value
  return list.value.filter((r) => r.tool === vehicle.value)
})

async function load() {
  loading.value = true
  try {
    const res = await getRiders({
      page: query.page,
      pageSize: query.pageSize,
      keyword: query.keyword || undefined,
      status: query.status || undefined,
    })
    list.value = res.list
    total.value = res.total
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  stats.value = await getRiderStats()
}

function onSearch(kw: string) {
  query.keyword = kw
  query.page = 1
  load()
}
function onReset() {
  query.keyword = ''
  query.status = ''
  vehicle.value = ''
  query.page = 1
  load()
}
function pickStatus(s: string) {
  query.status = s
  query.page = 1
  load()
}
function pickTool(t: string) {
  vehicle.value = t
}
function goPage(p: number) {
  query.page = p
  load()
}

/** 冻结 / 解冻（防重复提交） */
const freezeTarget = ref<RiderRow | null>(null)
const freezeOpen = ref(false)
const freezing = ref(false)
function askFreeze(row: RiderRow) {
  freezeTarget.value = row
  freezeOpen.value = true
}
async function onFreeze() {
  const row = freezeTarget.value
  if (!row || freezing.value) return
  const willFreeze = row.status !== '已冻结'
  freezing.value = true
  try {
    await freezeRider(row.id, willFreeze ? 'freeze' : 'unfreeze')
    row.status = willFreeze ? '已冻结' : '在线'
    toast.success(willFreeze ? `已冻结骑手 ${row.name}` : `已解冻骑手 ${row.name}`)
    freezeOpen.value = false
  } finally {
    freezing.value = false
  }
}

onMounted(() => {
  load()
  loadStats()
})
</script>

<template>
  <div class="flex flex-col gap-4">
    <!-- 指标行 -->
    <div class="grid grid-cols-4 gap-4">
      <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
        <div class="flex h-[46px] w-[46px] shrink-0 items-center justify-center rounded-xl" style="background:#E8F0FF;">
          <SvgIcon name="grid" :size="22" :stroke-width="1.9" style="color:#2B7FFF;" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="text-xs text-ink-3">骑手总数</div>
          <div class="mt-1 text-2xl font-bold text-ink-1">{{ stats?.total ?? '—' }}</div>
        </div>
        <div class="shrink-0 text-[11px] font-semibold" style="color:#00B578;">{{ stats?.weekDelta }}</div>
      </div>

      <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
        <div class="flex h-[46px] w-[46px] shrink-0 items-center justify-center rounded-xl" style="background:#E8F7F0;">
          <SvgIcon name="check-circle" :size="22" :stroke-width="1.9" style="color:#00B578;" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="text-xs text-ink-3">在线中</div>
          <div class="mt-1 text-2xl font-bold text-ink-1">{{ stats?.online ?? '—' }}</div>
        </div>
        <div class="shrink-0 text-[11px] font-semibold" style="color:#D48806;">在线率 {{ stats?.onlineRate }}</div>
      </div>

      <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
        <div class="flex h-[46px] w-[46px] shrink-0 items-center justify-center rounded-xl" style="background:#FFF8E0;">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#D48806" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="12" /><line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
        </div>
        <div class="min-w-0 flex-1">
          <div class="text-xs text-ink-3">待审核入驻</div>
          <div class="mt-1 text-2xl font-bold" style="color:#D48806;">{{ stats?.pendingAudit ?? '—' }}</div>
        </div>
        <div class="shrink-0 cursor-pointer text-[11px] font-semibold" style="color:#FF6B2C;" @click="router.push('/riders/audit')">去审核</div>
      </div>

      <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
        <div class="flex h-[46px] w-[46px] shrink-0 items-center justify-center rounded-xl" style="background:#FFE8E8;">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#FF4D4F" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" /><path d="M4.9 4.9l14.2 14.2" />
          </svg>
        </div>
        <div class="min-w-0 flex-1">
          <div class="text-xs text-ink-3">已冻结 / 封禁</div>
          <div class="mt-1 text-2xl font-bold" style="color:#FF4D4F;">{{ stats?.frozen ?? '—' }}</div>
        </div>
        <div class="shrink-0 text-[11px] text-ink-4">{{ stats?.frozenNote }}</div>
      </div>
    </div>

    <!-- 标题行 -->
    <div class="flex items-center">
      <div class="text-[16px] font-bold text-ink-1">骑手列表</div>
      <div class="ml-auto flex gap-3">
        <AppButton variant="secondary" icon="download">导出列表</AppButton>
        <AppButton variant="primary" icon="plus" class="shadow-[0_4px_12px_rgba(255,107,44,.3)]">新增骑手</AppButton>
      </div>
    </div>

    <!-- 筛选行 -->
    <div class="flex flex-wrap items-center gap-3">
      <div class="flex items-center gap-2">
        <span class="text-[13px] text-ink-2">状态</span>
        <span
          v-for="s in STATUS_TABS"
          :key="s || '全部'"
          class="cursor-pointer rounded-md px-3.5 py-1.5 text-[13px]"
          :class="query.status === s ? 'font-semibold' : 'bg-head text-ink-2'"
          :style="query.status === s ? 'background:#FFF3E8;color:#FF6B2C;' : ''"
          @click="pickStatus(s)"
        >{{ s || '全部' }}</span>
      </div>
      <div class="h-6 w-px bg-line" />
      <div class="flex items-center gap-2">
        <span class="text-[13px] text-ink-2">配送工具</span>
        <span
          v-for="t in TOOL_TABS"
          :key="t || '全部'"
          class="cursor-pointer rounded-md bg-head px-3.5 py-1.5 text-[13px] text-ink-2"
          :class="vehicle === t ? 'font-semibold text-primary' : ''"
          @click="pickTool(t)"
        >{{ t || '全部' }}</span>
      </div>
      <div class="ml-auto">
        <SearchBar placeholder="搜索骑手姓名 / 手机号" @search="onSearch" @reset="onReset" />
      </div>
    </div>

    <!-- 表格卡片 -->
    <div class="rounded-card bg-white">
      <DataTable :columns="columns" :rows="rows" :loading="loading" row-key="id">
        <template #col-id="{ row }">
          <span class="text-ink-3">{{ (row as RiderRow).id }}</span>
        </template>
        <template #col-info="{ row }">
          <div class="flex items-center gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-[15px] font-bold"
              :style="{ background: (row as RiderRow).avatarBg, color: (row as RiderRow).avatarColor }"
            >
{{ (row as RiderRow).name.charAt(0) }}
</div>
            <div>
              <div class="flex items-center gap-2 text-[14px] font-semibold text-ink-1">
                {{ (row as RiderRow).name }}
                <StatusTag :tone="statusTone[(row as RiderRow).status]">{{ (row as RiderRow).status }}</StatusTag>
              </div>
              <div class="mt-[3px] text-[12px] text-ink-3">{{ (row as RiderRow).tool }} · 等级 {{ (row as RiderRow).level }}</div>
            </div>
          </div>
        </template>
        <template #col-todayOrders="{ row }">
          <span v-if="(row as RiderRow).todayOrders != null" class="font-semibold" style="color:#FF6B2C;">{{ (row as RiderRow).todayOrders }}</span>
          <span v-else class="text-ink-4">—</span>
        </template>
        <template #col-rating="{ row }">
          <span v-if="(row as RiderRow).rating != null" class="font-semibold" style="color:#D48806;">{{ (row as RiderRow).rating }}</span>
          <span v-else class="text-ink-4">—</span>
        </template>
        <template #col-action="{ row }">
          <div class="flex justify-end gap-3 text-[13px]">
            <router-link class="font-semibold text-primary" :to="`/riders/${(row as RiderRow).id}`">详情</router-link>
            <span
              v-if="(row as RiderRow).status === '已冻结'"
              class="cursor-pointer font-semibold text-primary"
              @click="askFreeze(row as RiderRow)"
            >解冻</span>
            <span
              v-else
              class="cursor-pointer"
              style="color:#FF4D4F;"
              @click="askFreeze(row as RiderRow)"
            >冻结</span>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- 分页 -->
    <div class="flex items-center text-[13px] text-ink-3">
      <span>共 {{ stats?.total ?? total }} 位骑手 · 每页 {{ query.pageSize }} 位</span>
      <div class="ml-auto">
        <Pagination :page="query.page" :page-size="query.pageSize" :total="total" @change="goPage" />
      </div>
    </div>

    <Confirm
      :open="freezeOpen"
      :title="freezeTarget?.status === '已冻结' ? '解冻骑手' : '冻结骑手'"
      :message="`确认${freezeTarget?.status === '已冻结' ? '解冻' : '冻结'}骑手 ${freezeTarget?.name}（${freezeTarget?.id}）吗？`"
      :confirm-text="freezeTarget?.status === '已冻结' ? '确认解冻' : '确认冻结'"
      @update:open="(v) => (freezeOpen = v)"
      @confirm="onFreeze"
    />
  </div>
</template>
