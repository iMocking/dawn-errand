<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import StatusTag from '@/components/StatusTag.vue'
import MessageCenter from '@/components/MessageCenter.vue'
import { formatToday } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'
import { useWorkspaceStore } from '@/stores/workspace'
import { resetRouter } from '@/router'
import { BACKEND_ROLES } from '@/api/users'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const workspace = useWorkspaceStore()
const messageOpen = ref(false)
const crumb = computed(() => (route.meta?.title as string) ?? '')

/* 头像面板（交互对齐哔哩哔哩顶栏：hover 展开 / 移出关闭 / 点击切换） */
const open = ref(false)

const initial = computed(() => auth.user?.name?.charAt(0) ?? '陈')
const roleName = computed(() => {
  const role = auth.user?.role ?? ''
  return BACKEND_ROLES.find((r) => r.value === role)?.label ?? '普通用户'
})

function roleTone(role: string | undefined): 'orange' | 'blue' | 'green' | 'gray' | 'red' | 'yellow' {
  switch (role) {
    case 'super_admin':
      return 'orange'
    case 'cs_manager':
    case 'cs_agent':
    case 'region_station':
      return 'blue'
    case 'merchant':
    case 'driver':
      return 'green'
    default:
      return 'gray'
  }
}

function toggle() {
  open.value = !open.value
}
function close() {
  open.value = false
}

async function onLogout() {
  close()
  await auth.logout()
  resetRouter(router)
  router.push('/login')
}
</script>

<template>
  <header class="flex h-14 shrink-0 items-center border-b border-line bg-white px-6">
    <div class="flex items-center gap-2 text-[15px] font-semibold text-ink-1">
      {{ crumb }}
      <SvgIcon name="chevron-right" :size="14" class="text-ink-4" />
      <span class="text-[13px] font-normal text-ink-3">{{ formatToday() }}</span>
    </div>

    <div class="ml-auto flex items-center gap-4">
      <div class="flex h-[34px] w-[220px] items-center gap-2 rounded-lg bg-head px-3">
        <SvgIcon name="search" :size="15" class="text-ink-3" />
        <span class="text-[13px] text-ink-4">搜索订单 / 商家 / 骑手</span>
      </div>

      <button
        type="button"
        class="relative flex h-8 w-8 cursor-pointer items-center justify-center rounded-lg text-ink-2 transition-colors hover:bg-head hover:text-ink-1"
        title="消息中心"
        @click="messageOpen = true"
      >
        <SvgIcon name="bell" :size="20" />
        <span
          v-if="workspace.unread > 0"
          class="absolute -right-1 -top-1 flex h-[15px] min-w-[15px] items-center justify-center rounded-full px-1 text-[10px] font-semibold text-white"
          style="background:#FF4D4F;"
        >
          {{ workspace.unread > 99 ? '99+' : workspace.unread }}
        </span>
      </button>

      <!-- 头像：B 站式交互（hover 展开面板，移出关闭，点击切换） -->
      <div
        class="relative"
        @mouseenter="open = true"
        @mouseleave="open = false"
      >
        <button
          type="button"
          class="flex h-[34px] w-[34px] cursor-pointer items-center justify-center rounded-full text-[14px] font-bold text-white outline-none transition-transform duration-100"
          :class="open ? 'scale-95' : 'hover:scale-105'"
          style="background:linear-gradient(135deg,#FF8A3D,#FF6B2C);"
          @click="toggle"
        >
          {{ initial }}
        </button>

        <Transition
          enter-active-class="transition duration-100 ease-out"
          enter-from-class="opacity-0 translate-y-1"
          leave-active-class="transition duration-80 ease-in"
          leave-to-class="opacity-0 translate-y-1"
        >
          <div
            v-if="open"
            class="absolute right-0 top-[calc(100%+12px)] w-[280px] rounded-xl border border-line bg-white shadow-[0_8px_24px_rgba(0,0,0,0.10)]"
            @click="close"
          >
            <!-- 指向头像的三角 -->
            <div class="absolute -top-[6px] right-4 h-3 w-3 rotate-45 border-l border-t border-line bg-white" />

            <!-- 信息卡：头像 / 昵称 / 角色 -->
            <div class="flex items-center gap-3 px-4 pt-4">
              <div
                class="flex h-11 w-11 shrink-0 items-center justify-center rounded-full text-[16px] font-bold text-white"
                style="background:linear-gradient(135deg,#FF8A3D,#FF6B2C);"
              >
                {{ initial }}
              </div>
              <div class="min-w-0">
                <div class="truncate text-[14px] font-semibold text-ink-1">{{ auth.user?.name ?? '管理员' }}</div>
                <div class="mt-1">
                  <StatusTag :tone="roleTone(auth.user?.role)">{{ roleName }}</StatusTag>
                </div>
              </div>
            </div>

            <!-- 功能组 -->
            <div class="mx-4 mt-3 h-px bg-line" />
            <div class="py-1.5">
              <router-link
                to="/settings"
                class="flex cursor-pointer items-center gap-2.5 px-4 py-2.5 text-[13px] text-ink-2 transition-colors hover:bg-head hover:text-ink-1"
              >
                <SvgIcon name="settings" :size="15" class="shrink-0 text-ink-3" />
                系统设置
              </router-link>
              <router-link
                to="/finance/settlement"
                class="flex cursor-pointer items-center gap-2.5 px-4 py-2.5 text-[13px] text-ink-2 transition-colors hover:bg-head hover:text-ink-1"
              >
                <SvgIcon name="money" :size="15" class="shrink-0 text-ink-3" />
                财务结算
              </router-link>
            </div>

            <!-- 底部：退出登录 -->
            <div class="mx-4 h-px bg-line" />
            <div class="py-1.5">
              <button
                type="button"
                class="flex w-full cursor-pointer items-center gap-2.5 px-4 py-2.5 text-left text-[13px] transition-colors hover:bg-[#FFF1F0]"
                style="color:#FF4D4F;"
                @click="onLogout"
              >
                <SvgIcon name="logout" :size="15" class="shrink-0" />
                退出登录
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </div>
  </header>

  <MessageCenter :open="messageOpen" @close="messageOpen = false" />
</template>
