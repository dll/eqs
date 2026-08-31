# EQS 外部通道接入指南（结项用）

> 文档版本：V1.0
> 更新日期：2026-08-14
> 适用：`deploy/.env.example` 中的全部外部通道凭据；**代码已开发完成，仅剩填写凭据与联调验证**

---

## 0. 总览

| 通道 | 代码状态 | 所需凭据 | 填好后行为 | 留空时降级 |
|------|----------|----------|------------|------------|
| 腾讯云短信 | ✅ 已开发（TC3 签名黄金测试通过） | `TENCENT_SMS_SECRET_ID/KEY` + `SMS_SDK_APP_ID/SIGN_NAME/TEMPLATE_ID` | 生产登录真实发送验证码 | 演示号固定码；非演示号拒绝并提示"短信服务未配置" |
| 腾讯云 OCR | ✅ 已开发（TC3 签名复用） | `TENCENT_OCR_SECRET_ID/KEY` | 资质审核自动识别扫描件（审核备注前缀"OCR识别:"） | 规则/AI 辅助审核 |
| 微信支付 v3 | ✅ 已开发（签名+回调解密黄金测试通过，**未联调**） | `WXPAY_APPID/MCHID/API_V3_KEY/MCH_SERIAL_NO/私钥/平台证书/通知URL` + `PAYMENT_PROVIDER=wechat` | `/pay/create` 返回微信 Native 二维码；回调验签后订单自动进入"已支付" | `PAYMENT_PROVIDER=mock` 模拟通道 |
| 微信小程序登录 | ✅ 已开发（code2session；随附黄金/单测） | `WX_MINI_APPID` + `WX_MINI_SECRET`（可选 `WX_MINI_MOCK`） | 真实 openid 落库并登录 | 未配置或 `WX_MINI_MOCK=1` 降级 mock（openid_<code>） |
| App 推送（uni-push/个推） | ✅ 已开发（token 签名实现） | `PUSH_APP_ID/APP_KEY/MASTER_SECRET` | 通知创建后按用户手机号别名离线推送 | SSE 实时推送（H5）/30s 轮询（小程序/App） |
| 电子签 | 🔶 网关抽象就绪（Mock 实现） | 需选定服务商（法大大/e签宝/上上签）后开发适配器 | — | Mock 签署链接 |
| CAD 引擎 | ✅ 接入点就绪 | `CAD_CONVERT_API`（自建转换服务地址） | DWG 在线预览 | 提示下载 |
| 会员/保险/造价指数 | ✅ 会员已自研；保险/指数待合作方 | 商业立项 | 会员开通即生效（模拟支付） | — |

---

## 1. 腾讯云短信

### 步骤
1. 腾讯云控制台 → 短信 → 创建应用（获取 `SMS_SDK_APP_ID`）；
2. 创建签名（`SMS_SIGN_NAME`）与正文模板（验证码模板，含 `{1}` 参数，`SMS_TEMPLATE_ID`）；
3. 访问管理 → API 密钥管理 → 创建密钥（`TENCENT_SMS_SECRET_ID/KEY`）；
4. 填入 `deploy/.env`（生产为 `/opt/eqs/.env`），重启服务。

### 验证
```bash
curl -X POST http://127.0.0.1:8090/api/v1/sms/send -H "Content-Type: application/json" \
  -d '{"phone":"13800000001"}'
# 手机收到验证码；Redis 中 sms:<phone> 存在且 5 分钟有效
```

### 说明
- 代码：`internal/channel/sms.go`（TC3-HMAC-SHA256 签名，黄金测试 `TestTencentCloudSign_Golden`）。
- 演示手机号（1390000 前缀）始终走固定码，不受短信服务影响。

---

## 2. 腾讯云 OCR

### 步骤
1. 腾讯云控制台 → OCR → 开通通用印刷体识别；
2. 创建 API 密钥（可与短信共用）；
3. 填入 `TENCENT_OCR_SECRET_ID/KEY`。

### 验证
管理后台审核一条带扫描件的资质：审核备注应出现 `OCR识别:...` 前缀（识别出的证书文字）。

### 说明
- 代码：`internal/channel/ocr.go`；调用点 `qualification.go` 的 `ocrAssistQualification`。
- 识别失败/未配置不影响审核流程（静默跳过）。

---

## 3. 微信小程序登录（code2session，V11 新增）

### 步骤
1. 微信公众平台 → 小程序 → 开发管理 → 开发设置 → 获取 `AppID` 与 `AppSecret`；
2. 填入 `deploy/.env`：`WX_MINI_APPID` / `WX_MINI_SECRET`；
3. 开发/CI 可设 `WX_MINI_MOCK=1` 强制 mock（openid_<code>），不真实请求；
4. 重启服务。客户端在登录页点「微信一键登录」即可走真实登录。

### 验证
```bash
curl -X POST http://127.0.0.1:8090/api/v1/auth/wechat-login -H "Content-Type: application/json" \
  -d '{"code":"<真实验证临时code>","user_type":2}'
# 返回真实 openid 对应的用户 token（不再出现 openid_<code> 前缀）
```

### 说明
- 代码：`internal/channel/wxlogin.go`（code2session 交换器 + mock 降级）；调用点 `auth.go WxLogin`；
- 未配置或 mock 时维持旧仿真行为（openid_<code>），单测覆盖成功/失败/mock 三路径；
- 客户端：`pages/login/index.vue` 的「微信一键登录」按 `#ifdef MP-WEIXIN` 条件编译，H5 不受影响。

---

## 4. 微信支付 v3（唯一"未联调"通道）

