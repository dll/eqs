# EQS 非生产运维准备 —— 隔离启动/停止/日志/备份/回滚演练报告

> **任务**：EQS-OPS-DRYRUN-20260908（ccit-ceo → ccit-ops）
> **日期**：2026-09-08
> **执行**：ccit-ops（公司级运维）
> **性质**：非生产、隔离范围演练；**零生产写、零发布、零真实凭据/客户数据/DNS/外呼/短信/支付/签约，零系统服务或 nginx 改动**。
> **CTO 状态**：**Conditional**（非 Approved）；EQS 尚未形成合格单一交付 commit。

---

## 0. 结论速览

| 项 | 结果 |
|----|------|
| 隔离启动/探活 | ✅ 成功（SQLite 独立文件 + 独立端口 18081） |
| 探活接口 | ✅ `GET /api/v1/config/public` → HTTP 200 `{"configs":{}}` |
| 鉴权边界 | ✅ `/member/levels` → 401（未登录正确拦截） |
| 异常停止 | ✅ kill 后监听释放（进程退出，仅残留 TIME_WAIT） |
| 日志定位 | ✅ 启动日志明确输出「SQLite 模式 + 文件路径 + 端口」 |
| 数据清理 | ✅ 演练产物全部删除，既有 `eqs.db` 未动（299008 字节恒定） |
| 仓库污染 | ✅ 无新增/无改动源码（7 个 `.go` 变更为 feature 分支既有工作区改动，非本演练产生） |

---

## 1. 实际收到/使用的环境基线

读取自 ccit-env 已回报 + 本机实测，口径一致：

| 项 | 实测值 | 来源 |
|----|--------|------|
| 仓库 HEAD | `ce4936a`（`feature/archived-project-inbound`） | `git rev-parse HEAD` |
| Go | `go1.26.0 windows/amd64` | `go version` |
| Node / pnpm | `v24.19.0` / `10.34.5` | `node -v` / `pnpm -v` |
| 服务默认端口 | `8080`（`config.go` `getEnv("SERVER_PORT","8080")`） | `internal/config/config.go` |
| 8080 占用 | **Tomcat11.exe pid 6432**（GitAIOps），`0.0.0.0:8080` LISTENING | `netstat -ano` / `tasklist` |
| 数据库驱动默认 | `mysql`（非 sqlite）；`DB_DRIVER=sqlite` 时走本地文件 `eqs.db` | `internal/config/config.go` + `internal/model/db.go` |
| 权威占位模板 | `deploy/.env.example`（V3.0）；根目录 `.env` 不得误用 | 交接清单 §3.2 + 实测 |

### 关键代码事实（支撑演练设计）

1. **SQLite 模式完全支持**：`model.InitDB` 中 `if cfg.DBDriver == "sqlite"` → `gorm.Open(sqlite.Open(name))`；`DB_NAME` 为空/`eqs`/`eqs.db` 时回落到 `eqs.db`，否则用自定义文件名。
2. **SQLite 模式下 Redis 降级**：`main.go` 中 `if cfg.DBDriver == "sqlite"` → 打印「Redis 校验降级为内置模拟」，**不连真实 Redis**（`InitRedis` 仅在非 sqlite 分支调用）。
3. **`isProduction` 门禁**：仅 `APP_ENV=production` 时触发强校验（拒绝弱 `JWT_SECRET`/`DB_PASSWORD`/必填 `DATA_ENCRYPTION_KEY`）。演练设 `APP_ENV=test`，天然避开生产强校验，不落真实密钥。
4. **模型已扩容**：`AutoMigrate` 现注册 **37 个模型**（L0 基线「28 模型」口径 + 本分支新增 `ArchivedProject/Milestone/Risk/Todo`、`Organization/OrgMember/AgentPrincipal`、`PMAOpportunity/PMAProposal`）。⚠️ 提示：L0 报告（20260907）的「28 模型」锚点在本 feature 分支已漂移，runbook 需以当前 `model/db.go` 实际清单为准。

---

## 2. 隔离演练方案与实测证据

### 2.1 选定参数（不冲突 + 隔离）

| 参数 | 值 | 理由 |
|------|----|------|
| 工作目录 | `packages/server/`（仓库内，只读不改源码） | 与 `cmd/server` 相对路径一致 |
| 隔离目录 | `packages/server/.dryrun-ops-20260908/` | 前缀 `.` 不入 git；独立于任何既有产物 |
| 独立 SQLite | `.dryrun-ops-20260908/eqs-dryrun.db` | 与既有 `eqs.db`（299008 字节）完全隔离 |
| 端口 | `18081` | 复核空闲；避开 8080(Tomcat)、8090/8091 |
| 环境变量 | `APP_ENV=test`、`DB_DRIVER=sqlite`、`DB_NAME=…/eqs-dryrun.db`、`SERVER_PORT=18081`、`JWT_SECRET=dryrun-local-secret-only`、`WX_MINI_MOCK=1`、`PAYMENT_PROVIDER=mock` | 全 mock/本地，不碰真实凭据 |

