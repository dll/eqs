# 工程快捷服务（EQS）软件审核报告 V6.0

> 文档版本：V6.0
> 审核日期：2026-08-11
> 审核基线：`docs/SAR/SAR-v5.0.md`、当前工作区代码与配置、生产服务器实测
> 代码基线：Git `89debe9`（距 `v0.1.0` 共 31 个提交）+ 工作区未提交改动
> 审核范围：安全专项复审（认证授权、CORS、错误处理、凭据管理、部署加固）、部署层实测、测试与构建验证
> 审核方式：静态代码审查、单元测试与构建实测、生产服务器（CVM 129.211.223.113）配置与接口实测；未执行渗透测试、压测或容灾演练

---

## 0. 执行摘要

### 0.1 总体结论

**相比 SAR-v5.0，生产阻断项（P0）继续保持为 0；本轮安全专项复审修复 6 处 P1/P2 问题，其中 1 处为本轮引入的生产回归（CORS 白名单导致 8091 同源登录 403），已修复并部署验证通过。**

项目当前状态：

- **后端**：全部 5 个包 `go test ./...` 通过，`go vet`、`go build` 通过；CORS 白名单、错误不泄露、认证授权、敏感字段加密均已落地并实测。
- **部署**：凭据已从 systemd 明文收敛到 `/opt/eqs/.env`（600 权限，EnvironmentFile 加载）；nginx 安全头/上传限制/健康检查已生效；备份 cron 从 .env 读取凭据，备份文件权限 600。
- **测试**：Admin Vitest 11/11 通过；后端全量测试通过；CD 已通过 `workflow_run` 依赖 CI 成功。
- **生产接口**：同源/白名单 CORS 200、恶意来源 403、`/health` 200、`/admin` 200、登录接口恢复可用。

### 0.2 核心判断

| 维度 | 审核结论 | 主要依据 |
|------|----------|----------|
| 需求完整性 | 部分实现 | 与 V5.0 一致；交易闭环仍依赖 Mock/外部服务 |
| 后端实现 | 可开发验证 | 全量测试通过；对象授权、状态机、事务已部分加强 |
| 管理后台 | 基础运营可用 | 路由守卫 + 后端 Admin 中间件 + 角色读库 |
| 安全与合规 | 生产阻断已清零，遗留 P2 | 本轮修复 CORS/错误泄露/凭据管理；CSP 等 P2 遗留 |
| 测试质量 | 接近门槛 | 后端全绿；Admin 11/11；覆盖率门禁已入 CI |
| 构建与依赖 | 可构建 | 四端构建通过；19 个依赖漏洞已登记例外（P1-15） |
| CI/CD | 有发布门禁 | CD 依赖 CI workflow_run success；缺自动回滚 |
| 性能与可用性 | 无法验收 | 无压测/容灾证据（同 V5.0） |

### 0.3 风险数量

| 等级 | 数量 | 说明 |
|------|------|------|
| P0 | 0 | V5.0 的 7 项已全部关闭，本轮未新产生 |
| P1 | 3 遗留 | P1-01~04 外部依赖；本轮新 P1 已全部修复 |
| P2 | 8 遗留 | CSP、会话管理、监控告警、迁移/回滚、App 原生等 |

---

## 1. 审核口径

### 1.1 状态定义

与 V5.0 一致：已实现 / 部分实现 / 未实现 / 生产阻断 / 未验证。

### 1.2 审核限制

- 未执行渗透测试、SAST、密钥扫描、许可证扫描（同 V5.0）。
- 未执行真实支付/签约/OCR/COS 联调。
- 未执行压测、备份恢复演练、RTO/RPO 验证。
- 生产实测基于服务器本机 `127.0.0.1` 与公网路径 `curl`，未使用真实浏览器会话做端到端 UI 验收。

---

## 2. 基线复核

### 2.1 工程基线

| 项目 | V5.0 | V6.0 实测 | 结论 |
|------|------|-----------|------|
| 后端路由 | 91 条 | 91 条（未新增） | 一致 |
| AutoMigrate 模型 | 25 个 | 25 个 | 一致 |
| 根包版本 | 0.1.0 | 0.1.0 | 一致 |
| 当前提交 | `facfc91` | `89debe9`（+2 提交：P1-05~08 完成） | 持续推进 |
| 未提交改动 | — | deploy 3 文件 + 后端 5 文件 + migrate.go/server.linux | 见 2.2 |

### 2.2 工作区未提交改动（审核对象）

