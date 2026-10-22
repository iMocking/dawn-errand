<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import DataTable, { type Column } from '@/components/DataTable.vue'
import StatusTag from '@/components/StatusTag.vue'
import Pagination from '@/components/Pagination.vue'
import SearchBar from '@/components/SearchBar.vue'
import Modal from '@/components/Modal.vue'
import Confirm from '@/components/Confirm.vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import AppButton from '@/components/AppButton.vue'
import AppInput from '@/components/AppInput.vue'
import { useToast } from '@/composables/useToast'
import { formatTime } from '@/utils/format'
import {
  getAnnouncements,
  saveAnnouncement,
  type AnnouncementRow,
} from '@/api/settings'

const toast = useToast()

const loading = ref(false)
const allRows = ref<AnnouncementRow[]>([])
const keyword = ref('')
const page = ref(1)
const pageSize = 10

const filtered = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return allRows.value
  return allRows.value.filter((r) => r.title.toLowerCase().includes(kw))
})
const pagedRows = computed(() => filtered.value.slice((page.value - 1) * pageSize, page.value * pageSize))
const publishedCount = computed(() => allRows.value.filter((r) => r.status === '已发布').length)

/** 右侧预览默认选中「双十一乡村集市配送补贴活动」（设计稿） */
const activeId = ref('ann-002')
const active = computed<AnnouncementRow>(
  () => allRows.value.find((r) => r.id === activeId.value) ?? allRows.value[0],
)

async function load() {
  loading.value = true
  try {
    const res = await getAnnouncements({ page: 1, pageSize: 100 })
    allRows.value = res.list
  } finally {
    loading.value = false
  }
}
onMounted(load)

function onSearch(kw: string) {
  keyword.value = kw
  page.value = 1
}
function onReset() {
  keyword.value = ''
  page.value = 1
}

const columns: Column[] = [
  { key: 'title', title: '标题' },
  { key: 'type', title: '类型', width: '90px' },
  { key: 'audience', title: '发布对象', width: '130px' },
  { key: 'status', title: '状态', width: '90px' },
  { key: 'published_at', title: '发布时间', width: '150px' },
  { key: 'action', title: '操作', width: '120px', align: 'right' },
]

function typeStyle(tone: string) {
  const map: Record<string, { bg: string; color: string }> = {
    orange: { bg: '#FFF3E8', color: '#FF6B2C' },
    blue: { bg: '#E8F4FF', color: '#2B7FFF' },
    yellow: { bg: '#FFF7E6', color: '#D48806' },
    purple: { bg: '#F3F0FF', color: '#722ED1' },
  }
  return map[tone] ?? map.orange
}
function statusTone(status: string) {
  if (status === '已发布') return 'green'
  if (status === '审核中') return 'yellow'
  return 'gray'
}

/* 发布 / 编辑公告弹窗 */
const formOpen = ref(false)
const submitting = ref(false)
const editingId = ref<string | null>(null)
const form = reactive({ title: '', type: '活动', audience: '全部用户', content: '' })
const FORM_TYPES = ['费率', '活动', '提醒', '协议']
const FORM_AUDIENCES = ['全部用户', '商家 / 骑手', '骑手', '商家', '新入驻商家', '骑手（12 人）']

function openCreate() {
  editingId.value = null
  form.title = ''
  form.type = '活动'
  form.audience = '全部用户'
  form.content = ''
  formOpen.value = true
}
function openEdit(row: AnnouncementRow) {
  editingId.value = row.id
  form.title = row.title
  form.type = row.type
  form.audience = row.audience
  form.content = row.content
  formOpen.value = true
}
async function onSubmit() {
  if (submitting.value) return
  if (!form.title.trim()) {
    toast.warning('请输入公告标题')
    return
  }
  submitting.value = true
  try {
    await saveAnnouncement({
      id: editingId.value,
      action: editingId.value ? 'update' : 'create',
      title: form.title.trim(),
      type: form.type,
      audience: form.audience,
      content: form.content,
    })
    toast.success(editingId.value ? '公告已保存，提交审核' : '公告已发布')
    formOpen.value = false
  } finally {
    submitting.value = false
  }
}

/* 下线 / 发布 / 撤回 */
const confirmOpen = ref(false)
const confirmTitle = ref('')
const confirmMessage = ref('')
const confirmRow = ref<AnnouncementRow | null>(null)
const actionSubmitting = ref(false)

