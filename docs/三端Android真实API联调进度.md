# 三端 Android 真实 API 联调进度

> 日期：2026-10-26  
> 后端：`http://127.0.0.1:8080`（Android 模拟器访问 `http://10.0.2.2:8080`）  
> 范围：`apps/client` 用户端、`apps/store` 商家端、`apps/driver` 骑手端

## 1. 登录方式：邮箱优先

三端统一默认使用「邮箱 + 密码」登录，手机号验证码作为备用入口：

| 端 | 默认 Tab | 默认账号（seed） | 状态 |
|---|---|---|---|
| 用户端 client | 邮箱登录 | `client.test@dawnerrand.local` / `Client123456` | ✅ |
| 商家端 store | 邮箱密码 | `store.owner@dawnerrand.local` / `Store123456` | ✅ |
| 骑手端 driver | 邮箱登录 | `rider.owner@dawnerrand.local` / `Rider123456` | ✅（原本已邮箱优先） |

- Debug 构建自动预填上述联调账号，Release 不预填。
- 登录统一走 `POST /api/auth/login`，body `{account,password}`；成功后持久化 `token` / 用户信息。
- 用户端新增邮箱密码输入区；商家端默认模式由 `PHONE` 改为 `EMAIL`；骑手端保持 `emailMode=true`。
- 验证码备用：用户端/商家端/骑手端均调用真实 `/api/auth/sms/send`、`/api/auth/sms/login`，不再有“任意验证码”演示登录。

## 2. 各端真实接口接入

### 2.1 用户端 client

| 页面/能力 | 接口 | 状态 |
|---|---|---|
| 邮箱登录 | `POST /api/auth/login` | ✅ 实测 |
| 首页商家列表 | `GET /api/app/client/merchants` | ✅ 实测（村口老陈家常菜 4.6 分/305 单） |
| 个人资料 | `GET /api/app/client/me` | ✅ 登录后刷新昵称 |
| 我的订单 | `GET /api/app/client/orders` | ✅ 已接入（当前联调账号无订单，显示空态） |
| 地址簿 | `GET /api/app/client/addresses` | ✅ 已接入 |
| 下单/支付/退款/评价/优惠券 | `/api/app/client/orders*` | ⏳ 页面仍为本地流程，待下一步接入 |

### 2.2 商家端 store

| 页面/能力 | 接口 | 状态 |
|---|---|---|
| 邮箱登录 | `POST /api/auth/login` | ✅ 实测 |
| 工作台店铺/今日数据 | `GET /api/app/store/me` | ✅ 实测（村口老陈家常菜、4.62 分） |
| 订单列表 | `GET /api/app/store/orders?tab=all` | ✅ 实测（渲染 `SG202610220086`） |
| 接单/拒单/出餐 | `POST /api/app/store/orders/:id/{accept,reject,ready}` | ✅ 已接入按钮 |
| 菜品列表 | `GET /api/app/store/menu` | ✅ 已接入（分类/菜品/价格/库存/上下架） |
| 收入流水 | `GET /api/app/store/finance/{summary,records}` | ✅ 已接入 |
| 营业开关 | `POST /api/app/store/status` | ✅ 已接入 |
| 菜品新增/编辑/上下架、营业参数保存 | `POST/PUT /api/app/store/*` | ⏳ 页面仍有本地持久化逻辑，待补 |

### 2.3 骑手端 driver

| 页面/能力 | 接口 | 状态 |
|---|---|---|
| 邮箱登录 | `POST /api/auth/login` | ✅ 实测 |
| 骑手资料 | `GET /api/app/driver/me` | ✅ 实测（赵铁柱、4.9 分） |
| 接单大厅 | `GET /api/app/driver/order-pool` | ✅ 已接入（当前无待抢单，显示空态） |
| 我的订单 | `GET /api/app/driver/orders` | ✅ 已接入 |
| 收入/余额 | `GET /api/app/driver/finance/{summary,records}` | ✅ 已接入 |
| 上/下线 | `PUT /api/app/driver/status` | ✅ 已接入开关 |
| 抢单/取货/送达 | `POST /api/app/driver/orders/:id/{grab,pickup,deliver}` | ✅ API 方法已封装；页面动作待接 |
| 位置上报 | `POST /api/app/driver/location` | ⏳ 待接 |

## 3. 联调验证记录

| 端 | 操作 | 结果 |
|---|---|---|
| client | 清数据 → 邮箱登录 → 首页 | 进入 `MainActivity`；昵称「联调用户」；商家「村口老陈家常菜」来自后端 |
| store | 清数据 → 邮箱登录 → 工作台 | 进入 `MainActivity`；店铺名/评分来自后端；今日 0 单 |
| store | 订单 → 配送中 | 显示后端订单 `SG202610220086` / ¥43.00 |
| driver | 清数据 → 邮箱登录 → 工作台 | 进入 `MainActivity`；姓名「赵铁柱」、评分 4.9 来自后端 |

截图：

- `design/client/login-email-real.png`
- `design/client/home-real-api.png`
- `design/store/login-email-real.png`
- `design/store/home-real-api.png`
- `design/store/orders-real-api.png`
- `design/driver/login-email-real.png`
- `design/driver/home-real-api.png`

## 4. 构建验证

```powershell
$env:JAVA_HOME='C:\Program Files\Android\Android Studio\jbr'
cd apps/client; .\gradlew.bat :app:assembleDebug
cd apps/store;  .\gradlew.bat :app:assembleDebug
cd apps/driver; .\gradlew.bat :app:assembleDebug
```

三端均 `BUILD SUCCESSFUL`，APK：

- `apps/client/app/build/outputs/apk/debug/app-debug.apk`
- `apps/store/app/build/outputs/apk/debug/app-debug.apk`
- `apps/driver/app/build/outputs/apk/debug/app-debug.apk`

## 5. 已知边界

1. 用户端购物车、下单、支付、售后、评价、优惠券仍沿用本地 `DemoRepository` 流程，未全部切到后端。
2. 商家端菜品 CRUD、营业参数保存、提现仍为本地/演示逻辑；订单动作已真实调用。
3. 骑手端抢单/取货/送达 API 方法已封装，但页面按钮尚未全部替换；位置上报待接。
4. 模拟器通过 `10.0.2.2:8080` 访问后端；真机需将 `API_BASE_URL` 改为局域网 IP 或 HTTPS 域名。
5. 后端 JWT 刷新尚未在 Android 端实现自动单飞刷新；token 过期后需重新登录。

## 6. 下一步

- 用户端：下单/支付/售后/评价/优惠券接口替换本地流程。
- 商家端：菜品 CRUD、营业设置保存、提现接口。
- 骑手端：抢单/取货/送达按钮与位置上报、导航联动。
- 三端统一 `401 → refresh → 重放` 拦截器。
