<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import DataTable, { type Column } from '@/components/DataTable.vue'
import StatusTag from '@/components/StatusTag.vue'
import Pagination from '@/components/Pagination.vue'
import SearchBar from '@/components/SearchBar.vue'
import Modal from '@/components/Modal.vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { useToast } from '@/composables/useToast'
import { formatDate, formatNumber } from '@/utils/format'
import {
  getUsers,
  getUsersOverview,
  banUser,
  unbanUser,
  type UserRow,
  type UsersOverview,
} from '@/api/users'

const toast = useToast()

const loading = ref(false)
const list = ref<UserRow[]>([])
const total = ref(0)
const overview = ref<UsersOverview | null>(null)

const query = reactive({
  page: 1,
  pageSize: 10,
  keyword: '',
  status: '' as '' | '正常' | '已封禁',
})

const STATUS_TABS = [
  { label: '全部', value: '' },
  { label: '正常', value: '正常' },
  { label: '已封禁', value: '已封禁' },
] as const

const AREA_TABS = ['王家湾镇', '李庄村', '柳树沟村'] as const
const areaFilter = ref<(typeof AREA_TABS)[number] | ''>('')

/** 常驻区域为前端演示筛选（mock 中间件不支持该参数），仅作用于当前页数据 */
const displayRows = ref<UserRow[]>([])
function applyArea(rows: UserRow[]) {
  if (!areaFilter.value) return rows
  return rows.filter((r) => r.area.includes(areaFilter.value.replace('镇', '')))
}

