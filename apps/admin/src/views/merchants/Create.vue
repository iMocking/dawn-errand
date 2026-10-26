<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppButton from '@/components/AppButton.vue'
import AppInput from '@/components/AppInput.vue'
import ImageUpload from '@/components/ImageUpload.vue'
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
import { useToast } from '@/composables/useToast'
import { useWorkspaceStore } from '@/stores/workspace'
import {
  createMerchant,
  getMerchantDetail,
  updateMerchantApplication,
  uploadMerchantImage,
  type CreateMerchantPayload,
  type MerchantApplication,
} from '@/api/merchants'
import { getSettings } from '@/api/settings'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const workspace = useWorkspaceStore()

const editId = computed(() => (typeof route.params.id === 'string' ? route.params.id : ''))
const isEdit = computed(() => !!editId.value)
const loading = ref(false)
const submitting = ref(false)
const uploadingStore = ref(false)
const regions = ref<{ id: string; name: string }[]>([])
const rejectReason = ref('')

const CATEGORIES = ['餐饮', '商超', '药店', '其他']

function emptyApplication(): MerchantApplication {
  return {
    licenseNo: '', licenseImage: '', legalName: '', legalIdCard: '',
    idCardFront: '', idCardBack: '', foodLicenseNo: '', foodLicenseImage: '',
    storefrontImages: [], bankName: '', bankAccountName: '', bankAccountNo: '',
    accountEmail: '', accountPhone: '', notes: '',
  }
}

const form = reactive({
  name: '',
  category: '餐饮',
  contact: '',
  phone: '',
  servicePhone: '',
  regionId: '',
  address: '',
  addressSub: '',
  openHours: '',
  lat: 0,
  lng: 0,
  deliveryRange: '',
  deliveryMethod: '平台配送',
  commission: 12,
  minPrice: 0,
  packingFee: 0,
  deliveryFee: 0,
  autoAccept: false,
  announcement: '',
  application: emptyApplication(),
})

function numberOrZero(v: unknown): number {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}

async function load() {
  loading.value = true
  try {
    const settings = await getSettings()
    regions.value = settings.areas.map((a) => ({ id: a.id, name: a.name }))
    if (isEdit.value) {
      const d = await getMerchantDetail(editId.value)
      form.name = d.name
      form.category = d.category || '餐饮'
      form.contact = d.basic.owner
      form.phone = d.basic.phone
      form.address = d.basic.address
      form.openHours = d.basic.openHours
      form.deliveryRange = d.basic.deliveryRange
      form.commission = Number(d.basic.commission.replace('%', '')) || 0
      form.deliveryMethod = d.basic.deliveryMethod
      form.servicePhone = d.application.accountPhone || d.basic.phone
      form.application = { ...emptyApplication(), ...(d.application ?? {}) }
      form.application.storefrontImages = d.application?.storefrontImages ?? []
      form.application.accountPhone = d.application?.accountPhone || d.basic.phone
      rejectReason.value = d.rejectReason || ''
      form.regionId = regions.value[0]?.id ?? ''
    } else if (regions.value.length > 0) {
      form.regionId = regions.value[0].id
    }
  } finally {
    loading.value = false
  }
}

async function onStorefrontPick(e: Event) {
  const files = Array.from((e.target as HTMLInputElement).files ?? [])
  if (files.length === 0) return
  if (form.application.storefrontImages.length + files.length > 6) {
    toast.warning('门头/店内照片最多 6 张')
    return
  }
  uploadingStore.value = true
  try {
    for (const file of files) {
      if (!file.type.startsWith('image/')) continue
      const res = await uploadMerchantImage(file, 'store')
      form.application.storefrontImages.push(res.url)
    }
  } finally {
    uploadingStore.value = false
    ;(e.target as HTMLInputElement).value = ''
  }
}

function removeStorefront(index: number) {
  form.application.storefrontImages.splice(index, 1)
}