function askAction(row: AnnouncementRow) {
  confirmRow.value = row
  confirmTitle.value = row.action_name
  if (row.action_name === '下线') {
    confirmMessage.value = `确认下线公告「${row.title}」？下线后用户端将不再展示。`
  } else if (row.action_name === '发布') {
    confirmMessage.value = `确认发布公告「${row.title}」？发布后即时推送至${row.audience}。`
  } else {
    confirmMessage.value = `确认撤回公告「${row.title}」？撤回后回到草稿状态。`
  }
  confirmOpen.value = true
}
async function onConfirmAction() {
  if (!confirmRow.value || actionSubmitting.value) return
  actionSubmitting.value = true
  try {
    const row = confirmRow.value
    await saveAnnouncement({ action: row.action_name, id: row.id })
    toast.success(`已${row.action_name}公告「${row.title}」`)
    confirmOpen.value = false
  } finally {
    actionSubmitting.value = false
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <!-- 顶部操作条 -->
    <div class="flex shrink-0 items-center">
      <div class="flex items-center gap-[10px]">
        <span class="text-[16px] font-bold text-ink-1">公告管理</span>
        <span class="text-[12px] text-ink-3">
          共 {{ allRows.length }} 条公告 · 生效中 {{ publishedCount }} 条
        </span>
      </div>
      <div class="ml-auto flex items-center gap-3">
        <SearchBar placeholder="搜索公告标题" @search="onSearch" @reset="onReset" />
        <AppButton
          icon="plus"
          :icon-size="14"
          class="!h-[34px] !px-[18px]"
          @click="openCreate"
        >
          发布公告
        </AppButton>
      </div>
    </div>

    <!-- 主体：左列表 + 右预览 -->
    <div class="flex min-h-0 flex-1 gap-4">
      <!-- 公告列表 -->
      <div class="flex min-w-0 flex-[1.3] flex-col gap-3">
        <div class="overflow-hidden rounded-card bg-white">
          <DataTable :columns="columns" :rows="pagedRows" :loading="loading" row-key="id">
            <template #col-title="{ row }">
              <span
                class="cursor-pointer font-semibold text-ink-1"
                :class="row.id === activeId ? 'text-primary' : ''"
                @click="activeId = row.id"
              >
                {{ row.title }}
              </span>
            </template>
            <template #col-type="{ row }">
              <span
                class="inline-block rounded-tag px-2 py-0.5 text-[12px] font-semibold"
                :style="{ background: typeStyle(row.type_tone).bg, color: typeStyle(row.type_tone).color }"
              >
                {{ row.type }}
              </span>
            </template>
            <template #col-audience="{ row }">
              <span class="text-ink-2">{{ row.audience }}</span>
            </template>
            <template #col-status="{ row }">
              <StatusTag :tone="statusTone(row.status)">{{ row.status }}</StatusTag>
            </template>
            <template #col-published_at="{ row }">
              <span class="text-ink-3">
                {{ row.published_at ? formatTime(row.published_at) : row.planned_text }}
              </span>
            </template>
            <template #col-action="{ row }">
              <span class="flex items-center justify-end gap-1.5 text-[12px] font-semibold" style="color:#2B7FFF;">
                <span class="cursor-pointer" @click="openEdit(row)">{{ row.status === '审核中' ? '查看' : '编辑' }}</span>
                <span class="text-ink-4">/</span>
                <span class="cursor-pointer" @click="askAction(row)">{{ row.action_name }}</span>
              </span>
            </template>
          </DataTable>
        </div>

        <div class="flex shrink-0 items-center rounded-card bg-white px-[18px] py-2.5 text-[12px] text-ink-3">
          <span>共 {{ filtered.length }} 条公告</span>
          <div class="ml-auto">
            <Pagination :page="page" :page-size="pageSize" :total="filtered.length" @change="(p: number) => (page = p)" />
          </div>
        </div>
      </div>

      <!-- 右侧：预览 + 推送统计 -->
      <div class="flex min-w-0 flex-1 flex-col gap-4 overflow-y-auto pr-0.5">
        <div class="shrink-0 rounded-card bg-white p-5">
          <div class="flex items-center">
            <div class="text-[15px] font-semibold text-ink-1">公告预览</div>
            <span class="ml-auto max-w-[220px] truncate text-[12px] text-ink-3">{{ active?.title }}</span>
          </div>
          <div class="mt-4 rounded-[10px] border border-line bg-head p-[18px]">
            <div class="flex items-center gap-2.5">
              <div
                class="flex h-9 w-9 shrink-0 items-center justify-center rounded-[9px]"
                :style="{ background: 'linear-gradient(135deg,#FF8A3D,#FF6B2C)' }"
              >
                <SvgIcon name="bell" :size="20" :stroke-width="1.9" class="text-white" />
              </div>
              <div>
                <div class="text-[13px] font-bold text-ink-1">曙光跑腿 · 平台公告</div>
                <div class="text-[11px] text-ink-4">
                  {{ active?.published_at ? formatTime(active.published_at) : active?.planned_text }} 发布 ·
                  {{ active?.audience }}
                </div>
              </div>
            </div>
            <div class="mt-4 text-[15px] font-bold text-ink-1">{{ active?.title }}</div>
            <div class="mt-2 text-[13px] leading-[1.9] text-ink-2">{{ active?.content }}</div>
            <div class="mt-3.5 flex flex-wrap gap-2">
              <span
                v-for="tag in active?.tags ?? []"
                :key="tag"
                class="rounded-[6px] border border-line bg-white px-2.5 py-1 text-[12px] text-ink-2"
              >
                {{ tag }}
              </span>
            </div>
          </div>
          <div class="mt-4 flex justify-end gap-2.5">
            <AppButton
              variant="secondary"
              class="!h-[34px] !text-[12px]"
              @click="toast.info('App 端预览')"
            >
              预览 App 端
            </AppButton>
            <AppButton
              class="!h-[34px] !text-[12px]"
              @click="active && openEdit(active)"
            >
              编辑公告
            </AppButton>
          </div>
        </div>

        <div class="shrink-0 rounded-card bg-white p-5">
          <div class="text-[15px] font-semibold text-ink-1">推送统计</div>
          <div class="mt-3.5 flex flex-col gap-2.5 text-[13px]">
            <div
              v-for="s in active?.stats ?? []"
              :key="s.label"
              class="flex items-center justify-between rounded-lg bg-head px-3 py-2.5"
            >
              <span class="text-ink-3">{{ s.label }}</span>
              <span class="font-semibold" :style="s.highlight ? 'color:#FF6B2C;' : 'color:#1D2129;'">{{ s.value }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 发布 / 编辑公告弹窗 -->
    <Modal
      v-model:open="formOpen"
      :title="editingId ? '编辑公告' : '发布公告'"
      :confirm-text="editingId ? '保存' : '发布'"
      :loading="submitting"
      :width="520"
      @confirm="onSubmit"
    >
      <div class="flex flex-col gap-3.5 py-1">
        <label class="flex items-center gap-2 text-[13px] text-ink-2">
          <span class="w-20 shrink-0 text-right">公告标题</span>
          <AppInput v-model="form.title" class="flex-1" placeholder="请输入公告标题" />
        </label>
        <label class="flex items-center gap-2 text-[13px] text-ink-2">
          <span class="w-20 shrink-0 text-right">公告类型</span>
          <select v-model="form.type" class="h-9 flex-1 rounded-lg border border-line px-3 text-[13px] outline-none focus:border-primary">
            <option v-for="t in FORM_TYPES" :key="t" :value="t">{{ t }}</option>
          </select>
        </label>
        <label class="flex items-center gap-2 text-[13px] text-ink-2">
          <span class="w-20 shrink-0 text-right">发布对象</span>
          <select v-model="form.audience" class="h-9 flex-1 rounded-lg border border-line px-3 text-[13px] outline-none focus:border-primary">
            <option v-for="a in FORM_AUDIENCES" :key="a" :value="a">{{ a }}</option>
          </select>
        </label>
        <label class="flex items-start gap-2 text-[13px] text-ink-2">
          <span class="w-20 shrink-0 pt-2 text-right">公告正文</span>
          <textarea
            v-model="form.content"
            rows="4"
            class="flex-1 rounded-lg border border-line px-3 py-2 text-[13px] outline-none focus:border-primary"
            placeholder="请输入公告正文"
          />
        </label>
      </div>
    </Modal>

    <!-- 下线 / 发布 / 撤回确认 -->
    <Confirm
      v-model:open="confirmOpen"
      :title="confirmTitle"
      :message="confirmMessage"
      :confirm-text="confirmRow?.action_name ?? '确认'"
      @confirm="onConfirmAction"
    />
  </div>
</template>
