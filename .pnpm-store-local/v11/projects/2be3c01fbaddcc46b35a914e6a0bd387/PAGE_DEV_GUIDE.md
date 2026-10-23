# 曙光跑腿 · 后台管理端 开发指南（PAGE_DEV_GUIDE）

> 本文件由骨架 Agent 产出，供后续 6 个并行分片（交易/商家/骑手/用户/资金/系统）页面 Agent 直接遵守。
> **必须先读本文件再写页面。** 骨架已完成登录页 `/login` 与数据总览 `/dashboard`，其余页面按本指南新增。

---

## 0. 启动与验收（先跑通环境）

```bash
cd apps/admin
pnpm install
pnpm dev          # http://localhost:5173
pnpm typecheck    # vue-tsc --noEmit
pnpm lint         # eslint .
pnpm build        # 产物 dist/
```

- 默认账号：`admin@shuguang` / 密码 `shuguang2026` / 验证码 `8A6K`（任意值也可登录，dev mock 不校验）。
- 真实后端（proxy 模式）种子超管：`jsongkey@outlook.com` / `1076236bb`（当前库内，已于 2026-10-23 由 seed 输出的一次性账号变更而来；若重置 seed 则恢复为 `admin@shuguang.local` + 一次性密码）；验证码后端不校验，任意非空即可。
- 提/交/自测前 `pnpm typecheck && pnpm lint && pnpm build` 必须全绿；自测完 dev server 必须停止，**不要占用 8080**。

---

## 1. 设计 Token（逐字还原，禁止自造色值）

| 类别 | 值 |
|---|---|
| 主色 | `#FF6B2C`；渐变 `linear-gradient(135deg,#FF8A3D,#FF6B2C)`；深 `#F2531C`；浅橙底 `#FFF3E8` |
| 页面底 | `#F5F6F8`；卡片白 `#FFFFFF`，圆角 `12px`，细边框 `1px solid #E5E6EB` |
| 文字 | 主 `#1D2129` / 次 `#4E5969` / 弱 `#86909C` / 更弱 `#C9CDD4` |
| 边框 | `#E5E6EB`；表头底 `#F7F8FA`；分割线浅 `#F2F3F5` |
| 语义色 | 红 `#FF4D4F` / 蓝 `#2B7FFF` / 绿 `#00B578` / 黄 `#D48806` / 紫 `#722ED1` |
| 标签圆角 | `4px`；按钮/输入圆角 `8px` |
| 字体 | `-apple-system, PingFang SC, Microsoft YaHei, sans-serif` |
| 图标 | 一律内联 SVG 线框 `stroke-width=2`、`stroke=currentColor`，**禁止 emoji** |

Tailwind 已注册好语义色：`text-primary` / `bg-primary-bg` / `text-ink-1..4` / `bg-page` / `border-line` / `bg-head` / `text-sem-red|blue|green|yellow|purple`。渐变仍用内联 `style`。

---

## 2. 布局、路由与菜单约定

### 2.1 Layout
- 业务页全部自动套 `layouts/Index.vue`（左侧 224px Sidebar + 顶栏 56px + 内容区）。
- 页面文件放 `src/views/<模块>/Index.vue`。
- **不要**在页面里写侧边栏/顶栏；只写内容区。内容区外层 `padding:20px 24px; display:flex; flex-direction:column; gap:16px;`（dashboard 已是范例）。

### 2.2 新增一个页面路由（契约）
1. 在 `src/router/routes/` 新建你模块的文件，例如 `orders.ts`：

```ts
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/orders',
    name: 'orders',
    component: () => import('@/views/orders/Index.vue'),
    meta: {
      title: '订单管理',          // 菜单名 + 面包屑
      icon: 'order',             // SvgIcon 注册名
      group: '交易管理',          // 概览/交易管理/主体管理/资金管理/系统管理
      badge: 2,                  // 可选角标
      badgeColor: 'red',         // red | orange | yellow
      permission: 'order:list',  // 可选；不传=登录可见
    },
  },
  {
    path: '/orders/:id',
    name: 'order-detail',
    component: () => import('@/views/orders/Detail.vue'),
    meta: { title: '订单详情', hidden: true, parent: '/orders' }, // hidden 不进菜单
  },
]
export default routes
```

