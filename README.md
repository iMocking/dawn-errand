# 曙光跑腿 · Dawn Errand

> SaaS 多租户外卖 / 跑腿服务平台 —— 一整套可商用的本地即时配送解决方案。

[![Go](https://img.shields.io/badge/Go-1.27-00ADD8)](https://go.dev)
[![Gin](https://img.shields.io/badge/Web%20Framework-Gin-00ADD8)](https://github.com/gin-gonic/gin)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791)](https://www.postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D)](https://redis.io)

曙光跑腿是一套覆盖 **用户下单、商家接单、骑手配送、平台运营** 全链路的跑腿 / 外卖 SaaS 系统，面向多租户（多城市、多区域站）运营场景，提供 Web 管理后台、Android 多端应用与对外官网的完整产品矩阵。

- 📦 **多端覆盖**：官网、管理后台、用户端、商家端、骑手端
- 🏗️ **DDD 四层架构**：领域模型、用例、接口、基础设施清晰分层
- 🔐 **完备的身份与权限**：邮箱 / 短信验证码登录、JWT 会话轮换、Casbin RBAC 七角色、数据隔离
- 📊 **可观测**：Prometheus 指标、审计日志、统一错误码

---

## 产品矩阵

| 端 | 目录 | 形态 | 说明 |
|---|---|---|---|
| 🌐 **官网** | `apps/web` | Web（Vue3 + Tailwind + Vite） | **项目对外门户与下载中心**：产品介绍、功能展示、多端下载、商务合作（已交付） |
| 🛠️ 管理后台 | `apps/admin` | Web（Vue3 + Tailwind + Vite） | 平台运营：用户 / 商家 / 骑手 / 订单 / 财务 / 设置 |
| 📱 用户端 | `apps/client` | Android（Kotlin · MVVM） | 用户下单、订单跟踪、实名认证 |
| 🏪 商家端 | `apps/store` | Android（Kotlin） | 门店管理、接单、营业状态 |
| 🛵 骑手端 | `apps/driver` | Android（Kotlin） | 抢单、配送、结算 |

### Web 官网（`apps/web`）

官网是项目的对外门面，**已完整交付**，覆盖常规官网能力：

- **首页**：Hero（PC / 平板 / 手机三屏设备展示，表达全端覆盖）、能力全景（9 项）、适用场景（社区跑腿、同城配送、餐饮外卖）、核心功能（6 项）、架构链路、多端矩阵、版本动态、FAQ 摘要与底部合作 CTA
- **下载中心**：用户端 / 商家端 / 骑手端三张下载卡，点击进入各端**历史版本子页**（`/download/{client,store,driver}`），提供 Android / iOS / 鸿蒙三平台下载入口与按时间倒序的历史版本列表、管理后台入口、扫码直达说明
- **常见问题**：10 条问答折叠面板
- **商务合作**：4 种合作模式（城市合伙人 / 商家 / 骑手 / 区域代理）、合作流程、联系方式与表单
- **多端适配**：响应式支持 PC 与移动端（断点 768 / 1024），核心交互（汉堡菜单、FAQ 折叠、下载 / 表单 toast）均已可用
- **多语言适配**：中 / 英 / 西班牙语三语切换（导航栏 `中 / EN / ES`，移动端在汉堡菜单内），语言选择持久化（localStorage）；全站导航、页脚、首页、下载中心、版本子页、FAQ、商务合作表单文案均随语言切换，`<html lang>` 与页面标题同步更新

实现要点：

- **技术栈**：Vue 3 + TypeScript（strict）+ TailwindCSS + vue-router + pinia + 统一 axios（拦截器解包 `{message, body, request_id, code}`，`code === 0` 为成功）
- **文本数据全部 JSON 化**：页面文案零硬编码，统一提取至 `apps/web/mock/{common,home,download,faq,contact}.json`，由 Vite mock 中间件（`GET /api/site/*`，模拟 80–240ms 延迟）经请求渲染；多语言按 `?lang=zh|en|es` 返回 `mock/{name}.{lang}.json`，缺失回退默认语言；关闭 mock 时自动代理到后端 `:8080`
- **设计稿**：一阶段响应式设计稿位于 `design/web/`（index / download / download-app / faq / contact 五页，其中 download-app 为各端历史版本子页设计稿），与工程实现保持同步，均通过双视口（1440×900 / 390×844）自检
- **质量门禁**：typecheck / lint / build 三查通过；Playwright 双视口渲染无 console 错误，核心交互断言通过

本地运行（无需后端，纯前端 + mock 即可）：

```powershell
cd apps/web
pnpm install --store-dir "apps/web/node_modules/.store"   # 本机 pnpm 全局 store 无权限时的安装方式
pnpm dev        # :5173，mock 数据渲染；与 apps/admin 同时开发时请错开端口
```

管理后台（只对接真实后端）：

```powershell
cd apps/admin
pnpm install
pnpm dev        # :5173，/api 代理到 http://localhost:8080；需先启动后端
```

管理端已移除 mock 中间件与 `mock/*.json`，所有页面数据均来自真实 `/api`。

超管账号由 `go run ./cmd/seed` 初始化（默认 `admin@shuguang`，一次性密码见 seed 输出；联调库另有测试超管）。管理端与官网同时开发时请错开端口：`pnpm exec vite --port 5174`。

---

## 系统架构

- 📐 **架构图（draw.io）**：[跑腿服务架构图（M0–M5 已落地 + 外部接入待办）](docs/跑腿服务架构图.drawio)
- 🏢 **竞品架构参考**：[美团外卖 · 即时配送整体架构](docs/美团外卖架构图.drawio)

核心链路：

```
多端（官网 / 管理后台 / 用户 / 商家 / 骑手）
        │  HTTPS / REST（管理端 /api，App /api/app）
        ▼
Go 后端（Gin · :8080）
  身份与账号 · 商家/骑手入驻审核 · 商品目录 · 订单交易
  配送履约（抢单/追踪/ETA） · 钱包结算/提现/对账 · 评价/退款/优惠券
  消息中心 · 定时任务 · 审计 · 限流 · /readyz /metrics
        │  DDD 四层：interfaces → application → domain
        ▼
PostgreSQL 16（migration 001–009） · Redis 7（会话/限流/GEO/锁）
本地/OSS 文件存储 · SMTP / 短信 · Prometheus 指标
```

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.27 · Gin · ent（ORM）· Casbin（RBAC）· zap · cron |
| 存储 | PostgreSQL 16（migration 001–009）· Redis 7 · 本地存储（可切 OSS） |
| 身份与安全 | JWT（access + refresh 轮换 · 会话版本失效）· AES-GCM 敏感数据加密 · 邮箱 / 短信验证码 |
| 数据模型 | ent 31 张表 · ULID 主键 · UTC int64 时间戳 |
| 可观测 | Prometheus（`/metrics`）· 审计日志（`audit_logs`） |
| 前端 | Vue 3 · TailwindCSS · Vite（admin / web）· Android Kotlin · MVVM（client / store / driver） |

## 快速开始

前置依赖：Go 1.27+、Docker、pnpm（前端）。

```powershell
# 1. 启动基础设施（PostgreSQL 16 + Redis 7）
docker compose up -d

# 2. 数据库迁移（幂等，按版本 001/002…）
go run ./cmd/migrate up

# 3. 初始化种子数据（输出一次性超管密码，务必抄录）
go run ./cmd/seed --email admin@shuguang.local

# 4. 启动后端服务（:8080）
go run ./cmd/server

# 5. 启动管理后台（:5173，/api 自动代理到 :8080）
cd apps/admin
pnpm install
pnpm dev
```

环境配置见 `.env.example`（JWT 密钥、数据库连接、Redis、SMTP、存储目录等）。

## 目录结构

```
dawn-errand/
├── apps/                  # 多端应用
│   ├── web/               # 🌐 官网（Vue3 + Tailwind + Vite，mock 数据驱动，已交付）
│   ├── admin/             # 管理后台（Vue3 + Tailwind + Vite）
│   ├── client/            # 用户端（Android Kotlin · MVVM）
│   ├── store/             # 商家端（Android Kotlin）
│   └── driver/            # 骑手端（Android Kotlin）
├── cmd/                   # 后端入口（server / migrate / seed）
├── internal/              # 后端 DDD 四层
│   ├── domain/            # 领域模型 · 端口 · 错误码
│   ├── application/       # 用例编排 · 事务 · 权限
│   ├── infrastructure/    # ent 持久化 · Casbin · Redis · token · mailer · storage
│   └── interfaces/        # Gin 路由 · 中间件 · DTO
├── migrations/            # 数据库迁移脚本
├── docs/                  # 架构图与交付文档
├── design/                # 设计资产
│   └── web/               # 官网响应式设计稿（index / download / download-app / faq / contact）
├── .agent/                # AI 开发规范（前端 / 后端）
├── docker-compose.yml     # 本地基础设施编排
└── .env.example           # 环境变量样例
```

## 文档

- [后端交付报告](docs/后端交付报告.md)
- [后端商用化完善计划（M0–M5）](docs/后端商用化完善计划.md)
- [商家地推入驻业务闭环](docs/商家入驻业务闭环.md)
- [商用运维手册](docs/商用运维手册.md)
- [后端能力审阅与联调准备](docs/后端能力审阅与联调准备.md) ｜ [三端联调接口契约（v0.1 草案）](docs/三端联调接口契约.md)
- [客户端交付说明](docs/客户端交付说明.md) ｜ [骑手端交付说明](docs/骑手端交付说明.md) ｜ [商家端交付说明](docs/商家端交付说明.md)
- [三端 Android 真实 API 联调进度](docs/三端Android真实API联调进度.md)
- [客户端高德地图适配说明](docs/客户端高德地图适配说明.md)
- [系统身份设计分析](docs/系统身份设计分析.md) ｜ [实施进度记录](docs/实施进度记录.md)

## 开发规范

本项目由 AI 辅助开发，代码与协作遵循 `.agent` 下的规范：

- [后端开发规则](.agent/BACKEND_AI_DEVELOPMENT_RULES.md)
- [前端开发规则](.agent/FRONTEND_AI_DEVELOPMENT_RULES.md)

## 路线图

- [ ] 三端 App 业务接口（商品/地址/订单/配送/资金）：**M0–M5 代码完成**（身份会话、账号开通、上传、菜品目录、地址簿、营业设置、下单/支付/接单/取送/确认、位置追踪、通知、超时任务、钱包/提现/评价/售后、优惠券、支付回调、对账、发券、就绪探针）；生产接入待办见 [商用运维手册](docs/商用运维手册.md) §10；缺口基线与契约见 [能力审阅](docs/后端能力审阅与联调准备.md) 与 [接口契约](docs/三端联调接口契约.md)
- [ ] 支付网关 / 退款 / 对账、商家结算分账
- [x] 配送派单调度引擎（地图 LBS · 路线 · ETA）：客户端配送追踪已落地（v1.2.0 高德地图）；骑手端接单/导航与后端调度引擎待实现
- [ ] 消息推送与异步消息队列、订单超时任务
- [ ] HTTPS 与域名备案、监控告警、CI/CD、数据库备份
- [ ] 官网对接真实后端（替换 mock 数据源）、内容运营后台

## License

本项目当前未配置开源许可证，版权归项目方所有，商用授权请通过官网商务合作渠道联系。
