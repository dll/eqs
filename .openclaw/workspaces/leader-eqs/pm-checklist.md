# PM 需求核对清单 — EQS 项目重构范围核对

> 核对人：pm-eqs（需求核对专员）
> 核对日期：2026-08-16
> 核对方式：只读核对，未修改任何源码
> 关键文档基线：`docs/PRD/PRD-v12.0.md`（正式第二版，V12.0）、`docs/EQS项目需求描述.md`
> 代码基线：master 分支，尚未有正式 commit（git 显示 "No commits yet"），源码目录 `packages/`、`shared/`

---

## 一、项目概览

### 1.1 产品定位
**工程快捷服务（Engineering Quick Service, EQS）**——面向工程服务业的**一站式智慧交易与协作平台**，连接设计、造价、地勘、监理四类核心服务商，覆盖「发单—匹配—签约—交付—结算—评价—沉淀」全流程线上化。差异化定位为"中小项目敏捷交易 + AI 辅助决策 + 资金担保 + 工具化（计价/CAD预览）"的轻量垂直平台。

### 1.2 技术栈
| 层 | 技术 |
|----|------|
| 后端 | Go + Gin + GORM + JWT(v5, HS256) + AES-256-GCM |
| 外部通道适配层 | 纯标准库（crypto/rsa、crypto/hmac、crypto/aes），零新增第三方依赖 |
| 管理后台 | Vue3 + Element Plus（PC + H5 管理端） |
| 用户端 | uni-app Vue3（H5 / 微信小程序 / App 三端） |
| 数据库 | MySQL（生产）/ SQLite（开发） |
| 缓存 | Redis（验证码/频控；SQLite 模式降级内置模拟） |
| 部署 | 腾讯云 CVM + systemd + Nginx + Caddy |

### 1.3 仓库结构（pnpm monorepo）
```
eqs/
├── packages/
│   ├── server/          Go 后端
│   │   ├── cmd/server/      入口（main.go 集中注册 133 条路由）
│   │   ├── internal/
│   │   │   ├── channel/     外部通道适配（短信/OCR/支付/推送/电子签）
│   │   │   ├── config/      配置加载
│   │   │   ├── dxf/         DXF 自研渲染 SVG
│   │   │   ├── handler/     业务处理器（最大代码集中点）
│   │   │   ├── middleware/  认证/CORS/日志
│   │   │   └── model/       GORM 模型 + 迁移 + 字段加密
│   │   └── migrations/      版本化迁移 SQL
│   ├── client/          uni-app 用户端（H5/小程序/App）
│   └── admin/           Vue3 + Element Plus 管理后台
├── shared/              @eqs/shared 共享 TS 类型包（注意：当前未被前端引用）
├── docs/                PRD / SAR / 安全 / 部署文档
└── deploy/              nginx/caddy/systemd/备份脚本
```

### 1.4 数据与接口规模
- **数据模型**：28 张表（用户/交易/协作/风控/运营/增值六大域）
- **接口路由**：133 条（公开 ~15 + 登录态 ~70 + 管理端 ~25）
- **单元/集成测试**：后端 handler/model/channel/dxf 均带 *_test.go；前端含 vitest（*.spec.ts）与 Playwright e2e（client）

---

## 二、业务功能清单

### 2.1 用户端（Client，`packages/client`）
| 功能域 | 页面/模块 | 说明 |
|--------|-----------|------|
| 认证 | `pages/login` | 手机号登录、微信登录、短信验证码 |
| 首页/服务超市 | `pages/index`、`pages/provider` | 分类展示服务商、服务商主页、信用/案例 |
| 发单 | `pages/project`（create/list/mine/detail） | 智能发单、服务清单、附件上传、立项批文校验、编辑/撤销/废除 |
| 报价/抢单 | `pages/bid/mine` | 提交报价、撤回、我的报价 |
| 订单/里程碑 | `pages/order`（list/detail） | 订单管理、里程碑、交付上传/验收 |
| 争议 | `pages/dispute`（list/detail） | 发起争议、证据、专家评审、结案 |
| 打卡 | —（接口在 server `attendance`） | GPS 定位打卡 |
| 资质 | `pages/qualification` | 资质提交/审核（含扫描件上传） |
| 案例沉淀 | `pages/case/mine` | 已完成订单沉淀为案例 |
| 计价工具 | `pages/tools/estimate` | 工程量清单计价估算 |
| 会员 | `pages/member/index` | 等级、权益、开通/续费 |
| 消息通知 | `pages/message/index` | 站内消息/通知、SSE 实时推送角标 |
| 我的 | `pages/mine/index` | 个人中心 |
| 模板 | `pages/template/list` | 交付模板/合同模板 |
| 管理端入口 | `pages/admin`（H5 简易管理：audit/disputes/orders/projects/settlement/users） | 与 PC 后台并存 |

