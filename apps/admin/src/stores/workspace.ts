import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { getMessages, type BadgeCounts, type WorkspaceMessage } from '@/api/messages'

const READ_KEY = 'admin_message_read_ids'
const EMPTY_BADGES: BadgeCounts = { complaints: 0, merchantAudits: 0, riderAudits: 0, withdrawals: 0 }

function loadReadIds(): Set<string> {
  try {
    const raw = localStorage.getItem(READ_KEY)
    return new Set<string>(raw ? (JSON.parse(raw) as string[]) : [])
  } catch {
    return new Set<string>()
  }
}

export const useWorkspaceStore = defineStore('workspace', () => {
  const badges = ref<BadgeCounts>({ ...EMPTY_BADGES })
  const messages = ref<WorkspaceMessage[]>([])
  const readIds = ref<Set<string>>(loadReadIds())
  const loading = ref(false)
  const loaded = ref(false)

  const unread = computed(() => messages.value.filter((m) => !readIds.value.has(m.id)).length)
  const todoCount = computed(
    () => badges.value.complaints + badges.value.merchantAudits + badges.value.riderAudits + badges.value.withdrawals,
  )

  function persistRead() {
    localStorage.setItem(READ_KEY, JSON.stringify([...readIds.value]))
  }

  function markRead(id: string) {
    if (readIds.value.has(id)) return
    readIds.value = new Set([...readIds.value, id])
    persistRead()
  }

  function hasRead(id: string) {
    return readIds.value.has(id)
  }

  function markAllRead() {
    readIds.value = new Set([...readIds.value, ...messages.value.map((m) => m.id)])
    persistRead()
  }

  async function refresh() {
    if (loading.value) return
    loading.value = true
    try {
      const res = await getMessages()
      messages.value = res.list ?? []
      badges.value = { ...EMPTY_BADGES, ...(res.badges ?? {}) }
      loaded.value = true
    } catch {
      // 请求层已统一 toast；保留旧数据，避免侧边栏闪烁。
    } finally {
      loading.value = false
    }
  }

  /** 菜单路径 → 真实角标数（0 不展示）。 */
  function badgeCount(path: string): number {
    switch (path) {
      case '/complaints':
        return badges.value.complaints
      case '/merchants/audit':
        return badges.value.merchantAudits
      case '/riders/audit':
        return badges.value.riderAudits
      case '/finance':
        return badges.value.withdrawals
      default:
        return 0
    }
  }

  return { badges, messages, unread, todoCount, loading, loaded, refresh, badgeCount, markRead, markAllRead, hasRead }
})
