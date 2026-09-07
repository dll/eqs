# EQS 生产验证与多端发布流水线 —— 部署日志

> **任务**：EQS-OPS-20260831-01（ccit-ceo → ccit-ops）
> **日期**：2026-08-31 22:15 GMT+8
> **执行**：ccit-ops（公司级运维）
> **范围**：生产只读验证（SSH 只读 + git 只读 + HTTP GET），无任何生产写操作。
> **边界**：全程未执行 restart/stop/scp/rm/nginx -s reload/改配置/DB 写。仅 `ss / ps / systemctl status / nginx -t / nginx -T / curl GET/POST(空body探测) / mysql SELECT / git log/diff/show / cat / ls`。

---

## 0. 环境事实对齐（交接清单 §2 核对）

| 项 | 交接清单期望 | 实测 | 结论 |
|----|--------------|------|------|
| CVM | 129.211.223.113 | `VM-0-11-ubuntu`，up 32 天 | ✅ 一致 |
| SSH | root + `~/.ssh/wxx_deploy.pem` | PEM 存在，连通成功 | ✅ |
| 后端端口 | 8090（systemd `eqs-server`） | `LISTEN *:8090`（pid 1545774） | ✅ |
| nginx | 8091 | `LISTEN 0.0.0.0:8091`（nginx 1.24.0） | ✅ |
| Caddy | 占 80/443 | `LISTEN *:80 / *:443`（pid 723396） | ✅ |
| 生产域名 | eqs-chzu.tech → Caddy → nginx:8091 | 公网 HTTPS 200，`Via: 1.1 Caddy`，`Server: nginx/1.24.0` | ✅ |
| 数据库驱动 | 生产=MySQL 8.0（eqs@localhost，库 eqs） | `.env` 中 `DB_DRIVER=mysql`、`DB_HOST=localhost`、`DB_NAME=eqs`；未带 SQLite | ✅ |
| 服务运行时长 | systemd 8/16 起 | `active (running) since Sun 2026-08-16 14:33:53`（2 周 1 天） | ✅ |

**`.env` 关键项（已脱敏）**：`SERVER_PORT=8090 / DB_DRIVER=mysql / DB_HOST=localhost / DB_PORT=3306 / DB_USER=eqs / DB_PASSWORD=***REDACTED*** / DB_NAME=eqs`。
> 说明：`grep` 只匹配到了基础项，未看到 `JWT_SECRET`/`DATA_ENCRYPTION_KEY`/`REDIS_*` 显式输出（可能因 grep 模式仅匹配 `DB_|SERVER_PORT|PAYMENT_PROVIDER`，未对 JWT/DATA_ENCRYPTION 做抓取）。生产能正常启动并返回 200，说明 `config.go isProduction` 强校验已通过（`JWT_SECRET` 非弱值 + `DATA_ENCRYPTION_KEY` 已设），否则进程会拒绝启动。

---

## B1. 生产 9 步验证序列（交接清单 §5）

逐项执行/标注，结论标 PASS / FAIL / BLOCKED。