### 2.2 管理后台（Admin，`packages/admin`）
| 功能域 | 视图 | 说明 |
|--------|------|------|
| 首页看板 | `views/dashboard` | 运营数据/漏斗/7 天活跃 |
| 资质审核 | `views/audit` | 资质审核 + AI 辅助 + OCR |
| 信用评分 | `views/credit` | 五维权重信用评分 |
| 项目监管 | `views/project` | 全量项目 + 甘特图/看板（GanttChart/KanbanChart） |
| 订单管理 | `views/order` | 全量订单 |
| 结算中心 | `views/settlement` | 资金流水/托管台账/佣金 |
| 纠纷仲裁 | `views/dispute` | 争议管理/专家指派 |
| 用户管理 | `views/user` | 用户列表/详情/禁启用/会员列表 |
| 系统配置 | `views/settings` | 配置中心/版本发布 |
| 日志 | `views/log` | 审计日志/配置恢复 |
| 通用 | `layout`、`login`、`router` | 布局、登录取、路由守卫 |

### 2.3 后端 API（`packages/server/cmd/server/main.go`，核心业务能力）
- **Auth**：`SendSMS`、`PhoneLogin`、`WxLogin`
- **Project**：`Create/List/Mine/Get/Update/Withdraw/Abolish/Delete`、`UploadProjectFile`、`GetRecommendations`、`InviteSuppliers`、`ListProjectReviews`、`GetProjectProgress`、`AIAnalyzeProject`、`GetServiceChecklist`
- **Bid**：`SubmitBid`、`ListBids`、`ListMyBids`、`WithdrawBid`、`SelectBid`
- **Order**：`ListMyOrders`、`Get/Update/CancelOrder`、`SetMilestones`、`UploadDeliverable`、`ConfirmAcceptance`
- **Contract**：`GenerateContract`、`SignContract`、`DownloadContract`、模板列表
- **Payment/Escrow**：`CreatePayment`、`RefundPayment`、`SettleMilestone`、流水/余额、`GetOrderEscrow`、`PaymentNotify`（微信 v3 验签）
- **Dispute**：创建/证据/专家指派/意见/结案
- **Attendance**：`CheckIn`、`ListAttendance`
- **Qualification**：资质 CRUD + 审核 + 扫描件上传
- **File/Annotation**：上传/下载/预览（DXF 渲染/DWG 适配器）/批注
- **Review/Case**：评价、案例沉淀
- **Tools**：`CostEstimate`
- **Member**：等级/信息/升级
- **Message/Notification**：站内消息 + 通知 + SSE 实时流
- **Admin**：统计/用户/订单/交易/争议/资质/演示数据/配置/版本/监控/佣金/项目进度/AI 分析/日志/托管台账/会员
- **外部通道**（`internal/channel`）：腾讯云短信/OCR、微信支付 v3、uni-push、电子签、DXF 渲染

---

## 三、重构范围清单（本次可优化对象）

> 分类：**R1（代码结构）** / **R2（性能/可维护性）** / **R3（可读性/规范）**
> 均以"不改业务行为、不改接口契约、不破坏测试"为前提。

### R1 代码结构
- **[R1-1] handler 层超大文件拆分（重点）**
  - `handler/project.go`（561 行）、`handler/config.go`（530 行）、`handler/demo.go`（450 行）、`handler/dispute.go`（442 行）、`handler/qualification.go`（434 行）、`handler/payment.go`（421 行）、`handler/file.go`（403 行）、`handler/order.go`（378 行）等文件过大。
  - 建议：按业务域拆分（如 config.go 混入 公共配置/用户偏好/主题/i18n/版本/平台链接 多个无关领域；可拆为 config、i18n、version、theme 等子文件）。**保持函数导出名与路由绑定不变**。