2. **不要**改 `router/routes/index.ts` —— 它用 `import.meta.glob('./*.ts',{eager:true})` 自动聚合。
3. 可用 `group` 固定值：`概览 / 交易管理 / 主体管理 / 资金管理 / 系统管理`。
4. 详情页（`:id`）一律 `hidden:true`，不要出现在菜单。

### 2.3 已有路由清单（骨架已登记，勿重复声明）
`/login` `/dashboard` `/403` `/404`。
其余路径由各分片自行新增：`/orders(/:id)`、`/complaints(/:id)`、`/merchants(/audit|/:id)`、`/riders(/audit|/:id)`、`/users(/:id)`、`/finance(/settlement|/withdraw/:id)`、`/settings(/fee-rules|/announcements)`。

### 2.4 按钮级权限
```vue
<button v-permission="'merchant:audit'">审核通过</button>
<button v-permission="['merchant:audit','rider:audit']">批量审核</button>
```
无权限元素直接被移除。权限常量命名空间见 `mock/auth.json`：`merchant:audit`、`rider:audit`、`withdraw:audit`、`complaint:handle`。演示账号 permissions 含 `*`（全量），按钮权限对所有人可见；接真实后端后再收窄。

---

## 3. 公共组件 API（src/components/）

> 全部无组件库，Tailwind 手写。颜色/尺寸与设计稿对齐。

### 3.1 `<DataTable>`
```vue
<script setup lang="ts" generic="T extends Record<string, unknown>">
import DataTable, { type Column } from '@/components/DataTable.vue'
const columns: Column<Row>[] = [
  { key: 'id', title: '订单号', width: '160px' },
  { key: 'status', title: '状态' },
  { key: 'action', title: '操作', slot: 'action' },
]
</script>
<template>
  <DataTable :columns="columns" :rows="list" :loading="loading" row-key="id">
    <template #col-status="{ row }">
      <StatusTag tone="orange">{{ row.status }}</StatusTag>
    </template>
    <template #col-action="{ row }">
      <router-link :to="`/orders/${row.id}`">详情</router-link>
    </template>
  </DataTable>
</template>
```
- props：`columns: Column[]`、`rows: T[]`、`loading?: bool`、`rowKey?: string='id'`。
- 列 `Column<T>`：`{key,title,width?,align?,slot?}`。
- 默认插槽名 `col-<key>`；作用域参数 `{ row, index }`。
- 内置空态/loading 态。

### 3.2 `<StatusTag>`
```vue
<StatusTag tone="orange">待审核</StatusTag>
```
`tone`：`orange`(处理中/待审核) `green`(完成/通过) `red`(失败/拒绝/异常) `blue`(待处理/待接单) `yellow`(警告) `gray`(停用/冻结)。

### 3.3 `<StatCard>`
props：`label, value, subValue?, delta?, deltaColor?, foot, icon, iconBg, iconColor`。dashboard 顶部 4 卡即范例。

### 3.4 `<Modal>` / `<Confirm>`
```vue
<Modal v-model:open="show" title="审核" @confirm="onPass" :loading="submitting">
  <p>确认通过该商家入驻？</p>
</Modal>
<Confirm :open="del" title="删除" message="确认删除？" @confirm="onDel" />
```
`v-model:open`；`confirm-text/cancel-text`；默认按钮橙渐变。

### 3.5 `<Pagination>`
```vue
<Pagination :page="query.page" :page-size="query.pageSize" :total="total" @change="(p)=>{query.page=p; load()}" />
```

### 3.6 `<SearchBar>`
```vue
<SearchBar placeholder="搜索订单号 / 收件人" @search="onSearch" @reset="onReset">
  <template #extra>
    <select v-model="query.status"><option value="">全部状态</option></select>
  </template>
</SearchBar>
```