### 2.2 启动命令与输出摘要

```powershell
cd packages/server
$env:APP_ENV="test"; $env:SERVER_PORT="18081"; $env:DB_DRIVER="sqlite";
$env:DB_NAME=".dryrun-ops-20260908/eqs-dryrun.db";
$env:JWT_SECRET="dryrun-local-secret-only"; $env:WX_MINI_MOCK="1";
$env:PAYMENT_PROVIDER="mock"; $env:LOG_LEVEL="info";
# 净化污染变量（wxx/蔚小芯 注入的真实 JWT/WX/ZHIPU 等，避免假失败）
Remove-Item env:REDIS_ADDR,env:REDIS_PASS,env:DB_HOST,env:DB_PORT,env:DB_USER,env:DB_PASSWORD,env:WX_MINI_APPID,env:WX_MINI_SECRET,env:ZHIPU_API_KEY,env:DEEPSEEK_API_KEY -ErrorAction SilentlyContinue
go run ./cmd/server
```

**启动日志关键行**（证据）：
```
2026/09/08 07:06:21 SQLite 模式：使用本地文件库 .dryrun-ops-20260908/eqs-dryrun.db，Redis 校验降级为内置模拟
2026/09/08 07:06:21 Server starting on port 18081
```
路由注册含本分支新增端点（`/admin/archive/projects*`、`/org/*`、`/admin/agent-principals`、`/admin/pma/*`），与本 feature 分支实际代码一致。

### 2.3 探活（启动验证）

```powershell
curl -sS -o - -w "\nHTTP:%{http_code}\n" http://127.0.0.1:18081/api/v1/config/public
# → {"configs":{}}  HTTP:200   ✅
```

```powershell
curl -sS -o - -w "\nHTTP:%{http_code}\n" http://127.0.0.1:18081/api/v1/member/levels
# → {"error":"unauthorized"}  HTTP:401   ✅（未登录正确拦截）
```

**隔离证据**：
```powershell
Get-Item packages/server/eqs.db            # 既有库 → 299008 字节（未动）
Get-Item packages/server/.dryrun-ops-20260908/eqs-dryrun.db  # 演练库 → 385024 字节（独立）
netstat -ano | Select-String ":18081 "     # LISTENING (pid 11476) → 独占新端口
```

### 2.4 异常停止（kill）验证

`process kill` 终止 `go run` 会话 → 进程退出，监听释放：
```powershell
netstat -ano | Select-String ":18081 "
# → 仅剩 127.0.0.1 的 TIME_WAIT（socket 关闭残留，秒级自清），无 LISTENING
Get-Process server,server-dryrun           # → 无残留进程
```
> 说明：Windows 下 `go run` 封装进程被 `kill` 属强杀（非 SIGTERM 优雅停机），未打印 "Shutting down gracefully" / "Server exited"。生产 systemd 下 SIGTERM 会走 `srv.Shutdown(ctx)` 优雅停机（代码已实现，10s 宽限），路径一致但不在此非生产演练强杀路径中复现。

### 2.5 数据清理

```powershell
Remove-Item -Recurse -Force ".dryrun-ops-20260908"
Remove-Item -Force "bin/server-dryrun.exe"
# 复验：Test-Path → False / False；eqs.db 仍 299008
```

---

## 3. 未来测试/生产部署 runbook（建议稿，待 CTO Final Approved 后启用）

### 3.1 测试环境（非生产）启动 runbook

```powershell
# 1) 端口冲突预检（三步）
netstat -ano | findstr ":8080"     # EQS 默认端口
netstat -ano | findstr ":18081"    # 拟用端口（若占用则换，如 18082/18083）
tasklist | findstr "Tomcat"        # 确认 8080 为 GitAIOps Tomcat，勿抢占

# 2) 隔离启动（SQLite + 独立端口 + 净化变量）
cd packages/server
$env:APP_ENV="test"; $env:SERVER_PORT="18081"; $env:DB_DRIVER="sqlite"; $env:DB_NAME=".dryrun/eqs.db"; $env:WX_MINI_MOCK="1"; $env:PAYMENT_PROVIDER="mock"
Remove-Item env:JWT_SECRET,env:DB_HOST,env:DB_USER,env:DB_PASSWORD,env:WX_MINI_APPID,env:WX_MINI_SECRET -ErrorAction SilentlyContinue
go run ./cmd/server

# 3) 探活
curl -sS -w "%{http_code}\n" http://127.0.0.1:18081/api/v1/config/public   # 期望 200

# 4) 停止
# Ctrl+C / （后台用 taskkill 该 go 子进程）

# 5) 清理（演练后）
Remove-Item -Recurse -Force ".dryrun"
```