- **[R1-2] handler 层分层缺失**：当前业务逻辑直接在 `func Xxx(c *gin.Context)` 中内联 DB 查询，无 service 层。重构可选引入轻量 service/repository 分层，或将纯计算逻辑提取为可测函数。**注意：改动面大，需评估回归成本，建议作为选做项。**
- **[R1-3] 路由集中注册**：`main.go setupRouter` 内联 133 条路由，可提取为路由分组注册器，提升可读性。
- **[R1-4] shared 包利用率不足**：`shared/src/index.ts` 定义了 User/Project/Order/Deliverable/Payout 类型，但当前 client/admin 均未引用 `@eqs/shared`；且后端为 Go 无法引用。**属冗余包，可评估整合进 client 或后续统一 TS 类型。**
- **[R1-5] 前端页面组件提取**：client `order/detail.vue`（594 行）、`qualification/index.vue`（513 行）、`project/detail.vue`（462 行）、`project/create.vue`（428 行）与 admin 多个大 `index.vue`（最大 590 行）逻辑密集，可将区块拆分为子组件复用。

### R2 性能/可维护性
- **[R2-1] N+1 查询与 DB 访问**：`scope.go` 中 `isOrderParticipant` 在授权判断内多次 `model.DB.First` 查询；多处鉴权内嵌 DB 查询，可能在循环中重复命中。建议批量预载（Preload）或缓存，并对热点查询建索引。
- **[R2-2] 配置热点缓存**：`handler/config.go` 已有 `loadPublicCache/invalidatePublicCache`，可核对缓存失效与并发安全（RWMutex）是否到位。
- **[R2-3] Redis Key/策略**：验证码/频控 Redis key 管理与过期策略，SQLite 模式下模拟降级是否正确，可复核。
- **[R2-4] 前端编译/包体积**：uni-app 多端构建，可核对公共依赖（vue 3.4.21 overrides 版本锁定）与按需引入 Element Plus（admin）。
- **[R2-5] 测试体积与组织**：handler 下存在大量 `*_test.go`（`coverage_extra_test.go` 1068 行、`cover_more_test.go` 474 行、`extended/final/flow/v9/v10` 等），测试文件混杂且大，可整理分组。**重构不得删改这些测试的断言语义。**

### R3 可读性/规范
- **[R3-1] 魔法数字与状态码**：模型中的 `status: 0|1|2|3|4` 等魔法数字散落（如 order/project 状态），建议收敛为常量/枚举，提升可读性。**不改数据库落库值。**
- **[R3-2] 统一错误码管理**：`response.go` 的 `fail/badRequest/unauthorized/notFound/serverError` 已统一，但各 handler 内 `message_key` 字符串散落，可抽取错误码常量表。
- **[R3-3] 注释/命名规范化**：遵循根 `AGENTS.md`（中文注释、英文标识符、golangci-lint / ESLint / Vue 风格），可统一检查。
- **[R3-4] 字段加密读写法规范**：`model/crypto.go`（216 行）集中 AES-256-GCM 加解密，可核对覆盖的敏感字段范围是否一致（证书号/经纬度）。

---

## 四、不可改动边界（核心业务逻辑红线）

> 以下任何改动都可能破坏契约、数据安全或市场/验收合规，**除非单独立项，否则严禁修改**。