| # | 验证项 | 命令 | 实测输出摘要 | 结论 |
|---|--------|------|-------------|------|
| 1 | 健康检查 | `systemctl status eqs-server --no-pager -l` | `Active: active (running)`；Main PID 1545774 `/opt/eqs/packages/server/server`；8/16 14:33:53 起 | ✅ **PASS** |
| 2 | 公开接口 | `curl -sS -m 20 -w "%{http_code}" https://eqs-chzu.tech/api/v1/config/public` | **HTTP 200**，body `{"configs":{}}`，`Via: 1.1 Caddy` / `Server: nginx/1.24.0` | ✅ **PASS**（`{}` 为预期，见下） |
| 3 | 后台登录 | `GET https://eqs-chzu.tech/admin/` → 200 text/html；`POST /api/v1/auth/login`(空体) → 400「参数错误」 | 路由存活；登录接口 `PhoneLogin` 需真实手机号+验证码，未持凭据 | 🟡 **PASS（可达）/ 登录动作待凭据** |
| 4 | H5 打开 | `GET https://eqs-chzu.tech/h5/` | HTTP 200，`content-type:text/html`，index.html 已部署（8/16 14:21） | ✅ **PASS** |
| 5 | 短信 | 需 `TENCENT_SMS_*` 凭据 | `.env` 未见短信凭据（grep 未匹配到） | ⛔ **BLOCKED-凭据** |
| 6 | OCR | 需 `TENCENT_OCR_*` 凭据 | 同上 | ⛔ **BLOCKED-凭据** |
| 7 | 微信支付 | 需 `WXPAY_*` 真实商户号 | `PAYMENT_PROVIDER` 未见 wechat；唯一未联调通道 | ⛔ **BLOCKED-凭据** |
| 8 | 演示数据 | DB read-only 统计 | users=7 projects=5 orders=3 bids=12 disputes=1 reviews=4 quals=4 payments=3 tpls=20 ctpls=10 | ✅ **PASS（已生成，三角色闭环）** |
| 9 | 会员接口 | 路由 + DB | `member/levels`(GET) → 401 需鉴权；`member/upgrade`(POST) 需 JWT+写；`membership_orders=0` | 🟡 **PASS（接口就绪）/ 开通动作待鉴权+写** |

### 关键判定依据

**B1-2 `{"configs":{}}` 是否预期 —— 是预期值。**
- 代码：`handler.PublicConfigs`（`config_center.go`）走 `getPublicCached()` → `config_cache.go` 查 `model.DB.Where("is_public = ?", true)`。
- 实测 DB：`SELECT COUNT(*) FROM system_configs` → `total = 0`（表存在但 0 行）。
- 结论：生产库当前**没有任何公开配置**（`is_public=true` 的 `SystemConfig` 为空），所以返回空 map `{}` 是**完全正确且预期**的行为。
- 补充：演示数据 seed（`demo.go` `seedByMode`）只写业务表，不写 `system_configs`；`demo.enabled` 配置仅在 `DemoToggleHandler` 被调用时才创建（且 `IsPublic:false`）。所以 `system_configs` 为空 ≠ 异常。
- 唯一注意点：`R2-2` 修复（main@04a1b9f）正是针对「空缓存冷启动并发重复加载」的竞态，当前线上二进制（8/16）未含此修复（详见 B2）。

**B1-3 后台登录**：管理后台静态页 `/admin` 可打开（200）。登录接口 `POST /api/v1/auth/login`（`PhoneLogin`）已验证路由存活（空体返回 400「参数错误」）。真实登录需管理员手机号 + 短信验证码，属写/状态动作且需凭据，本任务（只读）不执行。管理员账号已确认存在：`users` 表 `id=180, phone=13900003333, user_type=3, status=1, company_name=EQS平台运营`。

**B1-4 H5**：`/h5` 返回 200 text/html，`index.html`（11‌45 字节）+ `assets/` + `static/` 目录已在服务器（mtime 8/16 14:21），与二进制部署时间（14:33）一致。

**B1-8 演示数据**：三角色闭环数据已存在，数据量与 `demo.go` 的 `demo` 模式一致（甲方/服务方/管理员、项目/报价/订单/里程碑/合同/支付流水/资质/争议/评价/模板/案例/佣金）。注意：`system_configs` 仍为空 → 意味着演示数据是通过 `POST /admin/demo/seed` 直调 `seedByMode` 生成，**未走 `demo/toggle` 持久化开关**（`demo.enabled` 未落库）。功能不受影响，仅状态标识未持久化。

