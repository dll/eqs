# EQS 多端开发计划（微信小程序多端真实化）

> **文档版本**：V1.0
> **日期**：2026-08-31
> **负责人**：leader-eqs（项目级）
> **上游指令**：ccit-ceo @ EQS-MP-20260831-01（P0）
> **分支**：`feature/mp-weixin-real`（基于 `main@04a1b9f`）
> **范围**：微信小程序等多端真实化；本档为「现状→缺口→改动→人工事项→验证」总计划。

---

## 1. 各端现状总览

| 端 | 构建命令 | 现状 | 上线状态 |
|----|----------|------|----------|
| 管理后台 Admin | `pnpm --filter @eqs/admin build` | ✅ Vue3+ElementPlus，`dist/` | ✅ 已部署 CVM |
| 用户端 H5 | `pnpm --filter @eqs/client build:h5` | ✅ uni-app Vue3，`dist/build/h5` | ✅ 已部署 CVM |
| **微信小程序** | `pnpm --filter @eqs/client build:mp-weixin` | 🔶 可构建，但登录/支付/AppID 未真实化 | ⏳ 本次开发 |
| Android App | `android.yml`（需 `apps/android` 原生工程）| 🔶 仅骨架（manifest app-plus） | ⏳ 待原生工程+签名 |
| iOS App | `ios.yml`（需 `apps/ios` + 证书）| 🔶 仅骨架（manifest ios 空）| ⏳ 待原生工程+证书 |

---

## 2. 小程序现状缺口（复核 CEO 预审计）

| # | 项 | 现状 | 缺口 |
|---|----|------|------|
| 1 | `manifest.json` root/mp-weixin `appid` | 空 | 需真实微信小程序 AppID |
| 2 | 服务端 `WxLogin` | 仿真：`openid = "openid_"+code`，无 code2session，无 WX_MINI_* 变量 | 需 code2session 真实接入 + mock 开关 |
| 3 | 客户端登录 | 仅手机号+验证码；无 `uni.login`/`wx.login` | 需微信一键登录流程 |
| 4 | 小程序 JSAPI 支付 | `CreateJSAPIOrder` 已存在（需 openid，`payment.go`）| 需在 `CreatePayment` 挂接 openid + 开关隔离 |
| 5 | CI/CD | android/ios/release.yml 就绪；签名 Secrets 待配 | 见 A5 |

---

## 3. 本次改动清单（已完成）

### A1 服务端真实小程序登录 ✅
- **`internal/config/config.go`**：新增 `WXMiniAppID` / `WXMiniSecret` / `WXMiniMock`（映射 `WX_MINI_APPID` / `WX_MINI_SECRET` / `WX_MINI_MOCK`）。
- **`internal/channel/wxlogin.go`**（新增）：`WxExchanger` 接口 + `mockExchanger`（`openid_<code>`，兼容旧行为）+ `code2sessionExchanger`（调 `api.weixin.qq.com/sns/jscode2session`）。`NewWxExchanger(appID, secret, useMock)`：mock 或凭据缺失→mock；否则真实。
- **`internal/handler/auth.go WxLogin`**：改用 exchanger 取真实 openid；真实交换失败返回 400（不 panic）。
- **降级规则**：未配凭据或 `WX_MINI_MOCK=1`→mock（保留 `openid_<code>`），开发/CI/生产裸跑均可用。

### A2 客户端小程序登录流 ✅
- **`packages/client/src/store/user.ts`**：新增 `wechatLogin(code, userType)`（POST `/api/v1/auth/wechat-login`，token 入 store+storage）。
- **`packages/client/src/pages/login/index.vue`**：新增「微信一键登录」按钮；`#ifdef MP-WEIXIN` 内 `uni.login({provider:'weixin'})` → store.wechatLogin → `switchTab`。
- **条件编译保障**：小程序包含、H5 剥离（已验证 H5 构建不受影响）。

### A3 小程序支付路径打通 ✅（不真实联调）
- **`internal/handler/payment.go CreatePayment`**：新增 `channel="jsapi"` 分支——`PAYMENT_PROVIDER=wechat` 时，取甲方 `WxOpenID`（A1 登录写入），调 `CreateJSAPIOrder` 返回 `prepay_id`；无 openid/未配置→400 业务提示。
- **隔离**：`PAYMENT_PROVIDER=mock` 时 jsapi 走模拟通道，不触发真实微信调用。
- **真实联调** ⏳ 待人工提供商户号（见 §5）。

---

## 4. 测试与构建验证（A4）

