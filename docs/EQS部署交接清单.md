# EQS 发布部署交接清单（部署/运维智能体专用）

> **文档版本**：V1.0
> **交接日期**：2026-08-31
> **交接人**：leader-eqs（已全权接管 EQS 代码层）
> **接收方**：公司级部署/运维智能体
> **本档定位**：把分散在 `docs/EQS构建部署.md`、`docs/EQS外部通道接入指南.md`、`docs/SAR/*` 的发布信息**汇总成一份可执行清单**，并标出当前真实状态。详细背景以引用的源文档为准。

---

## 0. 现状一句话

**代码功能层已完整、测试与双端生产构建全绿、达到可发布状态；剩下的全是「填凭据 + 部署 + 联调」的非代码项，由你（部署/运维智能体）接手。** 参考审核报告 `docs/SAR/SAR-v10.1.md`。

---

## 1. 发布前置：把当前代码变成可发布产物

> 本机已全量验证通过；部署代理在干净环境照此重建即可。

| 步骤 | 命令 | 预期 |
|------|------|------|
| 后端构建 | `cd packages/server && go build -o bin/server ./cmd/server` | EXIT 0 |
| 后端测试 | `cd packages/server && go vet ./... && go test ./...` | 7 包全 ok |
| 后台构建 | `pnpm --filter @eqs/admin build` | `packages/admin/dist` |
| 用户端 H5 | `pnpm --filter @eqs/client build:h5` | `packages/client/dist/build/h5` |
| 小程序 | `pnpm --filter @eqs/client build:mp-weixin` | `dist/build/mp-weixin` |
| App 资源 | `pnpm --filter @eqs/client build:app` | `dist/build/app` |

> ⚠️ **版本号同步**：发布前把 4 处版本号一并升位（根/`client`/`admin` 的 `package.json` + `client/src/manifest.json` 的 `versionName`/`versionCode`），再 `git tag -a vX.Y.Z` 触发 `release.yml`。详见 `docs/EQS构建部署.md` 第 3、9 章。

---

## 2. 关键环境事实（务必对齐，防踩坑）

| 项 | 值 | 说明 |
|----|----|------|
| CVM | `129.211.223.113`（root，`~/.ssh/wxx_deploy.pem`） | 与 wxx 同机 |
| 生产域名 | `eqs-chzu.tech` / `www.eqs-chzu.tech` → `129.211.223.113` | DNS 须提前生效 |
| HTTPS | Caddy 自动签（wxx 占用 80/443） | EQS 由 Caddy 转 `nginx:8091` |
| nginx | 监听 **8091**（勿抢 80/443） | 按路径分发 /admin /h5 /api |
| 后端端口 | systemd `eqs-server` 监听 **8090**（wxx 占 8080，错开） | — |
| 数据库 | **生产=MySQL 8.0**（库 `eqs`，用户 `eqs@localhost`）；**开发=SQLite** | 见下节「驱动切分」|
| Redis | `localhost:6379`（验证码/频控） | — |

### 2.1 ⚠️ 数据库驱动的开发/生产差异（重点）

- **代码默认**：`DB_DRIVER=mysql`（未设则连 MySQL）。
- **本机开发运行**：我此前用 `DB_DRIVER=sqlite` + `DB_NAME=eqs.db` 起的本地库，**仅限开发**。
- **生产**：`deploy/.env.example` 走 MySQL（`DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME`），systemd 经 `EnvironmentFile` 注入 `DB_PASSWORD`。
- ➡️ 部署代理部署生产时必须用 **MySQL 连接**，不要带 SQLite 变量；否则连接 MySQL 会失败。
- 生产强校验（`config.go isProduction`）：`JWT_SECRET` 不可为弱值、**必须设 `DATA_ENCRYPTION_KEY`**（AES-256-GCM 字段加密），否则拒绝启动。

---

## 3. 凭据 / Secrets 全清单（填好即启用）

> 完整步骤见 `docs/EQS外部通道接入指南.md`。下表是**一处汇总**。

### 3.1 GitHub Actions Secrets（仓库 → Settings → Secrets）

| Secret | 用途 | 状态 |
|--------|------|------|
| `CVM_HOST` / `CVM_USERNAME` / `CVM_SSH_KEY` | CD 部署三件套 | ✅ 已配（129.211.223.113 / root）|
| `ANDROID_SIGN_BASE64` + `ANDROID_KEYSTORE_PWD/ALIAS/ALIAS_PWD` | Android APK 签名 | ☐ 出实包时 |
| `IOS_DEV_P12_BASE64` / `IOS_DEV_PROVISION_BASE64` / `IOS_DIST_*` / `IOS_P12_PASSWORD` / `IOS_BUNDLE_ID` | iOS init / release | ☐ 手动触发时 |

### 3.2 服务器 `.env`（/opt/eqs/.env，systemd EnvironmentFile）

| 分组 | 变量 | 填好即 | 状态 |
|------|------|--------|------|
| 基础 | `SERVER_PORT`、`DB_*`、`REDIS_*`、`JWT_SECRET`、`DATA_ENCRYPTION_KEY` | 正确起点 | ☐ 填真实值 |
| 短信 | `TENCENT_SMS_SECRET_ID/KEY` + `SMS_SDK_APP_ID/SIGN_NAME/TEMPLATE_ID` | 真实发验证码 | ☐ |
| OCR | `TENCENT_OCR_SECRET_ID/KEY` | 审核自动识别扫描件 | ☐ |
| 微信支付 | `WXPAY_APPID/MCHID/API_V3_KEY/MCH_SERIAL_NO/私钥/平台证书/通知URL` + `PAYMENT_PROVIDER=wechat` | Native 扫码支付（**唯一未联调通道**）| ☐ 需真实商户号 |
| App 推送 | `PUSH_APP_ID/APP_KEY/MASTER_SECRET` | 离线推送 | ☐ + 客户端上报 clientid |
| CAD | `CAD_CONVERT_API` | DWG 在线预览 | ☐ 采购/部署引擎 |
| 电子签 | — | — | ☐ 商务决策后开发适配器 |