function validate(): string {
  const a = form.application
  const checks: [boolean, string][] = [
    [!!form.name.trim(), '商家名称'],
    [!!form.category, '经营类型'],
    [!!form.contact.trim(), '联系人'],
    [!!form.phone.trim(), '联系电话'],
    [!!form.regionId, '服务区域'],
    [!!form.address.trim(), '详细地址'],
    [!!form.openHours.trim(), '营业时间'],
    [!!a.licenseNo.trim(), '营业执照号'],
    [!!a.licenseImage, '营业执照照片'],
    [!!a.legalName.trim(), '法人姓名'],
    [!!a.legalIdCard.trim(), '法人身份证号'],
    [!!a.idCardFront, '身份证人像面'],
    [!!a.idCardBack, '身份证国徽面'],
    [a.storefrontImages.length > 0, '门头/店内照片'],
    [!!a.bankName.trim(), '开户银行'],
    [!!a.bankAccountName.trim(), '开户名'],
    [!!a.bankAccountNo.trim(), '结算账号'],
  ]
  if (form.category === '餐饮') {
    checks.push([!!a.foodLicenseNo.trim(), '食品经营许可证号'])
    checks.push([!!a.foodLicenseImage, '食品经营许可证照片'])
  }
  const miss = checks.filter(([ok]) => !ok).map(([, label]) => label)
  if (miss.length) return `请补充：${miss.join('、')}`
  if (form.commission < 0 || form.commission > 100) return '佣金比例需在 0-100 之间'
  return ''
}

function payload(): CreateMerchantPayload {
  const a = { ...form.application }
  a.storefrontImages = [...a.storefrontImages]
  if (!a.accountPhone) a.accountPhone = form.phone.trim()
  return {
    name: form.name.trim(),
    category: form.category,
    contact: form.contact.trim(),
    phone: form.phone.trim(),
    address: form.address.trim(),
    addressSub: form.addressSub.trim(),
    openHours: form.openHours.trim(),
    deliveryRange: form.deliveryRange.trim(),
    commission: numberOrZero(form.commission),
    deliveryMethod: form.deliveryMethod.trim(),
    regionId: form.regionId,
    lat: numberOrZero(form.lat),
    lng: numberOrZero(form.lng),
    autoAccept: form.autoAccept,
    minPrice: numberOrZero(form.minPrice),
    deliveryFee: numberOrZero(form.deliveryFee),
    packingFee: numberOrZero(form.packingFee),
    announcement: form.announcement.trim(),
    servicePhone: form.servicePhone.trim(),
    application: a,
  }
}

