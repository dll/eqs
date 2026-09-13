# CCIT-ENV 环境事实与 OPC 自动运维边界 v1

- 任务：`CCIT-ENV-BASELINE-20260909-R1`
- 维护角色：ccit-env（本地开发环境/工具链）
- 盘点日期：2026-09-09（Asia/Shanghai）
- 仓库：`E:\2026-2027\2026-2027-1\MyProjects\eqs`
- 盘点方式：只读检查；未安装依赖、未修改宿主机全局配置、未停止进程、未触碰生产或密钥。

## 1. P0 环境事实基线

### 1.1 已确认工具链

| 项目 | 事实 | 证据/命令 |
|---|---|---|
| Gateway | `127.0.0.1:18789 Ready` | CEO 已确认事实（本任务不重复操作） |
| CCIT Feishu | 8 个账号 running/connected | CEO 已确认事实（不向群发送测试消息） |
| Node | `v24.19.0` | `node --version` |
| npm | `11.17.0` | `npm --version` |
| pnpm | `10.34.5` | `pnpm --version` |
| Go | `go1.26.0 windows/amd64` | `go version` |
| Git | `2.55.0.windows.5` | `git --version` |
| OpenClaw | `2026.9.2 (3928bad)` | `openclaw --version` |
| Docker | 不在 PATH | `Get-Command docker` 无结果；禁止安装 |
| WSL | `C:\WINDOWS\system32\wsl.exe` 可定位，但 CEO 事实为不可用 | `Get-Command wsl`；禁止安装/启用/修改 |

### 1.2 EQS 仓库事实

- 当前 HEAD：`c1d793c2b72e6339278db1429b2c3e600035eb4a`。
- 当前分支：`feature/archived-project-inbound`。
- 工作树已有其他协作改动和未追踪文件；本任务未清理、未覆盖、未回滚。
- Go 服务端：`packages/server`，`go.mod` 声明 Go 1.22，module `github.com/eqs/server`。
- 管理端：`packages/admin`；客户端：`packages/client`；共享包：`shared`。
- 服务端 SQLite 文件：`packages/server/eqs.db`（已有本地文件，按规则不复制、不读取业务数据）。SQLite 文件与服务端二进制已被 `.gitignore` 排除。

## 2. 端口归属与隔离

### 2.1 EQS 预期端口

| 组件 | 端口/路由 | 依据 | 当前状态 |
|---|---|---|---|
| Go server | 默认 `8080`；`SERVER_PORT` 可改 | `packages/server/internal/config/config.go`、`cmd/server/main.go` | EQS 当前未启动 |
| Admin Vite | `3001`；base `/admin/` | `packages/admin/vite.config.*` | 未监听 |
| Client Vite/H5 | `3005`，`strictPort`，host `127.0.0.1` | `packages/client/vite.config.js` | 未监听 |
| Admin/Client API proxy | 均指向 `http://localhost:8090` | 两个 Vite 配置 | 与 server 默认 8080 不一致，需修正/显式设置 |

### 2.2 外部端口事实

CEO 已确认：`8080/Tomcat`、`3000/Gitea`、`5432/PostgreSQL` 为其他服务；EQS 不得占用、停止或修改这些服务。只读复核时 EQS 目标端口无监听；当前不做杀进程或端口抢占。

隔离原则：

1. 开发端口必须显式指定并与前端 proxy 一致（建议为本地 EQS 单独选择未占用端口，或将三处统一到同一端口；变更须由项目开发角色审查）。
2. 测试端口优先使用 httptest/内存调用，不启动真实监听；需要监听时使用独立端口并在验证后由启动进程自行退出。
3. 隔离生产端口、反代、DNS、对外暴露均未就绪，且必须等待 CTO Final Approved。
4. 禁止 ccit-env 杀外部进程、占用外部服务端口或修改宿主机防火墙/全局变量。

## 3. EQS 依赖与目录边界

### 3.1 服务端依赖

- `gin v1.9.1`、`gorm.io/gorm v1.30.0`、`gorm.io/driver/sqlite v1.6.0`、`gorm.io/driver/mysql v1.5.6`、`go-sqlite3 v1.14.22`、`go-redis/v9 v9.5.1`、JWT v5、zap。
- `DB_DRIVER=sqlite`：使用 `packages/server/eqs.db`（或 `DB_NAME` 指定隔离文件），AutoMigrate；Redis 校验降级为内置模拟。
- `DB_DRIVER=mysql`：需要独立 MySQL 8 实例/库，启动执行 `packages/server/migrations/*.sql`；Redis 7 为运行依赖之一。
- 当前本地默认/仓库根 `docker-compose.yml` 含弱默认开发配置，仅可视为 local-dev 参考，不得用于生产。
- `deploy/.env.example` 是 EQS 配置模板；服务端 `config.go` 直接读取进程环境变量，不自动读取仓库根 `.env`。禁止将其他系统 `.env` 当 EQS 配置。

### 3.2 目录边界