### 3.7 `<SvgIcon>`
```vue
import SvgIcon from '@/components/SvgIcon/SvgIcon.vue'
<SvgIcon name="order" :size="16" :stroke-width="2" class="text-primary" />
```
- 已注册图标见 `components/SvgIcon/icons.ts`：`clock search chevron-right logout bell download plus check-circle close chart order complaint merchant shield grid rider-audit user money list withdraw settings tag lock eye captcha checkbox`。
- **新图标**：直接在 `icons.ts` 追加一行（从设计稿复制 `<path>`，去掉外层 svg 属性）。

### 3.8 `<Empty>`
`<Empty text="暂无订单" />`。

### 3.9 Toast（全局）
```ts
import { useToast } from '@/composables/useToast'
const toast = useToast()
toast.success('操作成功'); toast.error('失败'); toast.warning('…'); toast.info('…')
```

---

## 4. API 层与统一返回

- 所有请求走 `src/utils/request.ts` 的 `http`（baseURL `/api`），**禁止在页面里直接用 axios**。
- 统一响应包 `{ message, body, request_id, code }`；`code===0` 时拦截器直接返回 `body`，否则 toast 报错并带 `request_id`。
- 401 自动跳登录；403 自动跳 `/403`。
- ID 一律 `string`；时间字段后端按 **int64 Unix 秒** 返回，展示用 `utils/format.ts`：
  - `formatTime(sec)` → `2026-10-22 12:03`
  - `formatDate(sec)` → `2026-10-22`
  - `formatClock(sec)` → `12:03`
  - `formatMoney(yuan)` → `¥32,540.00`；`formatNumber(n)` → `1,286`

### 写一个模块 api（范例 `src/api/orders.ts`）
```ts
import { http } from '@/utils/request'
import type { PageResult, PageQuery } from '@/types'

export interface OrderRow {
  id: string
  type: string
  route: string
  status: string
  amount: number   // 元
  created_at: number // Unix 秒
}

export function getOrders(q: PageQuery) {
  return http.get<PageResult<OrderRow>>('/orders', q)
}
export function getOrderDetail(id: string) {
  return http.get<OrderRow>(`/orders/${id}`)
}
```

### 系统配置 API（`src/api/system.ts`，设置页「邮箱 SMTP / 短信服务 / 提现规则」）
- SMTP 已对接真实后端：`getSmtpConfig`（GET `/system/config/smtp`，password 已配置时为掩码 `******`）、`updateSmtpConfig`（PUT，密码留空/掩码时后端保留原值）、`testSmtpConfig`（POST `/system/config/smtp/test`，不落库）。
- 短信已对接真实后端：`getSmsConfig/updateSmsConfig/testSmsConfig`（GET/PUT/POST `/system/config/sms*`）。返回字段 `{enabled, provider, access_key, secret_key(掩码******), sign_name, template_code}`；PUT 时 secret_key 留空/掩码保留原密文；必填 access_key+sign_name。权限 `system:config`（仅超管）。
- 提现规则：`getWithdrawRule/updateWithdrawRule`（GET/PUT `/system/config/withdraw-rule`）。字段 `{enabled, fee_rate, min_amount, max_amount, daily_limit, mode: auto|manual, settle_days}`；未配置时后端返回懒默认 `{enabled:true, fee_rate:0.006, min:10, max:5000, daily:5, manual, T+1}`。权限 `system:config`。
- 设置页 tab 约定：配送费率、公告管理已在左侧菜单有独立入口（`/settings/fee-rules`、`/settings/announcements`），**不要在「系统设置」页重复放置这两个 tab**；账号权限已并入用户管理，不要新增账号权限 tab。系统设置页 tabs = 服务区域 / 邮箱 SMTP 服务配置 / 短信服务配置 / 提现规则。