| 文件 | 内容 | 评价 |
|------|------|------|
| `deploy/scripts/backup.sh` | 凭据来源改为环境变量/凭据文件，备份文件 chmod 600，pipefail | ✅ 安全改进 |
| `deploy/systemd/eqs-server.service` | 增加 DATA_ENCRYPTION_KEY 环境项 | ✅ 与 P1-09 配套 |
| `deploy/.env.example` | 增加 DATA_ENCRYPTION_KEY 说明 | ✅ |
| `cmd/server/main.go` | 生产拒绝弱密钥启动 | ✅ |
| `internal/config/config.go` | CORS 白名单配置 + config.Get() 缓存 + ResetCache | ✅ 本轮新增 |
| `internal/handler/response.go` | serverError 生产不泄露内部错误 | ✅ 本轮新增 |
| `internal/middleware/cors.go` | 生产 CORS 白名单 + 同源按主机名判定 | ✅ 本轮新增/修复 |
| `internal/model/db.go` | DSN 增加 multiStatements（配合迁移脚本） | ⚠️ 需评估注入面 |
| `internal/model/migrate.go` | P2-08 版本化迁移（未提交，本轮修复编译） | ⚠️ 未接 main 启动链 |
| `packages/server/server.linux` | 本轮构建的 Linux 部署二进制 | 不应入库 |

> 说明：`db.go` 启用 `multiStatements=true` 仅服务于迁移脚本执行多语句 SQL；若未把迁移接口暴露为网络入口，风险可控，但建议确认无用户可控 SQL 进入该 DSN。

---

## 3. 本轮安全专项修复清单（V6.0 新增）

### 3.1 生产回归：CORS 白名单导致 8091 同源登录 403

| 项 | 内容 |
|----|------|
| 现象 | `POST http://129.211.223.113:8091/api/v1/auth/login` 返回 403，前端登录失败 |
| 根因 | nginx 反代用 `$host` 转发时**去掉端口**（`129.211.223.113`），而浏览器 Origin 带端口（`http://129.211.223.113:8091`）；原实现按整串相等比较 `origin == "http://"+c.Request.Host`，同源请求被误判为跨域 → 403 |
| 修复 | `cors.go` 新增 `isSameOrigin`/`originHost`：按**主机名**（忽略端口、scheme、IPv6 括号）判定同源；生产白名单 `CORS_ALLOW_ORIGINS` 保留 |
| 验证 | 单元测试新增 `TestCORS_ProductionSameOriginIgnoringPort`；生产实测：同源 200、白名单 200、恶意 403 |
| 状态 | ✅ 已部署（server.linux 交叉编译，systemd 重启 active） |

### 3.2 后端代码修复

| 编号 | 问题 | 修复 | 状态 |
|------|------|------|------|
| P1 | `serverError` 回显 `err.Error()` 泄露内部细节 | 生产环境统一返回"服务器内部错误"，原始错误经 `c.Error` 记录；开发环境保留详情 | ✅ |
| P1 | CORS 通配 `*` 允许任意来源跨域 | 生产环境白名单 + 同源放行，非白名单 403；`config.Get()` 缓存热路径配置 | ✅ |
| P1 | `migrate.go` 未提交且缺 `time` 导入，`go test ./...` 编译失败 | 补 `time` 导入，全量测试恢复通过 | ✅ |

### 3.3 部署层修复

| 项 | 修复前 | 修复后 | 状态 |
|----|--------|--------|------|
| systemd 凭据 | `Environment=DB_PASSWORD=...` 明文 + 文件 644 | 删除全部明文 Environment，改用 `EnvironmentFile=/opt/eqs/.env`；unit 权限 600 | ✅ |
| `.env` | 不存在 | `/opt/eqs/.env` 权限 600 root:root，12 项配置（DB/Redis/JWT/加密密钥/服务参数） | ✅ |
| nginx | 缺安全头、50m、健康检查 | `client_max_body_size 50m` + nosniff/SAMEORIGIN/Referrer-Policy + `/health` 代理；`nginx -t` 通过并 reload | ✅ |
| 备份凭据 | backup.sh 内置默认密码；cron 无凭据来源 | backup.sh 支持 `DB_PASSWORD` 环境变量或 `MYSQL_DEFAULTS_FILE`；cron 用 `set -a; . /opt/eqs/.env; set +a` 加载凭据 | ✅ |
| 备份文件 | 644（全用户可读） | 生成后强制 `chmod 600`；历史 644 文件已统一收紧 | ✅ |
| 备份实测 | 未验证 | 实测生成 `eqs_20260811_212255.sql.gz`（8.0K），gzip 校验通过 | ✅ |

### 3.4 测试修复

