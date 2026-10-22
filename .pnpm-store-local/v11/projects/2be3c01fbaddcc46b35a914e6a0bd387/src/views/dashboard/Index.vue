<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { getDashboard, type DashboardDto } from '@/api/dashboard'
import StatCard from '@/components/StatCard.vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { greetByHour } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const data = ref<DashboardDto | null>(null)
const loading = ref(true)

const greeting = computed(() => `${greetByHour()}，${auth.user?.name ?? '陈运营'}`)

onMounted(async () => {
  try {
    data.value = await getDashboard()
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-if="loading" class="p-6 text-ink-3">加载中…</div>
  <div v-else-if="data" class="flex flex-col gap-4 p-5">
    <!-- 欢迎行 -->
    <div class="flex items-center">
      <div>
        <div class="text-[18px] font-bold text-ink-1">{{ greeting }}</div>
        <div class="mt-1 text-[12px] text-ink-3">
          全平台今日经营一切正常，有 <span class="font-semibold text-primary">{{ data.pendingCount }}</span> 项待处理事项
        </div>
      </div>
      <div class="ml-auto flex gap-3">
        <button class="flex h-9 items-center gap-2 rounded-lg border border-line bg-white px-4 text-[13px] text-ink-2">
          <SvgIcon name="download" :size="15" />
          导出报表
        </button>
        <button
          class="flex h-9 items-center gap-2 rounded-lg border-none text-[13px] font-semibold text-white"
          style="background:linear-gradient(135deg,#FF8A3D,#FF6B2C);box-shadow:0 4px 12px rgba(255,107,44,.3);"
        >
          <SvgIcon name="plus" :size="15" :stroke-width="2.2" />
          发布公告
        </button>
      </div>
    </div>

    <!-- 指标卡 -->
    <div class="grid grid-cols-4 gap-4">
      <StatCard
        v-for="s in data.stats"
        :key="s.key"
        :label="s.label"
        :value="s.value"
        :sub-value="s.subValue"
        :delta="s.delta"
        :delta-color="s.deltaType === 'green' ? '#00B578' : s.deltaType === 'yellow' ? '#D48806' : '#FF4D4F'"
        :foot="s.foot"
        :icon="s.icon"
        :icon-bg="s.iconBg"
        :icon-color="s.iconColor"
      />
    </div>

    <!-- 中部：趋势 + 分布 -->
    <div class="grid gap-4" style="grid-template-columns:1.8fr 1fr;">
      <!-- 订单趋势 -->
      <div class="rounded-card bg-white px-5 py-4">
        <div class="flex items-center">
          <div class="text-[15px] font-semibold text-ink-1">订单趋势</div>
          <div class="ml-auto flex items-center gap-4">
            <div class="flex items-center gap-1.5 text-[12px] text-ink-3">
              <span class="inline-block h-[10px] w-[10px] rounded-[3px]" style="background:#FF6B2C;" />
              订单量
            </div>
            <div class="flex items-center gap-1.5 text-[12px] text-ink-3">
              <span class="inline-block h-[10px] w-[10px] rounded-[3px]" style="background:#2B7FFF;" />
              交易额
            </div>
            <div class="flex gap-1">
              <span class="rounded-md px-2.5 py-1 text-[12px] font-semibold" style="background:#FFF3E8;color:#FF6B2C;">近7日</span>
              <span class="rounded-md px-2.5 py-1 text-[12px] text-ink-3">近30日</span>
            </div>
          </div>
        </div>

        <svg viewBox="0 0 640 250" width="100%" height="250" class="mt-3 block">
          <defs>
            <linearGradient id="g1" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#2B7FFF" stop-opacity="0.18" />
              <stop offset="100%" stop-color="#2B7FFF" stop-opacity="0" />
            </linearGradient>
            <linearGradient id="g2" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#FF6B2C" stop-opacity="0.15" />
              <stop offset="100%" stop-color="#FF6B2C" stop-opacity="0" />
            </linearGradient>
          </defs>
          <line x1="40" y1="20" x2="620" y2="20" stroke="#F2F3F5" stroke-width="1" />
          <line x1="40" y1="70" x2="620" y2="70" stroke="#F2F3F5" stroke-width="1" />
          <line x1="40" y1="120" x2="620" y2="120" stroke="#F2F3F5" stroke-width="1" />
          <line x1="40" y1="170" x2="620" y2="170" stroke="#F2F3F5" stroke-width="1" />
          <line x1="40" y1="220" x2="620" y2="220" stroke="#F2F3F5" stroke-width="1" />
          <text x="32" y="24" font-size="11" fill="#C9CDD4" text-anchor="end">1600</text>
          <text x="32" y="74" font-size="11" fill="#C9CDD4" text-anchor="end">1200</text>
          <text x="32" y="124" font-size="11" fill="#C9CDD4" text-anchor="end">800</text>
          <text x="32" y="174" font-size="11" fill="#C9CDD4" text-anchor="end">400</text>
          <text x="32" y="224" font-size="11" fill="#C9CDD4" text-anchor="end">0</text>

          <!-- 交易额（蓝）面积+线 -->
          <polygon
            :points="`${data.trend.gmvLine.map(p => `${p.x},${p.y}`).join(' ')} 620,220 40,220`"
            fill="url(#g1)"
          />
          <polyline
            :points="data.trend.gmvLine.map(p => `${p.x},${p.y}`).join(' ')"
            fill="none" stroke="#2B7FFF" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"
          />
          <!-- 订单量（橙）面积+线 -->
          <polygon
            :points="`${data.trend.orderLine.map(p => `${p.x},${p.y}`).join(' ')} 620,220 40,220`"
            fill="url(#g2)"
          />
          <polyline
            :points="data.trend.orderLine.map(p => `${p.x},${p.y}`).join(' ')"
            fill="none" stroke="#FF6B2C" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"
          />
          <circle
            v-for="p in data.trend.orderLine"
            :key="`o${p.x}`"
            :cx="p.x" :cy="p.y" r="4" fill="#fff" stroke="#FF6B2C" stroke-width="2"
          />

          <rect :x="data.trend.orderTip.x - data.trend.orderTip.w / 2" :y="data.trend.orderTip.y - 16"
            :width="data.trend.orderTip.w" height="24" rx="6" fill="#FF6B2C" />
          <text :x="data.trend.orderTip.x" :y="data.trend.orderTip.y" font-size="11" fill="#fff"
            text-anchor="middle" font-weight="600">{{ data.trend.orderTip.text }}</text>

          <rect :x="data.trend.gmvTip.x - data.trend.gmvTip.w / 2" :y="data.trend.gmvTip.y - 16"
            :width="data.trend.gmvTip.w" height="24" rx="6" fill="#2B7FFF" />
          <text :x="data.trend.gmvTip.x" :y="data.trend.gmvTip.y" font-size="11" fill="#fff"
            text-anchor="middle" font-weight="600">{{ data.trend.gmvTip.text }}</text>

          <text
            v-for="(label, i) in data.trend.xLabels"
            :key="label"
            :x="data.trend.orderLine[i].x" y="242" font-size="11"
            :fill="i === data.trend.xLabels.length - 1 ? '#FF6B2C' : '#86909C'"
            text-anchor="middle" :font-weight="i === data.trend.xLabels.length - 1 ? 600 : 400"
          >{{ label }}</text>
        </svg>
      </div>

      <!-- 业务类型分布 -->
      <div class="flex flex-col rounded-card bg-white px-5 py-4">
        <div class="text-[15px] font-semibold text-ink-1">业务类型分布</div>
        <div class="flex flex-1 flex-col justify-center gap-[14px] pt-2.5">
          <div v-for="d in data.distribution" :key="d.label">
            <div class="mb-1.5 flex items-center text-[12px] text-ink-2">
              <span>{{ d.label }}</span>
              <span class="ml-auto font-semibold text-ink-1">{{ d.percent }}%</span>
            </div>
            <div class="h-[10px] rounded-[5px]" style="background:#F2F3F5;">
              <div class="h-[10px] rounded-[5px]" :style="{ width: `${d.percent}%`, background: d.color }" />
            </div>
          </div>
          <div class="flex items-center gap-2 rounded-lg bg-head px-3 py-2.5 text-[12px] text-ink-3">
            <SvgIcon name="check-circle" :size="14" class="shrink-0" style="color:#00B578;" />
            配送准时率 <span class="ml-1 font-semibold text-ink-1">{{ data.onTimeRate }}</span>
            <span class="ml-auto text-ink-4">目标 {{ data.onTimeTarget }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 下部：实时订单流 + 待办 -->
    <div class="grid h-[238px] shrink-0 gap-4" style="grid-template-columns:1.8fr 1fr;">
      <!-- 实时订单流 -->
      <div class="flex flex-col rounded-card bg-white px-5 py-[14px]">
        <div class="flex items-center">
          <div class="text-[15px] font-semibold text-ink-1">实时订单流</div>
          <span class="ml-3 flex items-center gap-[5px] text-[11px]" style="color:#00B578;">
            <span class="inline-block h-1.5 w-1.5 rounded-full" style="background:#00B578;" />
            实时更新
          </span>
          <a class="ml-auto cursor-pointer text-[12px] text-primary">查看全部</a>
        </div>
        <div class="flex flex-1 flex-col justify-around pt-1.5">
          <div v-for="o in data.recentOrders" :key="o.id" class="flex items-center gap-2.5">
            <span class="w-[120px] shrink-0 text-[12px] text-ink-3">#{{ o.id }}</span>
            <span
              class="rounded-tag px-2 py-0.5 text-[12px] font-semibold"
              :style="{ background: o.typeBg, color: o.typeColor }"
            >{{ o.type }}</span>
            <span class="flex-1 truncate text-[12px] text-ink-2">{{ o.route }}</span>
            <span class="text-[12px] font-semibold" :style="{ color: o.statusColor }">{{ o.status }}</span>
            <span class="text-[12px] text-ink-4">{{ o.time }}</span>
          </div>
        </div>
      </div>

      <!-- 待办事项 -->
      <div class="flex flex-col rounded-card bg-white px-5 py-[14px]">
        <div class="flex items-center">
          <div class="text-[15px] font-semibold text-ink-1">待办事项</div>
          <span class="ml-auto rounded-full px-2.5 py-0.5 text-[12px] font-semibold" style="background:#FFF3E8;color:#FF6B2C;">
            {{ data.pendingCount }} 项
          </span>
        </div>
        <div class="flex flex-1 flex-col justify-around pt-1.5">
          <div v-for="t in data.todos" :key="t.label" class="flex items-center gap-2.5">
            <div
              class="flex h-[30px] w-[30px] items-center justify-center rounded-lg"
              :style="{ background: t.bg }"
            >
              <SvgIcon :name="t.icon" :size="15" :style="{ color: t.color }" />
            </div>
            <div class="flex-1 text-[13px] text-ink-2">{{ t.label }}</div>
            <span class="text-[13px] font-bold" :style="{ color: t.color }">{{ t.count }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
