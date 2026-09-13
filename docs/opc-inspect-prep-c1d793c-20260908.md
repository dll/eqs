# EQS 第一单 · OPC 非生产隔离复核 —— 准备与技术首验报告

> **任务**：EQS-FIRST-ORDER-SIM-20260907-01（OPC 非生产隔离复核）
> **精确验收 commit**：`c1d793c2b72e6339278db1429b2c3e600035eb4a`（CTO **Approved**）
> **父提交**：`ce4936ac8ef7667153812541facaf9bce6f1af68`
> **执行**：ccit-ops（就绪准备 + 技术首验 + 环境交付）
> **日期**：2026-09-08
> **性质**：非生产、隔离；零生产部署/迁移/真实凭据/systemd/nginx/Caddy/DNS/支付/短信/外呼。

---

## 0. 结论速览

| 项 | 结果 |
|----|------|
| 精确 commit 定位 | ✅ HEAD `c1d793c`，独立 clean detached worktree 已建 |
| 端口预检 | ✅ `18992` 空闲（8080 仍被 Tomcat11 占用，勿碰） |
| 构建/静态检查 | ✅ `go build` / `go vet` EXIT 0 |
| 全量回归 | ✅ `go test -count=1 -p 1 ./...` 7 包全绿 EXIT 0（与 CTO 独立回归一致） |
| 隔离启动 | ✅ SQLite 独立文件 + 端口 18992 + 全 mock，成功监听 |
| 未授权探针 | ✅ 200 /config/public、401 /pma、404 /pma/opportunities（与 CTO §四一致） |
| 环境清理 | ✅ 进程停止、端口释放、隔离 SQLite 已删、worktree 保持 clean |

---

## 1. 关键边界发现（务必知悉，防混入）

**主工作区 `E:\...\MyProjects\eqs` 是「脏工作区」**，存在未提交变更，**不在 Approved 范围**：

```
 M packages/server/internal/handler/audit.go
 M packages/server/internal/handler/audit_query_test.go
 M packages/server/internal/handler/config_center.go
 M packages/server/internal/handler/config_test.go
 M packages/server/internal/model/template.go
?? packages/server/internal/handler/config_security.go
?? packages/server/internal/middleware/request_id.go
?? packages/server/internal/middleware/request_id_test.go
```

这正对应 CTO §八「原始工作区脏变更：不在签核范围，禁止混入发布」。因此 **OPC 复核必须在独立 clean worktree 进行，不得直接在主工作区跑**。

### 已交付的 clean worktree

```
E:\2026-2027\2026-2027-1\MyProjects\eqs-opc-c1d793c
```
- detached HEAD == `c1d793c2b72e6339278db1429b2c3e600035eb4a`
- `git status --porcelain` 为空（clean）
- 已含 `packages/server/internal/handler/pma.go` / `pma_test.go`

> 另：QA 的 `C:\Users\ldl\AppData\Local\Temp\eqs-qa-c1d793c` 与 `E:\...\eqs-qa-c1d793c2` 亦为同 commit clean worktree，OPC 可任选，建议用本次新建的 `eqs-opc-c1d793c` 避免与他方串扰。

---

## 2. 技术首验实测（ccit-ops 已完成，非生产）

### 2.1 环境与参数

| 参数 | 值 |
|------|----|
| worktree | `E:\...\eqs-opc-c1d793c`（clean detached @ c1d793c） |
| 端口 | `18992`（预检空闲） |
| DB | `DB_DRIVER=sqlite`，独立文件 `packages/server/.opc-dryrun/eqs-opc.sqlite` |
| 环境 | `APP_ENV=test`、`WX_MINI_MOCK=1`、`PAYMENT_PROVIDER=mock`、不设真实凭据 |
| 变量净化 | 清除 wxx/蔚小芯 注入的 `JWT_SECRET`/`WX_*`/`ZHIPU_*`/`DEEPSEEK_*` 等，避免假失败 |

### 2.2 回归（与 CTO §三一致）

```
go build ./...              EXIT 0
go vet ./...                EXIT 0
go test -count=1 -p 1 ./... 7 包 ok，EXIT 0
```

### 2.3 启动日志关键行（证据）

```
2026/09/08 07:58:49 SQLite 模式：使用本地文件库 .opc-dryrun/eqs-opc.sqlite，Redis 校验降级为内置模拟
2026/09/08 07:58:49 Server starting on port 18992
[GIN-debug] GET /api/v1/pma/projects --> PMAListProjects (7 handlers)
[GIN-debug] GET /api/v1/pma/projects/:id/overview --> PMAProjectOverview (7 handlers)
[GIN-debug] PUT /api/v1/pma/projects/:id/todos/:tid/decision --> PMADecideTodo (7 handlers)
```

### 2.4 未授权探针（与 CTO §四逐项吻合）

```
GET  /api/v1/config/public                 → 200 {"configs":{}}
GET  /api/v1/pma/projects                  → 401 {"error":"unauthorized"}
GET  /api/v1/pma/projects/1/overview       → 401 {"error":"unauthorized"}
PUT  /api/v1/pma/projects/1/todos/1/decision → 401 {"error":"unauthorized"}
GET  /api/v1/pma/opportunities             → 404 page not found（商机入口未注册）
```