async function load() {
  loading.value = true
  try {
    const res = await getUsers({
      page: query.page,
      pageSize: query.pageSize,
      keyword: query.keyword || undefined,
      status: query.status || undefined,
    })
    list.value = res.list
    displayRows.value = applyArea(res.list)
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function onSearch(kw: string) {
  query.keyword = kw
  query.page = 1
  load()
}
function onReset() {
  query.keyword = ''
  query.status = ''
  areaFilter.value = ''
  query.page = 1
  load()
}
function pickStatus(value: '' | '正常' | '已封禁') {
  query.status = value
  query.page = 1
  load()
}
function pickArea(area: (typeof AREA_TABS)[number]) {
  areaFilter.value = areaFilter.value === area ? '' : area
  displayRows.value = applyArea(list.value)
}
function onPage(p: number) {
  query.page = p
  load()
}

/* 封禁 / 解封 */
const confirmOpen = ref(false)
const submitting = ref(false)
const current = ref<UserRow | null>(null)

const confirmTitle = () => (current.value?.status === '已封禁' ? '解封用户' : '封禁用户')
const confirmMessage = () =>
  current.value?.status === '已封禁'
    ? `确认解封用户 ${current.value?.name}（${current.value?.id}）？解封后该账号可正常登录下单。`
    : `确认封禁用户 ${current.value?.name}（${current.value?.id}）？封禁后该账号将无法登录下单。`

function askBan(row: UserRow) {
  current.value = row
  confirmOpen.value = true
}
async function onConfirm() {
  if (!current.value) return
  submitting.value = true
  try {
    const willBan = current.value.status === '正常'
    if (willBan) {
      await banUser(current.value.id)
    } else {
      await unbanUser(current.value.id)
    }
    current.value.status = willBan ? '已封禁' : '正常'
    toast.success(willBan ? `已封禁用户 ${current.value.name}` : `已解封用户 ${current.value.name}`)
    confirmOpen.value = false
  } finally {
    submitting.value = false
  }
}

const columns: Column[] = [
  { key: 'id', title: '用户 ID', width: '110px' },
  { key: 'info', title: '用户信息' },
  { key: 'phone', title: '手机号', width: '150px' },
  { key: 'area', title: '常驻区域', width: '160px' },
  { key: 'orderCount', title: '累计下单', width: '100px' },
  { key: 'status', title: '状态', width: '90px' },
  { key: 'action', title: '操作', width: '110px', align: 'right' },
]

onMounted(async () => {
  overview.value = await getUsersOverview()
  await load()
})
</script>

<template>
  <div class="flex flex-col gap-4 px-6 py-5">
    <!-- 指标行 -->
    <div class="grid shrink-0 grid-cols-4 gap-4">
      <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl" style="background:#FFF3E8;">
          <SvgIcon name="user" :size="20" :stroke-width="1.9" style="color:#FF6B2C;" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="text-xs text-ink-3">注册用户</div>
          <div class="mt-1 text-[22px] font-bold leading-none text-ink-1">
            {{ overview ? formatNumber(overview.total) : '—' }}
          </div>
        </div>
        <div class="shrink-0 text-[11px] font-semibold" style="color:#00B578;">
          +{{ overview?.todayNew ?? 0 }} 今日
        </div>
      </div>

      <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl" style="background:#E8F7F0;">
          <SvgIcon name="check-circle" :size="20" :stroke-width="1.9" style="color:#00B578;" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="text-xs text-ink-3">本月活跃</div>
          <div class="mt-1 text-[22px] font-bold leading-none text-ink-1">
            {{ overview ? formatNumber(overview.activeMonth) : '—' }}
          </div>
        </div>
        <div class="shrink-0 text-[11px] font-semibold" style="color:#00B578;">
          {{ overview?.activeRate ?? 0 }}%
        </div>
      </div>

      <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl" style="background:#FFF8E0;">
          <SvgIcon name="clock" :size="20" :stroke-width="1.9" style="color:#D48806;" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="text-xs text-ink-3">人均下单次数</div>
          <div class="mt-1 text-[22px] font-bold leading-none text-ink-1">
            {{ overview?.avgOrders ?? '—' }}
          </div>
        </div>
        <div class="shrink-0 text-[11px] font-semibold" style="color:#00B578;">次 / 月</div>
      </div>

      <div class="flex items-center gap-3.5 rounded-card bg-white p-4">
        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl" style="background:#FFE8E8;">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#FF4D4F" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" />
            <path d="M4.9 4.9l14.2 14.2" />
          </svg>
        </div>
        <div class="min-w-0 flex-1">
          <div class="text-xs text-ink-3">已封禁</div>
          <div class="mt-1 text-[22px] font-bold leading-none" style="color:#FF4D4F;">
            {{ overview?.banned ?? '—' }}
          </div>
        </div>
        <div class="shrink-0 text-[11px] text-ink-4">近 30 天</div>
      </div>
    </div>

    <!-- 标题行 -->
    <div class="flex shrink-0 items-center">
      <div class="text-[16px] font-bold text-ink-1">用户列表</div>
      <div class="ml-auto flex gap-3">
        <button
          class="flex h-9 items-center gap-2 rounded-lg border border-line bg-white px-4 text-[13px] text-ink-2"
          @click="toast.info('导出列表')"
        >
          <SvgIcon name="download" :size="15" />
          导出列表
        </button>
      </div>
    </div>

    <!-- 筛选行 -->
    <div class="flex shrink-0 flex-wrap items-center gap-3">
      <SearchBar placeholder="搜索用户姓名 / 手机号" @search="onSearch" @reset="onReset">
        <template #extra>
          <span class="flex items-center gap-2 text-[13px] text-ink-2">
            状态
            <button
              v-for="t in STATUS_TABS"
              :key="t.value"
              class="rounded-md px-3.5 py-1.5 text-[13px]"
              :class="query.status === t.value ? 'bg-primary-bg font-semibold text-primary' : 'bg-head text-ink-2'"
              @click="pickStatus(t.value)"
            >
              {{ t.label }}
            </button>
          </span>
          <span class="h-6 w-px bg-line" />
          <span class="flex items-center gap-2 text-[13px] text-ink-2">
            常驻区域
            <button
              v-for="a in AREA_TABS"
              :key="a"
              class="rounded-md px-3.5 py-1.5 text-[13px]"
              :class="areaFilter === a ? 'bg-primary-bg font-semibold text-primary' : 'bg-head text-ink-2'"
              @click="pickArea(a)"
            >
              {{ a }}
            </button>
          </span>
        </template>
      </SearchBar>

      <div class="ml-auto flex h-[36px] w-60 items-center gap-2 rounded-lg bg-head px-3 text-[13px] text-ink-4">
        <SvgIcon name="clock" :size="15" class="shrink-0" />
        <span class="flex-1">注册时间：全部</span>
        <SvgIcon name="chevron-right" :size="10" class="shrink-0" style="transform:rotate(90deg);" />
      </div>
    </div>

    <!-- 表格 -->
    <div class="rounded-card bg-white">
      <DataTable :columns="columns" :rows="displayRows" :loading="loading" row-key="id">
        <template #col-info="{ row }">
          <div class="flex items-center gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-[15px] font-bold"
              :style="{ background: row.avatarBg, color: row.avatarColor }"
            >
              {{ row.name.charAt(0) }}
            </div>
            <div>
              <div class="text-[14px] font-semibold text-ink-1">{{ row.name }}</div>
              <div class="mt-0.5 text-[12px] text-ink-3">
                注册于 {{ formatDate(row.registeredAt) }}
              </div>
            </div>
          </div>
        </template>
        <template #col-orderCount="{ row }">
          <span class="text-[13px] font-semibold text-primary">{{ row.orderCount }}</span>
        </template>
        <template #col-status="{ row }">
          <StatusTag :tone="row.status === '已封禁' ? 'red' : 'green'">{{ row.status }}</StatusTag>
        </template>
        <template #col-action="{ row }">
          <div class="flex items-center justify-end gap-3 text-[13px]">
            <router-link :to="`/users/${row.id}`" class="font-semibold text-primary">详情</router-link>
            <span
              v-if="row.status === '正常'"
              v-permission="'user:ban'"
              class="cursor-pointer"
              style="color:#FF4D4F;"
              @click="askBan(row)"
            >
              封禁
            </span>
            <span
              v-else
              v-permission="'user:ban'"
              class="cursor-pointer text-primary"
              @click="askBan(row)"
            >
              解封
            </span>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- 分页 -->
    <div class="flex shrink-0 items-center text-[13px] text-ink-3">
      <span>共 {{ total }} 位用户 · 每页 {{ query.pageSize }} 位</span>
      <div class="ml-auto">
        <Pagination :page="query.page" :page-size="query.pageSize" :total="total" @change="onPage" />
      </div>
    </div>

    <!-- 封禁 / 解封确认 -->
    <Modal
      v-model:open="confirmOpen"
      :title="confirmTitle()"
      :confirm-text="current?.status === '已封禁' ? '确认解封' : '确认封禁'"
      :loading="submitting"
      @confirm="onConfirm"
    >
      <p class="text-sm text-ink-2">{{ confirmMessage() }}</p>
    </Modal>
  </div>
</template>