| 项 | 内容 | 状态 |
|----|------|------|
| `TestCORS_Headers` | 适配新逻辑：开发环境带 Origin 应回显来源 | ✅ |
| `TestCORS_NoOrigin` | 新增：无 Origin 通配放行 | ✅ |
| `TestCORS_ProductionWhitelist` | 新增：生产白名单放行 / 恶意来源 403 | ✅ |
| `TestCORS_ProductionSameOriginIgnoringPort` | 新增：nginx 去端口 Host + 带端口 Origin 同源放行（回归用例） | ✅ |
| `config.ResetCache()` | 新增测试辅助，解决 sync.Once 缓存导致的环境变量隔离问题 | ✅ |

---

## 4. 后端安全复审

### 4.1 认证与授权（复核通过）

- JWT：HS256 白名单校验 + 用户状态实时查库（禁用即失效）+ 角色以数据库为准（`middleware/auth.go`）。
- 公共注册仅允许甲方/服务方（`publicRoles{1,2}`），管理员/专家受控创建（`auth.go`）。
- 固定验证码仅测试环境与演示手机号（1390000 前缀）受控例外；短信频控 60s、失败锁定 5 次（`auth.go`）。
- 对象级授权：订单/项目/争议统一 scope 辅助（`handler/scope.go`）；Admin 路由统一 `RequireAdmin`。
- 生产弱密钥启动拒绝 + 必须配置 `DATA_ENCRYPTION_KEY`（`main.go`）。

### 4.2 敏感数据（复核通过）

- AES-256-GCM 透明加密：经纬度（EncryptedFloat）、证书号（EncryptedString），密文带 `v1:` 版本前缀，兼容历史明文（`model/crypto.go`）。
- 手机号脱敏：`MaskPhone` 应用于管理端列表与服务商列表（`admin.go`、`project.go`）。
- 密钥派生 SHA-256 → 32 字节，进程内单次构造复用。

### 4.3 残余观察（非阻断）

| 项 | 说明 | 建议 |
|----|------|------|
| `db.go` multiStatements | 迁移脚本需要；确认无网络入口直达 | 迁移接口仅限受控发布 |
| `migrate.go` 未接 main | P2-08 迁移未挂到启动链 | 下一迭代接入并加版本测试 |
| 支付/签约 Provider | 仍为 Mock（回调验签已就绪） | 外部依赖 P1-01/02 |

---

## 5. 部署与运维现状（生产实测）

### 5.1 端口暴露面

| 端口 | 监听 | 公网 | 说明 |
|------|------|------|------|
| 22 | 0.0.0.0 | 开放（SSH） | 安全组建议限制办公 IP |
| 80/443 | Caddy | 开放 | wxx 与 EQS 共用，自动 HTTPS |
| 8080 | wxx-server | 开放 | **非 EQS 服务，未改动** |
| 8090 | EQS server | 未放行 | 仅本机反代访问 |
| 8091 | nginx | 放行 | EQS 站点入口 |
| 3306/33060/6379 | 127.0.0.1 | 未放行 | MySQL/Redis 仅本机 |

### 5.2 加固项（已生效）

- systemd：`User=eqs`、`NoNewPrivileges=true`、`ProtectSystem=strict`、`ProtectHome=true`、`ReadWritePaths` 受限；unit 权限 600。
- 凭据：`/opt/eqs/.env` 600；systemd 与 cron 均从该文件读取，无散落明文。
- nginx：安全头、50m 上传、`/health` 健康检查（后端 `/api/v1/config/public` 200）。
- 备份：每日 02:30 cron；gzip 校验；7 天保留；文件 600；可选异地 rsync。

### 5.3 备份链路验证

```text
30 2 * * * set -a; . /opt/eqs/.env; set +a; BACKUP_DIR=/opt/eqs/backup/mysql bash /opt/eqs/backup.sh
实测：备份完成 /opt/eqs/backup/mysql/eqs_20260811_212255.sql.gz (8.0K) ✅
```

### 5.4 遗留风险

| 项 | 风险 | 建议 |
|----|------|------|
| Caddy 转发 443 未做 SNI 域名维度放行 | 备案通过后 80/8091 明文自动解除拦截 | 备案生效后复测 |
| 无备份恢复演练 | RTO/RPO 未验证 | P1-14 补恢复演练 |
| 无监控告警 | 进程内统计重启丢失 | P2-06 落地 Prometheus/告警 |

---

## 6. 前端 / 测试 / CI-CD / 依赖

### 6.1 前端

- Admin：Vue3 + Element Plus；token 存 localStorage，请求层注入 `Authorization`，401 清 token 跳登录（`store/user.ts`、`utils/request.ts`）。
- Client：uni-app H5/小程序/App；token 存 uni storage。
- **P2 遗留**：index.html 无 CSP；token 依赖页面脚本隔离，XSS 会放大令牌泄露风险（同 V5.0，未纳入本轮）。

### 6.2 测试（本轮实测）

