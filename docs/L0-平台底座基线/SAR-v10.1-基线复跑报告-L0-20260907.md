# SAR-v10.1 基线复跑报告（L0 平台底座 · leader-eqs 实测）

> 文档版本：L0-BASELINE-20260907
> 复跑人：leader-eqs（repository `eqs`，dev 窗口）
> 复跑日期：2026-09-07
> 承接：ccit-ceo 转单 → ccit-cto「EQS-平台底座路线与排期」L0 基线盘点
> 复跑基线：SAR-v10.1（文档 `docs/SAR/SAR-v10.1.md`，源码锚点 commit `04a1b9f`）
> 复跑对象：当前工作区 `main` 分支（HEAD `d00bcd5`）
> 结论方向：平台底座建设前的**交易域基线冻结确认**（只读实测，零改动）

---

## 0. 执行摘要

本轮 L0 承接 CTO 路线，利用 leader-eqs 仓库源码可达、可实测的优势，对 **SAR-v10.1「133 路由 / 28 模型」基线**在**当前工作区**做一次强制全新复跑（`-count=1`），并核验该基线自 SAR-v10.1 锚点以来是否发生漂移。

复跑结论：
- **133 路由 / 28 模型两项核心指标在当前 `main` HEAD 上仍然精确成立**，与 SAR-v10.1 报告口径完全一致。
- **路由注册表（`cmd/server/main.go`）与持久化模型集（`internal/model` 的 AutoMigrate）自 SAR-v10.1 锚点（`04a1b9f`）至 HEAD 零改动**（git diff 为空）——平台底座最关键的交易域"路由面 / 模型面"未发生任何漂移。
- 全链路回归（`go build` / `go vet` / `go test -count=1 -p 1 ./...`）7 包全绿，EXIT 0，符合 SAR-v10.1「可发布/未改破」判据。
- SAR-v10.1 之后合入的后端改动仅有**微信小程序真实化（通道适配层加法）**，均为新增适配文件/内部增强，未增删任何路由端点、未增删/改动任何持久化模型、未触及落库值与状态机语义红线。

---

## 1. 复跑口径与证据来源

### 1.1 两个核心指标的精确定义（口径锚定，供平台底座引用）

| 指标 | 精确口径 | 权威文件 |
|------|----------|----------|
| **133 路由** | `cmd/server/main.go::setupRouter` 内集中注册的 REST 端点行数（每行一个 HTTP 动作端点，非 group 数） | `packages/server/cmd/server/main.go` |
| **28 模型** | `internal/model/db.go::AutoMigrate` 全量迁移清单的模型个数 | `packages/server/internal/model/db.go` |

> 注：全仓仅 `main.go` 一处完成生产路由注册（`gin.Default`→`/api/v1`）；其余 `gin.New/Group` 均出现在 `_test.go` 的测试性路由粘贴中，不属于产品路由面。此口径与 SAR-v10.1 一致，可作为平台底座后续对"路由/模型面是否漂移"的机器可校验锚点。

### 1.2 当前工作区状态核验

- 未跟踪源码改动：`git status` 显示**无任何 `.go` 源码文件为 modified**（工作区源码 == HEAD）。
- 未提交项仅为历史遗留的 `.openclaw/workspaces/*` 删除与新增治理文档（非源码，不在本轮作用域）。
- 源码锚点一致性：`04a1b9f`(SAR-v10.1 R2-2 修复) → `HEAD(d00bcd5)` 期间，`cmd/server/main.go` 与 `internal/model/` 目录 **diff 为空**。

---

## 2. 路由 / 模型计数实测

### 2.1 133 路由 —— 复跑命中

| 分组 | 端点数 |
|------|--------|
| 公开 `api.`（含 V7/V8/V9 公开浏览、回调、公开预览） | 16 |
| 登录 `auth.`（认证后各业务域操作） | 90 |
| 管理员 `admin.`（后台管控/看板/对账） | 27 |
| **合计** | **133** |

方法分布复核：`GET 67 / POST 45 / PUT 16 / DELETE 5`（=133，自查自洽）。

### 2.2 28 模型 —— 复跑命中

`AutoMigrate(db)` 迁移清单逐一清点 = **28 个业务模型**：

用户/认证 `User`；项目域 `Project`；`SupplierQualification`；`Bid`；订单交易 `Order`、`PaymentMilestone`、`Deliverable`；`Contract`；文件 `ProjectFile`、`FileAnnotation`；资金 `PaymentTransaction`、`EscrowLedger`、`CommissionRecord`；打卡 `AttendanceRecord`；模板 `DeliveryTemplate`、`ContractTemplate`；争议 `Dispute`、`DisputeEvidence`、`DisputeExpertAssignment`；社交 `Review`、`Message`、`Notification`；审计 `AuditLog`；配置域 `SystemConfig`、`UserSetting`、`SystemVersion`；案例 `CaseShowcase`；会员 `MembershipOrder`。

> 附注：`model/migrate.go` 的 `SchemaMigration`（迁移版本记录）非 AutoMigrate 业务清单成员，不计入"28 模型"口径；已在迁移架构中作为版本化辅助表独立存在。

---

## 3. 回归验证实录（强制全新复跑，`-count=1`）

| 项 | 命令 | 结果 |
|----|------|------|
| 后端构建 | `go build ./...` | ✅ EXIT 0 |
| 后端静态检查 | `go vet ./...` | ✅ EXIT 0 |
| 后端全量测试 | `go test -count=1 -p 1 ./...` | ✅ 7 包全 ok，EXIT 0 |