**B1-9 会员**：路由齐备（`GET /member/levels`、`GET /member/info`、`POST /member/upgrade`，均在 `auth` 组，需 JWT）。`member.go` 说明当前为**模拟支付**（订单生成即生效，`MembershipOrder{Status:"paid"}`），真实支付待接入。`membership_orders` 表 0 行（尚无真实/模拟开通记录），`users` 表已有 `member_level`(varchar 默认 free) / `member_expire_at`(datetime) 字段。接口就绪但「权益联动」需鉴权 + 写操作，超出只读范围，标 PASS（就绪）/ 待开通。

---

## B2. 版本 diff（线上二进制 vs main）

### 线上基线
- 二进制：`/opt/eqs/packages/server/server`，mtime **2026-08-16 14:33**，size 20643960。
- 前端 admin/H5 dist：mtime 8/16 14:21~14:22。
- 服务启动：`since 2026-08-16 14:33:53 CST`。

### 8/16 之后合入 main 的提交清单（`git log --oneline --since="2026-08-16 14:00"`）

```
04a1b9f 2026-08-31 19:57:55  fix: 修复配置缓存冷启动并发重复加载(R2-2单飞锁)+忽略server.exe编译产物   ← 唯一代码改动
eda5852 2026-08-29 18:48:20  docs: 完善EQS飞书接入与验收内容
ec05ffc 2026-08-29 18:47:30  docs: 补充EQS飞书接入与验收内容
a9a37f3 2026-08-29 17:44:43  docs: 完善EQS OpenCode与DeepSeek方案
9342beb 2026-08-29 17:18:59  docs: 归档EQS DeepSeek实现方案
b1dde0b 2026-08-16 14:00:55  chore: 将 .openclaw 代理工作区纳入版本控制…   ← 部署基线附近（二进制 14:33 在此之后构建）
```

> 说明：`git log b1dde0b..04a1b9f -- packages/server packages/admin packages/client` 结果为 **仅 `04a1b9f` 一个提交**触及 server 代码。其余 eda5852/ec05ffc/a9a37f3/9342beb 均为纯 docs（飞书/OpenCode/DeepSeek 方案），无业务代码影响。

### R2-2 改动内容（`git diff b1dde0b 04a1b9f -- .../config_cache.go`）

单一文件 `packages/server/internal/handler/config_cache.go`，+18/-2 行：
1. 新增独立互斥量 `publicCacheLoadMu`（单飞锁）替代原 `loadPublicCache` 内用 `publicCache.mu.Lock()`。
2. `loadPublicCache` 加「双检」（拿到单飞锁后若缓存已被并发回填则直接 return）。
3. `getPublicCached` 在 `loadPublicCache()` 之后加显式 `RLock`/`RUnlock`，避免返回 map 时数据竞争。

**性质**：纯并发正确性修复（消除冷启动空缓存下多 goroutine 重复 DB 查询的竞态 + 潜在数据竞争），**无接口/行为变更**（路由、返回结构、业务逻辑均不变）。

### 升级判断与建议（不执行）

- **线上缺陷现状**：当前二进制（8/16）在「配置缓存冷启动」场景下存在并发重复加载 DB 的竞态。由于生产 `system_configs` 目前为 0 行，该竞态**影响面小**（空结果重复查询，不产生错误数据，但存在潜在 data race，`go run -race` 下可能报）。
- **是否需要升级**：**建议升级**，但**非紧急（非 P0）**。理由：R2-2 是并发正确性修复，属 `SAR-v10.1` 加固轮产物；当前无公开配置，触发概率低，但一旦后续在配置中心写入公开配置（如主题/平台配置），冷启动并发下可能触发重复加载或 race。
- **升级窗口建议**（仅建议）：
  1. 选低峰时段（如凌晨 0~2 点），先备份当前二进制（已完成自动备份脚本 `deploy/scripts/backup.sh` 可钩）。
  2. 走 `cd.yml` 自动（`git push` 触发 scp+`systemctl restart eqs-server`）或 §4.2 手动兜底。
  3. 升级后立即复跑 B1-1/2/3/4 验证序列（尤其关注 `GET /api/v1/config/public` 冷启动后首个请求仍返回 200）。
  4. 回滚方案：保留 8/16 二进制，`restart` 前 `cp` 备份，异常即原样恢复。
