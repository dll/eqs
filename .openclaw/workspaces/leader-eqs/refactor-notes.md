# EQS 重构改动记录（refactor-notes）

> 重构专员：dev-refactor-eqs
> 日期：2026-08-16
> 分支：`refactor/dev-cleanup`（基于 `main` 创建）
> 依据：`pm-eqs` 的 `pm-checklist.md`（R1/R2/R3 范围清单 + 不可改动边界 B-1~B-9）
> 总原则：**只移动不改写**——保持函数导出名、路由绑定、接口契约、数据库落库值、业务逻辑完全不变；以既有 Go 测试套件回归兜底。

---

## 0. 基线建立（解决清单「风险 1：无版本基线」）

- 项目实际已有完整 commit 历史（main 上 7+ 次提交，HEAD=b327a26），工作区原先干净。
- 但为隔离本次重构、便于按 commit 回滚，新建专用分支 **`refactor/dev-cleanup`**，并提交**零净改动基线** commit `6e9fa1a`。
- 基线提交同时将 `.openclaw/`（各 agent 工作区，含各自嵌套 `.git`）加入 `.gitignore`——这是运行代理基础设施，并非项目源码，不应纳入版本库。改动仅添加 3 行 gitignore，不影响源码。
- **基线验证**：创建分支前跑通 `go build ./...` 与 `go test ./...`（全部包 ok），作为回归对照。

---

## 1. [R1-1] 拆分超大 handler 文件（纯移动，不改逻辑）

### 1.1 config.go（481 行）→ 7 个单一职责子文件 ✨ 重点
- 原 `config.go` 混入 5 个无关领域，按业务域垂直拆分为（均在 `package handler` 内，无需改 import / 路由绑定）：
  - `config_cache.go`：`configCache`、`publicCache`、`loadPublicCache`、`invalidatePublicCache`、`getPublicCached`（配置热点缓存）
  - `config_center.go`：`AdminListConfigs`、`AdminUpsertConfig`、`AdminDeleteConfig`、`PublicConfigs`、`parseConfigValue`、`UpsertConfigRequest`
  - `config_user.go`：`GetUserPrefs`、`UpdateUserPrefs`、`UpdatePrefsRequest`
  - `config_theme.go`：`ThemeList`、`SetProjectTheme`
  - `config_i18n.go`：`I18nMessages`、`loadI18n`
  - `config_platform.go`：`PlatformLinks`
  - `config_version.go`：`VersionCheck`、`VersionLatest`、`compareVersions`、`splitVersion`、`AdminPublishVersion`、`AdminListVersions`
- **移出的全部 21 个函数**与原代码逐字节一致（仅调整文件内 import 分组），函数名、签名、逻辑、`message_key` 文案未变动。
- **不改业务逻辑验证**：
  - `go build ./...` 通过；
  - handler 定向测试 `TestConfigCenter_*`、`TestPlatformLinks`、`TestVersionLatest`、`TestAdminPublishVersion`、`TestSetProjectTheme*`、`TestParseConfigValue` 全部通过；
  - 全文 `go test ./...` 全绿。
- 拆分后 config 相关文件从 1 个 mixed 文件（530 行含测试）变为 7 个单一职责文件（不含测试），可读性与可维护性显著提升。

### 1.2 file.go（363 行）→ 批注独立为新文件 `annotation.go`
- 原 `file.go` 混合同一资源的两类操作：**文件上传/列表/下载/预览（DXF/CAD 渲染）** 与 **文件批注**。
- 将批注域 `AddAnnotationRequest`、`AddAnnotation`、`ListAnnotations`、`ResolveAnnotation` 移入新文件 `annotation.go`；file.go 保留上传/下载/预览/公开预览/CAD 转换逻辑。
- **不改业务逻辑验证**：`go build ./...` 通过；handler 定向测试 `*Annotation*`/`*File*`/`*Upload*`/`*Preview*`/`*Download*`/`*CAD*` 通过；全文测试全绿。

> **为何不拆分其余大文件**（project/order/payment/dispute/qualification/demo）：
> - 均属**单一业务域**，内部逻辑高度相互引用，拆分需触碰状态机/金额/加密等红线上下文（B-2/B-5/B-6/B-7）。
> - 这些文件虽大但主题内聚；按「只移动不改写」原则强行拆分风险 > 收益，故保持原状，符合清单「改法需走独立评审」的告诫。

---

## 2. [R1-3] 路由集中注册 —— 评估后不改
- `main.go setupRouter` 已按业务域分组（公开 / 登录 / 管理端），且含清晰注释。
- 133 条路由内联但结构良好；抽为分组注册器收益边际、风险偏高，**决定不动**（避免触碰 B-1 接口契约）。

---

## 3. 已核实但未改动（受红线保护）

| 清单项 | 结论 |
|--------|------|
| B-3 `scope.go` 对象级授权 / R2-1 N+1 | 鉴权内 DB 查询属安全边界，只读不改；批量预载/索引涉及 DB 双驱动（SQLite/MySQL）验证，超出本次安全范围，建议单独立项（触发 PM 评审）。 |
| B-5 支付/资金/结算/争议流程 | `payment.go`/`order.go`/`escrow.go`/`dispute.go` 属资金与仲裁，**一律未改**。 |
| B-6 AI 审核 / B-7 演示数据 | `ai.go`/`qualification.go`/`demo.go` 涉及审核门禁与种子逻辑，**只读不改**。 |
| B-4 外部通道加密/签名 | `internal/channel` 含黄金测试，**零改动**。 |
| R3-1 魔法数字收敛 | 状态码多为订单/项目/资金状态机（B-2/B-5），收敛为常量需跨多文件替换，回归风险与改动面大；保留为后续单独立项。现有 `response.go` 错误码助手已统一（`fail/badRequest/unauthorized/notFound/serverError/forbidden`），`message_key` 传播本体已良好。 |

