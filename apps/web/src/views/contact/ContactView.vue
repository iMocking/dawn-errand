<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { siteApi } from '@/api/site'
import type { ContactData } from '@/types'
import SvgIcon from '@/components/SvgIcon.vue'

const data = ref<ContactData | null>(null)
const error = ref(false)
const toastMsg = ref('')
const form = reactive<Record<string, string>>({})
let toastTimer: ReturnType<typeof setTimeout> | null = null

onMounted(async () => {
  try {
    data.value = await siteApi.contact()
    // 初始化下拉字段默认值
    const intent = data.value.contactSection.form.fields.find((f) => f.type === 'select')
    if (intent?.options?.length) form[intent.name] = intent.options[0]
  } catch {
    error.value = true
  }
})

function showToast(msg: string) {
  toastMsg.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMsg.value = ''
  }, 2400)
}

function onSubmit() {
  if (!data.value) return
  showToast(data.value.contactSection.form.demo)
}
</script>

<template>
  <div v-if="error" class="mx-auto max-w-3xl px-4 py-24 text-center">
    <p class="text-base text-ink-2">内容加载失败，请稍后刷新重试。</p>
  </div>

  <template v-else-if="data">
    <!-- 页头 -->
    <section class="page-head">
      <div class="wrap">
        <span class="kicker">{{ data.pageHead.kicker }}</span>
        <h1>{{ data.pageHead.title }}</h1>
        <p>{{ data.pageHead.desc }}</p>
      </div>
    </section>

    <!-- 合作模式 -->
    <section class="block">
      <div class="wrap">
        <div class="sec-head">
          <h2>{{ data.modes.title }}</h2>
          <p>{{ data.modes.desc }}</p>
        </div>
        <div class="grid-mode">
          <div v-for="item in data.modes.items" :id="item.id" :key="item.title" class="mode">
            <div class="ico" :class="`i-${item.tone}`"><SvgIcon :name="item.icon" :size="22" /></div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.desc }}</p>
            <span class="tag">{{ item.tag }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- 合作流程 -->
    <section class="block bg-soft">
      <div class="wrap">
        <div class="sec-head">
          <h2>{{ data.steps.title }}</h2>
          <p>{{ data.steps.desc }}</p>
        </div>
        <div class="grid-step">
          <div v-for="step in data.steps.items" :key="step.num" class="step">
            <span class="num">{{ step.num }}</span>
            <div>
              <h3>{{ step.title }}</h3>
              <p>{{ step.desc }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 联系我们 -->
    <section class="block">
      <div class="wrap">
        <div class="sec-head">
          <h2>{{ data.contactSection.title }}</h2>
          <p>{{ data.contactSection.desc }}</p>
        </div>
        <div class="contact-layout">
          <div class="contact-card">
            <div v-for="card in data.contactSection.cards" :key="card.label" class="contact-row">
              <span class="ico"><SvgIcon :name="card.icon" :size="20" /></span>
              <span>
                <span class="lb">{{ card.label }}</span>
                <span class="vl" :class="{ sub: card.icon === 'chat' || card.icon === 'clock' }">
                  {{ card.value }}
                </span>
              </span>
            </div>
          </div>
          <form class="form-card" @submit.prevent="onSubmit">
            <div v-for="field in data.contactSection.form.fields" :key="field.name" class="field">
              <label :for="field.name">{{ field.label }}</label>
              <select
                v-if="field.type === 'select'"
                :id="field.name"
                v-model="form[field.name]"
                :name="field.name"
              >
                <option v-for="opt in field.options" :key="opt" :value="opt">{{ opt }}</option>
              </select>
              <textarea
                v-else-if="field.type === 'textarea'"
                :id="field.name"
                v-model="form[field.name]"
                :name="field.name"
                :placeholder="field.placeholder"
                rows="3"
              ></textarea>
              <input
                v-else
                :id="field.name"
                v-model="form[field.name]"
                :name="field.name"
                type="text"
                :placeholder="field.placeholder"
              />
            </div>
            <button type="submit" class="btn btn-primary">
              {{ data.contactSection.form.submitLabel }}
            </button>
            <div class="form-tip">{{ data.contactSection.form.submitTip }}</div>
          </form>
        </div>

        <div class="notice">
          <b>说明：</b>{{ data.notice }}
        </div>
      </div>
    </section>

    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="translate-y-4 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="translate-y-4 opacity-0"
    >
      <div v-if="toastMsg" class="toast" role="status">{{ toastMsg }}</div>
    </transition>
  </template>
</template>

<style scoped>
.wrap {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}
.block {
  padding: 44px 0;
}
.bg-soft {
  background: #fff;
}
.sec-head {
  margin-bottom: 26px;
}
.sec-head h2 {
  font-size: 20px;
  font-weight: 800;
  color: #1d2129;
}
.sec-head p {
  margin-top: 8px;
  font-size: 13px;
  color: #86909c;
}

/* 页头 */
.page-head {
  background: linear-gradient(150deg, #ff8a3d 0%, #ff6b2c 60%, #f2531c 100%);
  color: #fff;
  padding: 52px 0 48px;
  position: relative;
  overflow: hidden;
}
.page-head::before {
  content: '';
  position: absolute;
  right: -90px;
  top: -90px;
  width: 280px;
  height: 280px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
}
.page-head .kicker {
  position: relative;
  display: inline-block;
  font-size: 12px;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.18);
  padding: 5px 12px;
  border-radius: 20px;
  letter-spacing: 1px;
}
.page-head h1 {
  position: relative;
  margin-top: 14px;
  font-size: 28px;
  font-weight: 800;
  letter-spacing: 0.5px;
}
.page-head p {
  position: relative;
  margin-top: 10px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.9);
  max-width: 560px;
}

/* 合作模式 */
.grid-mode {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
}
.mode {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  padding: 22px;
  display: flex;
  flex-direction: column;
}
.mode .ico {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.mode h3 {
  font-size: 15px;
  font-weight: 700;
  margin-top: 14px;
}
.mode p {
  margin-top: 8px;
  font-size: 13px;
  color: #4e5969;
  line-height: 1.7;
  flex: 1;
}
.mode .tag {
  margin-top: 12px;
  font-size: 11px;
  color: #86909c;
  background: #f5f6f8;
  border: 1px solid #e5e6eb;
  display: inline-block;
  padding: 3px 10px;
  border-radius: 6px;
}
.i-orange {
  background: #fff3e8;
  color: #ff6b2c;
}
.i-blue {
  background: #eef4ff;
  color: #2b7fff;
}
.i-green {
  background: #e8f9f0;
  color: #00b578;
}
.i-purple {
  background: #f4eeff;
  color: #722ed1;
}

/* 合作流程 */
.grid-step {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0;
}
.step {
  display: flex;
  gap: 16px;
  position: relative;
  padding-bottom: 22px;
}
.step::before {
  content: '';
  position: absolute;
  left: 17px;
  top: 40px;
  bottom: 0;
  width: 2px;
  background: #e5e6eb;
}
.step:last-child::before {
  display: none;
}
.step .num {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #ff8a3d, #ff6b2c);
  color: #fff;
  font-size: 15px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}
.step h3 {
  font-size: 15px;
  font-weight: 700;
  color: #1d2129;
}
.step p {
  margin-top: 5px;
  font-size: 13px;
  color: #4e5969;
  line-height: 1.7;
}

/* 联系方式 */
.contact-layout {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
  align-items: start;
}
.contact-card {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  padding: 22px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.contact-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
.contact-row .ico {
  width: 40px;
  height: 40px;
  border-radius: 9px;
  background: #fff3e8;
  color: #ff6b2c;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.contact-row .lb {
  display: block;
  font-size: 12px;
  color: #86909c;
}
.contact-row .vl {
  display: block;
  font-size: 14px;
  color: #1d2129;
  font-weight: 600;
  margin-top: 2px;
}
.contact-row .vl.sub {
  font-size: 13px;
  color: #4e5969;
  font-weight: 400;
}

/* 表单 */
.form-card {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  padding: 22px;
}
.form-card label {
  display: block;
  font-size: 13px;
  color: #4e5969;
  margin-bottom: 6px;
}
.form-card .field {
  margin-bottom: 16px;
}
.form-card input,
.form-card textarea,
.form-card select {
  width: 100%;
  height: 46px;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  padding: 0 14px;
  font-size: 16px;
  font-family: inherit;
  color: #1d2129;
  background: #fff;
  outline: none;
}
.form-card textarea {
  height: 110px;
  padding: 12px 14px;
  resize: vertical;
}
.form-card input:focus,
.form-card textarea:focus,
.form-card select:focus {
  border-color: #ff6b2c;
  box-shadow: 0 0 0 3px rgba(255, 107, 44, 0.12);
}
.form-card .btn {
  width: 100%;
  height: 46px;
  font-size: 15px;
}
.form-tip {
  font-size: 12px;
  color: #86909c;
  margin-top: 12px;
  text-align: center;
}

/* 说明 */
.notice {
  margin-top: 24px;
  background: #fff3e8;
  border: 1px solid #ffd9bd;
  border-radius: 14px;
  padding: 18px 20px;
  font-size: 13px;
  color: #4e5969;
  line-height: 1.8;
}
.notice b {
  color: #ff6b2c;
}

/* 按钮 */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: none;
  cursor: pointer;
  font-family: inherit;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  transition:
    filter 0.15s,
    box-shadow 0.15s,
    background 0.15s;
}
.btn:active {
  transform: translateY(1px);
}
.btn-primary {
  background: linear-gradient(135deg, #ff8a3d, #ff6b2c);
  color: #fff;
  box-shadow: 0 6px 16px rgba(255, 107, 44, 0.32);
}
.btn-primary:hover {
  filter: brightness(1.05);
  box-shadow: 0 8px 20px rgba(255, 107, 44, 0.4);
}

/* toast */
.toast {
  position: fixed;
  left: 50%;
  bottom: 28px;
  transform: translateX(-50%);
  background: #1d2129;
  color: #fff;
  font-size: 13px;
  padding: 11px 18px;
  border-radius: 10px;
  z-index: 100;
  max-width: 88vw;
  text-align: center;
}

@media (min-width: 768px) {
  .wrap {
    padding: 0 32px;
  }
  .page-head {
    padding: 64px 0 56px;
  }
  .page-head h1 {
    font-size: 34px;
  }
  .grid-mode {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .grid-mode {
    grid-template-columns: repeat(4, 1fr);
  }
  .grid-step {
    grid-template-columns: repeat(4, 1fr);
    gap: 18px;
  }
  .step {
    flex-direction: column;
    padding-bottom: 0;
  }
  .step::before {
    left: 18px;
    top: 44px;
    bottom: -14px;
    width: 2px;
  }
  .step:last-child::before {
    display: none;
  }
  .contact-layout {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
