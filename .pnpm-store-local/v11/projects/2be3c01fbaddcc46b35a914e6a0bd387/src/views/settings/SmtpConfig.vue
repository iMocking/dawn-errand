<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import StatusTag from '@/components/StatusTag.vue'
import AppButton from '@/components/AppButton.vue'
import AppInput from '@/components/AppInput.vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { useToast } from '@/composables/useToast'
import {
  getSmtpConfig,
  testSmtpConfig,
  updateSmtpConfig,
  SMTP_PASSWORD_MASK,
  type SmtpEncryption,
} from '@/api/system'

const toast = useToast()

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
/** 后端是否已配置密码（GET 返回掩码）；用于占位提示与「留空保持不变」 */
const hasPassword = ref(false)

const form = reactive({
  host: '',
  port: 465,
  from_email: '',
  password: '',
  encryption: 'ssl' as SmtpEncryption,
  timeout: 10,
})

const ENCRYPTIONS: Array<{ value: SmtpEncryption; label: string }> = [
  { value: 'ssl', label: 'SSL' },
  { value: 'tls', label: 'STARTTLS' },
  { value: 'none', label: '无' },
]

async function load() {
  loading.value = true
  try {
    const cfg = await getSmtpConfig()
    form.host = cfg.host
    form.port = cfg.port || 465
    form.from_email = cfg.from_email
    form.encryption = cfg.encryption || 'ssl'
    hasPassword.value = cfg.password === SMTP_PASSWORD_MASK
    form.password = ''
  } finally {
    loading.value = false
  }
}

function buildPayload() {
  return {
    host: form.host.trim(),
    port: form.port,
    from_email: form.from_email.trim(),
    // 留空 / 掩码：后端保留原密码
    password: form.password,
    encryption: form.encryption,
    timeout: form.timeout,
  }
}

function validate(): boolean {
  if (!form.host.trim()) {
    toast.warning('请输入 SMTP 服务器地址')
    return false
  }
  if (!form.port) {
    toast.warning('请输入端口号')
    return false
  }
  if (!form.from_email.trim()) {
    toast.warning('请输入发件邮箱')
    return false
  }
  return true
}

async function onSave() {
  if (saving.value || !validate()) return
  saving.value = true
  try {
    await updateSmtpConfig(buildPayload())
    toast.success('SMTP 配置已保存')
    await load()
  } finally {
    saving.value = false
  }
}

async function onTest() {
  if (testing.value || !validate()) return
  testing.value = true
  try {
    await testSmtpConfig(buildPayload())
    toast.success('SMTP 连接成功')
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
      <div class="flex h-[34px] w-10 shrink-0 items-center justify-center rounded-[10px]" style="background:#E8F0FF;">
        <SvgIcon name="mail" :size="20" :stroke-width="1.9" class="text-primary" />
      </div>
      <div class="ml-3 min-w-0 flex-1">
        <div class="text-[15px] font-bold text-ink-1">邮箱 SMTP 服务配置</div>
        <div class="mt-[2px] truncate text-[12px] text-ink-3">
          用于注册验证、找回密码等系统邮件发送，配置前请先确认邮箱已开启 SMTP 服务
        </div>
      </div>
      <div class="ml-auto flex shrink-0 items-center gap-3">
        <StatusTag :tone="form.host ? 'green' : 'gray'">
          {{ form.host ? '已启用' : '未启用' }}
        </StatusTag>
        <AppButton variant="secondary" :loading="testing" @click="onTest">测试连接</AppButton>
        <AppButton class="shadow-[0_4px_12px_rgba(255,107,44,.3)]" :loading="saving" @click="onSave">
          保存配置
        </AppButton>
      </div>
    </div>

    <!-- 表单 -->
    <div v-if="loading" class="py-16 text-center text-[13px] text-ink-3">加载中…</div>
    <div v-else class="grid max-w-3xl grid-cols-2 gap-x-10 gap-y-5 overflow-y-auto p-5">
      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">SMTP 服务器</span>
        <AppInput v-model="form.host" class="flex-1" placeholder="如 smtp.qq.com" />
      </label>
      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">端口号</span>
        <AppInput v-model.number="form.port" type="number" class="flex-1" />
      </label>

      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">发件邮箱</span>
        <AppInput v-model="form.from_email" class="flex-1" placeholder="如 noreply@shuguang.cn" />
      </label>
      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">连接超时</span>
        <AppInput v-model.number="form.timeout" type="number" class="flex-1" />
        <span class="text-[12px] text-ink-4">秒</span>
      </label>

      <div class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">加密方式</span>
        <div class="flex h-9 flex-1 overflow-hidden rounded-lg border border-line">
          <button
            v-for="opt in ENCRYPTIONS"
            :key="opt.value"
            type="button"
            class="flex-1 text-[13px] transition-colors"
            :class="
              form.encryption === opt.value
                ? 'bg-[#FFF3E8] font-semibold text-primary'
                : 'bg-white text-ink-3 hover:text-ink-1'
            "
            @click="form.encryption = opt.value"
          >
            {{ opt.label }}
          </button>
        </div>
      </div>
      <label class="flex items-center gap-3 text-[13px] text-ink-2">
        <span class="w-24 shrink-0 text-right">邮箱密码 / 授权码</span>
        <AppInput
          v-model="form.password"
          type="password"
          class="flex-1"
          :placeholder="hasPassword ? '已配置，留空保持不变' : '请输入邮箱授权码'"
        />
      </label>

      <p class="col-span-2 -mt-1 pl-[120px] text-[12px] text-ink-4">
        密码留空或保持原值将沿用已保存的配置；修改后点击「测试连接」可立即验证。
      </p>
    </div>
  </div>
</template>
