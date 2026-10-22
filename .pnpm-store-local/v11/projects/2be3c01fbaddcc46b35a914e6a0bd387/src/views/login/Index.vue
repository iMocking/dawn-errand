<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const toast = useToast()

const account = ref('admin@shuguang')
const password = ref('shuguang2026')
const captcha = ref('8A6K')
const loading = ref(false)

async function onSubmit() {
  if (loading.value) return
  if (!account.value || !password.value || !captcha.value) {
    toast.warning('请填写账号、密码和验证码')
    return
  }
  loading.value = true
  try {
    await auth.login(account.value, password.value, captcha.value)
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.push(redirect)
  } catch {
    // request 层已提示
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen bg-page">
    <!-- 左侧品牌区 -->
    <section
      class="relative flex w-[560px] shrink-0 flex-col overflow-hidden px-12 pb-10 pt-12"
      style="background:linear-gradient(150deg,#FF8A3D 0%,#FF6B2C 55%,#F2531C 100%);"
    >
      <div class="absolute -right-[140px] -top-[140px] h-[420px] w-[420px] rounded-full" style="background:rgba(255,255,255,.10);" />
      <div class="absolute -right-10 -top-10 h-[220px] w-[220px] rounded-full" style="background:rgba(255,255,255,.14);" />
      <div class="absolute -left-[90px] bottom-[120px] h-[260px] w-[260px] rounded-full" style="background:rgba(255,255,255,.08);" />

      <!-- Logo 行 -->
      <div class="relative z-[2] flex items-center gap-3">
        <div class="flex h-11 w-11 items-center justify-center rounded-xl" style="background:rgba(255,255,255,.22);">
          <SvgIcon name="clock" :size="26" :stroke-width="1.9" class="text-white" />
        </div>
        <div>
          <div class="text-[22px] font-bold tracking-wider text-white">曙光跑腿</div>
          <div class="mt-0.5 text-[12px]" style="color:rgba(255,255,255,.85);">平台运营管理端</div>
        </div>
      </div>

      <!-- 标语 -->
      <div class="relative z-[2] mt-24">
        <div class="text-[30px] font-bold leading-[1.45] tracking-wider text-white">
          让每个村庄的订单<br />都被认真送达
        </div>
        <div class="mt-[18px] text-[14px] leading-[1.8]" style="color:rgba(255,255,255,.88);">
          覆盖乡镇农村的跑腿与外卖一体化平台，统一管理<br />订单、商家、骑手与资金结算。
        </div>
      </div>

      <!-- 数据条 -->
      <div class="relative z-[2] mt-auto flex gap-8">
        <div>
          <div class="text-[26px] font-bold text-white">12,800+</div>
          <div class="mt-1 text-[12px]" style="color:rgba(255,255,255,.8);">覆盖乡镇</div>
        </div>
        <div class="h-10 w-px" style="background:rgba(255,255,255,.3);" />
        <div>
          <div class="text-[26px] font-bold text-white">8.6万+</div>
          <div class="mt-1 text-[12px]" style="color:rgba(255,255,255,.8);">服务用户</div>
        </div>
        <div class="h-10 w-px" style="background:rgba(255,255,255,.3);" />
        <div>
          <div class="text-[26px] font-bold text-white">4,320+</div>
          <div class="mt-1 text-[12px]" style="color:rgba(255,255,255,.8);">合作骑手</div>
        </div>
      </div>
    </section>

    <!-- 右侧登录区 -->
    <section class="flex flex-1 items-center justify-center bg-page">
      <div class="w-[420px]">
        <div class="text-[26px] font-bold text-ink-1">欢迎回来</div>
        <div class="mt-2 text-[14px] text-ink-3">登录曙光跑腿平台运营管理端</div>

        <div class="mt-8">
          <label class="mb-2 block text-[13px] text-ink-2">账号</label>
          <div class="flex h-12 items-center rounded-lg border border-line bg-white px-[14px]">
            <SvgIcon name="user" :size="20" :stroke-width="1.8" class="shrink-0 text-ink-3" />
            <input
              v-model="account"
              class="ml-2.5 flex-1 border-none text-[14px] text-ink-1 outline-none"
              placeholder="请输入管理员账号"
            />
          </div>

          <label class="mb-2 mt-5 block text-[13px] text-ink-2">密码</label>
          <div class="flex h-12 items-center rounded-lg border border-line bg-white px-[14px]">
            <SvgIcon name="lock" :size="20" :stroke-width="1.8" class="shrink-0 text-ink-3" />
            <input
              v-model="password"
              type="password"
              class="ml-2.5 flex-1 border-none text-[14px] text-ink-1 outline-none"
              placeholder="请输入密码"
            />
            <SvgIcon name="eye" :size="20" :stroke-width="1.8" class="shrink-0 cursor-pointer text-ink-3" />
          </div>

          <label class="mb-2 mt-5 block text-[13px] text-ink-2">验证码</label>
          <div class="flex gap-3">
            <div class="flex h-12 flex-1 items-center rounded-lg border border-line bg-white px-[14px]">
              <SvgIcon name="captcha" :size="20" :stroke-width="1.8" class="shrink-0 text-ink-3" />
              <input
                v-model="captcha"
                class="ml-2.5 flex-1 border-none text-[14px] tracking-[4px] text-ink-1 outline-none"
                placeholder="请输入验证码"
              />
            </div>
            <div class="flex h-12 w-32 shrink-0 items-center justify-center rounded-lg" style="background:#FFF3E8;">
              <span class="text-[16px] font-bold tracking-[3px] text-primary">8A6K</span>
            </div>
          </div>

          <div class="mt-5 flex items-center justify-between">
            <div class="flex cursor-pointer items-center gap-2">
              <div class="flex h-4 w-4 items-center justify-center rounded" style="background:#FF6B2C;">
                <SvgIcon name="checkbox" :size="10" :stroke-width="3" class="text-white" />
              </div>
              <span class="text-[13px] text-ink-2">记住账号</span>
            </div>
            <span class="cursor-pointer text-[13px] text-primary">忘记密码？</span>
          </div>

          <button
            class="mt-7 h-12 w-full rounded-lg border-none text-[16px] font-semibold text-white disabled:opacity-70"
            style="background:linear-gradient(135deg,#FF8A3D,#FF6B2C);box-shadow:0 6px 16px rgba(255,107,44,.35);"
            :disabled="loading"
            @click="onSubmit"
          >
            {{ loading ? '登录中…' : '登 录' }}
          </button>
        </div>

        <div class="mt-8 flex items-center gap-3">
          <div class="h-px flex-1 bg-line" />
          <span class="text-[12px] text-ink-3">安全提示：请勿在公共设备上保存密码</span>
          <div class="h-px flex-1 bg-line" />
        </div>
      </div>
    </section>
  </div>
</template>