| 项 | 结果 |
|----|------|
| `go build ./...` / `go vet ./...` | ✅ EXIT 0 |
| `go test ./...`（7 包）| ✅ 全 ok |
| channel 单测（成功/失败/mock/网络错误）| ✅ `internal/channel/wxlogin_test.go` |
| handler 单测（mock 开关/真实不可达 400/JSAPI 无 openid 400/JSAPI mock 200）| ✅ `wxlogin_test.go`、`final_test.go` |
| 客户端 store 单测（wechatLogin）| ✅ 6/6 pass |
| 客户端 `vue-tsc --noEmit` | ✅ EXIT 0 |
| `build:mp-weixin` 产物 | ✅ `dist/build/mp-weixin`（含登录页 uni.login+wechatLogin 编译结果）|
| `build:h5`（确认不受影响）| ✅ EXIT 0 |

---

## 5. 所需人工事项（凭据/商务/真机，非代码）

| # | 事项 | 类别 | 阻塞内容 | 处理方 |
|---|------|------|----------|--------|
| 1 | 微信小程序 `AppID`/`AppSecret` → `WX_MINI_APPID/SECRET` | 凭据 | 真实 openid（A1） | 人工 |
| 2 | `manifest.json` 填真实小程序 `appid` | 凭据 | 微信开发者工具真机/统计 | 人工 |
| 3 | 微信支付商户号 + `WXPAY_*` 全套 + `PAYMENT_PROVIDER=wechat` | 凭据+联调 | JSAPI 真实拉起支付（A3）| 商务+开发联调 |
| 4 | 真机验证微信一键登录（微信开发者工具上传体验版）| 真机 | 端到端确认 | 测试 |
| 5 | 腾讯云短信/OCR、uni-push 等（见交接清单）| 凭据 | 非本任务范畴，已在部署交接清单 | 运维 |

> 兼容策略：缺 AppID/Secret 不阻塞——服务端 `WX_MINI_MOCK=1` 走 mock，前端可导入微信开发者工具验证 UI 与流程；AppID 到位即切换真实。

---

## 6. A5 App 端盘点（只出清单，不开发）

Android / iOS 出**安装包**（APK/IPA）目前缺以下（CI 已就绪，Secrets 与原生工程齐备后自动出包）：

### 6.1 Android
| 缺失项 | 说明 | 就绪后行为 |
|--------|------|-----------|
| `apps/android/` 离线 SDK 原生 Gradle 工程 | 需 DCloud 离线打包 SDK（一次性放入仓库）；当前无该目录，`android.yml` 因 `hashFiles('apps/android/**')` 自动跳过 | push 自动构建签名 APK |
| `ANDROID_SIGN_BASE64`（jks）| 签名文件，keytool 生成后 Base64 存 Secret | 签名 |
| `ANDROID_KEYSTORE_PWD` / `ANDROID_ALIAS` / `ANDROID_ALIAS_PWD` | keytool 参数 | 签名 |
| `manifest.json` 已含 minSdk21/targetSdk30/INTERNET/abiFilters | ✅ 已具备 | — |

### 6.2 iOS
| 缺失项 | 说明 | 就绪后行为 |
|--------|------|-----------|
| `apps/ios/` 离线 SDK Xcode 工程 | 需 DCloud 离线打包 SDK；当前无该目录，`ios.yml` 自动跳过 | 手动 workflow_dispatch 出 IPA |
| 开发证书 `p12` + 开发描述文件（含 UDID）→ `IOS_DEV_*` | Appuploader（Windows）生成 | init 调试基座 |
| Distribution `p12` + AppStore 描述文件 → `IOS_DIST_*` | 上架用 | release 上架 |
| `IOS_P12_PASSWORD` / `IOS_BUNDLE_ID` | 证书密码 / Bundle ID | 签名 |
| `manifest.json` ios 段配置（当前空）| 随原生工程补齐 | — |

### 6.3 通用
- 需向 DCloud 申请离线打包 SDK 授权，维护 `apps/android`、`apps/ios` 原生工程。
- 出包前置清单已在 `docs/EQS构建部署.md §5.2/§6/§7` 与 `docs/EQS部署交接清单.md §3.1` 详述。

---

## 7. 未来工作（本期未含，另立任务）

- 小程序端「上传项目附件/资质扫描件」等文件能力真机验证。
- JSAPI 支付前端拉起 `uni.requestPayment`（需 openid 与 paySign 签名，依赖真实商户号联调）。
- App 端 uni-push 离线推送端到端（客户端初始化上报 clientid）。

---

## 8. 结项状态

| 端 | 代码 | 构建 | 仅剩人工 |
|----|------|------|----------|
| 管理后台 / H5 | ✅ | ✅ | — |
| 小程序（登录+支付接入）| ✅ | ✅ | AppID/Secret/商户号 + 真机 |
| Android / iOS | 骨架 | CI 就绪 | 原生工程 + 签名 Secrets |

---

**文档结束**