| 项 | 结果 |
|----|------|
| 后端 `go test ./...` | ✅ 5 包全过（含新增 CORS 回归用例） |
| 后端 `go vet` / `go build` | ✅ 通过 |
| Admin Vitest | ✅ 11/11 通过 |
| CD 门禁 | ✅ `workflow_run` 依赖 CI success（V5.0 已修，复核通过） |

### 6.3 依赖

- 19 个依赖漏洞（1 low / 10 moderate / 8 high）全部来自 `@dcloudio/*` 构建期工具链，运行时不可达；已按 P1-15 正式例外登记（`docs/security/依赖漏洞例外登记.md`），季度复核。
- Go 侧 `govulncheck` 仍未安装，未纳入门禁（遗留）。

---

## 7. 分级问题清单（V6.0）

### 7.1 P0：生产阻断

| 编号 | 事项 | 状态 |
|------|------|------|
| P0-01~07 | V5.0 全部生产阻断项 | ✅ 保持关闭，本轮复核无复发 |

### 7.2 P1

| 编号 | 事项 | 状态 |
|------|------|------|
| P1-09 配套 | CORS 白名单（生产） | ✅ 本轮修复 |
| P1-09 配套 | serverError 生产不泄露 | ✅ 本轮修复 |
| P1-09 配套 | 凭据收敛 .env + systemd EnvironmentFile | ✅ 本轮修复 |
| P1-14 配套 | 备份凭据链路 + 备份文件 600 | ✅ 本轮修复 |
| P1-01~04 | 支付/电子签/COS/OCR 外部通道 | ⏳ 外部依赖（回调验签已就绪） |

### 7.3 P2

| 编号 | 事项 | 状态 |
|------|------|------|
| P2-01 | 覆盖率 80% 门禁 | ⏳ 持续（CI 已设 ≥60%） |
| P2-06 | Prometheus/告警/SLO | ⏳ 待办 |
| P2-07 | OpenAPI 契约 | ⏳ 待办 |
| P2-08 | 版本化迁移（migrate.go） | ⏳ 已起草，未接 main |
| P2-09 | App 原生工程 | ⏳ 外部依赖 |
| 新增 | CSP / 会话管理 / HSTS | ⏳ 待办（安全头已部分补） |

---

## 8. 结论

EQS 在 V5.0 安全止血基础上，本轮完成一次安全专项复审与生产加固：

1. **修复 1 处生产回归**（CORS 白名单误伤 8091 同源登录），并补充回归测试，生产验证 200；
2. **代码层**：错误不泄露、CORS 白名单、config 缓存、迁移文件编译修复，后端全量测试通过；
3. **部署层**：凭据收敛 `.env`（600）、systemd EnvironmentFile、nginx 安全头/健康检查、备份链路与备份文件权限加固，全部生产实测通过；
4. **P0 保持 0**，P1 中可自控项已清零，剩余为外部依赖与 P2 治理项。

**V6.0 审核结论：演示/试运行级安全基线达成，可进入公测前治理（P1 外部联调 + P2 监控/迁移/CSP）；生产资金交易能力仍需持牌通道接入后方可认定。**

---

## 9. 附录

### 9.1 关键证据文件

| 领域 | 文件 |
|------|------|
| CORS | `packages/server/internal/middleware/cors.go` |
| 错误处理 | `packages/server/internal/handler/response.go` |
| 配置 | `packages/server/internal/config/config.go` |
| 认证 | `packages/server/internal/middleware/auth.go`、`handler/auth.go` |
| 授权 | `packages/server/internal/handler/scope.go`、`admin.go` |
| 加密 | `packages/server/internal/model/crypto.go` |
| 迁移 | `packages/server/internal/model/migrate.go`（未提交） |
| 部署 | `deploy/systemd/eqs-server.service`、`deploy/scripts/backup.sh`、`deploy/nginx/eqs-cvm.conf` |
| 生产凭据 | `/opt/eqs/.env`（600） |
| 依赖例外 | `docs/security/依赖漏洞例外登记.md` |

### 9.2 本次主要验证命令

```text
cd packages/server && go test ./... && go vet ./... && go build ./...
cd packages/admin && npx vitest run
curl http://127.0.0.1:8091/admin/  → 200
curl http://127.0.0.1:8091/health  → 200
curl -H 'Origin: https://evil.example.com' .../api/v1/config/public → 403
bash /opt/eqs/backup.sh（从 .env 加载凭据）→ 备份成功，文件 600
```

### 9.3 文档变更记录

| 版本 | 日期 | 变更 |
|------|------|------|
| V6.0 | 2026-08-11 | 安全专项复审：修复 CORS 生产回归、错误泄露、凭据收敛 .env、备份链路加固；全量测试与生产实测；P0 保持 0 |

---

**文档结束**
