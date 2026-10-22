<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import StatusTag from '@/components/StatusTag.vue'
import AppButton from '@/components/AppButton.vue'
import AppInput from '@/components/AppInput.vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { useToast } from '@/composables/useToast'
import { getWithdrawRule, updateWithdrawRule } from '@/api/system'

const toast = useToast()

const loading = ref(false)
const saving = ref(false)

const form = reactive({
  enabled: true,
  /** 手续费率（%，界面值；提交时换算为小数） */
  fee_rate_percent: 0.6,
  min_amount: 10,
  max_amount: 5000,
  daily_limit: 5,
  mode: 'manual' as 'auto' | 'manual',
  settle_days: 1,
})

async function load() {
  loading.value = true
  try {
    const cfg = await getWithdrawRule()
    form.enabled = cfg.enabled
    form.fee_rate_percent = Math.round(cfg.fee_rate * 10000) / 100
    form.min_amount = cfg.min_amount
    form.max_amount = cfg.max_amount
    form.daily_limit = cfg.daily_limit
    form.mode = cfg.mode === 'auto' ? 'auto' : 'manual'
    form.settle_days = cfg.settle_days
  } finally {
    loading.value = false
  }
}

function validate(): boolean {
  if (form.fee_rate_percent < 0) {
    toast.warning('手续费率不能为负数')
    return false
  }
  if (form.min_amount < 0 || form.max_amount < 0) {
    toast.warning('提现金额不能为负数')
    return false
  }
  if (form.max_amount > 0 && form.min_amount > 0 && form.min_amount > form.max_amount) {
    toast.warning('最低提现金额不能高于最高金额')
    return false
  }
  if (form.daily_limit < 0 || form.settle_days < 0) {
    toast.warning('次数与到账周期不能为负数')
    return false
  }
  return true
}

async function onSave() {
  if (saving.value || !validate()) return
  saving.value = true
  try {
    await updateWithdrawRule({
      enabled: form.enabled,
      fee_rate: Math.round(form.fee_rate_percent * 10000) / 100 / 100,
      min_amount: Number(form.min_amount),
      max_amount: Number(form.max_amount),
      daily_limit: Number(form.daily_limit),
      mode: form.mode,
      settle_days: Number(form.settle_days),
    })
    toast.success('提现规则已保存')
    await load()
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="flex min-h-0 flex-col rounded-card bg-white">
    <!-- 顶部条 -->
    <div class="flex shrink-0 items-center border-b border-line px-5 py-4">
      <div class="flex h-[34px] w-10 shrink-0 items-center justify-center rounded-[10px]" style="background:#E8F7F0;">
        <SvgIcon name="money" :size="20" :stroke-width="1.9" class="text-[#00B578]" />
      </div>
      <div class="ml-3 min-w-0 flex-1">
        <div class="text-[15px] font-bold text-ink-1">提现规则</div>
        <div class="mt-[2px] truncate text-[12px] text-ink-3">
          平台向骑手 / 商家结算账户提现的统一规则，保存后全端同步生效
        </div>
      </div>
      <div class="ml-auto flex shrink-0 items-center gap-3">
        <StatusTag :tone="form.enabled ? 'green' : 'gray'">
          {{ form.enabled ? '已启用' : '已停用' }}
        </StatusTag>
        <AppButton
          class="shadow-[0_4px_12px_rgba(255,107,44,.3)]"
          :loading="saving"
          @click="onSave"
        >
          保存配置
        </AppButton>
      </div>
    </div>

    <!-- 启用开关 -->
    <div class="flex shrink-0 items-center gap-3 border-b border-line px-5 py-3.5">
      <span class="w-24 shrink-0 text-right text-[13px] text-ink-2">提现开关</span>
      <button
        type="button"
        class="flex h-[22px] w-10 items-center rounded-full transition-colors"
        :class="form.enabled ? 'justify-end bg-[#00B578]' : 'justify-start bg-[#C9CDD4]'"
        @click="form.enabled = !form.enabled"
      >
        <span class="mx-[3px] h-4 w-4 rounded-full bg-white shadow" />
      </button>
      <span class="text-[12px] text-ink-3">开启后骑手 / 商家可发起提现申请；关闭后入口隐藏</span>
    </div>

    <!-- 表单 -->
    <div v-if="loading" class="py-16 text-center text-[13px] text-ink-3">加载中…</div>
    <div v-else class="grid max-w-3xl grid-cols-2 gap-x-10 gap-y-5 overflow-y-auto p-5">
      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">手续费率</span>
        <AppInput v-model.number="form.fee_rate_percent" type="number" class="flex-1" placeholder="如 0.6" />
        <span class="text-[12px] text-ink-4">%</span>
      </label>
      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">单笔最低</span>
        <AppInput v-model.number="form.min_amount" type="number" class="flex-1" placeholder="如 10" />
        <span class="text-[12px] text-ink-4">元</span>
      </label>
      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">单笔最高</span>
        <AppInput v-model.number="form.max_amount" type="number" class="flex-1" placeholder="如 5000" />
        <span class="text-[12px] text-ink-4">元</span>
      </label>
      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">每日次数</span>
        <AppInput v-model.number="form.daily_limit" type="number" class="flex-1" placeholder="如 5" />
        <span class="text-[12px] text-ink-4">次</span>
      </label>
      <div class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">审核模式</span>
        <div class="flex h-9 flex-1 overflow-hidden rounded-lg border border-line">
          <button
            v-for="opt in [{ value: 'manual', label: '人工审核' }, { value: 'auto', label: '自动审核' }]"
            :key="opt.value"
            type="button"
            class="flex-1 text-[13px] transition-colors"
            :class="
              form.mode === opt.value
                ? 'bg-[#E8F7F0] font-semibold text-[#00B578]'
                : 'bg-white text-ink-3 hover:text-ink-1'
            "
            @click="form.mode = opt.value as 'auto' | 'manual'"
          >
            {{ opt.label }}
          </button>
        </div>
      </div>
      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">到账周期</span>
        <AppInput v-model.number="form.settle_days" type="number" class="flex-1" placeholder="如 1" />
        <span class="text-[12px] text-ink-4">T+{{ form.settle_days }} 工作日</span>
      </label>

      <p class="col-span-2 -mt-1 pl-[120px] text-[12px] text-ink-4">
        规则保存后立即生效；提现申请审核在「财务结算 → 提现审核」中处理，自动审核模式将按每日次数与金额上限直接放行。
      </p>
    </div>
  </div>
</template>