### 商家 / 骑手新增与用户详情 API（本轮新增，已对接真实后端）
- 商家新增：`createMerchant`（POST `/merchants`，`src/api/merchants.ts`）。请求体 `{name, phone, category?, contact?, address?, addressSub?, openHours?, deliveryRange?, commission?, deliveryMethod?, regionId?}`；**name 与 phone 必填**，后端缺省校验 400。返回 `{id, applyNo, auditStatus:"pending"}`——创建即提交入驻申请，进「入驻审核」tab，审核通过前不营业。
- 骑手新增：`createRider`（POST `/riders`，`src/api/riders.ts`）。请求体 `{name, phone, area?, tool?, orderTypes?, idCard?, serviceType?, regionId?}`；name/phone 必填。返回结构同上。
- 用户详情订单记录：`getUserOrders(id, q)`（GET `/users/:id/orders`，分页返回 `PageResult<UserRecentOrder>`）。
- 用户详情投诉售后：`getUserComplaints(id, q)`（GET `/users/:id/complaints`，分页返回 `PageResult<ComplaintRow>`）。
- 用户角色：列表与详情均带 `role`+`roleName`（`src/api/users.ts` 的 `BACKEND_ROLES` 映射展示）；`updateUserRole(id, role)`（POST `/users/:id/role`，body `{role}`）。角色即时生效；仅全局管理员可改（改自己 → 1105「不能修改自己的角色」）；role=system 禁改。权限 `user:role`。
- 用户实名认证：用户端 `submitRealName/realNameStatus`（POST/GET `/auth/realname-verification`，提交入参 **`{realName, idCard}`**，与返回字段统一驼峰；身份证号接口只回掩码 `前6********后4`；已实名重复提交 409）；管理端 `getRealNameVerifications(q)`（GET `/users/realname-verifications`，分页 + `status` 过滤）+ `auditRealName(id, 'approve'|'reject', reason?)`（POST `/users/realname-verifications/:id/audit`，通过后用户详情 `verified=已实名` 并同步 `realName`）。权限 `user:audit`。页面：`/users/realname`（views/users/RealName.vue，主体管理菜单「实名认证」：状态 tab 筛选/审核弹窗/驳回原因展示）；用户详情 `realNameVerify` 字段（null=无申请；status=待审核 显示「审核」入口；已驳回 显示原因）。
- 骑手审核身份信息：`RiderAuditRow.idCardMasked`（脱敏）；通过入驻 → 后端把 `id_check` 置「已通过」，驳回 → 「已驳回」，骑手重提恢复「待补充」；Audit.vue「查看」弹窗展示证件/材料/核验状态。
- 区域站长场景：创建接口不传 regionId 时后端自动用本区域；传了其他区域返回 403。
- 权限：`merchant:create` / `rider:create`（`merchant:*`/`rider:*` 通配覆盖，super_admin 全量）；用户维度查询复用 `user:detail`。

---

## 5. Mock 中间件契约（仅 VITE_USE_MOCK=1 时生效）

`vite.config.ts` 已按 `VITE_USE_MOCK` 环境变量切换请求链路：
- **默认（未设置或非 `'1'`）**：`/api/*` 由 `server.proxy` 转发到真实后端 `http://localhost:8080`，mock 中间件不注册。需后端已启动并完成 migrate + seed。
- **`VITE_USE_MOCK=1`**：注册 dev 中间件，按下面规则从 `mock/*.json` 取数，无需后端即可开发。

mock 模式下**你的页面不要在组件里硬编码业务数据**，把数据写进对应 mock json，再通过 api 拉。

### 5.1 文件 ↔ 路由映射（严格遵守）