async function onSubmit() {
  if (submitting.value) return
  const err = validate()
  if (err) {
    toast.warning(err)
    return
  }
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateMerchantApplication(editId.value, payload())
      toast.success('资料已保存并重新提交审核')
    } else {
      const res = await createMerchant(payload())
      toast.success(`已提交审核 · 申请单号 ${res.applyNo}`)
    }
    workspace.refresh()
    router.push('/merchants/audit')
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="mx-auto max-w-[1100px] pb-24">
    <div class="mb-4 flex items-center gap-3">
      <button
        type="button"
        class="flex h-8 w-8 items-center justify-center rounded-lg text-ink-3 transition-colors hover:bg-white hover:text-ink-1"
        @click="router.back()"
      >
        <SvgIcon name="chevron-right" :size="16" class="rotate-180" />
      </button>
      <div>
        <div class="text-[18px] font-bold text-ink-1">{{ isEdit ? '编辑入驻资料' : '新增商家' }}</div>
        <div class="mt-0.5 text-[12px] text-ink-3">
          业务员线下采集完整经营与资质信息，提交后进入平台审核
        </div>
      </div>
      <span
        v-if="isEdit"
        class="ml-auto rounded-full bg-[#FFF8E0] px-3 py-1 text-[12px] font-semibold"
        style="color:#D48806;"
      >补件重提</span>
    </div>

    <div v-if="rejectReason" class="mb-4 rounded-xl border border-[#FFE0E0] bg-[#FFF7F7] px-4 py-3 text-[13px] text-[#FF4D4F]">
      上次驳回原因：{{ rejectReason }}
    </div>

    <div v-if="loading" class="rounded-card bg-white py-16 text-center text-ink-3">加载中…</div>
    <div v-else class="flex flex-col gap-4">
      <!-- 基础信息 -->
      <section class="rounded-card bg-white p-5">
        <div class="mb-4 flex items-center gap-2">
          <span class="h-4 w-1 rounded-full bg-primary" />
          <span class="text-[15px] font-bold text-ink-1">基础信息</span>
          <span class="text-[12px] text-ink-4">用于商家列表、订单展示与配送</span>
        </div>
        <div class="grid grid-cols-3 gap-x-4 gap-y-4">
          <div class="col-span-2">
            <div class="mb-1.5 text-[13px] text-ink-2">商家名称 <span class="text-[#FF4D4F]">*</span></div>
            <AppInput v-model="form.name" placeholder="如：村口老陈家常菜" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">经营类型 <span class="text-[#FF4D4F]">*</span></div>
            <div class="flex flex-wrap gap-1.5">
              <button
                v-for="c in CATEGORIES"
                :key="c"
                type="button"
                class="rounded-md px-3 py-1.5 text-[12px]"
                :class="form.category === c ? 'bg-primary-bg font-semibold text-primary' : 'bg-head text-ink-2'"
                @click="form.category = c"
              >{{ c }}</button>
            </div>
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">联系人 <span class="text-[#FF4D4F]">*</span></div>
            <AppInput v-model="form.contact" placeholder="负责人姓名" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">联系电话 <span class="text-[#FF4D4F]">*</span></div>
            <AppInput v-model="form.phone" placeholder="用于商家端登录" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">客服电话</div>
            <AppInput v-model="form.servicePhone" placeholder="默认同联系电话" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">服务区域 <span class="text-[#FF4D4F]">*</span></div>
            <select v-model="form.regionId" class="h-9 w-full rounded-lg border border-line bg-white px-3 text-[13px] text-ink-1 outline-none focus:border-primary">
              <option v-for="r in regions" :key="r.id" :value="r.id">{{ r.name }}</option>
            </select>
          </div>
          <div class="col-span-2">
            <div class="mb-1.5 text-[13px] text-ink-2">详细地址 <span class="text-[#FF4D4F]">*</span></div>
            <AppInput v-model="form.address" placeholder="街道 / 门牌号" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">地址补充</div>
            <AppInput v-model="form.addressSub" placeholder="如：距镇政府 300m" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">营业时间 <span class="text-[#FF4D4F]">*</span></div>
            <AppInput v-model="form.openHours" placeholder="如 08:00-22:00" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">纬度</div>
            <AppInput v-model="form.lat" type="number" placeholder="如 30.123456" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">经度</div>
            <AppInput v-model="form.lng" type="number" placeholder="如 120.123456" clearable />
          </div>
        </div>
      </section>

      <!-- 经营设置 -->
      <section class="rounded-card bg-white p-5">
        <div class="mb-4 flex items-center gap-2">
          <span class="h-4 w-1 rounded-full bg-primary" />
          <span class="text-[15px] font-bold text-ink-1">经营设置</span>
          <span class="text-[12px] text-ink-4">审核通过后商家可登录调整</span>
        </div>
        <div class="grid grid-cols-4 gap-x-4 gap-y-4">
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">配送方式</div>
            <AppInput v-model="form.deliveryMethod" placeholder="平台配送 / 自配送" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">配送范围</div>
            <AppInput v-model="form.deliveryRange" placeholder="如 3km / 全镇" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">佣金比例（%）</div>
            <AppInput v-model="form.commission" type="number" placeholder="如 12" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">起送价（元）</div>
            <AppInput v-model="form.minPrice" type="number" placeholder="0" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">打包费（元）</div>
            <AppInput v-model="form.packingFee" type="number" placeholder="0" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">配送费（元）</div>
            <AppInput v-model="form.deliveryFee" type="number" placeholder="0" clearable />
          </div>
          <div class="col-span-2 flex items-end">
            <button
              type="button"
              class="flex h-9 items-center gap-2 rounded-lg px-3 text-[13px]"
              :class="form.autoAccept ? 'bg-[#E8F7F0] text-[#00B578]' : 'bg-head text-ink-3'"
              @click="form.autoAccept = !form.autoAccept"
            >
              <span class="relative h-[18px] w-[32px] rounded-full" :style="{ background: form.autoAccept ? '#00B578' : '#C9CDD4' }">
                <span class="absolute top-[2px] h-[14px] w-[14px] rounded-full bg-white transition-all" :style="form.autoAccept ? { right: '2px' } : { left: '2px' }" />
              </span>
              自动接单
            </button>
          </div>
          <div class="col-span-4">
            <div class="mb-1.5 text-[13px] text-ink-2">店铺公告</div>
            <textarea
              v-model="form.announcement"
              rows="2"
              placeholder="选填，展示在商家端与用户端门店页"
              class="w-full resize-none rounded-lg border border-line bg-white px-3 py-2 text-[13px] outline-none focus:border-primary"
            />
          </div>
        </div>
      </section>

      <!-- 资质认证 -->
      <section class="rounded-card bg-white p-5">
        <div class="mb-4 flex items-center gap-2">
          <span class="h-4 w-1 rounded-full bg-primary" />
          <span class="text-[15px] font-bold text-ink-1">资质认证</span>
          <span class="text-[12px] text-ink-4">需与营业执照、法人证件一致，平台审核后生效</span>
        </div>
        <div class="grid grid-cols-3 gap-4">
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">营业执照号 <span class="text-[#FF4D4F]">*</span></div>
            <AppInput v-model="form.application.licenseNo" placeholder="统一社会信用代码" clearable />
          </div>
          <div class="col-span-2 row-span-2">
            <div class="mb-1.5 text-[13px] text-ink-2">营业执照照片 <span class="text-[#FF4D4F]">*</span></div>
            <ImageUpload v-model="form.application.licenseImage" label="营业执照" ratio="wide" purpose="license" hint="支持 JPG/PNG，10MB 以内 · 存私有桶" />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">法人姓名 <span class="text-[#FF4D4F]">*</span></div>
            <AppInput v-model="form.application.legalName" placeholder="与身份证一致" clearable />
          </div>
          <div class="col-span-2">
            <div class="mb-1.5 text-[13px] text-ink-2">法人身份证号 <span class="text-[#FF4D4F]">*</span></div>
            <AppInput v-model="form.application.legalIdCard" placeholder="18 位身份证号" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">身份证人像面 <span class="text-[#FF4D4F]">*</span></div>
            <ImageUpload v-model="form.application.idCardFront" label="人像面" purpose="idcard" hint="需清晰可见证件号 · 存私有桶" />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">身份证国徽面 <span class="text-[#FF4D4F]">*</span></div>
            <ImageUpload v-model="form.application.idCardBack" label="国徽面" purpose="idcard" />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">门头/店内照片 <span class="text-[#FF4D4F]">*</span></div>
            <label class="flex h-20 cursor-pointer flex-col items-center justify-center gap-1 rounded-lg border border-dashed border-[#D0D3D9] bg-head text-ink-3 hover:border-primary hover:text-primary">
              <input type="file" accept="image/*" multiple class="hidden" @change="onStorefrontPick" />
              <SvgIcon name="plus" :size="18" />
              <span class="text-[12px]">{{ uploadingStore ? '上传中…' : '上传门头照' }}</span>
            </label>
            <div v-if="form.application.storefrontImages.length" class="mt-2 flex flex-wrap gap-1.5">
              <div v-for="(img, i) in form.application.storefrontImages" :key="img" class="group relative">
                <img :src="img" alt="" class="h-12 w-12 rounded-md object-cover" />
                <button type="button" class="absolute -right-1 -top-1 hidden h-4 w-4 items-center justify-center rounded-full bg-black/60 text-[10px] text-white group-hover:flex" @click="removeStorefront(i)">×</button>
              </div>
            </div>
          </div>

          <template v-if="form.category === '餐饮'">
            <div>
              <div class="mb-1.5 text-[13px] text-ink-2">食品经营许可证号 <span class="text-[#FF4D4F]">*</span></div>
              <AppInput v-model="form.application.foodLicenseNo" placeholder="许可证编号" clearable />
            </div>
            <div class="col-span-2">
              <div class="mb-1.5 text-[13px] text-ink-2">食品经营许可证照片 <span class="text-[#FF4D4F]">*</span></div>
              <ImageUpload v-model="form.application.foodLicenseImage" label="食品许可证" ratio="wide" purpose="cert" />
            </div>
          </template>

          <div class="col-span-3">
            <div class="mb-1.5 text-[13px] text-ink-2">补充说明</div>
            <textarea
              v-model="form.application.notes"
              rows="2"
              placeholder="选填，如经营范围、特殊资质说明"
              class="w-full resize-none rounded-lg border border-line bg-white px-3 py-2 text-[13px] outline-none focus:border-primary"
            />
          </div>
        </div>
      </section>

      <!-- 结算与账号 -->
      <section class="rounded-card bg-white p-5">
        <div class="mb-4 flex items-center gap-2">
          <span class="h-4 w-1 rounded-full bg-primary" />
          <span class="text-[15px] font-bold text-ink-1">结算与商家端账号</span>
          <span class="text-[12px] text-ink-4">审核通过后自动开通，初始密码仅展示一次</span>
        </div>
        <div class="grid grid-cols-3 gap-4">
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">开户银行 <span class="text-[#FF4D4F]">*</span></div>
            <AppInput v-model="form.application.bankName" placeholder="如 工商银行" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">开户名 <span class="text-[#FF4D4F]">*</span></div>
            <AppInput v-model="form.application.bankAccountName" placeholder="与营业执照/法人一致" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">结算账号 <span class="text-[#FF4D4F]">*</span></div>
            <AppInput v-model="form.application.bankAccountNo" placeholder="银行卡号" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">登录手机号</div>
            <AppInput v-model="form.application.accountPhone" placeholder="默认同联系电话" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">登录邮箱</div>
            <AppInput v-model="form.application.accountEmail" placeholder="选填，与手机号二选一" clearable />
          </div>
          <div>
            <div class="mb-1.5 text-[13px] text-ink-2">初始密码</div>
            <div class="flex h-9 items-center rounded-lg bg-head px-3 text-[12px] text-ink-4">
              留空由系统生成，审核通过后一次性展示
            </div>
          </div>
        </div>
        <div class="mt-3 rounded-lg bg-[#FFF8E0] px-3 py-2.5 text-[12px]" style="color:#D48806;">
          审核通过后自动创建商家端账号并绑定该商家；商家使用手机号/邮箱 + 初始密码登录，首次登录后请修改密码。
        </div>
      </section>
    </div>

    <!-- 底部操作 -->
    <div class="fixed inset-x-0 bottom-0 z-20 border-t border-line bg-white/95 px-6 py-3 backdrop-blur">
      <div class="mx-auto flex max-w-[1100px] items-center">
        <span class="text-[12px] text-ink-4">带 <span class="text-[#FF4D4F]">*</span> 为必填；提交后资料进入审核，审核通过前不可登录经营。</span>
        <div class="ml-auto flex gap-3">
          <AppButton variant="secondary" @click="router.back()">取消</AppButton>
          <AppButton :loading="submitting" @click="onSubmit">{{ isEdit ? '保存并重新提交' : '提交入驻审核' }}</AppButton>
        </div>
      </div>
    </div>
  </div>
</template>