---

## 3. OPC 需执行的验证步骤（§七 · 需测试账号凭据的部分）

> 以下「登录后」步骤需**公司提供的普通 EQS 测试账号口径**（密码/令牌不入报告、不入聊天）。ccit-ops 不接触凭据，由 OPC 亲自执行。若缺测试账号，请向 CEO 索取。

| # | 步骤 | 预期 |
|---|------|------|
| 0 | 端口预检 + 隔离启动（见 §4 复用命令） | 监听 18992，日志「SQLite 模式」 |
| 1 | 未登录访问 `/pma/projects` | **401**（已首验通过） |
| 2 | 登录后查看 `/pma/projects` | 200，返回存量项目列表 |
| 3 | 打开一个真实/模拟项目 `/pma/projects/:id/overview` | 200，含工期/里程碑/进度/风险/验收/待办 |
| 4 | 普通 internal **成员**（非 owner/admin）执行 `PUT .../decision` | **403**「仅OPC或平台管理员可处理项目决策」 |
| 5 | OPC/internal **owner** 执行 `PUT .../decision` | 200，`message:决策待办已更新` |
| 6 | 查 audit | 记录 `pma.todo.decide`，含主体/动作/对象/状态可追溯 |
| 7 | 访问 `/pma/opportunities` | **404**（已首验通过） |
| 8 | 结束停止进程 + 删除隔离 SQLite | 见 §4 |

**PMA 鉴权模型（源码 `pma.go`）**，供 OPC 判断 403/200 的正确预期：
- 读（`/projects`、`/overview`）：`isAdmin` **或** 启用中的 internal 组织成员（`requirePMARead`）。
- 决策（`/decision`）：`isAdmin` **或** internal 组织 `owner/admin` 角色（`requirePMADecision`, `ownersOnly=true`）。
- 非 internal 成员 / 普通成员 → `forbidden` 403。
- 未认证 → 401（`middleware.Auth`）。

---

## 4. OPC 复用命令（隔离启动 / 停止 / 清理）

```powershell
# 0) 端口预检
netstat -ano | findstr ":18992 "   # 空 = 可用；占用则换 18993/18994
tasklist | findstr "Tomcat"         # 确认 8080 是 GitAIOps Tomcat，勿碰

# 1) 进入 clean worktree + 隔离启动
cd E:\2026-2027\2026-2027-1\MyProjects\eqs-opc-c1d793c\packages\server
New-Item -ItemType Directory -Force -Path ".opc-dryrun" | Out-Null
$env:APP_ENV="test"; $env:SERVER_PORT="18992"; $env:DB_DRIVER="sqlite"
$env:DB_NAME=".opc-dryrun/eqs-opc.sqlite"; $env:JWT_SECRET="<测试JWT>"
$env:WX_MINI_MOCK="1"; $env:PAYMENT_PROVIDER="mock"; $env:DATA_ENCRYPTION_KEY=""
Remove-Item env:REDIS_ADDR,env:DB_HOST,env:DB_USER,env:DB_PASSWORD,env:WX_MINI_APPID,env:WX_MINI_SECRET,env:ZHIPU_API_KEY,env:DEEPSEEK_API_KEY -ErrorAction SilentlyContinue
go run ./cmd/server

# 2) 探活
curl -sS -w "%{http_code}\n" http://127.0.0.1:18992/api/v1/config/public   # 200
curl -sS -w "%{http_code}\n" http://127.0.0.1:18992/api/v1/pma/projects    # 401（未登录）

# 3) 停止
# Ctrl+C（或 kill 该 go 子进程）

# 4) 清理
Remove-Item -Recurse -Force ".opc-dryrun"
```

---

## 5. 硬边界（不得越界）

- ✅ 允许：非生产隔离启动/探活/停止/清理、独立端口、独立 SQLite、mock 配置、`git`/`go` 只读、`curl GET/PUT(测试账号)` 于 localhost。
- ⛔ 禁止：生产部署、生产迁移、真实凭据入库、systemd/nginx/Caddy、DNS、支付/短信/外呼、真实资金/签约副作用。
- ⛔ 禁止：把主工作区脏变更（`config_security.go`/`request_id.go` 等未提交文件）混入 OPC 复核或后续发布。
- ⚠️ 若需 owner 执行系统级操作（如安装 sqlite3 CLI、开放受限端口等）：仅提交审批请求，不自行执行。

## 6. 待 OPC 回传项

OPC 复核异常时，回传：**commit、步骤、HTTP 状态、日志、复现数据**，并暂停后续发布。本报告已备好 clean worktree + 已验证的技术基线（未授权探针全部通过），OPC 可直接在其上执行 §3 的鉴权闭环步骤。

---

*—— ccit-ops 就绪准备 + 技术首验，2026-09-08。精确 commit `c1d793c` 非生产隔离复核环境已交付。*
