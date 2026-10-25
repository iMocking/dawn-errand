<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { useWorkspaceStore } from '@/stores/workspace'
import type { WorkspaceMessage } from '@/api/messages'
import { formatTime } from '@/utils/format'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const router = useRouter()
const workspace = useWorkspaceStore()
const tab = ref<'all' | 'todo' | 'system'>('all')

const tabs = computed(() => [
  { key: 'all' as const, label: '全部', count: workspace.messages.length },
  {
    key: 'todo' as const,
    label: '待办',
    count: workspace.messages.filter((m) => m.type !== 'system').length,
  },
  {
    key: 'system' as const,
    label: '系统',
    count: workspace.messages.filter((m) => m.type === 'system').length,
  },
])

const filtered = computed(() => {
  if (tab.value === 'todo') return workspace.messages.filter((m) => m.type !== 'system')
  if (tab.value === 'system') return workspace.messages.filter((m) => m.type === 'system')
  return workspace.messages
})

const TYPE_META: Record<WorkspaceMessage['type'], { icon: string; bg: string; color: string }> = {
  complaint: { icon: 'complaint', bg: '#FFF1F0', color: '#FF4D4F' },
  merchant: { icon: 'merchant', bg: '#FFF3E8', color: '#FF6B2C' },
  rider: { icon: 'shield', bg: '#E8F7F0', color: '#00B578' },
  withdraw: { icon: 'withdraw', bg: '#FFF8E0', color: '#D48806' },
  system: { icon: 'message', bg: '#E8F0FF', color: '#2B7FFF' },
}

function meta(type: WorkspaceMessage['type']) {
  return TYPE_META[type] ?? TYPE_META.system
}

function isUnread(id: string) {
  return !workspace.hasRead(id)
}

function openMessage(msg: WorkspaceMessage) {
  workspace.markRead(msg.id)
  emit('close')
  if (msg.link) router.push(msg.link)
}

function markAll() {
  workspace.markAllRead()
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.open) emit('close')
}

watch(
  () => props.open,
  (v) => {
    if (v) {
      workspace.refresh()
      window.addEventListener('keydown', onKeydown)
    } else {
      window.removeEventListener('keydown', onKeydown)
    }
  },
)

onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0"
      leave-active-class="transition duration-100 ease-in"
      leave-to-class="opacity-0"
    >
      <div v-if="open" class="fixed inset-0 z-[100]">
        <div class="absolute inset-0 bg-black/25 backdrop-blur-[1px]" @click="emit('close')" />

        <aside
          class="absolute right-0 top-0 flex h-full w-[420px] flex-col bg-white shadow-[-8px_0_32px_rgba(0,0,0,0.10)]"
        >
          <!-- 头部 -->
          <div class="flex shrink-0 items-start border-b border-line px-5 pb-3 pt-[18px]">
            <div>
              <div class="flex items-center gap-2">
                <span class="text-[16px] font-bold text-ink-1">消息中心</span>
                <span
                  v-if="workspace.unread > 0"
                  class="rounded-full bg-[#FFF1F0] px-2 py-[1px] text-[11px] font-semibold text-[#FF4D4F]"
                >
                  {{ workspace.unread }} 条未读
                </span>
              </div>
              <div class="mt-1 text-[12px] text-ink-3">
                待办 {{ workspace.todoCount }} 项 · 最近更新 {{ workspace.loaded ? '刚刚' : '加载中' }}
              </div>
            </div>
            <button
              type="button"
              class="ml-auto flex h-8 w-8 items-center justify-center rounded-lg text-ink-3 transition-colors hover:bg-head hover:text-ink-1"
              title="关闭"
              @click="emit('close')"
            >
              <SvgIcon name="close" :size="16" />
            </button>
          </div>

          <!-- 标签页 -->
          <div class="flex shrink-0 items-center gap-5 border-b border-line px-5">
            <button
              v-for="t in tabs"
              :key="t.key"
              type="button"
              class="relative py-3 text-[13px] transition-colors"
              :class="tab === t.key ? 'font-semibold text-primary' : 'text-ink-3 hover:text-ink-1'"
              @click="tab = t.key"
            >
              {{ t.label }}
              <span class="ml-1 text-[11px] text-ink-4">{{ t.count }}</span>
              <span
                v-if="tab === t.key"
                class="absolute inset-x-0 bottom-0 h-[2px] rounded-full bg-primary"
              />
            </button>
            <button
              type="button"
              class="ml-auto text-[12px] text-ink-3 transition-colors hover:text-primary disabled:cursor-not-allowed disabled:opacity-40"
              :disabled="workspace.unread === 0"
              @click="markAll"
            >
              全部已读
            </button>
          </div>

          <!-- 列表 -->
          <div class="flex-1 overflow-y-auto p-3">
            <div v-if="filtered.length === 0" class="flex h-full flex-col items-center justify-center pb-20">
              <div class="flex h-14 w-14 items-center justify-center rounded-full bg-head">
                <SvgIcon name="bell" :size="26" class="text-ink-4" />
              </div>
              <div class="mt-3 text-[13px] font-semibold text-ink-2">暂无消息</div>
              <div class="mt-1 text-[12px] text-ink-4">有新的待办或公告时会出现在这里</div>
            </div>

            <button
              v-for="msg in filtered"
              :key="msg.id"
              type="button"
              class="mb-1 flex w-full items-start gap-3 rounded-xl p-3 text-left transition-colors hover:bg-head"
              @click="openMessage(msg)"
            >
              <span
                class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg"
                :style="{ background: meta(msg.type).bg, color: meta(msg.type).color }"
              >
                <SvgIcon :name="meta(msg.type).icon" :size="16" />
              </span>
              <span class="min-w-0 flex-1">
                <span class="flex items-center gap-2">
                  <span class="truncate text-[13px] font-semibold text-ink-1">{{ msg.title }}</span>
                  <span
                    v-if="isUnread(msg.id)"
                    class="h-1.5 w-1.5 shrink-0 rounded-full"
                    style="background:#FF6B2C;"
                  />
                </span>
                <span class="mt-0.5 block truncate text-[12px] text-ink-3">{{ msg.content }}</span>
                <span class="mt-1 block text-[11px] text-ink-4">{{ formatTime(msg.createdAt) }}</span>
              </span>
              <SvgIcon name="chevron-right" :size="14" class="mt-2 shrink-0 text-ink-4" />
            </button>
          </div>

          <!-- 底部 -->
          <div class="flex shrink-0 items-center border-t border-line px-5 py-3 text-[11px] text-ink-4">
            <span>仅展示最近 20 条 · 已读状态保存在本机</span>
            <button
              type="button"
              class="ml-auto text-[12px] text-ink-3 transition-colors hover:text-primary"
              @click="workspace.refresh()"
            >
              刷新
            </button>
          </div>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>
