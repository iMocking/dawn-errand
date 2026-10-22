<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import StatusTag from '@/components/StatusTag.vue'
import AppButton from '@/components/AppButton.vue'
import AppInput from '@/components/AppInput.vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { useToast } from '@/composables/useToast'
import { getSmsConfig, testSmsConfig, updateSmsConfig } from '@/api/system'

const toast = useToast()

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
/** 后端短信配置接口（/system/config/sms）尚未提供时置为 true，禁用保存/测试 */
const backendMissing = ref(false)

const form = reactive({
  provider: 'aliyun',
  access_key: '',
  secret_key: '',
  sign_name: '',
  template_code: '',
})

const PROVIDERS = [
  { value: 'aliyun', label: '阿里云' },
  { value: 'tencent', label: '腾讯云' },
  { value: 'other', label: '其他' },
]

async function load() {
  loading.value = true
  try {
    const cfg = await getSmsConfig()
    form.provider = cfg.provider || 'aliyun'
    form.access_key = cfg.access_key
    form.secret_key = cfg.secret_key
    form.sign_name = cfg.sign_name
    form.template_code = cfg.template_code
  } catch {
    // 后端尚未提供短信配置接口（HTTP 404）
    backendMissing.value = true
  } finally {
    loading.value = false
  }
}

function buildPayload() {
  return {
    provider: form.provider,
    access_key: form.access_key.trim(),
    secret_key: form.secret_key,
    sign_name: form.sign_name.trim(),
    template_code: form.template_code.trim(),
  }
}

function validate(): boolean {
  if (!form.access_key.trim()) {
    toast.warning('请输入 AccessKey')
    return false
  }
  if (!form.sign_name.trim()) {
    toast.warning('请输入短信签名')
    return false
  }
  return true
}

async function onSave() {
  if (saving.value || backendMissing.value || !validate()) return
  saving.value = true
  try {
    await updateSmsConfig(buildPayload())
    toast.success('短信服务配置已保存')
  } finally {
    saving.value = false
  }
}

async function onTest() {
  if (testing.value || backendMissing.value || !validate()) return
  testing.value = true
  try {
    await testSmsConfig(buildPayload())
    toast.success('短信服务配置校验通过')
  } finally {
    testing.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="flex min-h-0 flex-col rounded-card bg-white">
    <!-- 顶部条 -->
    <div class="flex shrink-0 items-center border-b border-line px-5 py-4">
      <div class="flex h-[34px] w-10 shrink-0 items-center justify-center rounded-[10px]" style="background:#E8F7F0;">
        <SvgIcon name="message" :size="20" :stroke-width="1.9" class="text-[#00B578]" />
      </div>
      <div class="ml-3 min-w-0 flex-1">
        <div class="text-[15px] font-bold text-ink-1">短信服务配置</div>
        <div class="mt-[2px] truncate text-[12px] text-ink-3">
          用于验证码短信、通知短信发送；请填写服务商提供的密钥与签名信息
        </div>
      </div>
      <div class="ml-auto flex shrink-0 items-center gap-3">
        <StatusTag :tone="backendMissing ? 'gray' : 'green'">
          {{ backendMissing ? '接口未提供' : '已启用' }}
        </StatusTag>
        <AppButton
          variant="secondary"
          :loading="testing"
          :disabled="backendMissing"
          :title="backendMissing ? '后端短信配置接口尚未提供' : undefined"
          @click="onTest"
        >
          测试连接
        </AppButton>
        <AppButton
          class="shadow-[0_4px_12px_rgba(255,107,44,.3)]"
          :loading="saving"
          :disabled="backendMissing"
          :title="backendMissing ? '后端短信配置接口尚未提供' : undefined"
          @click="onSave"
        >
          保存配置
        </AppButton>
      </div>
    </div>

    <!-- 后端接口未提供提示 -->
    <div
      v-if="backendMissing"
      class="mx-5 mt-4 flex items-center gap-2 rounded-lg px-3 py-2.5 text-[12px]"
      style="background:#FFF8E0;color:#D48806;"
    >
      <SvgIcon name="bell" :size="14" />
      后端短信配置接口（/system/config/sms）尚未提供，当前界面为 UI 预览；后端实现后此页即可保存与测试。
    </div>

    <!-- 表单 -->
    <div v-if="loading" class="py-16 text-center text-[13px] text-ink-3">加载中…</div>
    <div v-else class="grid max-w-3xl grid-cols-2 gap-x-10 gap-y-5 overflow-y-auto p-5">
      <div class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">服务商</span>
        <div class="flex h-9 flex-1 overflow-hidden rounded-lg border border-line">
          <button
            v-for="opt in PROVIDERS"
            :key="opt.value"
            type="button"
            class="flex-1 text-[13px] transition-colors"
            :class="
              form.provider === opt.value
                ? 'bg-[#E8F7F0] font-semibold text-[#00B578]'
                : 'bg-white text-ink-3 hover:text-ink-1'
            "
            @click="form.provider = opt.value"
          >
            {{ opt.label }}
          </button>
        </div>
      </div>
      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">短信签名</span>
        <AppInput v-model="form.sign_name" class="flex-1" placeholder="如 曙光跑腿" />
      </label>

      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">AccessKey</span>
        <AppInput v-model="form.access_key" class="flex-1" placeholder="服务商分配的 AccessKey" />
      </label>
      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">SecretKey</span>
        <AppInput v-model="form.secret_key" type="password" class="flex-1" placeholder="请输入 SecretKey" />
      </label>

      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">模板 CODE</span>
        <AppInput v-model="form.template_code" class="flex-1" placeholder="如 SMS_123456789" />
      </label>

      <p class="col-span-2 -mt-1 pl-[120px] text-[12px] text-ink-4">
        配置保存后，验证码短信与通知短信将通过该服务商发送；可在保存前点击「测试连接」验证密钥有效性。
      </p>
    </div>
  </div>
</template>