---

## 4. 验证与回归

- `go build ./...`：通过（重构前、中、后均验证）。
- `go vet ./internal/handler/`：通过。
- `go test ./...`：全部包 `ok`（handler 10.1s，channel/config/dxf/middleware/model/cmd 全绿）。
- 重构仅改动 Go 后端 handler 层文件；client/admin 前端（vitest/Playwright）未触碰，不受影响。
- 每次拆分均独立 commit，可单独回滚（见 §5）。

---

## 5. 提交清单（refactor/dev-cleanup 相对 main）

| commit | 内容 | 可回滚性 |
|--------|------|---------|
| `6e9fa1a` | 重构基线 + `.openclaw/` 加入 gitignore | 基线 |
| `f207a66` | config.go → 7 子文件（纯移动） | `git revert f207a66` 即还原 |
| `c09efd0` | file.go 批注 → annotation.go（纯移动） | `git revert c09efd0` 即还原 |
| `ef508bc` | config_i18n.go gofmt 对齐（纯格式化） | `git revert ef508bc` 即还原 |

工作区干净，无未提交改动。

---

## 6. 遗留建议（交由 PM 排期）

1. **R2-1 N+1/索引**：`scope.go` 鉴权内 DB 查询批量化 + 热点索引（需 SQLite+MySQL 双驱动验证）——单独立项。
2. **R1-5 前端大组件拆分**：client `order/detail.vue`(594) 等——本次未做（后端专项）。
3. **R1-4 shared 包收编**：确认前端改版计划后再决定整合/删除。
4. **R3-1 状态常量收敛**：建议随状态机专项评审一并推进。

*重构完成，业务逻辑与接口契约均未改变。*

---

## 7. EQS-MP-20260831-01：微信小程序多端真实化（leader-eqs 执行）

> 日期：2026-08-31 ｜ 分支：`feature/mp-weixin-real`（基于 main@04a1b9f）
> 上游：ccit-ceo（P0，@from ccit-ceo @to leader-eqs）
> 定位：在既有 133 路由/28 模型基础上，补小程序登录与支付真实化；未新增业务规则，未触碰资金/状态机红线。

### 7.1 服务端（A1）
- `internal/config/config.go`：新增 `WXMiniAppID/WXMiniSecret/WXMiniMock`（← `WX_MINI_APPID/WX_MINI_SECRET/WX_MINI_MOCK`）。
- `internal/channel/wxlogin.go`（新增）：`WxExchanger` 接口；`mockExchanger`（openid_<code>，兼容旧行为）；`code2sessionExchanger`（真调 api.weixin.qq.com/sns/jscode2session）。`NewWxExchanger(appID,secret,useMock)`：mock/缺凭据→mock，否则真实。
- `internal/handler/auth.go WxLogin`：改用 exchanger；真实交换失败返回 400（不 panic）。
- 单测：`internal/channel/wxlogin_test.go`（mock 3 例 + 真实成功 + 微信错误 + 网络错误）；`internal/handler/wxlogin_test.go`（mock 开关 200 + 真实不可达 400）。

### 7.2 客户端小程序登录（A2）
- `packages/client/src/store/user.ts`：新增 `wechatLogin(code,userType)`（POST /auth/wechat-login，token 入 store+storage）。
- `packages/client/src/pages/login/index.vue`：新增「微信一键登录」按钮；`#ifdef MP-WEIXIN` 内 `uni.login({provider:'weixin'})` → wechatLogin → switchTab。
- 条件编译：小程序含、H5 剥离（H5 构建已验证不受影响）。store 单测 `user.spec.ts` 新增 wechatLogin 用例。

### 7.3 小程序 JSAPI 支付（A3，不真实联调）
- `internal/handler/payment.go CreatePayment`：新增 `channel="jsapi"` 分支——`PAYMENT_PROVIDER=wechat` 时取甲方 `WxOpenID`，调 `CreateJSAPIOrder` 返回 prepay_id；无 openid/未配置→400。`PAYMENT_PROVIDER=mock` 时走模拟通道，不触发真实调用。
- 单测：`final_test.go` 新增 `TestCreatePayment_Jsapi_NoOpenid`（400）、`TestCreatePayment_Jsapi_Mock`（200）。

### 7.4 凭据与文档
- `deploy/.env.example`：新增 WX_MINI_APPID / WX_MINI_SECRET / WX_MINI_MOCK。
- `docs/EQS外部通道接入指南.md`：新增微信小程序登录章节 + 结项清单条目。
- `docs/EQS多端开发计划.md`（新增）：各端现状/缺口/改动/A5 App 盘点/所需人工事项/验证。

### 7.5 验证
- `go vet` / `go test ./...`：7 包全 ok。
- 客户端 `vue-tsc --noEmit`：EXIT 0；vitest store 6/6 pass。
- `build:mp-weixin`：产物 `dist/build/mp-weixin`，登录页已含 uni.login+wechatLogin；`build:h5` EXIT 0。

### 7.6 所需人工事项（非代码）
- 小程序 AppID/AppSecret（→ WX_MINI_* 与 manifest appid）；微信支付商户号 + WXPAY_*；真机上传体验版验证；Android/iOS 原生工程 + 签名 Secrets（仅出安装包用，见计划文档 §6）。

*本轮未改既有业务逻辑与接口契约；新增登录/支付能力均带开关隔离可回退。*