7 包与实耗：
```
ok  github.com/eqs/server/cmd/server         4.400s
ok  github.com/eqs/server/internal/channel   3.329s
ok  github.com/eqs/server/internal/config    1.544s
ok  github.com/eqs/server/internal/dxf       1.496s
ok  github.com/eqs/server/internal/handler   8.830s
ok  github.com/eqs/server/internal/middleware 3.812s
ok  github.com/eqs/server/internal/model     3.356s
```
与 SAR-v10.1 的包构成（cmd/channel/config/dxf/handler/middleware/model）完全一致，无新增包、无断链包。

### 3.1 ⚠️ 复跑环境陷阱（重要：自动化/验收跑测试的必读）

> 复跑过程中发现，**本开发主机每一个 exec/子进程都会继承其它项目（wxx/蔚小芯）注入的环境变量**，例如 `JWT_SECRET=79d1dee4dafe…cbab`（与本仓库根 `.env` 中 wxx 项目同名变量一致）、`WX_MINI_APPID/SECRET`、`ZHIPU_*/DEEPSEEK_*` 等。若直接跑 `go test ./...`，会触发**假失败**而非源码回归：

| 现象 | 根因 | 印证 |
|------|------|------|
| `config_test.go TestLoad_Defaults` FAIL：JWT 默认值异常 | 断言默认凭据 `eqs-secret-key`，但继承到真实 `JWT_SECRET` 使 `os.Getenv` 命中非默认值 | 净化后 `go test -count=1 ./internal/config/` → ok |
| `auth_extra_test.go TestWxLogin` 400 | 继承到真实 `WX_MINI_APPID/SECRET` 使 `NewWxExchanger` 走真实 code2session（`useMock=false`），无网/无效凭据 → 400（期望 mocks 200/400 分支见 config 导出测试） | 净化后 `-run Wx` → ok |

**正确复跑姿势**（在同一进程先清再测，勿依赖跨进程继承）：
```powershell
cd packages/server
'JWT_SECRET','JWT_EXPIRE_HOURS','APP_ENV','SERVER_PORT','DB_DRIVER','DB_NAME','DB_USER','DB_PASSWORD','DB_HOST','DB_PORT','REDIS_ADDR','REDIS_PASS','WX_MINI_APPID','WX_MINI_SECRET','WX_MINI_MOCK','DATA_ENCRYPTION_KEY','ZHIPU_API_KEY','DEEPSEEK_API_KEY' | ForEach-Object { Remove-Item "env:$_" -ErrorAction SilentlyContinue }
go test -count=1 -p 1 ./...
```

**判定**：本报告 §3 的 7 包全绿，即**在此净化环境下**测得（上面刚重跑：cmd/channel/config/dxf/handler/middleware/model 全 ok，EXIT 0）。此前一次直接跑（未净化）config、handler 两个包假失败，净化后均恢复 ok —— 说明**是环境变量污染、非源码回归**。后续 L1→L3 每笔合入回归也须采用此净化姿势；CI 若在干净容器跑则天然规避。

---

## 4. SAR-v10.1 锚点至今的源码漂移审计

### 4.1 交易域关键面：零改动（git diff 为空）

- `packages/server/cmd/server/main.go`（133 路由注册面）：**零改动**
- `packages/server/internal/model/*`（28 模型/AutoMigrate/DB 层）：**零改动**

### 4.2 该窗口内全部后端改动（均为小程序真实化通道加法）

`git diff --name-only 04a1b9f HEAD -- packages/server/` 仅 7 个文件：

| 文件 | 性质 |
|------|------|
| `internal/channel/wxlogin.go`(+89) | **新增**微信号通道适配层（code2session 验签） |
| `internal/channel/wxlogin_test.go`(+117) | 新增单测 |
| `internal/config/config.go`(+11) | 新增微信 appid/secret 配置项（字段加法） |
| `internal/handler/auth.go`(改) | 复用既有 `/auth/wechat-login` 端点，填入真实 channel 逻辑 |
| `internal/handler/payment.go`(+58) | 复用既有 `/pay/create` 端点，内部新增 JSAPI 支付分支 |
| `internal/handler/final_test.go`(+36)、`wxlogin_test.go`(+52) | 测试补充 |

**审计判定**：上述改动未新增/删除**任何**路由端点（注册行集合与 `04a1b9f` 完全一致，Compare 为空）、未增删/改动**任何**持久化模型、未改落库值规则与交易状态机语义（资金/dispute/AI 的现有路由与状态迁移不被触碰）。属"通道适配层 + 端点内部实现增强"的纯加法，正是平台底座期望在交易域外层收敛的形态。

---

## 5. L0 结论与建议（提交 CTO 门禁）

1. **基线冻结成立**：以「133 路由（公开 16/登录 90/管理员 27）/ 28 模型」作为平台底座 L1 横切加固、L2 可服务化的**现状基线锚**，可机器复算（锚点文件见第 1.1 节）。
2. **交易域回写式保护**：后续任何改造，凡改动 `main.go` 注册行集合或 `model/db.go` AutoMigrate 清单，即视为"交易域面漂移"，须经 CTO 门禁 + 全量回归（见《既有交易域保护清单》）。
3. **回归兜底有效**：既有 Go 测试回归（7 包）可作为贯穿 L1→L3 的兜底红线；本报告为"只做加法不动交易域"提供了可复现的绿证基线。

**附产出**（同目录 L0-平台底座基线）：
- `既有交易域保护清单-L0-20260907.md`
- `变更影响模板-L0-20260907.md`

---

*—— leader-eqs，L0 基线复跑（实测证据补足 CTO 工作区源码盲区），2026-09-07。*
