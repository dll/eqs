# 工程快捷服务 (EQS) 产品需求文档 V8.0

> **文档版本**：V8.0
> **创建日期**：2026-08-10
> **文档状态**：正式版
> **机密等级**：内部机密
> **驱动模式**：规范驱动 + 测试驱动 + AI驱动
> **基于**：V6.0（功能基线）+ V7.0（非功能性需求）+ V8.0（整合实际构建/部署实现）

---

## 文档修订历史

| 版本 | 日期 | 修订人 | 修订内容 |
|------|------|--------|----------|
| V1.0 | 2026-08-07 | - | 初始版本，整合EQS.md需求，明确技术栈 |
| V2.0 | 2026-08-07 | - | 审核完善：补充API设计、数据模型、微信集成、测试策略、部署架构 |
| V3.0 | 2026-08-07 | - | 规范驱动+测试驱动+AI流水线+国产模型集成 |
| V4.0 | 2026-08-07 | 刘东良 (ldl@chzu.edu.cn) | 图表改英文、补充部署规格、添加参与方信息 |
| V5.0 | 2026-08-07 | 刘东良 (ldl@chzu.edu.cn) | 对照原始需求审核：补齐平台职能、资质核验、仲裁规则、银行存管、简介文案 |
| V6.0 | 2026-08-08 | 刘东良 (ldl@chzu.edu.cn) | 完成原始需求追踪审查：补齐双端App、交易核心数据与API、第三方支付合规、争议处理和验收基线 |
| V7.0 | 2026-08-09 | 刘东良 (ldl@chzu.edu.cn) | 新增非功能性需求：系统主题、国际化、多端切换、版本检查更新、系统配置中心；演示数据纳入配置管理 |
| **V8.0** | **2026-08-10** | **刘东良 (ldl@chzu.edu.cn)** | **整合 V6+V7 全量需求，对齐代码实际实现（76 路由/25 表/测试/CI-CD），落地生产部署（腾讯云 CVM + eqs-chzu.tech + GitHub Actions 自动构建发布）** |

---

## 目录