### 步骤
1. 微信商户平台开通 Native 支付（扫码）与退款权限；
2. 下载商户私钥 `apiclient_key.pem` 与商户证书序列号（`WXPAY_MCH_SERIAL_NO`）；
3. 设置 APIv3 密钥（32 位，`WXPAY_API_V3_KEY`）；
4. 从微信支付平台下载平台证书（用于回调验签，`WXPAY_PLATFORM_CERT_FILE`）；
5. 配置回调地址 `WXPAY_NOTIFY_URL`（公网可达）；
6. `deploy/.env`：`PAYMENT_PROVIDER=wechat` + 全部 `WXPAY_*`。

### 验证（联调清单）
```bash
# 1) 下单：应返回 code_url（weixin://wxpay/bizpayurl?...）
curl -X POST http://127.0.0.1:8090/api/v1/pay/create -H "Authorization: Bearer <甲方token>" \
  -H "Content-Type: application/json" -d '{"order_id":1,"amount":3000,"channel":"wechat"}'
# 2) 扫码支付后，微信回调 /api/v1/pay/notify/wechat：
#    验签失败返回 401；成功返回 {"code":"SUCCESS","message":"成功"}，订单 status 0→1
# 3) 退款：POST /api/v1/pay/refund {"transaction_id":N}
```

### 说明
- 代码：`internal/channel/payment.go`（RSA-SHA256 签名、AES-256-GCM 回调解密均有黄金测试）；
- **必须使用真实商户号联调**：签名/解密算法已用独立向量验证，但协议交互（下单/回调/退款）未与微信服务器实测；
- 未配置或 `PAYMENT_PROVIDER=mock` 时保持原模拟通道（`/pay/create` 直接返回 paid=true）。

---

## 4. App 推送（uni-push）

### 步骤
1. DCloud 开发者中心创建应用，开通 uni-push 2.0（个推）；
2. 获取 `PUSH_APP_ID`（个推 appId）、`PUSH_APP_KEY`（appkey）、`PUSH_MASTER_SECRET`；
3. 客户端 `manifest.json` 配置 uni-push 模块与厂商推送（可选）；
4. 填入 `deploy/.env` 三项。

### 验证
客户端登录并触发任意业务通知（如报价中选），App 应收到离线推送。

### 说明
- 代码：`internal/channel/push.go`（个推 v2 REST API：鉴权 + 按别名单推，别名=用户手机号）；
- **端到端生效还需客户端完成 uni-push 初始化并上报 clientid**（客户端工程已基于 uni-app，按官方文档配置即可）；
- 未配置时通知仍通过站内消息/SSE/轮询送达。

---

## 5. 电子签（网关抽象）

### 当前状态
- `internal/channel/esign.go` 提供 `EsignGateway` 接口（创建签署/回调处理），现为 Mock 实现（返回演示链接）。
- 未选择具体服务商：法大大/e签宝/上上签 API 与签名算法差异大，写死一家等于赌服务商。

### 后续步骤（服务商选定后，约 1-2 天）
1. 按 `EsignGateway` 接口实现新适配器（如 `FadadaEsignGateway`）；
2. `NewEsignGateway(provider)` 按 `ESIGN_PROVIDER` 分发；
3. `GenerateContract`/`SignContract` handler 增加真实签署流程；`/sign/notify` 接入服务商回调验签。

---

## 6. CAD 引擎（DWG 在线预览）

### 当前状态
- `internal/channel` 之外，`file.go` 的 `/file/:id/preview` 已支持 `CAD_CONVERT_API` 转发（multipart file + format=svg）。

### 后续步骤
1. 采购/部署商用引擎（Aspose.CAD 自建服务 / 商业渲染网关），提供 HTTP 接口：
   `POST {CAD_CONVERT_API}`，接收 `file` 字段，返回 SVG；
2. 服务器 `.env` 填 `CAD_CONVERT_API=http://127.0.0.1:8899/convert`；
3. DWG 预览即生效（DXF 仍走自研渲染器，不依赖引擎）。

---

## 7. 会员体系（已自研完成）

- 模型：`User.member_level/member_expire_at` + `membership_orders`；
- API：`GET /member/levels`、`GET /member/info`、`POST /member/upgrade`、`GET /admin/members`；
- 权益联动：佣金折扣（银 9.5 折/金 9 折，`effectiveCommissionRate`）、派单推荐加权（银 +5/金 +10）；
- 前端：`pages/member/index.vue`（我的会员页）。
- **当前为模拟支付**：开通即生效并记录订单；真实支付接入后改为 `pending` + 回调生效。

---

## 8. 结项清单（验收时勾选）

| # | 项 | 状态 |
|---|----|------|
| 1 | `go test ./...` 全绿（含 channel 黄金测试） | ✅ 代码侧已通过 |
| 2 | 前端 vue-tsc / eslint 0 error 0 warning | ✅ 代码侧已通过 |
| 3 | 短信：真实手机号收到验证码 | ☐ 填凭据后联调 |
| 4 | OCR：审核备注出现识别文本 | ☐ 填凭据后联调 |
| 5 | 微信支付：下单→扫码→回调→订单已支付→退款 | ☐ 填凭据后联调（唯一未联调通道） |
| 5a | 小程序 JSAPI：登录拿真实 openid → 下单 → 拉起支付 → 回调 | ☐ 填凭据 + A1 登录后联调 |
| 6 | App 推送：业务通知离线送达 | ☐ 填凭据 + 客户端上报 clientid 后联调 |
| 7 | 电子签：选定服务商后开发适配器 | ☐ 商务决策 |
| 8 | DWG 预览：引擎部署后填 `CAD_CONVERT_API` | ☐ 采购/部署 |
| 9 | 保险/造价指数：商业立项 | ☐ 商务决策 |

---

**文档结束**