| 边界 | 说明 |
|------|------|
| **[B-1] 接口契约（133 条路由）** | 路径、方法、请求/响应字段（`error/message/message_key`）、认证方式（Bearer/SSE 签名参数）不得变更 |
| **[B-2] 数据模型与迁移** | 28 张表结构、字段名、枚举状态值、`AutoMigrate` 顺序、`migrations/001_init.sql`；**不改数据库落库值** |
| **[B-3] 安全/合规逻辑** | JWT HS256 白名单、角色以 DB 为准、对象级授权（`scope.go`）、`RequireAdmin`、CORS 白名单、AES-256-GCM 字段加密、手机号脱敏、备份文件 600、生产拒绝弱密钥/缺加密密钥启动（`main.go` P0-07/P1-09 检查） |
| **[B-4] 外部通道适配层** | `internal/channel`（腾讯云签名 TC3-HMAC-SHA256、微信支付 v3 平台证书 RSA 验签 + AES-GCM 解密、uni-push、电子签）——含"黄金测试"（crypto_test/payment_test），不得改动签名/解密算法 |
| **[B-5] 支付/资金/结算/争议仲裁流程** | 里程碑结算、资金托管台账（EscrowLedger）、佣金计算（含会员折扣）、争议资金冻结/解冻记账、三专家评审——涉及真实资金与对账，行为改动需严格评审 |
| **[B-6] AI 辅助判定逻辑** | 资质 AI 审核门禁（证书号/发证机关/有效期/附件完整）、规则降级、OCR 识别附加——改动影响审核准确性与合规 |
| **[B-7] 演示数据逻辑** | demo/test/training 三种模式种子与一键清理（保留管理员），供验收/培训，行为不改 |
| **[B-8] 部署与诊断脚本** | `deploy/`（nginx/caddy/systemd）、Dockerfile、.env 加载、优雅停机 | 
| **[B-9] 既有测试断言** | handler/model/channel/dxf 的所有 `*_test.go`、client 的 `*.spec.ts`、admin 的 KanbanChart.spec、Playwright e2e——重构不得使现有测试语义失效 |

> **重构总原则**：只允许"不改行为"的重排、拆分、提取、规范化、批处理与索引优化；涉及状态机转移、金额计算、加密、鉴权边界的一律只读不动，改动需走独立评审。

---

## 五、风险提示

| 风险 | 等级 | 说明与建议 |
|------|------|-----------|
| **无版本基线（git 无 commit）** | 🔴 高 | 项目尚未有正式 commit，重构前必须先行建立 git 基线提交，否则无法回滚、无法按 commit 比对回归。 |
| **handler 层重构破坏接口行为** | 🔴 高 | 分拆超大文件若误改 `func` 签名或路由绑定，将影响 133 条接口。强烈建议"只移动不改写"，并以现有测试为回归兜底。 |
| **无可运行 E2E/回归环境依赖** | 🟠 中 | 依赖既有 vitest/Go test 与 Playwright e2e。重构每步后需完整跑通 `go test ./...`、`vitest run`、`pnpm lint`。 |
| **大模型（Model）为请求偏好而非约束** | 🟢 提示 | 本清单基于代码与 PRD 静态核对；若本次重构有更具体的优化目标（性能/结构/规范权重），需任务发起方明确排序。 |
| **shared 包废弃倾向** | 🟢 低 | @eqs/shared 未被打包引用，整合或删除需确认后续 frontend 改版计划，避免误删未来契约。 |
| **数据库迁移与索引** | 🟢 低 | R2-1 若引入索引，SQLite 与 MySQL 需分别验证；SQLite AutoMigrate / MySQL migration 路径不同。 |
| **外包通道凭据缺失时的降级路径** | 🟠 中 | 未配置凭据时自动降级（Mock/规则/SSE 轮询）；重构不得破坏这些"代码就绪、待凭据"的降级分支。 |

---

## 六、建议重构落地顺序（供任务方参考，非强制）

1. **建立 git 基线**（先解决风险 1）。
2. **R1-1 / R1-3 handler 与路由拆分**（收益大、可回归测试兜底，"只移动不改写"）。
3. **R3-1 / R3-2 状态常量与错误码收敛**（纯可读性，风险低）。
4. **R2-1 鉴权/列表 N+1 与索引**（性能，需 DB 双驱动验证）。
5. **R1-5 前端大组件拆分子组件**。
6. **R1-4 shared 包收编决定**。
7. 其余（R1-2 引入分层）作为选做，评估回归成本后单独立项。

---

*核对完成。如需针对某一模块输出更细的"文件级重构明细"（如 config.go 拆分子文件方案、project.go 函数清单），可单独指派。*