### 3.2 端口冲突检查清单

| 端口 | 占用者 | 用途 | EQS 关系 |
|------|--------|------|----------|
| 80/443 | Caddy | TLS 反代 | 生产入口 |
| 8080 | Tomcat11（GitAIOps，pid 6432） | 第三方 | **默认端口冲突，EQS 本机测试必须改 `SERVER_PORT`** |
| 8090 | （生产）eqs-server | EQS 后端 | 生产专用，测试勿抢 |
| 8091 | （生产）nginx | 路径分发 | 生产专用 |
| 18081 等 | 空闲 | 测试自留 | 推荐测试端口段 |

> 规则：本机任何 EQS 测试启动前，先 `netstat -ano | findstr ":<port>"` 复核；不要用 8080（被 Tomcat 占用）。生产 8090/8091 严禁测试环境触碰。

### 3.3 数据库备份 / 恢复

- **测试（SQLite）**：直接 `Copy-Item eqs.db eqs.db.bak-<ts>` 即可，恢复即覆盖回。SQLite 单文件，备份=文件拷贝，无锁表风险（WAL 下建议先优雅停机或使用 sqlite3 `.backup`）。
- **生产（MySQL）**：走 `deploy/scripts/backup.sh`（仓库内已提供），需 `mysqldump --single-transaction --routines --triggers eqs > /backup/eqs-<ts>.sql`；恢复用 `mysql eqs < eqs-<ts>.sql`。**备份/恢复均属生产写，需 L2 审批**。
- **Redis**：生产 `BGSAVE` 到 `.rdb`，与 DB 备份同步时间窗快照；测试 SQLite 模式不依赖 Redis，无此步骤。

### 3.4 回滚门槛

- **测试**：SQLite 单文件回滚门槛 = 保留 `.db.bak` 文件即可，秒级。
- **生产**：回滚门槛 = ① 保留当次发布前 server 二进制备份 + ② DB 备份点 + ③ 前端 dist 备份。触发回滚条件：探活失败 / 关键接口 5xx / 数据损坏 / 关键业务回归。**回滚属 L2 写操作，必须先报 ccit-ceo 审批**。
- **版本锚**：当前仓库「合格交付 commit 未形成」，回滚目标 commit 需由 CTO Final Approved 指定（当前 Conditional 不构成可回滚基线）。

---

## 4. 必须等 CTO Final Approved 才能执行的动作（硬门槛）

> 当前 CTO 结论为 **Conditional**，以下动作**一律不得执行**，需 CTO 对精确功能 commit 给出 **Final Approved** 后逐项放行：

| 动作 | 为何必须等 |
|------|-----------|
| 任何生产写（`systemctl restart eqs-server`、`nginx -s reload`、scp、改 `.env`/nginx/caddy/systemd 配置） | 生产写默认关闭，L2 审批前提是「合格交付 commit + Final Approved」 |
| 生产部署 / 发布（`cd.yml` push 触发 or 手动兜底） | 无合格单一交付 commit，当前 feature 分支不可作为发布基线 |
| 数据库迁移执行（MySQL `ApplyMigrations`） | 本分支新增 10 个模型/表，迁移对生产 schema 的落库需 Final Approved |
| 填真实凭据（短信/OCR/微信支付/小程序/Android 签名/iOS） | 凭据涉及真实资金/外呼/签约通道，需 CTO 最终放行后由 owner 填入 Secrets |
| 修改系统服务 / nginx / Caddy | 属系统级写，且 B4 历史残留（`eqs.bak.20260811205051` conflicting 8091）清理也需审批 |
| 绑定 DNS / 外呼 / 支付回调 / 短信真实发送 | 触发真实外部副作用，必须 Final Approved |
| 回滚操作 | 回滚本身属 L2，且当前无可回滚的「合格基线」 |

---

## 5. 边界遵守声明

- ✅ 未执行任何生产写 / 发布 / 系统服务或 nginx 改动。
- ✅ 未接入真实凭据、客户数据、DNS、外呼/支付/签约/短信。
- ✅ 全程 `git` 只读 + `go build`/`go run`（本地编译/运行，不改源码）+ `curl GET` + `netstat`/`tasklist` 只读。
- ✅ 演练产物（`.dryrun-ops-20260908/`、`bin/server-dryrun.exe`）已全部清理，既有 `eqs.db` 未动，源码零改动（本仓库 7 个 `.go` 变更为 feature 分支既有工作区状态，非本演练引入）。
- ⚠️ **需 owner 执行的系统级操作**（若有）：本次演练全程免 owner 操作；如后续需「安装 sqlite3 CLI」「开放受限端口」等，将另行提交审批请求，不自行执行。

---

*—— ccit-ops 执行，2026-09-08。非生产隔离演练，零生产副作用。*