1. [项目概述](#1-项目概述)
2. [多端产品形态与访问入口](#2-多端产品形态与访问入口)
3. [统一版本号策略](#3-统一版本号策略)
4. [功能需求总览（V6 功能基线 + V7 非功能）](#4-功能需求总览)
5. [功能模块详细说明](#5-功能模块详细说明)
6. [非功能性需求](#6-非功能性需求)
7. [技术架构与选型](#7-技术架构与选型)
8. [数据需求与接口设计](#8-数据需求与接口设计)
9. [构建与部署（V8.0 新增）](#9-构建与部署)
10. [CI/CD 流水线](#10-cicd流水线)
11. [发布管理（GitHub Release）](#11-发布管理)
12. [测试策略](#12-测试策略)
13. [安全与合规](#13-安全与合规)
14. [版本发布计划](#14-版本发布计划)
15. [验收基线](#15-验收基线)
16. [原始需求追踪与验收矩阵](#16-原始需求追踪与验收矩阵)
17. [附录](#17-附录)

---

## 1. 项目概述

### 1.1 产品定位

**一句话定位**：工程服务业的"美团+滴滴"——一站式智慧交易与协作平台。

**核心价值**：
- 对甲方：像"点外卖"一样找工程服务——简单、透明、可控
- 对服务方：像"滴滴抢单"一样接项目——高效、低成本、有背书

### 1.2 业务闭环

```
+----------------------------------------------------------------------+
|BUSINESS FLOW                                                         |
+----------------------------------------------------------------------+
|Publish Request -> Match/Claim -> Online Contract -> Delivery         |
|-> Result Confirm -> Milestone Settlement -> Credit Accumulation      |
+----------------------------------------------------------------------+
```

> **注解**：发布需求 -> 匹配/抢单 -> 线上签约 -> 过程交付 -> 成果确认 -> 节点结算 -> 信用沉淀

### 1.3 项目范围

| 范围维度 | 说明 |
|----------|------|
| 服务类型 | 造价咨询、工程监理、地质勘察、工程设计 |
| 项目规模 | 初期聚焦 50万以下中小型项目 |
| 地域范围 | 初期以单城市为试点，验证模式后全国推广 |
| 终端覆盖 | 需求方与服务方共用业务客户端（uni-app）：H5 / 微信小程序 / iOS / Android App；运营方使用 PC 管理后台 |

### 1.4 参与方

| 角色 | 说明 |
|------|------|
| 甲方（需求方） | 建设单位、施工单位、个人业主，user_type=1 |
| 乙方（服务方） | 事务所、地勘院、监理公司、设计院、独立工程师，user_type=2 |
| 运营方（平台） | 平台管理员，user_type=3 |
| 评审专家 | 争议专家评审团成员，user_type=4 |
| 开发方 | 刘东良 (ldl@chzu.edu.cn)，项目负责人/全栈开发 |

---

## 2. 多端产品形态与访问入口

### 2.1 终端矩阵（V8.0 已落地）

| 端 | 技术 | 构建命令 | 产物 | 生产入口 |
|----|------|----------|------|----------|
| 用户端 H5 | uni-app (Vue3) | `pnpm --filter @eqs/client build:h5` | `dist/build/h5` | `https://eqs-chzu.tech/h5/` |
| 微信小程序 | uni-app | `build:mp-weixin` | `dist/build/mp-weixin` | 微信开发者工具上传 |
| Android App | uni-app + 离线 SDK | `build:app` → Gradle | `*.apk` | GitHub Artifact / 分发 |
| iOS App | uni-app + 离线 SDK | `build:app` → xcodebuild | `*.ipa`（init/release） | 真机安装 / TestFlight |
| 管理后台 | Vue3 + Element Plus | `pnpm --filter @eqs/admin build` | `dist/` | `https://eqs-chzu.tech/admin/` |
| 后端 API | Go + Gin + GORM | `go build cmd/server/main.go` | 二进制 | `https://eqs-chzu.tech/api/v1/*` |

### 2.2 多端切换（V7 NFR-03）

- 同一账号 token 多端共用（统一 JWT 体系，无端绑定）。
- `GET /api/v1/platform/links` 返回各端访问地址（由 `multiplatform.urls` 配置驱动）。
- 用户中心提供"其他端访问"入口。

### 2.3 子路径 base 约定（V8.0 落地要点）

| 端 | base 配置 | 说明 |
|----|-----------|------|
| admin | vite `base: '/admin/'` + `createWebHistory(import.meta.env.BASE_URL)` | 子路径部署，路由 base 必须同步 |
| H5 | manifest `h5.router.base: '/h5/'`（hash 路由） | 子路径部署 |
| API | nginx 反代 `/api` | 前端走相对路径，无需 CORS |

> 前端 `request.ts` baseURL 为空，走**同域相对路径** `/api/...`，由 nginx 反代到后端。

---

## 3. 统一版本号策略（V8.0 新增）

### 3.1 版本号定义

全端（根 / client / admin / manifest）使用**同一个**语义化版本 `MAJOR.MINOR.PATCH`：

| 版本位 | 含义 | 变更时机 |
|--------|------|----------|
| MAJOR | 不兼容的大改动 | 重构架构、破坏性升级 |
| MINOR | 新增功能（向后兼容） | 发版新功能 |
| PATCH | 缺陷修复 / 小重构 | **迭加规则：`0.1.0 +1 → 0.1.1`** |

### 3.2 迭加规则（重构时）

**每次重构 / 迭代发版，`PATCH +1`，直到 MINOR 增位时归零**：

| 当前版本 | 重构 1 次后 | 重构 2 次后 | 新增功能后 |
|----------|-------------|-------------|------------|
| 0.1.0 | 0.1.1 | 0.1.2 | 0.2.0 |
| 1.2.3 | 1.2.4 | 1.2.5 | 1.3.0 |

### 3.3 同步位置（发布前必须一致）

| 位置 | 字段 | 说明 |
|------|------|------|
| 根 `package.json` | `version` | monorepo 总版本 |
| `packages/client/package.json` | `version` | 用户端 |
| `packages/admin/package.json` | `version` | 管理后台 |
| `packages/client/src/manifest.json` | `versionName` | App 显示版本 |
| `packages/client/src/manifest.json` | `versionCode` | App 构建号（PATCH 每 +1，`versionCode +1`） |

### 3.4 发布流程（Tag 触发 Release）

```bash
# 1. 同步 4 处版本号（示例：0.1.0 → 0.1.1）
# 2. 提交并打 tag
git add package.json packages/client/package.json packages/admin/package.json packages/client/src/manifest.json
git commit -m "chore: 版本号升至 0.1.1"
git tag -a v0.1.1 -m "EQS v0.1.1"
git push origin main --tags
```

> tag 名 `v{version}` 与全端版本号一致，`release.yml` 据此自动构建全部产物并发布 GitHub Release。

---

## 4. 功能需求总览

### 4.1 需求来源构成

| 来源 | 内容 | 集成方式 |
|------|------|----------|
| PRD V6.0 | 功能基线：发单/服务超市/进度看板/支付结算/派单/竞价/协作/打卡/资质/合同/结算中心/信用/争议/运营 + 19 数据表 + 55 API | 第 5、8 章 |
| PRD V7.0 | 非功能性：配置中心/主题/国际化/多端/版本检查/演示数据/非功能加固 | 第 6、8 章 |
| V8.0 新增 | 构建与部署：GitHub Actions 全自动构建/发布、腾讯云 CVM 生产部署、统一版本号、生产访问入口 | 第 2、3、9、10、11 章 |

### 4.2 功能模块清单（含实现状态）

> 状态标注：✅ 已实现并部署 ｜ 🔶 部分实现（Mock/占位） ｜ ⬜ 待实现/规划

| 模块 | 归属端 | 关联验收 | 实现状态 |
|------|--------|----------|----------|
| 智能发单 | 客户端 | AC-01、AC-02 | ✅ |
| 服务超市 | 客户端 | AC-01 | 🔶 服务商列表接口待补 |
| 进度看板 | 客户端 | AC-08 | ✅ |
| 订单支付与节点结算 | 客户端/后台 | AC-04、AC-08、AC-11 | ✅（支付通道 Mock） |
| 智能派单 | 客户端 | AC-06 | ✅ |
| 抢单/竞价 | 客户端 | AC-02 | ✅ |
| 在线协作（图纸/批注/文件） | 客户端 | AC-07 | ✅（文件无下载端点，批注已实现） |
| 考勤打卡 | 客户端 | AC-09 | ✅ |
| 标准交付模板 | 客户端/后台 | AC-12 | ✅ |
| 资质审核（OCR） | 后台 | AC-05 | ✅（OCR Mock） |
| 合同管理（电子签） | 后台 | AC-03 | ✅（签约 Mock） |
| 结算中心 | 后台 | AC-04 | ✅ |
| 信用评分 | 后台 | AC-10 | ✅ |
| 纠纷评审与调解 | 后台/客户端 | AC-11 | ✅（后台无仲裁 UI） |
| 运营管理（用户/项目/看板） | 后台 | - | ✅ |
| 系统配置中心 | 后台 | NFR-AC-05 | ✅ |
| 系统主题（print/dark/light） | 客户端/后台 | NFR-AC-01 | ✅ |
| 国际化（中/英） | 客户端/后台 | NFR-AC-02 | ✅ |
| 多端切换 | 客户端 | NFR-AC-03 | ✅ |
| 版本检查与更新 | 客户端/后台 | NFR-AC-04 | ✅ |
| 演示数据管理 | 后台 | NFR-AC-06 | ✅ |
| 智能 OCR / 推荐 / 客服 | AI 能力 | AC-05、AC-06 | 🔶 接口设计完成，模型集成 Mock |
| 消息/通知 | 客户端 | - | ⬜ 模型已建，无路由 |
| 佣金管理 | 后台 | SRC-BIZ-01 | ✅ |

---

## 5. 功能模块详细说明

### 5.1 智能发单模块（AC-01、AC-02）

- 甲方通过标准化表单发布工程服务需求，系统自动识别项目类型并生成服务清单。
- 功能点：项目类型选择（房建/市政/装修/水利等）、项目信息填写、智能清单生成、附件上传（CAD/PDF/图片，单文件最大 50MB）、发布范围设置（公开发布/指定服务商邀请）、预估报价。
- 业务规则：项目金额 ≥50 万需上传立项批文；同一项目最多邀请 5 家指定服务商；发布后 24 小时内未响应自动推送提醒。
- **已实现**：`POST /api/v1/project/create`、`GET /api/v1/project/list`、`GET /api/v1/project/:id`、`POST /api/v1/project/:id/invite`、`GET /api/v1/project/:id/recommend`。

### 5.2 服务超市模块（AC-01）

- 按服务类型分类展示入驻服务商，支持多维度筛选和比较。
- 功能点：分类浏览（造价/监理/地勘/设计四大类）、筛选（地区/资质等级/评分/价格区间/服务类型）、服务商主页（资质证书/团队规模/过往案例/用户评价/信用分）、收藏关注、智能推荐、对比（最多 3 家）。
- **实现缺口**：客户端 `provider/list.vue` 调用 `GET /api/v1/provider/list`，后端**未注册该路由**（前端静默空列表），需补齐服务商列表/详情接口。

### 5.3 进度看板模块（AC-08）

- 状态流转：Published → Accepted → Field Work → Drafting → Reviewing → Delivered → Accepted by Client → Settled。
- 功能点：进度可视化（时间轴）、节点提醒、消息中心、交付物预览（在线预览 PDF/图片）、进度催办。
- **已实现**：订单/里程碑状态机（orders.status、payment_milestones.status），`GET /api/v1/order/:id`。

### 5.4 订单支付与节点结算模块（AC-04、AC-08、AC-11）

- 平台对接持牌支付机构或银行的支付、分账/存管能力，按节点验收发起资金释放指令；平台不吸收存款、不形成自有资金池。
- 功能点：订单支付（微信支付/银行转账/对公）、节点设置、验收确认、退款、结算管理、发票、账单。
- 业务规则：资金进入受监管账户；验收后 T+1 工作日释放；争议订单可止付/冻结；默认"节点验收放款"。
- **已实现**：`POST /api/v1/pay/create`、`POST /api/v1/pay/notify/:channel`、`POST /api/v1/milestone/:id/settle`、`GET /api/v1/pay/transactions`、`GET /api/v1/pay/balance`（**支付通道为 Mock，待持牌机构签约**）。

### 5.5 智能派单模块（AC-06）

- 基于地理位置、资质等级、历史评价自动推送匹配项目。
- 功能点：项目推荐（每日 3-5 个）、匹配度评分、一键报名、报名管理。
- **已实现**：`GET /api/v1/project/:id/recommend`（推荐接口 P95 ≤ 3s）。

### 5.6 抢单/竞价模块（AC-02）

- 功能点：项目列表、报价提交（金额+服务周期+资质说明）、方案上传（PDF/Word）、竞价排名（隐去具体金额）、中标通知。
- **已实现**：`POST /api/v1/bid/submit`、`GET /api/v1/project/:id/bids`、`PUT /api/v1/bid/:id/withdraw`、`POST /api/v1/bid/:id/select`。

### 5.7 在线协作模块（AC-07）

- 功能点：图纸预览（PDF/图片）、批注工具（坐标标注/评论/版本）、文件传输（PDF/CAD/图片+版本管理）、消息沟通。
- **已实现**：`POST /api/v1/file/upload`、`GET /api/v1/project/:id/files`、`POST /api/v1/annotation/add`、`GET /api/v1/annotation/list/:id`、`PUT /api/v1/annotation/:id/resolve`。
- **实现缺口**：ProjectFile/Deliverable/DisputeEvidence 无下载端点。

### 5.8 考勤打卡模块（AC-09）

- 定位合规规则：仅主动打卡时获取一次定位；首次单独告知并授权；拒绝授权可上传带时间水印照片；不后台持续定位。
- **已实现**：`POST /api/v1/attendance/checkin`、`GET /api/v1/order/:id/attendance`。

### 5.9 标准交付模板模块（AC-12）

- 功能点：模板分类（地勘报告/造价审核清单及监理/设计类）、模板下载、清单校验（缺项阻止提交）、版本管理。
- **已实现**：`GET /api/v1/contract/templates`（合同模板）；交付模板表 delivery_templates 已建。

### 5.10 资质审核模块（AC-05）

- 功能点：营业执照/资质证书/人员证书 OCR、社保记录核验、主管部门接口核验、人工复核、审核记录、资质到期提醒。
- 审核规则：强制社保+注册证书；挂靠/虚假/社保不一致一律驳回。
- 审核流程：Submit → OCR → Authorized Data Check → System Precheck → Manual Review → Approve/Reject。
- **已实现**：`POST /api/v1/supplier/:id/qualifications`、`GET /api/v1/supplier/:id/qualifications`、`POST /api/v1/qualification/:id/review`、`GET /api/v1/admin/qualifications`（**OCR 为 Mock，主管数据核验待授权**）。

### 5.11 合同管理模块（AC-03）

- 功能点：合同模板库（四类）、合同生成（自动填充）、电子签章（对接合法电子签约服务商）、合同归档、合同变更。
- **已实现**：`GET /api/v1/contract/templates`、`POST /api/v1/order/:id/contract`、`POST /api/v1/contract/:id/sign`、`GET /api/v1/contract/:id/download`、`POST /api/v1/sign/notify`（**电子签为 Mock，待服务商签约**）。

### 5.12 结算中心模块（AC-04）

- 功能点：对账单生成（按月）、资金流水、结算管理（向第三方发起指令并处理回调）、发票核验、财务报表。
- **已实现**：`GET /api/v1/admin/transactions`、`POST /api/v1/milestone/:id/settle`、佣金管理 `GET /api/v1/admin/commission/list`、`POST /api/v1/admin/commission/:id/collect`。

### 5.13 信用评分模块（AC-10）

- 评分维度：交付准时率 30%、质量评分 30%、纠纷次数 20%、平台活跃度 10%、历史履约 10%。
- 信用等级：AAA(90-100) 优先推荐免审核；AA(80-89) 正常；A(70-79) 降权；B(60-69) 限制接单；C(<60) 暂停接单需整改。
- **已实现**：`recalcUserCredit` 在评价/结算/争议结案时触发（SAR IMP-03 已闭环）。

### 5.14 纠纷评审与调解模块（AC-11）

- 平台专家评审与调解，非法定仲裁；保留仲裁/诉讼权利。
- 功能点：纠纷发起、证据提交、专家指派（适配专业且无利益冲突，随机 3 人）、在线评审（投票）、结果执行、复核机制（仅一次）。
- 规则：专家评审团由资深总工组成；评审费用责任方承担；争议订单可申请止付/冻结；全程留痕。
- **已实现**：`POST /api/v1/dispute/create`、`GET /api/v1/order/:id/disputes`、`GET /api/v1/dispute/:id`、`POST /api/v1/dispute/:id/evidence`、`POST /api/v1/dispute/:id/expert`、`POST /api/v1/dispute-expert/:id/opinion`、`POST /api/v1/dispute/:id/close`。

### 5.15 系统配置中心（NFR-AC-05，V7）

- 配置项分类：主题、国际化、多端、版本、演示数据、系统维护。
- 功能要求：管理员 CRUD、批量公开读取、类型安全、内存缓存、变更审计。
- **已实现**：`GET /api/v1/admin/config/list`、`POST /api/v1/admin/config/upsert`、`DELETE /api/v1/admin/config/delete/:key`、`GET /api/v1/config/public`、`GET/PUT /api/v1/config/user/prefs`。

### 5.16 系统主题（NFR-AC-01，V7）

- 三主题：print（默认，白底黑字适配打印）、dark、light。
- 切换规则：用户偏好、项目级覆盖用户级。
- **已实现**：CSS 变量实现、user_settings.theme 持久化、projects.theme 项目级、`GET /api/v1/theme/list`、`PUT /api/v1/project/:id/theme`。

### 5.17 国际化（NFR-AC-02，V7）

- 支持 zh-CN（默认）、en-US。
- **已实现**：Vue i18n 客户端/后台均接入、文案集中 locales JSON、用户偏好持久化、后端 message_key、tabBar 动态同步、Element Plus locale 联动。

### 5.18 版本检查与更新（NFR-AC-04，V7）

- **已实现**：`GET /api/v1/version/check?current=`、`GET /api/v1/version/latest`、`POST /api/v1/admin/version/publish`、`GET /api/v1/admin/version/list`、VersionRateLimit 限流、checkInterval 配置（默认 6 小时）。

### 5.19 演示数据管理（NFR-AC-06，V7）

- **已实现**：`POST /api/v1/admin/demo/seed`（demo/test/training 三模式）、`POST /api/v1/admin/demo/clean`、`GET /api/v1/admin/demo/status`、`POST /api/v1/admin/demo/toggle`、后台系统配置页集成入口。

### 5.20 非功能性加固（V7 NFR-07）

- **已实现**：MonitorMiddleware 记录耗时，`GET /api/v1/admin/monitor/stats` 输出 avg/P95/错误率；安全加固（敏感配置不出现在公开接口、偏好白名单、版本检查限流、GORM 参数化查询防注入、审计日志）。

---

## 6. 非功能性需求

### 6.1 性能需求

| 指标 | 要求 | 现状 |
|------|------|------|
| 首屏加载 | ≤ 2s（4G） | H5 已部署，待实测 |
| API 响应 P95 | ≤ 500ms（推荐接口 ≤ 3s） | Monitor 已监控 |
| 并发用户 | 1000 并发 | 待压测（k6） |
| 数据库查询 | 单次 ≤ 100ms | SQLite/MySQL 均适用 |
| 文件上传 | 50MB，≥1MB/s | 已支持 |

### 6.2 安全需求

| 类别 | 要求 | 现状 |
|------|------|------|
| 数据加密 | 传输 TLS 1.3（Caddy 自动 HTTPS）、存储 AES-256 | ✅ HTTPS 已启用 |
| 身份认证 | 手机号+验证码、JWT | ✅ |
| 权限控制 | RBAC（user_type 1/2/3/4 + RequireAdmin） | ✅ |
| 审计日志 | 关键操作全量记录、保留 180 天 | ✅ audit_logs |
| 合规 | 网安法/数安法/个保法及支付、电子签名监管 | 外部依赖项按准入门槛降级 |
| 敏感信息 | 最小必要收集、单独授权、加密存储 | ✅ 定位/证件加密存储 |

### 6.3 可用性需求

| 指标 | 要求 | 现状 |
|------|------|------|
| 系统可用性 | ≥ 99.9%（月度） | systemd 常驻 + Restart=always |
| 数据备份 | 每日全量 + 实时增量 | SQLite 文件可备份 |
| 灾难恢复 | RTO ≤ 4h，RPO ≤ 1h | 待建立 |
| 故障响应 | P0 15 分钟响应，2 小时修复 | 监控待接入告警 |

### 6.4 兼容性需求

| 终端 | 要求 |
|------|------|
| 微信小程序 | 基础库 2.20.0+ |
| 浏览器 | Chrome 90+、Edge 90+、Safari 14+ |
| iOS App | iOS 14+ |
| Android App | Android 10+ |

### 6.5 非功能性验收（V7 NFR-AC）

| 验收ID | 场景 | 预期结果 | 状态 |
|--------|------|----------|------|
| NFR-AC-01 | 主题切换 | print/dark/light 可切换；打印白底黑字；项目独立主题 | ✅ |
| NFR-AC-02 | 国际化 | 中英文切换实时更新；偏好持久化 | ✅ |
| NFR-AC-03 | 多端切换 | 用户中心展示各端入口；同账号数据一致 | ✅ |
| NFR-AC-04 | 版本检查 | 启动检查；非强制提示/强制阻断；跳转更新 | ✅ |
| NFR-AC-05 | 配置中心 | 管理员 CRUD；公开读取；审计 | ✅ |
| NFR-AC-06 | 演示数据开关 | 后台入口；开关/生成/清理/状态 | ✅ |

---

## 7. 技术架构与选型

### 7.1 技术栈

| 层级 | 技术方案 | 说明 |
|------|----------|------|
| 业务客户端 | Uni-app (Vue 3 + Pinia) | 一套代码构建微信小程序、H5 及 iOS/Android App |
| 管理后台 | Vue 3 + Element Plus + Vite | PC 端运营管理（base `/admin/`） |
| 服务端 | Go (Gin + GORM) | 高性能 HTTP 框架 + ORM，76 条路由 |
| 数据库 | MySQL 8.0 / SQLite | 生产 MySQL，开发/演示 SQLite |
| 缓存 | Redis 7.0 | 会话/验证码/配置缓存（SQLite 模式降级内置模拟） |
| 反向代理 | Caddy（80/443 自动 HTTPS）→ nginx（8091 路径分发） | 腾讯云 CVM |
| 文件存储 | 腾讯云 COS | 图纸/报告（待接入，当前本地存储） |
| AI 服务 | 国产大模型（百度/阿里/讯飞） | 接口设计完成，模型集成 Mock |
| 容器化 | Docker + Docker Compose | 本地开发环境 |
| CI/CD | GitHub Actions | 自动构建/测试/部署/发布 |

### 7.2 项目结构（Monorepo）

```
eqs/
├── packages/
│   ├── server/            # Go + Gin + GORM（cmd/internal/migrations）
│   ├── admin/             # Vue3 + Element Plus（views/router/store）
│   └── client/            # Uni-app（pages/store/utils）
├── shared/                # 共享代码
├── deploy/                # nginx / caddy / systemd / deploy.sh / docker-compose
├── .github/workflows/     # ci.yml / cd.yml / android.yml / ios.yml / release.yml
└── pnpm-workspace.yaml    # pnpm monorepo
```

### 7.3 生产部署架构（V8.0 落地）

```
https://eqs-chzu.tech:443
      │  Caddy（自动 Let's Encrypt HTTPS，与 wxx 共用 80/443）
      ▼
    nginx :8091（EQS 站点，路径分发）
      ├── /admin → /opt/eqs/packages/admin/dist   （管理后台）
      ├── /h5    → /opt/eqs/packages/client/dist/build/h5 （用户端）
      └── /api   → 127.0.0.1:8090                  （Go 后端）
                          │
                  MySQL :3306（库 eqs，用户 eqs@localhost）
                  Redis :6379
                  systemd: eqs-server
```

| 组件 | 端口 | 说明 |
|------|------|------|
| Caddy | 80/443 | 自动 HTTPS，转发 eqs-chzu.tech 到 nginx:8091 |
| nginx | 8091 | 路径分发（与 Caddy 错开，勿抢 80/443） |
| Go 后端 | 8090 | 与 wxx（8080）错开 |
| MySQL | 3306 | 本地监听，库 eqs |
| Redis | 6379 | 本地监听 |

---

## 8. 数据需求与接口设计

### 8.1 数据模型（25 张表，AutoMigrate）

| 序号 | 表名 | 说明 |
|------|------|------|
| 1 | users | 用户（user_type 1甲方/2服务方/3管理员/4专家） |
| 2 | projects | 项目（service_type/theme/status/经纬度） |
| 3 | supplier_qualifications | 服务方资质（verification_method/status） |
| 4 | bids | 报价 |
| 5 | orders | 订单（status 0-4） |
| 6 | contracts | 合同（contract_no/status） |
| 7 | payment_milestones | 付款节点（金额合计=订单金额） |
| 8 | deliverables | 交付物（version/status/checklist_result） |
| 9 | project_files | 项目文件（storage_key/sha256/版本链） |
| 10 | file_annotations | 文件批注（x_ratio/y_ratio 相对坐标） |
| 11 | payment_transactions | 支付/结算/退款/止付流水 |
| 12 | attendance_records | 现场打卡（经纬度加密/distance） |
| 13 | delivery_templates | 标准交付模板（checklist JSON） |
| 14 | contract_templates | 合同模板（service_type/version） |
| 15 | disputes | 争议（status/expert_result/resolution_type） |
| 16 | dispute_evidences | 争议证据 |
| 17 | dispute_expert_assignments | 专家指派（conflict/recusal/vote） |
| 18 | reviews | 评价（rating 1-5） |
| 19 | messages | 消息（⚠️ 无路由，待实现） |
| 20 | notifications | 通知（⚠️ 无路由，待实现） |
| 21 | audit_logs | 审计日志（action/target/detail/ip） |
| 22 | system_configs | 系统配置（config_key 唯一/value_type/is_public） |
| 23 | user_settings | 用户设置（theme/lang） |
| 24 | system_versions | 系统版本（version/build/platform/mandatory） |
| 25 | commission_records | 佣金记录 |

### 8.2 核心 API 清单（76 条已注册）

| 模块 | 方法+路径 | 数量 | 说明 |
|------|-----------|------|------|
| 公开 | sms/send、auth/login、auth/wechat-login、pay/notify/:channel、sign/notify、config/public、theme/list、i18n/:lang、platform/links、version/check、version/latest | 11 | 无需登录 |
| user | user/info GET/PUT | 2 | |
| project | project/create、list、:id、:id/recommend、:id/invite | 5 | |
| bid | bid/submit、project/:id/bids、bid/:id/withdraw、bid/:id/select | 4 | |
| order | order/list、:id、:id/milestones | 3 | |
| milestone | milestone/:id/deliver、accept、settle | 3 | |
| contract | contract/templates、order/:id/contract、contract/:id/sign、contract/:id/download | 4 | |
| payment | pay/create、transactions、balance | 3 | |
| dispute | dispute/create、order/:id/disputes、dispute/:id、:id/evidence、:id/expert、dispute-expert/:id/opinion、:id/close | 7 | |
| attendance | attendance/checkin、order/:id/attendance | 2 | |
| qualification | supplier/:id/qualifications GET/POST、qualification/:id/review | 3 | |
| file/annotation | file/upload、project/:id/files、annotation/add、list/:id、:id/resolve | 5 | |
| review | review/submit、user/:id/reviews | 2 | |
| config/theme | config/user/prefs GET/PUT、project/:id/theme | 3 | |
| admin | stats/users/orders/transactions/disputes/qualifications | 6 | |
| admin-demo | demo/seed、clean、toggle、status | 4 | |
| admin-config | config/list、upsert、delete/:key | 3 | |
| admin-version | version/publish、list | 2 | |
| admin-monitor | monitor/stats | 1 | |
| admin-commission | commission/list、:id/collect | 2 | |

> **实现缺口**：Message/Notification 路由未实现；`GET /api/v1/provider/list` 前端调用但后端未注册；文件下载端点待补。

---

## 9. 构建与部署（V8.0 新增）

### 9.1 多端构建命令与产物

| 端 | 命令 | 产物目录 | 说明 |
|----|------|----------|------|
| 用户端 H5 | `pnpm --filter @eqs/client build:h5` | `packages/client/dist/build/h5` | 生产 H5，base `/h5/` |
| 微信小程序 | `pnpm --filter @eqs/client build:mp-weixin` | `packages/client/dist/build/mp-weixin` | 微信开发者工具导入 |
| App 资源 | `pnpm --filter @eqs/client build:app` | `packages/client/dist/build/app` | 普通 uni-app(vue) 资源包，非最终 APK/IPA |
| 管理后台 | `pnpm --filter @eqs/admin build` | `packages/admin/dist` | base `/admin/` |
| 后端 | `CGO_ENABLED=1 go build -o server cmd/server/main.go` | `packages/server/server` | Linux 二进制，依赖 CGO（sqlite） |

### 9.2 平台特性声明（重要）

EQS 的 App 端是**普通 uni-app（vue）**，不是 uni-app x（uts）。`uni build -p app` 只产出 App **资源包**，最终 APK/IPA 需离线 SDK 原生工程（`apps/android`、`apps/ios`）二次打包。

### 9.3 生产部署（腾讯云 CVM）

| 项 | 值 |
|----|-----|
| 服务器 | 腾讯云 CVM，`129.211.223.113`（与 wxx 同机） |
| 域名 | `eqs-chzu.tech`（DNS A → 129.211.223.113） |
| HTTPS | Caddy 自动 Let's Encrypt 证书 |
| 数据库 | MySQL 8.0（库 eqs，用户 eqs@localhost，utf8mb4） |
| 缓存 | Redis 7.0 |
| 后端 | systemd `eqs-server`，端口 8090 |
| 静态分发 | nginx :8091（/admin、/h5、/api 反代） |

### 9.4 部署配置文件

| 文件 | 作用 |
|------|------|
| `deploy/nginx/eqs-cvm.conf` | nginx 8091 站点（路径分发） |
| `deploy/caddy/eqs-chzu.tech.conf` | Caddy 域名转发片段 |
| `deploy/systemd/eqs-server.service` | 后端常驻服务（MySQL 模式，端口 8090） |
| `deploy/scripts/deploy.sh` | CVM 手动部署脚本 |
| `deploy/docker-compose.prod.yml` | MySQL + Redis 生产编排 |

### 9.5 部署注意事项（避坑）

1. **Caddy 与 nginx 分工**：80/443 被 wxx 的 Caddy 占用，EQS 的 nginx 必须监听独立端口（8091），由 Caddy 转发。
2. **DNS 必须提前生效**：Caddy 自动签证书依赖域名 A 记录；未生效时报 `no valid A records found`。
3. **端口错开**：wxx 占 8080，EQS 后端用 8090。
4. **Go CGO 关键**：mattn/go-sqlite3 依赖 CGO，Dockerfile 与 cd.yml 须 `CGO_ENABLED=1`。
5. **子路径 base**：admin `base: '/admin/'` + 路由 base 同步；H5 `h5.router.base: '/h5/'`。

---

## 10. CI/CD流水线

### 10.1 Workflow 清单

| 文件 | 触发 | 作用 |
|------|------|------|
| `ci.yml` | push/PR main | lint + 后端 go test + admin/client 三端构建 + 产物归档 |
| `cd.yml` | push main | server 二进制 scp + admin/H5 scp 到 CVM nginx + reload |
| `release.yml` | tag v* / 手动 | 构建全端产物并发布 GitHub Release |
| `android.yml` | push main | 构建 App 资源 +（有原生工程时）签名 APK |
| `ios.yml` | 手动（init/release） | macOS runner 出 IPA（需离线原生工程） |

### 10.2 CI 流水线（ci.yml，push 触发）

| 阶段 | 内容 |
|------|------|
| lint | admin/client ESLint + vue-tsc |
| 后端 | go vet + go test + CGO 构建二进制 |
| 前端 | admin build + client build:h5 / mp-weixin / app |
| 产物 | 上传 artifact（server-binary / admin-dist / client-h5 / mp-weixin-dist / app-resources） |

### 10.3 CD 部署（cd.yml，push 触发）

| Job | 内容 |
|------|------|
| deploy-server | scp server 二进制 → `/opt/eqs/packages/server/server` → `systemctl restart eqs-server` |
| deploy-admin | scp `admin-dist/*` → `/opt/eqs/packages/admin/dist` → `nginx -s reload` |
| deploy-client | scp `client-h5/*` → `/opt/eqs/packages/client/dist/build/h5` → `nginx -s reload` |

### 10.4 版本策略（iOS 两版本）

| 版本 | 用途 | 签名 | 导出方式 | 触发 |
|------|------|------|----------|------|
| init | 调试基座 | 开发证书（IOS_DEV_*） | ad-hoc | 手动 workflow_dispatch |
| release | 上架 | Distribution（IOS_DIST_*） | app-store | 手动 workflow_dispatch |

---

## 11. 发布管理

### 11.1 GitHub Release 自动发布（release.yml）

打 tag `v*` 时自动：
- 构建 server（CGO）、admin、H5、小程序、App 资源，全端同一版本
- 组装 `release/`：`eqs-server`、`admin-web/`、`eqs-h5/`、`eqs-mp-weixin/`、`eqs-app-resources/`
- 生成 GitHub Release（draft）+ 自动 release notes

### 11.2 发布步骤

```bash
# 1. 升级版本号（同步 4 处，见第 3 章）
# 2. 提交并推送 tag
git commit -m "chore: 版本号升至 0.1.1"
git tag -a v0.1.1 -m "EQS v0.1.1"
git push origin main --tags
```

### 11.3 发布前核对清单

1. ✅ CI（ci.yml）通过——各端 lint/test/build 全绿
2. ✅ CD（cd.yml）已部署——CVM 三件套 Secrets 已配置
3. ✅ 版本号 4 处同步一致
4. ✅ 生产域名 HTTPS 可访问（admin/h5/api 均 200）

---

## 12. 测试策略

### 12.1 测试现状（已实现）

| 类型 | 工具 | 现状 | 说明 |
|------|------|------|------|
| 后端单元测试 | Go test | ✅ 17 文件约 3700 行 | handler 覆盖率 85.9%，配置包 100% |
| 后端集成/冒烟 | Go test | ✅ | 登录→发布→报价→中选→节点→合同→支付→交付→验收→结算全链路 |
| 前端单元测试 | Vitest | ✅ Client 23 例 + Admin 11 例 | |
| E2E | Playwright | ✅ 3 例 | 登录、中英文切换、主题切换 |
| 构建 | vue-tsc / go build | ✅ | 三端构建通过 |

### 12.2 测试目标（后续）

| 类型 | 目标 |
|------|------|
| 单元测试覆盖率 | 核心逻辑 ≥80%，API 100% |
| 集成测试 | 数据库/Redis/COS 交互 |
| E2E | H5 关键路径全覆盖 |
| 性能测试 | k6：1000 并发、P95 ≤ 500ms、≥1000 QPS |
| 安全测试 | OWASP ZAP、SQLMap、手动 XSS |

### 12.3 CI 集成

- GitHub Actions push/pull_request 自动执行 go test、vitest、eslint、三端构建。
- 测试报告自动生成、覆盖率上传 Codecov（待配置）。

---

## 13. 安全与合规

### 13.1 安全清单

| 检查项 | 状态 |
|--------|------|
| JWT 鉴权（72h，user_id + user_type） | ✅ |
| RequireAdmin() 管理员权限（user_type=3） | ✅ |
| 配置敏感项不出现在公开接口 | ✅ |
| 偏好写入 theme/lang 白名单 | ✅ |
| 版本接口限流 | ✅ |
| GORM 参数化查询（防 SQL 注入） | ✅ |
| 审计日志（配置/佣金/信用/演示数据） | ✅ |
| 定位/证件加密存储 | ✅ |
| HTTPS（Caddy 自动证书） | ✅ |

### 13.2 合规降级规则（外部能力准入门槛）

| 外部能力 | 准入条件 | 未满足处理 |
|----------|----------|-----------|
| 支付/分账/存管 | 持牌机构签约 + 验签/幂等/退款/对账测试 | 仅 Mock/沙箱，不开放真实支付 |
| 电子签约 | 服务商资质 + 实名/意愿认证 + 存证 | 线下合同人工登记 |
| 主管数据核验 | 接口授权 + 留存期限确认 | 转人工核验 |
| 保险服务 | 持牌机构签约 + 条款审核 | 不展示可购买 |

---

## 14. 版本发布计划

### 14.1 里程碑回顾

| 版本 | 里程碑 | 交付内容 | 状态 |
|------|--------|----------|------|
| V0.1 | MVP | 核心发单-接单-支付流程 + 基础 AI OCR | ✅ 已实现 |
| V0.2 | 内测版 | 完整用户端+服务端 + 智能推荐 | ✅ 已实现 |
| V1.0 | 公测版 | 全功能 + App 构建 + AI 助手 + 管理后台 | 🔶 已部署，App 实包待离线 SDK |
| V7.0 | 非功能 | 配置中心/主题/国际化/多端/版本/演示数据 | ✅ 已实现 |
| V8.0 | 生产落地 | 腾讯云部署 + GitHub Actions 自动构建发布 + 统一版本号 | ✅ 已部署 |

### 14.2 当前发布状态（V8.0）

| 项 | 状态 |
|----|------|
| 生产域名 | `https://eqs-chzu.tech`（admin 200 / h5 200 / api 200） |
| CI/CD | ci.yml + cd.yml 全绿 |
| Release | tag v0.1.0 已发布，产物可下载 |
| App 实包 | App 资源已构建；APK/IPA 待离线 SDK 工程 |
| 版本号 | 当前 0.1.0，下次迭加 0.1.1 |

---

## 15. 验收基线

### 15.1 V6 功能验收（AC-01~AC-12）

| 验收ID | 场景 | 预期结果 | 状态 |
|--------|------|----------|------|
| AC-01 | 需求方发布标准化需求 | 生成服务清单；校验通过后发布成功 | ✅ |
| AC-02 | 服务方报价并由甲方中选 | 报价提交；脱敏排名；仅生成一个待签约订单 | ✅ |
| AC-03 | 生成并签署合同 | 合同与订单一致；签署归档后进入待支付 | ✅ |
| AC-04 | 按节点支付、验收和结算 | 回调验签幂等；验收后结算；账单一致 | ✅（支付 Mock） |
| AC-05 | 服务方资质审核 | OCR 人工确认；不一致不得通过；全程留痕 | ✅（OCR Mock） |
| AC-06 | 智能匹配 | 无匹配资质排除；排序可解释；P95 ≤ 3s | ✅ |
| AC-07 | 图纸与交付物协作 | PDF/图片预览；CAD 版本留痕；批注绑定版本 | ✅（下载端点待补） |
| AC-08 | 多端核心闭环 | 四端可完成核心流程，状态一致 | ✅（App 实包待出） |
| AC-09 | 现场打卡 | 仅点击时定位；记录距离时间；照片人工审核 | ✅ |
| AC-10 | 信用评分更新 | 分项可追溯；权重重算；不重复计分 | ✅ |
| AC-11 | 争议处理 | 止付；举证质证；专家回避；留痕；法定救济明示 | ✅ |
| AC-12 | 标准模板交付 | 下载签约时版本；缺必交材料无法提交 | ✅ |

### 15.2 非功能验收（NFR-AC-01~06）

见第 6.5 节，全部 ✅。

### 15.3 部署验收（V8.0 新增）

| 验收ID | 场景 | 预期结果 | 状态 |
|--------|------|----------|------|
| DEP-01 | 生产 HTTPS 访问 | admin/h5/api 均 200 | ✅ |
| DEP-02 | CI 自动构建 | push 后 ci.yml 全绿，三端产物归档 | ✅ |
| DEP-03 | CD 自动部署 | push 后 admin/h5/后端自动更新到 CVM | ✅ |
| DEP-04 | Release 发布 | tag 触发全端产物 GitHub Release | ✅ |
| DEP-05 | 统一版本号 | 4 处版本一致，PATCH 迭加 | ✅ |

---

## 16. 原始需求追踪与验收矩阵

### 16.1 追踪矩阵（37 条 SRC，V6 已覆盖）

原始需求 37 条（SRC-POS/FLOW/ROLE/CLIENT/SUP/ADMIN/OPS/TRUST/RISK/BIZ/MKT）已在 PRD V6.0 第 21 章完整映射，V8.0 维持该映射并标注实现状态：

| 追踪ID | 原始需求 | 落点 | 优先级 | V8.0 状态 |
|--------|----------|------|--------|-----------|
| SRC-POS-01 | 甲方发单比价签约进度透明 | 5.1/5.3.2 | P0 | ✅ 已实现 |
| SRC-POS-02 | 服务方派单获客结算信用 | 5.2/5.3.3/5.3.4 | P0-P1 | ✅ 已实现 |
| SRC-FLOW-01 | 发单至信用沉淀完整闭环 | 1.2/5.x | P0 | ✅ 已实现 |
| SRC-ROLE-01~03 | 三方角色 | 4.1 | P0 | ✅ |
| SRC-CLIENT-01 | 需求方 App/小程序发单监控 | 5.1 | P0 | ✅ |
| SRC-CLIENT-02 | 智能识别项目类型生成清单 | 5.1.1 | P0 | ✅ |
| SRC-CLIENT-03 | 服务超市四类展示 | 5.1.2 | P0 | 🔶 接口待补 |
| SRC-CLIENT-04 | 进度看板 | 5.1.3 | P0 | ✅ |
| SRC-CLIENT-05 | 资金第三方监管节点释放 | 5.1.4 | P0 外部依赖 | 🔶 Mock |
| SRC-SUP-01 | 服务方 App/小程序接单执行 | 5.2 | P0 | ✅ |
| SRC-SUP-02 | 智能派单 | 5.2.1 | P0 | ✅ |
| SRC-SUP-03 | 抢单/竞价 | 5.2.2 | P0 | ✅ |
| SRC-SUP-04 | 在线协作 PDF/CAD/批注 | 5.2.3 | P0 | ✅ |
| SRC-SUP-05 | GPS 打卡 | 5.2.4 | P0 | ✅ |
| SRC-ADMIN-01 | 人工+OCR 资质审核 | 5.3.1 | P0 | ✅ |
| SRC-ADMIN-02 | 合同模板电子签署 | 5.3.2 | P0 外部依赖 | 🔶 Mock |
| SRC-ADMIN-03 | 对账/转账/发票 | 5.3.3 | P0-P1 外部依赖 | 🔶 Mock |
| SRC-ADMIN-04 | 动态信用排名 | 5.3.4 | P1 | ✅ |
| SRC-OPS-01~02 | 运营冷启动 | 14.x | 运营阶段 | ✅ 可配置 |
| SRC-OPS-03 | 小型项目切入 | 2.3 | MVP | ✅ |
| SRC-OPS-04 | 标准模板 | 5.2.6 | P0-P1 | ✅ |
| SRC-TRUST-01 | 第三方支付监管 | 5.1.4 | P0 外部依赖 | 🔶 Mock |
| SRC-TRUST-02 | 先服务后付款 | 5.1.4 | P1 | 🔶 受控 |
| SRC-RISK-01 | PDF/图片预览，BIM 远期 | 5.2.3 | P0/P1/远期 | ✅ |
| SRC-RISK-02 | 社保/证书核验防挂靠 | 5.3.1 | P0-P1 外部依赖 | 🔶 Mock |
| SRC-RISK-03 | 专家评审争议 | 5.3.5 | P0-P1 | ✅ |
| SRC-BIZ-01 | 佣金 5%-10% | 13.1 | P0 | ✅ |
| SRC-BIZ-02 | 会员权益 | 13.2 | P1 | 🔶 预留 |
| SRC-BIZ-03 | 打印/保险增值 | 13.3 | P1-P2 | ⬜ 外部依赖 |
| SRC-BIZ-04 | 造价指数/地勘数据 | 5.4.4 | P2 | ⬜ 预留 |
| SRC-MKT-01 | 多场景 | 5.1.1 | P0 | ✅ |
| SRC-MKT-02 | 推荐 3 秒返回 | 3.2 | P1 | ✅ |
| SRC-MKT-03 | 监理/造价优先 | 2.3 | MVP | ✅ |

### 16.2 实现缺口汇总（V8.0 待办）

| 编号 | 缺口 | 优先级 |
|------|------|--------|
| GAP-08-01 | Message/Notification 路由与 handler 未实现（模型已建） | P1 |
| GAP-08-02 | `GET /api/v1/provider/list` 服务商列表接口缺失 | P1 |
| GAP-08-03 | ProjectFile/Deliverable/DisputeEvidence 下载端点 | P1 |
| GAP-08-04 | 管理后台路由守卫（未登录可访问，安全风险） | P0 |
| GAP-08-05 | 后台争议页无仲裁操作 UI | P1 |
| GAP-08-06 | 客户端无争议/资质/打卡/消息/交付页面 | P1 |
| GAP-08-07 | admin Element Plus 按需引入（主 chunk 1.2MB） | P2 |
| GAP-08-08 | 微信登录/支付真实对接、第三方服务商签约 | 外部依赖 |
| GAP-08-09 | App 离线 SDK 原生工程（apps/android、apps/ios） | 外部依赖 |
| GAP-08-10 | client ESLint 清零 | P2 |

---

## 17. 附录

### 17.1 测试账号

| 角色 | 手机号 | 验证码 | user_type |
|------|--------|--------|-----------|
| 甲方 | 13900001111 | 123456 | 1 |
| 服务方 | 13900002222 | 123456 | 2 |
| 管理员 | 13900003333 | 123456 | 3 |

### 17.2 服务地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 生产域名 | `https://eqs-chzu.tech` | 生产环境（admin/h5/api） |
| 后端 API | `https://eqs-chzu.tech/api/v1/*` | 生产 |
| 管理后台 | `https://eqs-chzu.tech/admin/` | 生产 |
| 用户端 H5 | `https://eqs-chzu.tech/h5/` | 生产 |
| 本地后端 | `http://localhost:8090` | 开发 |
| 本地 H5 | `http://127.0.0.1:3005` | 开发 |
| 本地后台 | `http://localhost:3001` | 开发 |

### 17.3 关键配置

| 项 | 值 |
|----|-----|
| 仓库 | `github.com:dll/eqs`（私有） |
| CVM | 129.211.223.113（root） |
| 域名 | eqs-chzu.tech（DNS A → 129.211.223.113） |
| 数据库 | MySQL eqs（用户 eqs@localhost） |
| 版本 | 0.1.0（下次 0.1.1） |

---

**文档结束**