| 范围 | ccit-env 允许 | 禁止 |
|---|---|---|
| EQS repo `packages/server`、`packages/admin`、`packages/client`、`shared` | 只读检查、非生产 mock/SQLite 验证、事实文档 | 未授权业务代码重构、生产数据操作 |
| `deploy/` | 只读审阅模板和脚本 | 执行生产部署、改 systemd/nginx/caddy |
| 宿主机全局 | 只读检测版本、端口、进程 | 安装 Docker/WSL、改 PATH/注册表/全局变量、防火墙、服务配置 |
| 密钥/数据 | 只记录变量名和是否配置，不显示值 | 读取、复制、打印 Secret；导入客户/生产数据 |

## 4. 可重复 mock 与验证清单

### 4.1 推荐本地/CI 基线

使用独立 shell/进程环境，不改宿主机永久变量：

```text
APP_ENV=test
SERVER_PORT=<未占用的本地端口>
DB_DRIVER=sqlite
DB_NAME=<隔离工作目录>/eqs-test.db
JWT_SECRET=<仅测试占位值>
DATA_ENCRYPTION_KEY=<仅测试占位值>
PAYMENT_PROVIDER=mock
ESIGN_PROVIDER=mock
QUALIFICATION_VERIFY_PROVIDER=manual
WX_MINI_MOCK=1
UPLOAD_DIR=<隔离工作目录>/uploads
CORS_ALLOW_ORIGINS=http://localhost:<admin-port>,http://127.0.0.1:<client-port>
```

不得把占位值写入生产配置；不得使用真实客户数据。测试更优先使用 `model.InitTestDB()` 的随机内存 SQLite，避免持久化污染。

### 4.2 最小验证顺序

```text
cd E:\2026-2027\2026-2027-1\MyProjects\eqs\packages\server
go test ./...
go build -o bin/server cmd/server/main.go
pnpm --filter @eqs/admin test
pnpm --filter @eqs/client test
```

本轮实测：Go test 全部通过；Admin 3 files/13 tests 通过；Client 4 files/24 tests 通过。未重复执行 build（CEO 已提供本轮 admin/client 与 EQS build 已通过的事实）。

验收项目：

- [x] 工具链版本与仓库 HEAD/分支记录。
- [x] 端口归属记录，确认 EQS 未启动且不触碰外部服务。
- [x] 服务端 Go test 通过。
- [x] admin 13/13、client 24/24 测试通过。
- [x] mock 通道与 SQLite/内存库方案明确。
- [ ] 前端 proxy 与 server 端口统一（当前仓库风险，需 DEV/CTO 决策）。
- [ ] Docker/WSL 隔离环境（当前未就绪；本角色不得安装）。
- [ ] MySQL/Redis 隔离实例及恢复演练（需隔离资源与批准）。
- [ ] 日志、备份、回滚演练（生产边界，需 CTO Final Approved）。
- [ ] 真实外部通道、DNS、反向代理、对外发布（必须 CTO Final Approved）。

## 5. OPC 自动运维边界

### OPC 可自动执行（仅非生产、可逆、无秘密）

- 读取版本、Git 状态、端口监听、目录结构与配置键名。
- 运行 Go/前端测试和构建（产物限于 EQS 工作目录，且不覆盖他人改动）。
- 使用 mock 通道与隔离 SQLite/内存数据库进行验证。
- 更新本环境事实文档和 workspace `memory/facts`。

### 必须人工/CTO Final Approved 后才能执行

- 生产或隔离生产部署、发布、重启、回滚、systemd/nginx/caddy/DNS 改动。
- 任何真实凭据注入、外部支付/签约/短信/OCR/AI 调用、客户数据导入。
- MySQL/Redis 正式实例创建、库迁移、备份恢复、删除或跨环境复制。
- 对外端口、防火墙、证书、代理、域名、云资源的变更。
- 安装 Docker/WSL，修改宿主机 PATH、注册表、全局环境变量，或停止/杀死外部进程。

当前 CTO 状态：`Conditional`。在获得明确 `Final Approved` 前，上述高风险项均为阻塞，不得交给 ccit-ops 执行。

## 6. 本轮变更与阻塞

### 变更

1. 新增本文档：`docs/CCIT-ENV环境事实与OPC自动运维边界-v1-20260909.md`。
2. 更新 CCIT ENV workspace facts：`CCIT/.openclaw/workspaces/ccit-env/memory/facts/eqs-server-env-r1-20260909.md`。
3. 未修改业务代码、配置文件、宿主机全局设置、外部进程或生产资源。

### 阻塞/风险

- 隔离环境未就绪：Docker 不在 PATH，WSL 不可用；根据派单要求不安装、不启用。
- EQS 前端 proxy 目标 8090 与 Go server 默认 8080 不一致，尚未改码；需 DEV/CTO 决策端口基线。
- 工作树存在其他协作改动，未清理；后续变更必须避免覆盖。
- CTO 仍为 Conditional；生产/隔离生产运维交接、真实通道和数据动作全部阻塞。