| 路由 | mock 文件 | 说明 |
|---|---|---|
| `POST /api/auth/login` | `auth.json` | 返回 `{token,user,permissions}` |
| `GET /api/dashboard` | `dashboard.json` | 对象 |
| `GET /api/orders` | `orders.json`（数组） | 列表 |
| `GET /api/orders/:id` | `order-detail.json` | 对象（或数组按 id 查） |
| `GET /api/complaints` | `complaints.json` | 列表 |
| `GET /api/complaints/:id` | `complaint-detail.json` | |
| `GET /api/merchants` | `merchants.json` | 列表 |
| `GET /api/merchants/audit` | `merchant-audit.json` | 列表 |
| `GET /api/merchants/:id` | `merchant-detail.json` | |
| `GET /api/riders` | `riders.json` | 列表 |
| `GET /api/riders/audit` | `rider-audit.json` | 列表 |
| `GET /api/riders/:id` | `rider-detail.json` | |
| `GET /api/users` | `users.json` | 列表 |
| `GET /api/users/:id` | `user-detail.json` | |
| `GET /api/finance` | `finance.json` | 对象/列表 |
| `GET /api/finance/settlement` | `settlement.json` | 列表 |
| `GET /api/finance/withdraw/:id` | `withdraw-detail.json` | |
| `GET /api/settings` | `settings.json` | 对象 |
| `GET /api/settings/fee-rules` | `fee-rules.json` | 列表 |
| `GET /api/settings/announcements` | `announcements.json` | 列表 |

> 命名规则：多段路径自动尝试「末段连字符组合 + 单复数归一」。你按上表命名文件即可命中；新增模块时文件名用 `<entity>[-<sub>].json`，详情加 `-detail`。

### 5.2 列表查询参数（中间件已实现）
- `page`（默认 1）、`pageSize`（默认 10）、`keyword`（对整行 JSON 字符串化模糊匹配）、`status`（精确匹配行内 `status` 字段）。
- 数组响应：`body:{ list, total, page, pageSize }`，即 `PageResult<T>`。
- 写操作（POST/PUT/DELETE）统一回显 `{code:0, body:<请求体>}`，刷新即复位。
- 所有响应带 100–300ms 随机延迟 + `request_id`。

### 5.3 列表 mock 数据规范
- 数组每项是一个业务行，**必须含 `id: string` 和 `status: string`**（status 供筛选演示）。
- 金额存「元」数字；时间存 Unix 秒数字。
- 列表至少放 15+ 行以演示分页。

---

## 6. 页面还原规范

1. **以 `design/admin/<你的页面>.html` 为唯一视觉事实源，逐字还原**（间距/色值/文案/图标）。
2. 数据从 mock 经 api 拉，**禁止在 SFC 里写死业务数字/文案**。
3. 时间用 `formatTime/formatDate`；金额用 `formatMoney/formatNumber`；ID 当 string。
4. 列表页结构参考：页面头部 `<SearchBar/>` → 白色卡片内 `<DataTable/>` → 底部 `<Pagination/>`。
5. 图表一律内联 SVG（参考 dashboard 的折线图写法），**禁止 echarts/任何图表库**。
6. 按钮防重复提交：写操作加 loading 态 + disabled。

---

## 7. 你可以改 / 不可以改

**允许新增**：
- `src/views/<你的模块>/**`
- `src/router/routes/<你的模块>.ts`
- `src/api/<你的模块>.ts`
- `mock/<对应文件名>.json`
- 必要时在 `components/SvgIcon/icons.ts` **追加**图标（不要改已有项）。

**禁止改动**（骨架 Agent 维护，并行改会冲突）：
- `src/router/index.ts`、`src/router/routes/index.ts`、`src/router/types.ts`
- `src/layouts/**`、`src/permission/**`、`src/utils/request.ts`、`src/utils/format.ts`、`src/utils/permission.ts`
- `src/stores/auth.ts`、`types/index.ts`、`vite.config.ts`、`tailwind.config.js`
- 已存在的 `src/views/login/**`、`src/views/dashboard/**`、`src/views/error/**`、`src/api/auth.ts`、`src/api/dashboard.ts`、`mock/auth.json`、`mock/dashboard.json`

---

## 8. 提交流程
1. `pnpm typecheck`、`pnpm lint`、`pnpm build` 全绿。
2. `pnpm dev` 起 5173，浏览器走一遍你页面的：进列表 → 搜索/分页 → 进详情 → 按钮权限显隐。
3. **停止 dev server**，确认 5173 释放。
4. 输出交付报告：变更文件、新增 api、mock 文件、自测截图/结果、未决问题。