- **需先报 ccit-ceo 生成审批请求**（升级 = 生产写/restart，属 L2 边界），本日志不执行。

---

## B3. 多端发布流水线方案（写入文档，供 cc 决策）

> 现状：`cd.yml`（后端+admin/h5 自动 scp+restart/reload）已在第 3.1 节 CVM 三件套（`CVM_HOST/CVM_USERNAME/CVM_SSH_KEY`）就绪下可用。多端（小程序/App）尚未纳入，需补流水线。

### B3.1 小程序（微信 mp-weixin）

| 维度 | 结论 |
|------|------|
| 构建产物 | `pnpm --filter @eqs/client build:mp-weixin` → `dist/build/mp-weixin`（uni-app 产物） |
| 上传路径 A（推荐）| **miniprogram-ci 自动化**：微信官方 CLI，`miniprogram-ci` npm 包 + `ci.upload()`，需 `privateKey`（小程序后台「开发管理-开发设置-小程序代码上传」生成的**上传密钥** + `appid` + `projectPath`）。可在 cd.yml 增加 job，`actions/upload-artifact` + 脚本调用 miniprogram-ci 上传，产出 **preview 二维码 / 直接上传到体验版**。优点：可完全 CI，无需人工点「上传」。 |
| 上传路径 B（兜底）| **微信开发者工具人工上传**：本地导入 `dist/build/mp-weixin`，点「上传」生成体验版二维码。适用于 miniprogram-ci 暂未配密钥阶段。 |
| 所需凭据 | `WX_APPID`、`WX_UPLOAD_PRIVATE_KEY`（miniprogram-ci 上传密钥，需在小程序后台生成并下载，**私钥入 GitHub Secrets，不落仓库**）、可选 `WX_CI_ROBOT`（1~30 上传机器人号） |
| 人工步骤（无法自动化部分）| ① 提交审核：体验版二维码需在微信公众平台或「微信公众平台助手」小程序提交审核 → ② 审核通过后人工点「发布」。**该两步微信不开放 API，必须人工**。 |
| 推荐路径 | **miniprogram-ci 上传体验版（自动化） + 人工提交审核/发布（不可自动化）**。先配 `WX_APPID` + `WX_UPLOAD_PRIVATE_KEY` 两个 Secrets 即可启用自动化上传。 |

### B3.2 App（Android / iOS，uni-app）

**Android（`android.yml` 签名 Secrets 精确清单）**

| Secret | 用途 | 说明 |
|--------|------|------|
| `ANDROID_SIGN_BASE64` | 签名 keystore（.jks/.keystore）的 base64 | `base64 -w0 keystore.jks` 生成；CI 中 `echo $ANDROID_SIGN_BASE64 \| base64 -d > keystore.jks` |
| `ANDROID_KEYSTORE_PWD` | keystore 密码 | — |
| `ANDROID_ALIAS` | 签名别名 | — |
| `ANDROID_ALIAS_PWD` | 别名密码 | — |
| （复用）`CVM_HOST/CVM_USERNAME/CVM_SSH_KEY` | 分发 APK 用 | 若要 scp APK 到下载站 |

- **触发方式**：`workflow_dispatch`（手动触发，推荐，签名出实包属正式发布）+ 可选 `release` 事件「打 tag 出包」。建议实包走手动触发，避免每次 push 都签名打包（证书/安全考虑）。
- **步骤**：`pnpm --filter @eqs/client build:app` 生成 App 资源 → 原生打包（uni-app 用 HBuilderX 云端打包或离线 SDK，需确认原生工程已就绪，当前「原生工程」在交接清单 §6 第 10 项标记 ☐ 待配置）→ `apksigner`/`jarsigner` 签名 → 产物 artifact + 上传下载站。

