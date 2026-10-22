<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import AppButton from '@/components/AppButton.vue'
import { getComplaintDetail, type ComplaintDetail } from '@/api/complaints'
import { http } from '@/utils/request'
import { formatTime } from '@/utils/format'
import { useToast } from '@/composables/useToast'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const loading = ref(true)
const detail = ref<ComplaintDetail | null>(null)
const plan = ref('全额退款')
const remark = ref('')
const submitting = ref(false)

const PLANS = ['全额退款', '部分退款', '补送处理']

async function load() {
  loading.value = true
  try {
    detail.value = await getComplaintDetail(String(route.params.id))
  } finally {
    loading.value = false
  }
}

async function onConfirm() {
  if (!detail.value) return
  submitting.value = true
  try {
    await http.post(`/complaints/${detail.value.id}/resolve`, {
      plan: plan.value,
      remark: remark.value,
    })
    toast.success('退款已确认，工单完结')
    router.push('/complaints')
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>

<template>
  <div v-if="loading" class="text-ink-3">加载中…</div>
  <div v-else-if="detail" class="flex flex-col gap-3">
    <!-- 工单头部卡 -->
    <div class="flex items-center gap-3 rounded-card bg-white px-5 py-3">
      <div class="flex h-[38px] w-[56px] shrink-0 items-center justify-center rounded-[14px]" style="background:#FFF0E8;">
        <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="#FF6B2C" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><path d="M3 11h18M5 11V8a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v3M4 11l1 9h14l1-9" /></svg>
      </div>
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <span class="text-[18px] font-bold text-ink-1">{{ detail.title }}</span>
          <span class="rounded-tag px-2.5 py-[3px] text-[12px] font-semibold" style="background:#FFF7E6;color:#D48806;">{{ detail.status }}</span>
          <span class="rounded-tag px-2.5 py-[3px] text-[12px] font-semibold" style="background:#FFF0E8;color:#FF6B2C;">{{ detail.priority }}</span>
        </div>
        <div class="mt-2 flex flex-wrap items-center gap-6 text-[13px] text-ink-3">
          <span>工单号 <b class="font-semibold text-ink-1">{{ detail.id }}</b></span>
          <span>提交于 <b class="font-semibold text-ink-1">{{ formatTime(detail.created_at) }}</b></span>
          <span>发起人 <b class="font-semibold text-ink-1">{{ detail.initiator }}</b>（{{ detail.initiator_extra }}）</span>
          <span>责任方 <b class="font-semibold" style="color:#FF4D4F;">{{ detail.target }}</b></span>
          <span>关联订单 <b class="font-semibold" style="color:#2B7FFF;">#{{ detail.order_id }}</b></span>
        </div>
      </div>
      <div class="flex shrink-0 gap-2">
        <AppButton
          variant="secondary"
          class="!h-[30px] !px-[18px]"
          @click="toast.info('已打开转派处理人')"
        >
          转派处理人
        </AppButton>
        <AppButton
          v-permission="'complaint:handle'"
          class="!h-[30px] !px-[18px]"
          @click="onConfirm"
        >
          标记完成
        </AppButton>
      </div>
    </div>

    <!-- 主体 -->
    <div class="flex min-h-0 flex-1 gap-3">
      <!-- 左：处理流程 + 沟通记录 -->
      <div class="flex min-w-0 flex-1 flex-col gap-3" style="flex:1.2;">
        <!-- 处理流程 -->
        <div class="rounded-card bg-white p-3">
          <div class="text-[15px] font-semibold text-ink-1">处理流程</div>
          <div class="mt-2.5 flex flex-col">
            <div v-for="(f, i) in detail.flow" :key="f.title" class="flex gap-3.5">
              <div class="flex w-5 shrink-0 flex-col items-center">
                <!-- done -->
                <div v-if="f.state === 'done'" class="flex h-5 w-5 items-center justify-center rounded-full" style="background:#00B578;">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12" /></svg>
                </div>
                <!-- current -->
                <div
                  v-else-if="f.state === 'current'"
                  class="flex h-5 w-5 items-center justify-center rounded-full text-[11px] font-bold text-white"
                  style="background:#FF6B2C;"
                >
                  {{ f.step }}
                </div>
                <!-- pending -->
                <div
                  v-else
                  class="flex h-5 w-5 items-center justify-center rounded-full text-[11px] font-bold text-white"
                  style="background:#D48806;"
                >
                  {{ f.step }}
                </div>
                <div v-if="i < detail.flow.length - 1" class="h-[38px] w-0.5" style="background:#E5E6EB;"></div>
              </div>
              <div class="pb-1">
                <div class="text-[13px] font-semibold text-ink-1">{{ f.title }}</div>
                <div class="mt-0.5 text-[12px] leading-[1.6] text-ink-3">{{ f.desc }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 沟通记录 -->
        <div class="rounded-card bg-white p-3">
          <div class="flex items-center">
            <div class="text-[15px] font-semibold text-ink-1">沟通记录</div>
            <div class="ml-auto cursor-pointer text-[12px] font-semibold" style="color:#2B7FFF;" @click="toast.info('备注输入已打开')">
              + 添加备注
            </div>
          </div>
          <div class="mt-2 flex flex-col gap-3.5">
            <div v-for="n in detail.notes" :key="n.name + n.time" class="flex gap-2">
              <div
                class="flex h-[30px] w-[30px] shrink-0 items-center justify-center rounded-full text-[12px] font-bold"
                :style="{ background: n.avatar_bg, color: n.avatar_color }"
              >
                {{ n.avatar }}
              </div>
              <div class="flex-1">
                <div class="text-[12px] text-ink-2">
                  <b class="font-semibold text-ink-1">{{ n.name }}</b> ·
                  {{ formatTime(n.time) }}
                </div>
                <div
                  class="mt-1 rounded-lg p-2 text-[13px] leading-[1.7] text-ink-2"
                  :style="n.tone === 'yellow' ? 'background:#FFF7E6;border:1px solid #FFE7BA;' : 'background:#F7F8FA;'"
                >
                  {{ n.content }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 右：关联订单 + 处理操作 -->
      <div class="flex min-w-0 flex-1 flex-col gap-3">
        <!-- 关联订单 -->
        <div class="rounded-card bg-white p-3">
          <div class="text-[15px] font-semibold text-ink-1">关联订单</div>
          <div class="mt-2 flex flex-col gap-2 text-[13px]">
            <div v-for="o in detail.order" :key="o.label" class="flex justify-between">
              <span class="text-ink-3">{{ o.label }}</span>
              <span
                :style="
                  o.tone === 'blue'
                    ? 'color:#2B7FFF;font-weight:600;'
                    : o.tone === 'primary'
                      ? 'color:#1D2129;font-weight:700;'
                      : 'color:#4E5969;'
                "
              >
                {{ o.value }}
              </span>
            </div>
          </div>
        </div>

        <!-- 处理操作 -->
        <div class="rounded-card bg-white p-3">
          <div class="text-[15px] font-semibold text-ink-1">处理操作</div>
          <div class="mt-2">
            <div class="mb-1.5 text-[12px] text-ink-3">处理方案</div>
            <div class="flex gap-2">
              <div
                v-for="p in PLANS"
                :key="p"
                class="flex h-8 flex-1 cursor-pointer items-center justify-center rounded-lg text-[13px]"
                :class="plan === p ? 'font-semibold' : 'border border-line bg-white text-ink-2'"
                :style="plan === p ? 'border:2px solid #FF6B2C;background:#FFF3E8;color:#FF6B2C;' : ''"
                @click="plan = p"
              >
                {{ p === '全额退款' ? `全额退款 ¥${detail.pay_amount.toFixed(2)}` : p }}
              </div>
            </div>
            <div class="mb-1.5 mt-3.5 text-[12px] text-ink-3">定责与餐损分摊（商家建议方案）</div>
            <div class="flex h-8 items-center justify-between rounded-lg border border-line bg-head px-3 text-[13px] text-ink-1">
              <span>平台承担 70% · 骑手承担 30%</span>
              <SvgIcon name="chevron-right" :size="14" class="text-ink-3" style="transform:rotate(90deg);" />
            </div>
            <div class="mb-1.5 mt-3.5 text-[12px] text-ink-3">备注说明</div>
            <textarea
              v-model="remark"
              rows="2"
              placeholder="填写处理备注，将同步给用户与责任方…"
              class="box-border h-[50px] w-full resize-none rounded-lg border border-line bg-head p-2 text-[13px] text-ink-1 outline-none placeholder:text-ink-4"
            ></textarea>
          </div>
          <div class="mt-2 flex gap-2">
            <AppButton
              v-permission="'complaint:handle'"
              class="!h-8 flex-1"
              :disabled="submitting"
              @click="onConfirm"
            >
              {{ submitting ? '提交中…' : '确认退款并完结' }}
            </AppButton>
            <AppButton
              variant="secondary"
              class="!h-8 flex-1"
              @click="toast.info('已暂存草稿')"
            >
              暂存
            </AppButton>
          </div>
          <div class="mt-2 flex items-center gap-1.5 text-[12px] text-ink-4">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#C9CDD4" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10" /><path d="M12 16v-4M12 8h.01" /></svg>
            退款需风控复核，处理时长约 10 分钟
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
