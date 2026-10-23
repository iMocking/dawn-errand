<script setup lang="ts">
import { onMounted, ref } from 'vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { useMenus, useActiveMenu } from '@/permission/menus'

const groups = useMenus()
const active = useActiveMenu()

/* 面板收起 / 展开（默认展开，状态持久化到 localStorage） */
const SIDEBAR_KEY = 'admin_sidebar_collapsed'
const collapsed = ref(false)

onMounted(() => {
  collapsed.value = localStorage.getItem(SIDEBAR_KEY) === '1'
})

function toggleCollapse() {
  collapsed.value = !collapsed.value
  localStorage.setItem(SIDEBAR_KEY, collapsed.value ? '1' : '0')
}

const badgeColor: Record<string, string> = {
  red: 'background:#FF4D4F;',
  orange: 'background:#FF6B2C;',
  yellow: 'background:#D48806;',
}
</script>

<template>
  <aside
    class="flex h-full shrink-0 flex-col overflow-hidden border-r border-line bg-white transition-[width] duration-200 ease-out"
    :class="collapsed ? 'w-[64px]' : 'w-[200px]'"
  >
    <!-- Logo（收起时仅图标） -->
    <div
      class="flex shrink-0 items-center border-b border-[#F2F3F5] py-2"
      :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-4'"
    >
      <div
        class="flex h-[38px] w-[38px] shrink-0 items-center justify-center rounded-[10px]"
        style="background:linear-gradient(135deg,#FF8A3D,#FF6B2C);"
      >
        <SvgIcon name="clock" :size="22" :stroke-width="1.9" class="text-white" />
      </div>
      <div v-if="!collapsed">
        <div class="text-[16px] font-bold leading-tight text-ink-1">曙光跑腿</div>
        <div class="mt-px text-[10px] text-ink-3">平台运营管理端</div>
      </div>
    </div>

    <!-- 菜单（收起时仅图标） -->
    <nav class="flex-1 overflow-y-auto overflow-x-hidden py-2">
      <template v-for="g in groups" :key="g.group">
        <div
          v-if="!collapsed"
          class="px-4 pb-0.5 pt-[9px] text-[11px] tracking-wide text-ink-4"
        >{{ g.group }}</div>
        <div
          v-else
          class="mx-4 my-1.5 h-px bg-[#F2F3F5]"
        />
        <router-link
          v-for="item in g.items"
          :key="item.path"
          :to="item.path"
          :title="collapsed ? item.title : undefined"
          class="relative mb-0.5 flex items-center rounded-lg py-2 text-[13px]"
          :class="[
            collapsed ? 'mx-auto h-9 w-9 justify-center px-0' : 'mx-3 gap-2 px-3',
            active === item.path
              ? 'bg-primary-bg font-semibold text-primary'
              : 'text-ink-2 hover:bg-page',
          ]"
        >
          <SvgIcon :name="item.icon ?? 'list'" :size="16" :stroke-width="2" />
          <span v-if="!collapsed">{{ item.title }}</span>
          <span
            v-if="item.badge"
            class="flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[11px] font-semibold text-white"
            :class="collapsed
              ? 'absolute -right-1 -top-1 h-[15px] min-w-[15px] px-0.5'
              : 'ml-auto'"
            :style="badgeColor[item.badgeColor ?? 'orange']"
          >
            {{ item.badge }}
          </span>
        </router-link>
      </template>
    </nav>

    <!-- 底部：面板收起 / 展开 -->
    <div class="flex shrink-0 items-center justify-center border-t border-[#F2F3F5] py-[11px]">
      <button
        type="button"
        class="flex h-8 w-8 cursor-pointer items-center justify-center rounded-lg text-ink-3 outline-none transition-colors hover:bg-page hover:text-primary"
        :title="collapsed ? '展开面板' : '收起面板'"
        @click="toggleCollapse"
      >
        <SvgIcon :name="collapsed ? 'chevrons-right' : 'chevrons-left'" :size="18" />
      </button>
    </div>
  </aside>
</template>