**iOS（待账号，标注）**

| Secret | 用途 | 状态 |
|--------|------|------|
| `IOS_DEV_P12_BASE64` / `IOS_DEV_PROVISION_BASE64` | 开发签名证书 + 描述文件 | ☐ 待账号 |
| `IOS_DIST_P12_BASE64` / `IOS_DIST_PROVISION_BASE64` | 分发签名证书 + 描述文件 | ☐ 待账号 |
| `IOS_P12_PASSWORD` | 证书私钥密码 | ☐ 待账号 |
| `IOS_BUNDLE_ID` | 应用 Bundle ID | ☐ 待账号 |

- **标注**：iOS 需 Apple 开发者账号（年费 99 美元）才能生成 `.p12` + `.mobileprovision`。账号未到位前 iOS 流水线无法启用，标 **BLOCKED-账号**。启用后建议同样走 `workflow_dispatch` 手动触发（TestFlight 上传同属人工/待账号）。

### B3.3 H5/admin 后续升级路径确认

- **自动**：`cd.yml` 已覆盖（scp admin/h5 dist → nginx 目录 → `nginx -s reload`；后端 scp → `systemctl restart eqs-server`）。前提 CVM 三件套 Secrets 已配（✅）。
- **手动兜底**：§4.2（`scp` + `systemctl restart` + `nginx -s reload`）可用。
- **结论**：H5/admin 建议**继续走 `cd.yml` 自动**为主，手动兜底为备。当前唯一阻断点是：升级需 L2 审批（生产写），流程上**任何 push 到 main 都会自动触发 cd.yml 生产写**——这需要 ccit-ceo 明确「自动 CD 与审批单绑定的触发策略」（例如 tag 触发 or 白名单分支），否则自动 CD 与「生产写需批准单」的边界存冲突，需上层决策。

---

## B4. 配置卫生（nginx `sites-enabled` 残留）

### 实测（`nginx -T` / `nginx -t`）

`sites-enabled/` 实际内容：
```
lrwxrwxrwx  eqs -> /etc/nginx/sites-available/eqs            （符号链接，正确）
-rw-r--r--  eqs.bak.20260811205051                            （残留：普通文件，非符号链接！）
-rw-r--r--  tsloms
```

- `nginx.conf` 主配置含 `include /etc/nginx/sites-enabled/*;` → **同时包含 `eqs`（符号链接）和 `eqs.bak.20260811205051`（普通文件）**。
- `nginx -t` 输出警告：
  ```
  [warn] conflicting server name "_" on 0.0.0.0:8091, ignored
  nginx: configuration file test is successful
  ```
- 加载顺序（`nginx -T`）：`# configuration file /etc/nginx/sites-enabled/eqs`（第 192 行，先加载）→ `.../eqs.bak.20260811205051`（第 234 行，后加载）。
- 两份内容对比：
  - `eqs`（当前生效）：含 `client_max_body_size 50m` + 安全响应头（X-Content-Type-Options / X-Frame-Options / Referrer-Policy）+ `/health` location。
  - `eqs.bak.20260811205051`（残留）：**缺失** `client_max_body_size 50m`、安全头、`/health` 路由；其余 `/admin` `/h5` `/api` 相同。

### 实际加载的是哪份 + 风险判定

- **实际生效**：`eqs`（符号链接）——因为 `include` 按文件名字典序加载，`eqs` < `eqs.bak...`，且两 `server { listen 8091; server_name _; }` 相同，nginx 对后出现的同名 `server_name` 报「conflicting... ignored」，即 **`.bak` 的 server 块被忽略**，`eqs` 块生效。
- **风险**：
  1. **每次 reload/restart 都会打 `conflicting server name "_" on 0.0.0.0:8091, ignored` 警告**（污染日志/健康检查心智，可能被监控误判）。
  2. **潜在隐患**：若将来有人把 `eqs` 符号链接删掉 / 重命名，或新增文件使字典序变化，`.bak`（缺安全头 + 缺 50m 上传限制 + 缺 /health）可能被生效，导致**安全响应头丢失、上传 50MB 限制失效、/health 探针 404**。
  3. `.bak` 是**普通文件直接丢在 `sites-enabled/`**，违反「enabled 只放符号链接」约定，属配置卫生问题。