---

## 4. 部署运行手册（push 自动 / 手动兜底）

### 4.1 自动（推荐）
`git push origin main` → `cd.yml` 自动：scp server → `/opt/eqs/packages/server/server` → `systemctl restart eqs-server`；admin/h5 → nginx 目录 → `nginx -s reload`。需先配好第 3.1 节 CVM 三件套。

### 4.2 手动兜底（部署代理在 CVM 直接执行）
```bash
# 后端（以非 root 用户，systemd 已加固）
scp packages/server/bin/server root@129.211.223.113:/opt/eqs/packages/server/server
ssh root@129.211.223.113 "systemctl restart eqs-server && systemctl status eqs-server --no-pager"

# 前端
scp -r packages/admin/dist/*  → /opt/eqs/packages/admin/dist/
scp -r packages/client/dist/build/h5/* → /opt/eqs/packages/client/dist/build/h5/
ssh root@129.211.223.113 "nginx -s reload"
```
> 首次部署前检查：`deploy/systemd/eqs-server.service`（非 root、NoNewPrivileges、ProtectSystem、EnvironmentFile=/opt/eqs/.env）、`deploy/nginx/eqs-cvm.conf`（8091）、`deploy/caddy/eqs-chzu.tech.conf`、备份脚本 `deploy/scripts/backup.sh`。

---

## 5. 上线验证序列（部署后必过）

| # | 验证 | 命令/操作 | 通过标准 |
|---|------|-----------|----------|
| 1 | 健康检查 | 服务进程 + 日志无 fatal | systemd active |
| 2 | 公开接口 | `GET https://eqs-chzu.tech/api/v1/config/public` | HTTP 200 返回 JSON |
| 3 | 后台登录 | 管理后台 `/admin` | 可登录 |
| 4 | H5 | 用户端 `/h5` | 可打开 |
| 5 | 短信（填凭据后）| `POST /api/v1/sms/send` 真实手机号 | 收到验证码 |
| 6 | OCR（填凭据后）| 审核含扫描件资质 | 备注出现 `OCR识别:` |
| 7 | 微信支付（填凭据+联调）| 下单→扫码→回调→订单已支付→退款 | 全链路 |
| 8 | 演示数据 | 管理后台「演示数据」一键生成 | 三角色闭环数据 |
| 9 | 会员 | `POST /member/upgrade` | 权益联动生效 |

> 生产环境忌带弱 `JWT_SECRET` / 缺 `DATA_ENCRYPTION_KEY`（启动即拒，见 2.1）。

---

## 6. 交接状态总表（部署代理接手时逐项勾选）

| # | 项 | 归类 | 状态 |
|---|----|------|------|
| 1 | 代码层功能完整 + 测试/构建全绿 | 代码 | ✅ 完成 |
| 2 | 发布产物可构建（server/admin/h5/mp/app） | 代码 | ✅ 完成 |
| 3 | CVM 部署三件套 Secrets | 部署 | ✅ 已配 |
| 4 | 服务器 `.env` 基础项（JWT/加密密钥/MySQL） | 部署 | ☐ 待填 |
| 5 | 短信 / OCR 凭据 | 凭据 | ☐ 待申请填入 |
| 6 | 微信支付 v3 real 联调（唯一 P1）| 凭据+联调 | ☐ 待商户号 |
| 7 | App 推送（含客户端上报）| 凭据+客户端 | ☐ 待开通 |
| 8 | 电子签服务商适配器 | 商务+开发 | ☐ 待决策 |
| 9 | DWG 引擎采购部署 | 采购 | ☐ 待采购 |
| 10 | Android/iOS 签名 Secrets + 原生工程 | 移动端 | ☐ 待配置 |

---

## 7. 建议接手顺序（10~14 天量级）

1. **立即**：第 2 节环境对齐 + 第 4 节手动部署跑通（半天）→ 让产品先上线裸跑（Mock 支付/短信演示码可先用）。
2. **第一批**：第 5 节基础验证（1/2/3/4/8/9）全过。
3. **第二批**：短信 + OCR 凭据（1 天）。
4. **第三批**：微信支付联调（2-3 天，唯一协议风险点）。
5. **并行商务线**：电子签服务商、DWG 引擎、App 推送、移动端签名。

---

## 8. 文档索引（详细背景）

- `docs/EQS构建部署.md` —— CI/CD、版本号、Android/iOS、避坑清单。
- `docs/EQS外部通道接入指南.md` —— 各通道凭据步骤 + 联调清单 + 结项勾选。
- `docs/SAR/SAR-v10.0.md` / `SAR-v10.1.md` —— 外部通道完成度审核 + 发布前加固轮（含 R2-2 缓存并发修复）。
- `deploy/` —— nginx/caddy/systemd/backup 配置与 `.env.example`。
- 代码基线：`main` 分支，最近提交 `04a1b9f`（R2-2 修复）。

---

*—— leader-eqs 交接，2026-08-31。代码层已交付；部署/运维由接收方继续。*
