<script setup lang="ts">
import { useRouter } from 'vue-router'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { useMenus, useActiveMenu } from '@/permission/menus'
import { useAuthStore } from '@/stores/auth'
import { resetRouter } from '@/router'

const groups = useMenus()
const active = useActiveMenu()
const auth = useAuthStore()
const router = useRouter()

const badgeColor: Record<string, string> = {
  red: 'background:#FF4D4F;',
  orange: 'background:#FF6B2C;',
  yellow: 'background:#D48806;',
}

async function onLogout() {
  await auth.logout()
  resetRouter(router)
  router.push('/login')
}
</script>

<template>
  <aside class="flex h-full w-[224px] shrink-0 flex-col border-r border-line bg-white">
    <!-- Logo -->
    <div class="flex items-center gap-2.5 border-b border-[#F2F3F5] px-4 py-2">
      <div
        class="flex h-[38px] w-[38px] items-center justify-center rounded-[10px]"
        style="background:linear-gradient(135deg,#FF8A3D,#FF6B2C);"
      >
        <SvgIcon name="clock" :size="22" :stroke-width="1.9" class="text-white" />
      </div>
      <div>
        <div class="text-[16px] font-bold leading-tight text-ink-1">曙光跑腿</div>
        <div class="mt-px text-[10px] text-ink-3">平台运营管理端</div>
      </div>
    </div>

    <!-- 菜单 -->
    <nav class="flex-1 overflow-y-auto py-2">
      <template v-for="g in groups" :key="g.group">
        <div class="px-4 pb-0.5 pt-[9px] text-[11px] tracking-wide text-ink-4">{{ g.group }}</div>
        <router-link
          v-for="item in g.items"
          :key="item.path"
          :to="item.path"
          class="mx-3 mb-0.5 flex items-center gap-2 rounded-lg px-3 py-2 text-[13px]"
          :class="
            active === item.path
              ? 'bg-primary-bg font-semibold text-primary'
              : 'text-ink-2 hover:bg-page'
          "
        >
          <SvgIcon :name="item.icon ?? 'list'" :size="16" :stroke-width="2" />
          <span>{{ item.title }}</span>
          <span
            v-if="item.badge"
            class="ml-auto flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[11px] font-semibold text-white"
            :style="badgeColor[item.badgeColor ?? 'orange']"
          >
            {{ item.badge }}
          </span>
        </router-link>
      </template>
    </nav>

    <!-- 用户卡片 -->
    <div class="flex items-center gap-2.5 border-t border-[#F2F3F5] px-4 py-[14px]">
      <div
        class="flex h-[34px] w-[34px] items-center justify-center rounded-full text-[14px] font-bold"
        style="background:#FFF3E8;color:#FF6B2C;"
      >
        {{ auth.user?.name?.charAt(0) ?? '陈' }}
      </div>
      <div class="min-w-0 flex-1">
        <div class="truncate text-[13px] font-semibold text-ink-1">{{ auth.user?.name ?? '陈运营' }}</div>
        <div class="text-[11px] text-ink-3">{{ auth.user?.role ?? '平台管理员' }}</div>
      </div>
      <button class="text-ink-3 hover:text-primary" title="退出登录" @click="onLogout">
        <SvgIcon name="logout" :size="18" />
      </button>
    </div>
  </aside>
</template>