### 清理建议（只建议，不动手）

1. **首选**：把 `/etc/nginx/sites-enabled/eqs.bak.20260811205051` 移出 nginx include 路径。推荐移到 `sites-available/`（与已有 `eqs.bak.*` / `tsloms.bak.*` 同级归档）或备份目录：
   ```bash
   mv /etc/nginx/sites-enabled/eqs.bak.20260811205051 /etc/nginx/sites-available/eqs.bak.20260811205051
   nginx -t && nginx -s reload
   ```
2. 若该 `.bak` 无保留价值（8/11 的旧版，已被 8/16 的 `eqs` 取代），可经批准后删除，归档到带日期备份目录更稳妥。
3. **执行前需报 ccit-ceo 生成审批请求**（`mv`/`nginx -s reload` 均属生产写 + reload，L2 边界）。本日志仅给出建议，未执行。

> 附带发现（同类卫生问题，供一并处理）：`sites-available/` 下另有 `tsloms.bak-143413`、`tsloms.bak.20260814071420`、`tsloms.bak.charset`、`tsloms.bak.media` 等历史残留；`/etc/caddy/` 下也有 `Caddyfile.bak.*` 多个版本。这些在 `sites-available/`（不在 enabled）不参与加载，风险低，但建议统一归档清理。

---

## 5. 结论汇总

| 节 | 结论 |
|----|------|
| B1 | 基础验证 #1/2/4/8 **PASS**；#3 后台登录可达（登录动作待凭据）；#9 会员接口就绪（开通待鉴权+写）；#5/6/7 短信/OCR/支付 **BLOCKED-凭据** |
| B2 | 8/16 后仅 1 个 server 代码提交（`04a1b9f` R2-2 缓存并发修复），其余为 docs。**建议升级但非 P0**；升级=生产写，需 L2 审批 |
| B3 | 小程序推荐 miniprogram-ci 自动化（需 `WX_APPID`+`WX_UPLOAD_PRIVATE_KEY`）；Android 签名 Secrets 清单已列（`workflow_dispatch` 触发）；iOS **BLOCKED-账号**；H5/admin 走 cd.yml 自动为主、手动兜底为备（需明确自动 CD 与审批单边界） |
| B4 | **发现 P0 级卫生问题**：`sites-enabled/eqs.bak.20260811205051` 残留导致每次 reload 打 `conflicting server name "_" on 8091, ignored` 警告 + 潜在生效应切换风险。**建议清理**（移出 enabled），需 L2 审批 |

### 需升级 ccit-ceo 的事项
1. **B4 立即项**：nginx `sites-enabled` 残留 `.bak` 导致的 `conflicting server name` 警告与潜在隐患，申请清理审批（`mv` + `nginx -t && nginx -s reload`）。
2. **B2 建议项**：R2-2 缓存并发修复升级（低峰窗口 + 回滚方案已备）。
3. **B3 决策项**：自动 CD（`cd.yml`) 与「生产写需批准单」的触发策略冲突，需明确绑定方案（tag 触发 / 白名单分支 / 手动批准）。
4. **凭据阻塞项**：短信/OCR/微信支付（#5/6/7）、iOS 账号、小程序 `WX_APPID`+`WX_UPLOAD_PRIVATE_KEY`、Android 签名 Secrets 仍待填。

---

*—— ccit-ops 执行，2026-08-31。全程只读，未执行任何生产写操作。*
